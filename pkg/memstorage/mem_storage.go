package memstorage

import (
	"errors"
	"sync"
)

var ErrNotFound = errors.New("not found")
var mu = sync.RWMutex{}

type Storage[K comparable, V any] interface {
	Set(key K, value V) error
	Get(key K) (V, error)
	Delete(key K) error
	Has(key K) bool
	All() map[K]V
}

type MemStorage[K comparable, V any] struct {
	data map[K]V
}

func NewMemStorage[K comparable, V any]() *MemStorage[K, V] {
	return &MemStorage[K, V]{data: make(map[K]V)}
}

func (s *MemStorage[K, V]) Set(key K, value V) error {
	mu.Lock()
	s.data[key] = value
	mu.Unlock()
	return nil
}

func (s *MemStorage[K, V]) Get(key K) (V, error) {
	v, ok := s.data[key]
	if !ok {
		var zero V
		return zero, ErrNotFound
	}
	return v, nil
}
func (s *MemStorage[K, V]) All() map[K]V {
	return s.data
}

func (s *MemStorage[K, V]) Delete(key K) error {
	if _, ok := s.data[key]; !ok {
		return ErrNotFound
	}
	mu.Lock()
	delete(s.data, key)
	mu.Unlock()
	return nil
}

func (s *MemStorage[K, V]) Has(key K) bool {
	_, ok := s.data[key]
	return ok
}
