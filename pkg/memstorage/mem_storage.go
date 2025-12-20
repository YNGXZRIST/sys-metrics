package memstorage

import (
	"errors"
)

var ErrNotFound = errors.New("not found")

type Storage[K comparable, V any] interface {
	Set(key K, value V) error
	Get(key K) (V, error)
	Delete(key K) error
	Has(key K) bool
}

type MemStorage[K comparable, V any] struct {
	data map[K]V
}

func NewMemStorage[K comparable, V any]() *MemStorage[K, V] {
	return &MemStorage[K, V]{data: make(map[K]V)}
}

func (s *MemStorage[K, V]) Set(key K, value V) error {
	s.data[key] = value
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

func (s *MemStorage[K, V]) Delete(key K) error {
	delete(s.data, key)
	return nil
}

func (s *MemStorage[K, V]) Has(key K) bool {
	_, ok := s.data[key]
	return ok
}
