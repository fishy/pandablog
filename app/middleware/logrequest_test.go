package middleware

import (
	"net/http"
	"net/netip"
	"testing"
)

func TestReadIP(t *testing.T) {
	genReq := func(remoteAddr, xForwardedFor string) *http.Request {
		req := &http.Request{
			RemoteAddr: remoteAddr,
		}
		if xForwardedFor != "" {
			req.Header = make(http.Header)
			req.Header.Set("x-forwarded-for", xForwardedFor)
		}
		return req
	}
	for _, c := range []struct {
		label string
		req   *http.Request
		want  netip.Addr
	}{
		{
			label: "empty",
			req:   genReq("", ""),
		},
		{
			label: "2-x-forwarded-for",
			req:   genReq("", "8.8.4.4,8.8.8.8"),
			want:  netip.MustParseAddr("8.8.8.8"),
		},
		{
			label: "spaces",
			req:   genReq("", " 8.8.4.4 ,\t8.8.8.8\n"),
			want:  netip.MustParseAddr("8.8.8.8"),
		},
		{
			label: "remote-addr",
			req:   genReq("8.8.8.8:1234", "127.0.0.1"),
			want:  netip.MustParseAddr("8.8.8.8"),
		},
		{
			label: "skip-local",
			req:   genReq("", "8.8.8.8,127.0.0.1"),
			want:  netip.MustParseAddr("8.8.8.8"),
		},
	} {
		t.Run(c.label, func(t *testing.T) {
			got := realIP(c.req)
			if got.Compare(c.want) != 0 {
				t.Errorf("got %v want %v", got, c.want)
			}
		})
	}
}
