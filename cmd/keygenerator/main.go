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
	BlockPublicKey  = "RSA PUBLIC KEY"
	BlockPrivateKey = "RSA PRIVATE KEY"
)

func main() {
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Fatal(fmt.Errorf("error generating rsa key: %w", err))
	}
	err = writeRSAPrivateKeyPKCS1(key)
	if err != nil {
		log.Fatal(fmt.Errorf("error generating rsa private key: %w", err))
	}
	err = writeRSAPublicKeyPKCS1(&key.PublicKey)
	if err != nil {
		log.Fatal(fmt.Errorf("error generating rsa public key: %w", err))
	}

}
func writeRSAPrivateKeyPKCS1(key *rsa.PrivateKey) error {
	der := x509.MarshalPKCS1PrivateKey(key)
	block := &pem.Block{
		Type:  BlockPrivateKey,
		Bytes: der,
	}
	return os.WriteFile("private.pem", pem.EncodeToMemory(block), 0600)
}
func writeRSAPublicKeyPKCS1(key *rsa.PublicKey) error {
	der := x509.MarshalPKCS1PublicKey(key)
	block := &pem.Block{
		Type:  BlockPublicKey,
		Bytes: der,
	}
	return os.WriteFile("public.pem", pem.EncodeToMemory(block), 0644)
}
