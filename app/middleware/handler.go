package middleware

import (
	"go.yhsif.com/pandablog/app/lib/htmltemplate"
	"go.yhsif.com/pandablog/app/lib/router"
	"go.yhsif.com/pandablog/app/lib/websession"
)

// Handler -
type Handler struct {
	Router     *router.Mux
	Render     *htmltemplate.Engine
	Sess       *websession.Session
	SiteURL    string
	SiteScheme string
}

// NewHandler -
func NewHandler(te *htmltemplate.Engine, sess *websession.Session, mux *router.Mux, siteURL string, siteScheme string) *Handler {
	return &Handler{
		Render:     te,
		Router:     mux,
		Sess:       sess,
		SiteURL:    siteURL,
		SiteScheme: siteScheme,
	}

}
