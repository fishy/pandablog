package route

import (
	"log/slog"
	"net/http"
)

func (c *Core) registerBridyFedRedirect() {
	// Ref: https://fed.brid.gy/docs#fediverse-enhanced
	redir := func(w http.ResponseWriter, r *http.Request) (status int, err error) {
		site, err := c.Storage.Site.Load(r.Context())
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to load site", "err", err)
			return http.StatusInternalServerError, err
		}
		domain := site.BridgyFedDomain
		if domain == "" {
			return http.StatusNotFound, nil
		}
		if err := r.ParseForm(); err != nil {
			slog.ErrorContext(r.Context(), "failed to parse form", "err", err)
			return http.StatusInternalServerError, err
		}
		url := site.BridgyFedURL(r.URL.Path, r.Form.Encode())
		if url == "" {
			return http.StatusNotFound, nil
		}
		http.Redirect(w, r, url, http.StatusFound)
		return http.StatusFound, nil
	}
	c.Router.Get("/.well-known/host-meta", redir)
	c.Router.Get("/.well-known/webfinger", redir)
}
