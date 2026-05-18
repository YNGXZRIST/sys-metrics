package secure

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

func writeTestRSAPEM(t *testing.T) (pubPath, privPath string) {
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

func TestNewRequestEncryptor_emptyPath(t *testing.T) {
	e, err := NewRequestEncryptor("")
	if err != nil {
		t.Fatalf("NewRequestEncryptor: %v", err)
	}
	if e == nil || e.IsEnabled {
		t.Fatalf("want disabled encryptor, got %#v", e)
	}
	if _, err := e.Encrypt([]byte("x")); err == nil {
		t.Fatal("Encrypt without key: want error")
	}
}

func TestNewRequestEncryptor_missingFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "missing.pub.pem")
	_, err := NewRequestEncryptor(p)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewRequestEncryptor_success(t *testing.T) {
	pubPath, _ := writeTestRSAPEM(t)
	e, err := NewRequestEncryptor(pubPath)
	if err != nil {
		t.Fatal(err)
	}
	if !e.IsEnabled {
		t.Fatal("want enabled encryptor")
	}
	cipher, err := e.Encrypt([]byte("payload"))
	if err != nil {
		t.Fatal(err)
	}
	if len(cipher) == 0 {
		t.Fatal("empty ciphertext")
	}
}

func TestNewRequestDecryptor_emptyPath(t *testing.T) {
	d, err := NewRequestDecryptor("")
	if err != nil {
		t.Fatalf("NewRequestDecryptor: %v", err)
	}
	if d == nil || d.IsEnabled {
		t.Fatalf("want disabled decryptor, got %#v", d)
	}
	if _, err := d.Decrypt([]byte("x")); err == nil {
		t.Fatal("Decrypt without key: want error")
	}
}

func TestNewRequestDecryptor_missingFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "missing.pem")
	_, err := NewRequestDecryptor(p)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewRequestDecryptor_success(t *testing.T) {
	_, privPath := writeTestRSAPEM(t)
	d, err := NewRequestDecryptor(privPath)
	if err != nil {
		t.Fatal(err)
	}
	if !d.IsEnabled {
		t.Fatal("want enabled decryptor")
	}
}

func TestEncryptDecrypt_roundTrip(t *testing.T) {
	pubPath, privPath := writeTestRSAPEM(t)
	enc, err := NewRequestEncryptor(pubPath)
	if err != nil {
		t.Fatal(err)
	}
	dec, err := NewRequestDecryptor(privPath)
	if err != nil {
		t.Fatal(err)
	}
	plain := []byte(`{"id":"m","type":"gauge","value":1}`)
	blob, err := enc.Encrypt(plain)
	if err != nil {
		t.Fatal(err)
	}
	got, err := dec.Decrypt(blob)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if !bytes.Equal(got, plain) {
		t.Fatalf("plaintext mismatch: %q vs %q", got, plain)
	}
}

func TestDecrypt_invalidCiphertext(t *testing.T) {
	_, privPath := writeTestRSAPEM(t)
	dec, err := NewRequestDecryptor(privPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := dec.Decrypt([]byte("short")); err == nil {
		t.Fatal("expected error for garbage input")
	}
}

func TestDecrypt_nilPrivateKey(t *testing.T) {
	d := &RequestDecryptor{IsEnabled: true}
	if _, err := d.Decrypt([]byte("x")); err == nil {
		t.Fatal("expected error")
	}
}
