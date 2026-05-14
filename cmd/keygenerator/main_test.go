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
	if errChdir := os.Chdir(dir); errChdir != nil {
		t.Fatal(errChdir)
	}

	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	if errWrPr := writeRSAPrivateKeyPKCS1(key); errWrPr != nil {
		t.Fatal(errWrPr)
	}
	if errWrPb := writeRSAPublicKeyPKCS1(&key.PublicKey); errWrPb != nil {
		t.Fatal(errWrPb)
	}
	privateData, err := os.ReadFile(filepath.Join(dir, "private.pem"))
	if err != nil || len(privateData) < 100 {
		t.Fatalf("private.pem: len=%d err=%v", len(privateData), err)
	}
	pubData, err := os.ReadFile(filepath.Join(dir, "public.pem"))
	if err != nil || len(pubData) < 50 {
		t.Fatalf("public.pem: len=%d err=%v", len(pubData), err)
	}
}
