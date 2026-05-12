package main

import (
	"crypto/rand"
	"crypto/rsa"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteRSAPrivateKeyPKCS1_and_public(t *testing.T) {
	dir := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWD) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeRSAPrivateKeyPKCS1(key); err != nil {
		t.Fatal(err)
	}
	if err := writeRSAPublicKeyPKCS1(&key.PublicKey); err != nil {
		t.Fatal(err)
	}
	privData, err := os.ReadFile(filepath.Join(dir, "private.pem"))
	if err != nil || len(privData) < 100 {
		t.Fatalf("private.pem: len=%d err=%v", len(privData), err)
	}
	pubData, err := os.ReadFile(filepath.Join(dir, "public.pem"))
	if err != nil || len(pubData) < 50 {
		t.Fatalf("public.pem: len=%d err=%v", len(pubData), err)
	}
}
