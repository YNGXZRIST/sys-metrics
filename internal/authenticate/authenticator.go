package authenticate

import (
	"hash"
	"sync"
)

type authenticator struct {
	hash          hash.Hash
	hashType      string
	hashHeaderKey string
	mu            sync.Mutex
}

type Authenticator interface {
	Validate(header string, bytes []byte) (bool, error)
	GetHashHeaderKey() string
	SignBody(data []byte) string
}
