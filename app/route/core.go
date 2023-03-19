package route

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"go.yhsif.com/pandablog/app/lib/datastorage"
	"go.yhsif.com/pandablog/app/lib/htmltemplate"
	"go.yhsif.com/pandablog/app/lib/router"
	"go.yhsif.com/pandablog/app/lib/websession"
	"go.yhsif.com/pandablog/assets"
)

// Core -
type Core struct {
	Router  *router.Mux
	Storage *datastorage.Storage
	Render  *htmltemplate.Engine
	Sess    *websession.Session
}

// Register all routes.
func Register(storage *datastorage.Storage, sess *websession.Session, tmpl *htmltemplate.Engine) (*Core, error) {
	// Create core app.
	c := &Core{
		Router:  setupRouter(tmpl),
		Storage: storage,
		Render:  tmpl,
		Sess:    sess,
	}

	// Register routes.
	site, err := storage.Site.Load(context.Background())
	if err != nil {
		return nil, err
	}
	registerHomePost(&HomePost{c}, site.HomeURL)
	registerStyles(&Styles{c})
	registerAuthUtil(&AuthUtil{c})
	registerXMLUtil(&XMLUtil{c})
	registerAdminPost(&AdminPost{c})
	registerPost(&Post{c}, site.HomeURL)

	return c, nil
}

func setupRouter(tmpl *htmltemplate.Engine) *router.Mux {
	// Set the handling of all responses.
	customServeHTTP := func(w http.ResponseWriter, r *http.Request, status int, err error) {
		// Handle only errors.
		if status >= 400 {
			vars := make(map[string]any)
			vars["title"] = fmt.Sprint(status)
			errTemplate := "400"
			if status == 404 {
				errTemplate = "404"
			}
			status, err = tmpl.ErrorTemplate(w, r, "base", errTemplate, vars)
			if err != nil {
				log.Println(err.Error())
				http.Error(w, "500 internal server error", http.StatusInternalServerError)
				return
			}
		}

		// Display server errors.
		if status >= 500 {
			if err != nil {
				log.Println(err.Error())
			}
		}
		// Only add the content type for non assets files.
		if !strings.HasPrefix(r.URL.Path, "/assets/") {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		}

	}

	// Send all 404 to the customer handler.
	notFound := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		customServeHTTP(w, r, http.StatusNotFound, nil)
	})

	// Set up the router.
	rr := router.New(customServeHTTP, notFound)

	// Static assets.
	rr.Get("/assets...", func(w http.ResponseWriter, r *http.Request) (status int, err error) {
		// Don't allow directory browsing.
		if strings.HasSuffix(r.URL.Path, "/") {
			return http.StatusNotFound, nil
		}

		// Use the root directory.
		fsys, err := fs.Sub(assets.CSS, ".")
		if err != nil {
			return http.StatusInternalServerError, err
		}

		// Get the requested file name.
		fname := strings.TrimPrefix(r.URL.Path, "/assets/")

		// Open the file.
		f, err := fsys.Open(fname)
		if err != nil {
			return http.StatusNotFound, nil
		}
		defer f.Close()

		// Get the file time.
		st, err := f.Stat()
		if err != nil {
			return http.StatusInternalServerError, err
		}

		http.ServeContent(w, r, fname, st.ModTime(), f.(io.ReadSeeker))
		return
	})

	return rr
}
