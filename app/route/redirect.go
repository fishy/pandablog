package route

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

func (c *Core) registerBridyFedRedirect() {
	// Ref: https://fed.brid.gy/docs#fediverse-enhanced
	redir := func(w http.ResponseWriter, r *http.Request) (status int, err error) {
		site, err := c.Storage.Site.Load(r.Context())
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to load site", "err", err)
			return http.StatusInternalServerError, err
		}
		domain := site.BridgyFedRedirect
		if domain == "" {
			return http.StatusNotFound, nil
		}
		if err := r.ParseForm(); err != nil {
			slog.ErrorContext(r.Context(), "failed to parse form", "err", err)
			return http.StatusInternalServerError, err
		}
		status = http.StatusFound
		url := (&url.URL{
			Scheme:   "https",
			Host:     domain,
			Path:     r.URL.Path,
			RawQuery: r.Form.Encode(),
		}).String()
		w.Header().Set("Location", url)
		w.WriteHeader(status)
		io.WriteString(w, fmt.Sprintf("%s: %s", http.StatusText(status), url))
		return status, nil
	}
	c.Router.Get("/.well-known/host-meta", redir)
	c.Router.Get("/.well-known/webfinger", redir)
}
