package websession

import (
	"bytes"
	"testing"
)

func TestEncrypt(t *testing.T) {
	key := "59f3726ba3f8271ddf32224b809c42e9ef4523865c74cb64e9d7d5a031f1f706"
	raw := []byte("hello")

	en := NewEncryptedStorage(key)

	enc, err := en.Encrypt(raw)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	dec, err := en.Decrypt(enc)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}
	if got, want := dec, raw; !bytes.Equal(got, want) {
		t.Errorf("Decrypted got %q want %q", dec, raw)
	}
}
