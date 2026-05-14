package secure

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"io"
	"sys-metrics/internal/secure/internal"

	"golang.org/x/crypto/chacha20"
)

// RequestEncryptor encrypts request bodies for transport to a server that has the matching private key.
// NewRequestEncryptor loads a PKCS#1 PEM public key from path; an empty path yields a disabled encryptor.
type RequestEncryptor struct {
	publicKey *rsa.PublicKey
	IsEnabled bool
}

// NewRequestEncryptor builds an encryptor. If path is empty, returns a disabled encryptor (IsEnabled false).
func NewRequestEncryptor(path string) (*RequestEncryptor, error) {
	encryptor := &RequestEncryptor{IsEnabled: false}
	if path != "" {
		key, err := internal.ReadPublicKey(path)
		if err != nil {
			return nil, err
		}
		encryptor.IsEnabled = true
		encryptor.publicKey = key

	}
	return encryptor, nil
}

// Encrypt returns ciphertext: RSA-OAEP-wrapped AES key, nonce, and GCM-sealed plaintext.
func (e *RequestEncryptor) Encrypt(plaintext []byte) ([]byte, error) {
	if e.publicKey == nil {
		return nil, fmt.Errorf("public key not set")
	}
	key, err := generateKeyAES256()
	if err != nil {
		return nil, fmt.Errorf("failed to generate AES-256 key: %w", err)
	}
	nonce, err := generateNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES-256 block: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}
	sealed := gcm.Seal(nil, nonce, plaintext, nil)
	cipherKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, e.publicKey, key, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt key: %w", err)
	}
	out := bytes.Join([][]byte{cipherKey, nonce, sealed}, nil)
	return out, nil
}

// generateKeyAES256 returns a random 32-byte key (AES-256); uses chacha20.KeySize for length constant.
func generateKeyAES256() ([]byte, error) {
	key := make([]byte, chacha20.KeySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}

// generateNonce returns a random nonce suitable for AES-GCM (12 bytes; chacha20.NonceSize matches GCM standard nonce length here).
func generateNonce() ([]byte, error) {
	nonce := make([]byte, chacha20.NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return nonce, nil
}
