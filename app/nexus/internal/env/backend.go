//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package env

import (
	"os"
	"strings"
)

// StoreType represents the type of backend storage to use.
type StoreType string

const (
	// S3 indicates an Amazon S3 (or compatible) storage backend
	S3 StoreType = "s3"

	// Sqlite indicates a SQLite database storage backend
	Sqlite StoreType = "sqlite"

	// Memory indicates an in-memory storage backend
	Memory StoreType = "memory"
)

// BackendStoreType determines which storage backend type to use based on the
// SPIKE_NEXUS_BACKEND_STORE environment variable. The value is case-insensitive.
//
// Valid values are:
//   - "file": Uses file-based storage
//   - "sqlite": Uses SQLite database storage
//   - "memory": Uses in-memory storage
//
// If the environment variable is not set or contains an invalid value,
// it defaults to Memory.
func BackendStoreType() StoreType {
	st := os.Getenv("SPIKE_NEXUS_BACKEND_STORE")

	switch strings.ToLower(st) {
	case string(S3):
		return S3
	case string(Sqlite):
		return Sqlite
	case string(Memory):
		return Memory
	default:
		return Memory
	}
}