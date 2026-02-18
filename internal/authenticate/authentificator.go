package authenticate

import "hash"

type authenticator struct {
	hash     hash.Hash
	hashType string
	validator
}
type validator interface {
	validate(string) (bool, error)
}
