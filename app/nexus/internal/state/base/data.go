//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"sync"

	"github.com/spiffe/spike-sdk-go/kv"

	"github.com/spiffe/spike/app/nexus/internal/env"
)

var (
	secretStore = kv.New(kv.Config{
		MaxSecretVersions: env.MaxSecretVersions(),
	})
	secretStoreMu sync.RWMutex
)

var policies sync.Map

var (
	rootKey   []byte
	rootKeyMu sync.RWMutex
)

func RootKey() []byte {
	rootKeyMu.RLock()
	defer rootKeyMu.RUnlock()
	return rootKey
}

func SetRootKey(rk []byte) {
	rootKeyMu.Lock()
	defer rootKeyMu.Unlock()
	rootKey = rk
}
