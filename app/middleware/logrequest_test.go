package middleware

import (
	"net/http"
	"testing"
)

func TestGoodHeadersInRequest(t *testing.T) {
	for _, c := range []struct {
		label   string
		headers map[string]string
		want    string
	}{
		{
			label: "empty",
			want:  "(none)",
		},
		{
			label: "gzip",
			headers: map[string]string{
				"accept-encoding": "gzip",
			},
			want: "Accept-Encoding",
		},
		{
			label: "if-modified-since",
			headers: map[string]string{
				"if-modified-since": "foo",
			},
			want: "If-Modified-Since",
		},
		{
			label: "both",
			headers: map[string]string{
				"accept-encoding":   "gzip",
				"if-modified-since": "foo",
			},
			want: "Accept-Encoding, If-Modified-Since",
		},
	} {
		t.Run(c.label, func(t *testing.T) {
			r := &http.Request{
				Header: make(http.Header),
			}
			for k, v := range c.headers {
				r.Header.Set(k, v)
			}
			if got := goodHeadersInRequest(r); got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}
