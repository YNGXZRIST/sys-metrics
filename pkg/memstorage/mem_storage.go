package memstorage

import (
	"errors"
	"sync"
)

var ErrNotFound = errors.New("not found")

type Storage[K comparable, V any] interface {
	Set(key K, value V) error
	Get(key K) (V, error)
	Delete(key K) error
	Has(key K) bool
	All() map[K]V
}

type MemStorage[K comparable, V any] struct {
	data map[K]V
	mu   sync.RWMutex
}

func NewMemStorage[K comparable, V any]() *MemStorage[K, V] {
	return &MemStorage[K, V]{data: make(map[K]V), mu: sync.RWMutex{}}
}

func (s *MemStorage[K, V]) Set(key K, value V) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return nil
}

func (s *MemStorage[K, V]) Get(key K) (V, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	if !ok {
		var zero V
		return zero, ErrNotFound
	}
	return v, nil
}
func (s *MemStorage[K, V]) All() map[K]V {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r := make(map[K]V, len(s.data))
	for k, v := range s.data {
		r[k] = v
	}
	return r
}

func (s *MemStorage[K, V]) Delete(key K) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[key]; !ok {
		return ErrNotFound
	}
	delete(s.data, key)
	return nil
}

func (s *MemStorage[K, V]) Has(key K) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.data[key]
	return ok
}
