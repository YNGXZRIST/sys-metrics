package authenticate

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"sys-metrics/internal/common"
	"sys-metrics/internal/errors/labelerrors"
)

type SHA256 struct {
	authenticator
}

func NewSha256(secretKey *string) *SHA256 {
	if secretKey == nil {
		return nil
	}
	key := []byte(*secretKey)
	h := hmac.New(sha256.New, key)
	return &SHA256{
		authenticator{
			hash:          h,
			hashType:      common.SHA256,
			hashHeaderKey: common.HashSHA256,
			mu:            sync.Mutex{},
		},
	}
}
func (s *SHA256) GetHashHeaderKey() string {
	return s.hashHeaderKey
}

func (s *SHA256) SignBody(data []byte) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hash.Reset()
	s.hash.Write(data)
	return hex.EncodeToString(s.hash.Sum(nil))
}

func (s *SHA256) Validate(header string, body []byte) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	decoded, err := hex.DecodeString(header)
	if err != nil {
		return false, labelerrors.NewLabelError(s.hashType, fmt.Errorf("cannot decode string %v, error:%w", header, err))
	}
	s.hash.Reset()
	s.hash.Write(body)
	got := s.hash.Sum(nil)
	return hmac.Equal(decoded, got), nil
}
