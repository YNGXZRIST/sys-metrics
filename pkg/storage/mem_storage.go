package storage

import (
	"context"
	"errors"
	"sync"
)

// ErrNotFound is returned when a key is missing from MemStorage.
var ErrNotFound = errors.New("not found")

// MemStorage is an in-memory Storage implementation protected by a mutex.
type MemStorage[K comparable, V any] struct {
	data map[K]V
	mu   sync.Mutex
}

// NewMemStorage returns an empty in-memory store.
func NewMemStorage[K comparable, V any]() *MemStorage[K, V] {
	return &MemStorage[K, V]{data: make(map[K]V), mu: sync.Mutex{}}
}

// Set stores value under key.
func (s *MemStorage[K, V]) Set(ctx context.Context, key K, value V) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return nil
}

// Get returns the value for key or ErrNotFound.
func (s *MemStorage[K, V]) Get(ctx context.Context, key K) (V, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.data[key]
	if !ok {
		var zero V
		return zero, ErrNotFound
	}
	return v, nil
}

// All returns a copy of all key-value pairs.
func (s *MemStorage[K, V]) All(ctx context.Context) map[K]V {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := make(map[K]V, len(s.data))
	for k, v := range s.data {
		r[k] = v
	}
	return r
}

// Delete removes key; it returns ErrNotFound if the key is absent.
func (s *MemStorage[K, V]) Delete(ctx context.Context, key K) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[key]; !ok {
		return ErrNotFound
	}
	delete(s.data, key)
	return nil
}

// Has reports whether key exists.
func (s *MemStorage[K, V]) Has(ctx context.Context, key K) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.data[key]
	return ok
}
