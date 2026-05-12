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

type RequestEncryptor struct {
	IsEnabled bool
	publicKey *rsa.PublicKey
}

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
func generateKeyAES256() ([]byte, error) {
	key := make([]byte, chacha20.KeySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}
func generateNonce() ([]byte, error) {
	nonce := make([]byte, chacha20.NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return nonce, nil
}
