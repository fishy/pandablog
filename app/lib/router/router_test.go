package router

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// defaultServeHTTP is the default ServeHTTP function that receives the status and error from
// the function call.
var defaultServeHTTP = func(w http.ResponseWriter, r *http.Request, status int,
	err error) {
	if status >= 400 {
		if err != nil {
			http.Error(w, err.Error(), status)
		} else {
			http.Error(w, "", status)
		}
	}
}

func TestParams(t *testing.T) {
	mux := New(defaultServeHTTP, nil)
	mux.Get("/user/:name", HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) (status int, err error) {
			if got, want := mux.Param(r, "name"), "john"; got != want {
				t.Errorf("name got %q want %q", got, want)
			}
			return http.StatusOK, nil
		}))

	r := httptest.NewRequest("GET", "/user/john", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
}

func TestInstance(t *testing.T) {
	mux := New(defaultServeHTTP, nil)

	mux.Get("/user/:name", HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) (status int, err error) {
			if got, want := mux.Param(r, "name"), "john"; got != want {
				t.Errorf("name got %q want %q", got, want)
			}
			return http.StatusOK, nil
		}))

	r := httptest.NewRequest("GET", "/user/john", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
}

func TestPostForm(t *testing.T) {
	mux := New(defaultServeHTTP, nil)

	form := url.Values{}
	form.Add("username", "jsmith")

	mux.Post("/user", HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) (status int, err error) {
			r.ParseForm()
			if got, want := r.FormValue("username"), "jsmith"; got != want {
				t.Errorf("username got %q want %q", got, want)
			}
			return http.StatusOK, nil
		}))

	r := httptest.NewRequest("POST", "/user", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
}

func TestPostJSON(t *testing.T) {
	mux := New(defaultServeHTTP, nil)

	j, err := json.Marshal(map[string]any{
		"username": "jsmith",
	})
	if err != nil {
		t.Fatalf("Failed to marshal json: %v", err)
	}

	mux.Post("/user", HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) (status int, err error) {
			b, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("Failed to read body: %v", err)
			}
			r.Body.Close()
			if got, want := string(b), `{"username":"jsmith"}`; got != want {
				t.Errorf("Body got %q want %q", got, want)
			}
			return http.StatusOK, nil
		}))

	r := httptest.NewRequest("POST", "/user", bytes.NewBuffer(j))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
}

func TestGet(t *testing.T) {
	mux := New(defaultServeHTTP, nil)

	called := false

	mux.Get("/user", HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) (status int, err error) {
			called = true
			return http.StatusOK, nil
		}))

	r := httptest.NewRequest("GET", "/user", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if got, want := called, true; got != want {
		t.Errorf("called got %v want %v", got, want)
	}
}

func TestDelete(t *testing.T) {
	mux := New(defaultServeHTTP, nil)

	called := false

	mux.Delete("/user", HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) (status int, err error) {
			called = true
			return http.StatusOK, nil
		}))

	r := httptest.NewRequest("DELETE", "/user", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if got, want := called, true; got != want {
		t.Errorf("called got %v want %v", got, want)
	}
}

func TestHead(t *testing.T) {
	mux := New(defaultServeHTTP, nil)

	called := false

	mux.Head("/user", HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) (status int, err error) {
			called = true
			return http.StatusOK, nil
		}))

	r := httptest.NewRequest("HEAD", "/user", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if got, want := called, true; got != want {
		t.Errorf("called got %v want %v", got, want)
	}
}

func TestOptions(t *testing.T) {
	mux := New(defaultServeHTTP, nil)

	called := false

	mux.Options("/user", HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) (status int, err error) {
			called = true
			return http.StatusOK, nil
		}))

	r := httptest.NewRequest("OPTIONS", "/user", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if got, want := called, true; got != want {
		t.Errorf("called got %v want %v", got, want)
	}
}

func TestPatch(t *testing.T) {
	mux := New(defaultServeHTTP, nil)

	called := false

	mux.Patch("/user", HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) (status int, err error) {
			called = true
			return http.StatusOK, nil
		}))

	r := httptest.NewRequest("PATCH", "/user", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if got, want := called, true; got != want {
		t.Errorf("called got %v want %v", got, want)
	}
}

func TestPut(t *testing.T) {
	mux := New(defaultServeHTTP, nil)

	called := false

	mux.Put("/user", HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) (status int, err error) {
			called = true
			return http.StatusOK, nil
		}))

	r := httptest.NewRequest("PUT", "/user", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if got, want := called, true; got != want {
		t.Errorf("called got %v want %v", got, want)
	}
}

func Test404(t *testing.T) {
	mux := New(defaultServeHTTP, nil)

	called := false

	mux.Get("/user", HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) (status int, err error) {
			called = true
			return http.StatusOK, nil
		}))

	r := httptest.NewRequest("GET", "/badroute", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if got, want := called, false; got != want {
		t.Errorf("called got %v want %v", got, want)
	}
	if got, want := w.Code, http.StatusNotFound; got != want {
		t.Errorf("HTTP status code got %v want %v", got, want)
	}
}

func Test500NoError(t *testing.T) {
	mux := New(defaultServeHTTP, nil)

	called := true

	mux.Get("/user", HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) (status int, err error) {
			called = true
			return http.StatusInternalServerError, nil
		}))

	r := httptest.NewRequest("GET", "/user", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if got, want := called, true; got != want {
		t.Errorf("called got %v want %v", got, want)
	}
	if got, want := w.Code, http.StatusInternalServerError; got != want {
		t.Errorf("HTTP status code got %v want %v", got, want)
	}
}

func Test500WithError(t *testing.T) {
	mux := New(defaultServeHTTP, nil)

	called := true
	specificError := errors.New("specific error")

	mux.Get("/user", HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) (status int, err error) {
			called = true
			return http.StatusInternalServerError, specificError
		}))

	r := httptest.NewRequest("GET", "/user", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if got, want := called, true; got != want {
		t.Errorf("called got %v want %v", got, want)
	}
	if got, want := w.Code, http.StatusInternalServerError; got != want {
		t.Errorf("HTTP status code got %v want %v", got, want)
	}
	if got, want := strings.TrimSpace(w.Body.String()), strings.TrimSpace(specificError.Error()); got != want {
		t.Errorf("Body got %q want %q", got, want)
	}
}

func Test400(t *testing.T) {
	notFound := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
	)

	mux := New(defaultServeHTTP, notFound)

	r := httptest.NewRequest("GET", "/unknown", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if got, want := w.Code, http.StatusNotFound; got != want {
		t.Errorf("HTTP status code got %v want %v", got, want)
	}
}

func TestNotFound(t *testing.T) {
	mux := New(defaultServeHTTP, nil)

	r := httptest.NewRequest("GET", "/unknown", nil)
	w := httptest.NewRecorder()
	mux.NotFound(w, r)

	if got, want := w.Code, http.StatusNotFound; got != want {
		t.Errorf("HTTP status code got %v want %v", got, want)
	}
}

func TestBadRequest(t *testing.T) {
	mux := New(defaultServeHTTP, nil)

	r := httptest.NewRequest("GET", "/unknown", nil)
	w := httptest.NewRecorder()
	mux.BadRequest(w, r)

	if got, want := w.Code, http.StatusBadRequest; got != want {
		t.Errorf("HTTP status code got %v want %v", got, want)
	}
}
