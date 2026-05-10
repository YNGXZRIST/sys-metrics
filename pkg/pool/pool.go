// Package pool provides a small generic wrapper around sync.Pool for reusing temporary objects.
//
// The pool stores values of one concrete type T. To make reuse safe, values must implement Reset(),
// and Pool.Put always calls Reset() before returning the value back to the underlying sync.Pool.
//
// Note: sync.Pool is a cache for optimization only. The runtime may drop pooled values at any time
// (e.g. during garbage collection), so Get may call the factory function even if values were Put before.
package pool

import "sync"

// Resettable is implemented by types that can reset their internal state,
// making them safe to reuse when returned back to a Pool.
type Resettable interface {
	Reset()
}

// Pool is a typed wrapper over sync.Pool for values of type T.
type Pool[T Resettable] struct {
	p sync.Pool
}

// New constructs a new Pool that uses fn to create new values when the pool is empty.
func New[T Resettable](fn func() T) *Pool[T] {
	return &Pool[T]{
		p: sync.Pool{
			New: func() any { return fn() },
		},
	}
}

// Put resets x and returns it back to the pool for reuse.
func (p *Pool[T]) Put(x T) {
	x.Reset()
	p.p.Put(x)
}

// Get returns a value from the pool or creates one using the factory passed to New.
func (p *Pool[T]) Get() T {
	return p.p.Get().(T)
}
