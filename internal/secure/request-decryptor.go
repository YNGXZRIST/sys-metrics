package secure

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"sys-metrics/internal/secure/internal"
)

type RequestDecryptor struct {
	IsEnabled  bool
	privateKey *rsa.PrivateKey
}

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
func (d *RequestDecryptor) Decrypt(plain []byte) ([]byte, error) {
	if d.privateKey == nil {
		return nil, fmt.Errorf("private key is nil")
	}
	rsaLen := (d.privateKey.N.BitLen() + 7) / 8
	aesBlock := plain[:rsaLen]
	nonceBlock := plain[rsaLen : rsaLen+12]
	sealedBlock := plain[rsaLen+12:]
	_, _ = nonceBlock, sealedBlock
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
