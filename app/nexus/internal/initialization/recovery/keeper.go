//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"encoding/hex"
	"encoding/json"
	"net/url"
	"os"

	"github.com/cloudflare/circl/secretsharing"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiUrl "github.com/spiffe/spike-sdk-go/api/url"

	"github.com/spiffe/spike/app/nexus/internal/env"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/config"
	"github.com/spiffe/spike/internal/log"
)

func iterateKeepersToBootstrap(
	keepers map[string]string, rootShares []secretsharing.Share,
	successfulKeepers map[string]bool, source *workloadapi.X509Source,
) bool {
	const fName = "iterateKeepersToBootstrap"

	for keeperId, keeperApiRoot := range keepers {
		u, err := url.JoinPath(
			keeperApiRoot, string(apiUrl.SpikeKeeperUrlContribute),
		)
		if err != nil {
			log.Log().Warn(
				fName, "msg", "Failed to join path", "url", keeperApiRoot,
			)
			continue
		}

		data := shardContributionResponse(
			keeperId, keepers, u, rootShares, source,
		)
		if len(data) == 0 {
			log.Log().Info(fName, "msg", "No data; moving on...")
			continue
		}

		var res reqres.ShardContributionResponse
		err = json.Unmarshal(data, &res)
		if err != nil {
			log.Log().Info(fName, "msg", "Failed to unmarshal response", "err", err)
			continue
		}

		successfulKeepers[keeperId] = true
		log.Log().Info(fName, "msg", "Success", "keeper_id", keeperId)

		if len(successfulKeepers) == env.ShamirShares() {
			log.Log().Info(fName, "msg", "All keepers initialized")

			tombstone := config.SpikeNexusTombstonePath()

			// Create the tombstone file to mark SPIKE Nexus as bootstrapped.
			err = os.WriteFile(
				tombstone,
				[]byte("spike.nexus.bootstrapped=true"), 0644,
			)
			if err != nil {
				// Although the tombstone file is just a marker, it's still important.
				// If SPIKE Nexus cannot create the tombstone file, or if someone
				// deletes it, this can slightly change system operations.
				// To be on the safe side, we let SPIKE Nexus crash because not being
				// able to write to the data volume (where the tombstone file would be)
				// can be a precursor of other problems that can affect the reliability
				// of the backing store.
				log.FatalLn(fName + ": failed to create tombstone file: " + err.Error())
			}

			log.Log().Info(fName, "msg", "Tombstone file created successfully")

			return true
		}
	}

	return false
}

func iterateKeepersAndTryRecovery(
	source *workloadapi.X509Source,
	successfulKeeperShards map[string]string,
) bool {
	const fName = "iterateKeepersAndTryRecovery"

	for keeperId, keeperApiRoot := range env.Keepers() {
		log.Log().Info(fName, "id", keeperId, "url", keeperApiRoot)

		u := shardUrl(keeperApiRoot)
		if u == "" {
			continue
		}

		data := shardResponse(source, u, keeperId)
		if len(data) == 0 {
			continue
		}

		res := unmarshalShardResponse(data)
		if res == nil {
			continue
		}

		shard := res.Shard
		if len(shard) == 0 {
			log.Log().Info(fName, "msg", "No shard")
			continue
		}

		successfulKeeperShards[keeperId] = shard
		if len(successfulKeeperShards) != env.ShamirThreshold() {
			continue
		}

		ss := rawShards(successfulKeeperShards)
		if len(ss) != env.ShamirThreshold() {
			continue
		}

		binaryRec := RecoverRootKey(ss)
		encoded := hex.EncodeToString(binaryRec)
		state.Initialize(encoded)

		state.SetRootKey(binaryRec)

		// System initialized: Exit infinite loop.
		return true
	}

	return false
}
