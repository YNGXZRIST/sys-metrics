// Package internal loads RSA PEM keys (PKCS#1) for package secure.
package internal

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// ReadPublicKey reads a PEM file and parses an RSA public key in PKCS#1 form.
func ReadPublicKey(path string) (*rsa.PublicKey, error) {
	publicKeyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(publicKeyBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the public key")
	}
	pub, errParse := x509.ParsePKCS1PublicKey(block.Bytes)
	if errParse != nil {
		return nil, fmt.Errorf("failed to parse public key: %s", errParse)
	}
	return pub, nil
}

// ReadPrivateKey reads a PEM file and parses an RSA private key in PKCS#1 form.
func ReadPrivateKey(path string) (*rsa.PrivateKey, error) {
	privateKeyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(privateKeyBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the private key")
	}
	pub, errParse := x509.ParsePKCS1PrivateKey(block.Bytes)
	if errParse != nil {
		return nil, fmt.Errorf("failed to parse private key: %s", errParse)
	}
	return pub, nil
}
