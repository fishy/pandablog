package websession_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.yhsif.com/pandablog/app/lib/datastorage"
	"go.yhsif.com/pandablog/app/lib/websession"
)

func TestNewJSONSession(t *testing.T) {
	// Use local filesytem when developing.
	f := filepath.Join(t.TempDir(), "data.bin")
	if err := os.WriteFile(f, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create file %q: %v", f, err)
	}
	ss := datastorage.NewLocalStorage(f)

	// Set up the session storage provider.
	secretkey := "82a18fbbfed2694bb15d512a70c53b1a088e669966918d3d474564b2ac44349b"
	en := websession.NewEncryptedStorage(secretkey)
	store, err := websession.NewJSONSession(ss, en)
	if err != nil {
		t.Fatalf("Failed to create json session: %v", err)
	}

	token := "abc"
	data := "hello"
	now := time.Now()

	if err := store.Commit(token, []byte(data), now); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	b, exists, err := store.Find(token)
	if err != nil {
		t.Fatalf("store.Find failed: %v", err)
	}
	if !exists {
		t.Error("store.Find returned false on exists")
	}
	if got, want := string(b), data; got != data {
		t.Errorf("store.Find got %q want %q", got, want)
	}

	if err := store.Delete(token); err != nil {
		t.Fatalf("store.Delete failed: %v", err)
	}

	_, exists, err = store.Find(token)
	if err != nil {
		t.Fatalf("store.Find failed: %v", err)
	}
	if exists {
		t.Errorf("store.Find returned true on exits")
	}
}
