package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"go.yhsif.com/pandablog/app/lib/envdetect"
)

// Redirect will handle all redirects required.
func (c *Handler) Redirect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Currently the following code will cause trouble when this application is
		// running behind a reverse proxy and have a domain name. The reverse proxy
		// will set the Host header to the domain name, and the code below will
		// redirect to the domain name. This will cause the reverse proxy to
		// redirect to itself, and the redirect will be infinite.
		// To fix it, we need to add a configuration to tell the application that
		// it is running behind a reverse proxy, and the reverse proxy will set
		// the Host header to the domain name. Then we can skip the redirect
		// code below.
		// Redirect to the correct website.


		// if !envdetect.RunningLocalDev() && len(c.SiteURL) > 0 && !strings.Contains(r.Host, c.SiteURL){
		// 	http.Redirect(w, r, fmt.Sprintf("%v://%v%v", c.SiteScheme, c.SiteURL, r.URL.Path), http.StatusPermanentRedirect)
		// 	return
		// }
		
		if !envdetect.RunningLocalDev() && len(c.SiteURL) > 0 {
			var host string
			if len(r.Header.Get("X-Forwarded-Host")) > 0 {
			  host = r.Header.Get("X-Forwarded-Host")
			} else {
			  host = r.Host
			}
			if !strings.Contains(host, c.SiteURL) {
			  http.Redirect(w, r, fmt.Sprintf("%v://%v%v", c.SiteScheme, c.SiteURL, r.URL.Path), http.StatusPermanentRedirect)
			  return
			}
		}

		// Don't allow access to files with a slash at the end.
		if strings.Contains(r.URL.Path, ".") && strings.HasSuffix(r.URL.Path, "/") {
			c.Router.NotFound(w, r)
			return
		}

		// Strip trailing slash.
		if r.URL.Path != "/" && strings.HasSuffix(r.URL.Path, "/") {
			http.Redirect(w, r, strings.TrimRight(r.URL.Path, "/"), http.StatusPermanentRedirect)
			return
		}

		next.ServeHTTP(w, r)
	})
}
