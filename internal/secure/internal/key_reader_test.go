package internal

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

func rsaPEMFiles(t *testing.T) (pubPath, privPath string) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	privPath = filepath.Join(dir, "private.pem")
	pubPath = filepath.Join(dir, "public.pem")
	privBlock := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}
	if err := os.WriteFile(privPath, pem.EncodeToMemory(privBlock), 0o600); err != nil {
		t.Fatal(err)
	}
	pubBlock := &pem.Block{Type: "RSA PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(&priv.PublicKey)}
	if err := os.WriteFile(pubPath, pem.EncodeToMemory(pubBlock), 0o600); err != nil {
		t.Fatal(err)
	}
	return pubPath, privPath
}

func TestReadPublicKey(t *testing.T) {
	pubPath, _ := rsaPEMFiles(t)
	key, err := ReadPublicKey(pubPath)
	if err != nil {
		t.Fatalf("ReadPublicKey: %v", err)
	}
	if key == nil || key.N == nil {
		t.Fatal("expected non-nil public key")
	}
}

func TestReadPublicKey_notFound(t *testing.T) {
	_, err := ReadPublicKey(filepath.Join(t.TempDir(), "nope.pem"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReadPublicKey_invalidPEM(t *testing.T) {
	p := filepath.Join(t.TempDir(), "bad.pem")
	if err := os.WriteFile(p, []byte("not pem"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := ReadPublicKey(p)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReadPrivateKey(t *testing.T) {
	_, privPath := rsaPEMFiles(t)
	key, err := ReadPrivateKey(privPath)
	if err != nil {
		t.Fatalf("ReadPrivateKey: %v", err)
	}
	if key == nil || key.D == nil {
		t.Fatal("expected non-nil private key")
	}
}

func TestReadPrivateKey_notFound(t *testing.T) {
	_, err := ReadPrivateKey(filepath.Join(t.TempDir(), "missing.pem"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReadPrivateKey_invalidPEM(t *testing.T) {
	p := filepath.Join(t.TempDir(), "bad.pem")
	if err := os.WriteFile(p, []byte("-----BEGIN RSA PRIVATE KEY-----\nAAAA\n-----END RSA PRIVATE KEY-----"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := ReadPrivateKey(p)
	if err == nil {
		t.Fatal("expected error")
	}
}
