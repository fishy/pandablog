package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.yhsif.com/pandablog/app/middleware"
)

func TestNewSession(t *testing.T) {
	r := httptest.NewRequest("HEAD", "/", nil)
	w := httptest.NewRecorder()
	mux := http.NewServeMux()
	mw := middleware.Head(mux)
	mw.ServeHTTP(w, r)
	if got, want := w.Result().StatusCode, http.StatusOK; got != want {
		t.Errorf("StatusCode got %v want %v", got, want)
	}
}
