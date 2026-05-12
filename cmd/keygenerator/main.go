package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
)

const (
	BlockPublicKey  = "PUBLIC KEY"
	BlockPrivateKey = "PRIVATE KEY"
)

func main() {
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Fatal(fmt.Errorf("error generating rsa key: %w", err))
	}
	err = writeRSAPrivateKeyPKCS8(key)
	if err != nil {
		log.Fatal(fmt.Errorf("error generating rsa private key: %w", err))
	}
	err = writeRSAPublicKeyPEM(&key.PublicKey)
	if err != nil {
		log.Fatal(fmt.Errorf("error generating rsa public key: %w", err))
	}

}
func writeRSAPrivateKeyPKCS8(key *rsa.PrivateKey) error {
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return err
	}
	block := &pem.Block{
		Type:  BlockPrivateKey,
		Bytes: der,
	}
	return os.WriteFile("private.pem", pem.EncodeToMemory(block), 0600)
}
func writeRSAPublicKeyPEM(key *rsa.PublicKey) error {
	der, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return err
	}
	block := &pem.Block{
		Type:  BlockPublicKey,
		Bytes: der,
	}
	return os.WriteFile("public.pem", pem.EncodeToMemory(block), 0644)
}
