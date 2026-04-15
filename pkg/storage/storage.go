// Package storage defines a key-value abstraction for storing metrics and auxiliary data.
package storage

import "context"

// Storage is the storage contract with read, write, and enumeration operations.
type Storage[K comparable, V any] interface {
	Set(ctx context.Context, key K, value V) error
	Get(ctx context.Context, key K) (V, error)
	Delete(ctx context.Context, key K) error
	Has(ctx context.Context, key K) bool
	All(ctx context.Context) map[K]V
}
