package authenticate

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"sys-metrics/internal/common"
	"testing"
)

func TestNewSha256(t *testing.T) {
	s := NewSha256(common.TypeModeTest)
	if s == nil {
		t.Fatalf("NewSHA256()=%v, want not nil", s)
	}
	key := []byte(common.TypeModeTest)
	h := hmac.New(sha256.New, key)

	expected := h.Sum(nil)
	got := s.hash.Sum(nil)
	if !hmac.Equal(expected, got) {
		t.Errorf("HMAC digest mismatch: got %s, want %s", hex.EncodeToString(got), hex.EncodeToString(expected))
	}
}

func TestSha256_validate(t *testing.T) {
	s := NewSha256(common.TypeModeTest)
	expectedValid := hmac.New(sha256.New, []byte(common.TypeModeTest)).Sum(nil)
	expectedInvalid := hmac.New(sha256.New, []byte("invalid")).Sum(nil)
	validate, err := s.validate(hex.EncodeToString(expectedValid))
	if err != nil {
		t.Fatalf("Validate()=%v, want nil", err)
	}
	if !validate {
		t.Errorf("Validate()=false, want true")
	}
	validate, err = s.validate(hex.EncodeToString(expectedInvalid))
	if err != nil {
		t.Fatalf("Validate()=%v, want nil", err)
	}
	if validate {
		t.Fatalf("Validate()=true, want false")
	}
}
