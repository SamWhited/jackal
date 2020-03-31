/*
 * Copyright (c) 2020 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package cluster

import "context"

// KV represents a generic cluster key-value store.
type KV interface {
	// Put stores a new value associated to a given key.
	// This value will expire after ttlInSeconds expiration time is reached.
	Put(ctx context.Context, key string, value string, ttlInSeconds int64) error

	// Del removes from store the value associated to key.
	Del(ctx context.Context, key string) error

	// GetPrefix retrieves all values whose key matches prefix.
	GetPrefix(ctx context.Context, prefix string) (map[string]string, error)
}
