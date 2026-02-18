package authenticate

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sys-metrics/internal/common"
	"sys-metrics/internal/errors/labelerrors"
)

type SHA256 struct {
	authenticator
}

func NewSha256(secretKey string) *SHA256 {
	key := []byte(secretKey)
	h := hmac.New(sha256.New, key)
	return &SHA256{
		authenticator{
			hash:     h,
			hashType: common.SHA256,
		},
	}
}

func (s *SHA256) validate(string string) (bool, error) {
	decodeString, err := hex.DecodeString(string)
	if err != nil {
		return false, labelerrors.NewLabelError(s.hashType, fmt.Errorf("cannot decode string %v, error:%w", string, err))
	}
	return hmac.Equal(decodeString, s.hash.Sum(nil)), nil
}
