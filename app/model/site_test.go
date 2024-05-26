package model_test

import (
	"testing"

	"go.yhsif.com/pandablog/app/model"
)

func TestSiteURL(t *testing.T) {
	s := new(model.Site)
	s.Scheme = "http"
	s.URL = "localhost"
	if got, want := s.SiteURL(nil /* post */), "http://localhost"; got != want {
		t.Errorf("SiteURL() got %q want %q", got, want)
	}
	if got, want := s.SiteURL(&model.Post{URL: "foo"}), "http://localhost/foo"; got != want {
		t.Errorf("SiteURL() got %q want %q", got, want)
	}
}
