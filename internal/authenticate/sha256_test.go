package authenticate

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"sys-metrics/internal/common"
	"testing"
)

func TestNewSha256(t *testing.T) {

	s := NewSha256(new(common.TypeModeTest))
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

func TestSha256_Validate(t *testing.T) {
	s := NewSha256(new(common.TypeModeTest))
	body := []byte(common.PollCount)
	hValid := hmac.New(sha256.New, []byte(common.TypeModeTest))
	hValid.Write(body)
	expectedValid := hValid.Sum(nil)
	hInvalid := hmac.New(sha256.New, []byte("invalid"))
	hInvalid.Write(body)
	expectedInvalid := hInvalid.Sum(nil)

	validate, err := s.Validate(hex.EncodeToString(expectedValid), body)
	if err != nil {
		t.Fatalf("Validate() err = %v, want nil", err)
	}
	if !validate {
		t.Errorf("Validate() = false, want true")
	}
	validate, err = s.Validate(hex.EncodeToString(expectedInvalid), body)
	if err != nil {
		t.Fatalf("Validate() err = %v, want nil", err)
	}
	if validate {
		t.Errorf("Validate() = true, want false")
	}
}
