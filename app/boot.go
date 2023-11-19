package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/alexedwards/scs/v2"

	"go.yhsif.com/pandablog/app/lib/blocklist"
	"go.yhsif.com/pandablog/app/lib/datastorage"
	"go.yhsif.com/pandablog/app/lib/envdetect"
	"go.yhsif.com/pandablog/app/lib/htmltemplate"
	"go.yhsif.com/pandablog/app/lib/websession"
	"go.yhsif.com/pandablog/app/middleware"
	"go.yhsif.com/pandablog/app/route"
	"go.yhsif.com/pandablog/html"
)

var (
	storageSitePath    = "storage/site.json"
	storageSessionPath = "storage/session.bin"
	sessionName        = "session"
)

// Boot -
func Boot(ctx context.Context) (http.Handler, error) {
	// Set the storage and session environment variables.
	sitePath := os.Getenv("PBB_SITE_PATH")
	if len(sitePath) > 0 {
		storageSitePath = sitePath
	}

	sessionPath := os.Getenv("PBB_SESSION_PATH")
	if len(sessionPath) > 0 {
		storageSessionPath = sessionPath
	}

	sname := os.Getenv("PBB_SESSION_NAME")
	if len(sname) > 0 {
		sessionName = sname
	}

	// Get the environment variables.
	secretKey := os.Getenv("PBB_SESSION_KEY")
	if len(secretKey) == 0 {
		return nil, fmt.Errorf("environment variable missing: %v", "PBB_SESSION_KEY")
	}

	bucket := os.Getenv("PBB_GCP_BUCKET_NAME")
	if len(bucket) == 0 {
		return nil, fmt.Errorf("environment variable missing: %v", "PBB_GCP_BUCKET_NAME")
	}

	allowHTML, err := strconv.ParseBool(os.Getenv("PBB_ALLOW_HTML"))
	if err != nil {
		return nil, fmt.Errorf("environment variable not able to parse as bool: %v", "PBB_ALLOW_HTML")
	}

	// Create new store object with the defaults.
	var ds datastorage.Datastorer
	var ss websession.Sessionstorer

	if !envdetect.RunningLocalDev() {
		// Use Google when running in GCP.
		ds = datastorage.NewGCPStorage(bucket, storageSitePath)
		ss = datastorage.NewGCPStorage(bucket, storageSessionPath)
	} else {
		// Use local filesytem when developing.
		ds = datastorage.NewLocalStorage(storageSitePath)
		ss = datastorage.NewLocalStorage(storageSessionPath)
	}

	// Set up the data storage provider.
	storage, err := datastorage.New(ds)
	if err != nil {
		return nil, err
	}

	// Set up the session storage provider.
	en := websession.NewEncryptedStorage(secretKey)
	store, err := websession.NewJSONSession(ss, en)
	if err != nil {
		return nil, err
	}

	// Initialize a new session manager and configure the session lifetime.
	sessionManager := scs.New()
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.Persist = false
	sessionManager.Store = store
	sess := websession.New(sessionName, sessionManager)

	// Set up the template engine.
	tm := html.NewTemplateManager(storage, sess)
	tmpl := htmltemplate.New(tm, allowHTML)

	// Load blocklist
	b := loadBlocklist(ctx)

	// Setup the routes.
	c, err := route.Register(storage, sess, tmpl)
	if err != nil {
		return nil, err
	}

	// Set up the router and middleware.
	site, err := c.Storage.Site.Load(ctx)
	if err != nil {
		return nil, err
	}
	var mw http.Handler
	mw = c.Router
	h := middleware.NewHandler(c.Render, c.Sess, c.Router, site.URL, site.Scheme)
	mw = h.Redirect(mw)
	mw = middleware.Head(mw)
	mw = h.DisallowAnon(mw)
	mw = sessionManager.LoadAndSave(mw)
	mw = b.Middleware(mw)
	mw = middleware.Gzip(mw)
	mw = h.LogRequest(mw)

	return mw, nil
}

func loadBlocklist(ctx context.Context) blocklist.Blocklist {
	const path = "blocklist.yaml"
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// no blocklist, return an empty one
			slog.InfoContext(ctx, "blocklist file does not exist", "path", path)
			return blocklist.Blocklist{}
		} else {
			slog.ErrorContext(ctx, "failed to open blocklist file", "err", err, "path", path)
			return blocklist.Blocklist{}
		}
	}
	defer f.Close()
	b, err := blocklist.ParseYAML(ctx, f)
	if err != nil {
		slog.ErrorContext(ctx, "failed to parse blocklist file", "err", err, "path", path)
		return blocklist.Blocklist{}
	}
	return b
}
