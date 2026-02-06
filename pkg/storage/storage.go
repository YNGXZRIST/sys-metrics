package storage

type Storage[K comparable, V any] interface {
	Set(key K, value V) error
	Get(key K) (V, error)
	Delete(key K) error
	Has(key K) bool
	All() map[K]V
}
