// Package authenticate defines signing and verification of HTTP request bodies (HMAC, etc.).
package authenticate

import (
	"hash"
	"sync"
)

// generate:reset
type authenticator struct {
	hash          hash.Hash
	hashType      string
	hashHeaderKey string
	mu            sync.Mutex
}

// Authenticator signs response bodies and validates the hash header on incoming requests.
type Authenticator interface {
	Validate(header string, bytes []byte) (bool, error)
	GetHashHeaderKey() string
	SignBody(data []byte) string
}
