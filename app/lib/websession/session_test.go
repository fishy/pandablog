package websession_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"

	"go.yhsif.com/pandablog/app/lib/datastorage"
	"go.yhsif.com/pandablog/app/lib/websession"
)

func TestNewSession(t *testing.T) {
	// Set up the session storage provider.
	f := filepath.Join(t.TempDir(), "data.bin")
	if err := os.WriteFile(f, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create %q: %v", f, err)
	}
	ss := datastorage.NewLocalStorage(f)
	secretkey := "82a18fbbfed2694bb15d512a70c53b1a088e669966918d3d474564b2ac44349b"
	en := websession.NewEncryptedStorage(secretkey)
	store, err := websession.NewJSONSession(ss, en)
	if err != nil {
		t.Fatalf("Failed to create json session: %v", err)
	}

	// Initialize a new session manager and configure the session lifetime.
	sessionManager := scs.New()
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.Persist = false
	sessionManager.Store = store
	sess := websession.New("session", sessionManager)

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Test user
		u := "foo"
		sess.SetUser(r, u)
		user, found := sess.User(r)
		if !found {
			t.Errorf("Did not find user")
		}
		if got, want := user, u; got != want {
			t.Errorf("User got %q want %q", got, want)
		}

		// Test Logout
		sess.Logout(r)
		_, found = sess.User(r)
		if found {
			t.Errorf("Should not find user")
		}

		// Test persistence
		if sessionManager.Cookie.Persist {
			t.Error("sessionManager.Cookie.Persist should not be true")
		}
		sess.RememberMe(r, true)
		if !sessionManager.Cookie.Persist {
			t.Error("sessionManager.Cookie.Persist should not be false")
		}

		// Test CSRF
		if got, want := sess.CSRF(r), false; got != want {
			t.Errorf("sess.CSRF(r) got %v want %v", got, want)
		}
		token := sess.SetCSRF(r)
		r.Form = url.Values{}
		r.Form.Set("token", token)
		if got, want := sess.CSRF(r), true; got != want {
			t.Errorf("sess.CSRF(r) got %v want %v", got, want)
		}
	})

	mw := sessionManager.LoadAndSave(mux)
	mw.ServeHTTP(w, r)
}
