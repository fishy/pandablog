package model_test

import (
	"testing"
	"time"

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

func TestSiteLastModified(t *testing.T) {
	now := time.Now()
	for _, c := range []struct {
		label string
		want  time.Time
		site  func() *model.Site
	}{
		{
			label: "empty",
			site: func() *model.Site {
				return new(model.Site)
			},
		},
		{
			label: "site",
			want:  now,
			site: func() *model.Site {
				return &model.Site{
					Updated: now,
				}
			},
		},
		{
			label: "post-after-site",
			want:  now,
			site: func() *model.Site {
				return &model.Site{
					Updated: now.Add(-time.Second),
					Posts: map[string]model.Post{
						"foo": {Updated: now},
					},
				}
			},
		},
		{
			label: "site-after-post",
			want:  now,
			site: func() *model.Site {
				return &model.Site{
					Updated: now,
					Posts: map[string]model.Post{
						"foo": {Updated: now.Add(-time.Second)},
					},
				}
			},
		},
	} {
		t.Run(c.label, func(t *testing.T) {
			got := c.site().LastModified()
			if !got.Equal(c.want) {
				t.Errorf("got %v, want %v", got, c.want)
			}
			if name, offset := got.Zone(); offset != 0 {
				t.Errorf("want UTC, got %v", name)
			}
		})
	}
}
