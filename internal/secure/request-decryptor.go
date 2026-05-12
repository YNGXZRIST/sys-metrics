// Package secure implements optional RSA+AES-GCM encryption of HTTP request bodies between the
// metrics agent and server. Empty key paths yield disabled encryptor/decryptor; middleware then
// passes plaintext through unchanged.
package secure

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"sys-metrics/internal/secure/internal"

	"golang.org/x/crypto/chacha20"
)

// RequestDecryptor decrypts request bodies produced by RequestEncryptor when IsEnabled is true.
// NewRequestDecryptor loads a PKCS#1 PEM private key from path; an empty path yields a disabled
// instance (IsEnabled false) suitable for servers that do not expect ciphertext.
type RequestDecryptor struct {
	IsEnabled  bool
	privateKey *rsa.PrivateKey
}

// NewRequestDecryptor builds a decryptor. If path is empty, returns a disabled decryptor with no key loaded.
func NewRequestDecryptor(path string) (*RequestDecryptor, error) {
	decryptor := &RequestDecryptor{IsEnabled: false}
	if path != "" {
		key, err := internal.ReadPrivateKey(path)
		if err != nil {
			return nil, err
		}
		decryptor.IsEnabled = true
		decryptor.privateKey = key
	}
	return decryptor, nil
}

// Decrypt unwraps ciphertext: RSA-OAEP block, GCM nonce, then GCM-sealed payload.
// Returns an error if the key is missing, the blob is too short, or decryption fails.
func (d *RequestDecryptor) Decrypt(plain []byte) ([]byte, error) {
	if d.privateKey == nil {
		return nil, fmt.Errorf("private key is nil")
	}
	rsaLen := (d.privateKey.N.BitLen() + 7) / 8
	if len(plain) < rsaLen+chacha20.NonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	aesBlock := plain[:rsaLen]
	nonceBlock := plain[rsaLen : rsaLen+chacha20.NonceSize]
	sealedBlock := plain[rsaLen+chacha20.NonceSize:]
	bytes, errD := rsa.DecryptOAEP(sha256.New(), rand.Reader, d.privateKey, aesBlock, nil)
	if errD != nil {
		return nil, errD
	}
	cipherBlock, err := aes.NewCipher(bytes)
	if err != nil {
		return nil, fmt.Errorf("error in aes.NewCipher: %v", err)
	}
	gcm, err := cipher.NewGCM(cipherBlock)
	if err != nil {
		return nil, fmt.Errorf("NewGCM(): %v", err)
	}
	plaintext, err := gcm.Open(nil, nonceBlock, sealedBlock, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}
	return plaintext, nil
}
