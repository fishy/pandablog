package datastorage

import (
	"context"
	"encoding/json"
	"os"
	"sync/atomic"
	"time"

	"go.yhsif.com/stalecache"

	"go.yhsif.com/pandablog/app/lib/envdetect"
	"go.yhsif.com/pandablog/app/model"
)

// Datastorer reads and writes data to an object.
type Datastorer interface {
	Save([]byte) error
	Load() ([]byte, error)
}

// Storage represents a writable and readable object.
type Storage struct {
	Site *stalecache.Cache[model.Site]

	datastorer Datastorer
	reload     atomic.Int32
}

const (
	defaultCacheTTL = 1 * time.Minute
	cacheTTLEnv     = "PBB_CACHE_TTL"
)

func validateSite(site *model.Site) {
	// Set the defaults for the site object.
	// Save to storage. Ensure the posts exists first so it doesn't error.
	if site.Posts == nil {
		site.Posts = make(map[string]model.Post)
	}
	// Ensure redirects don't try to happen if the scheme is empty.
	if site.Scheme == "" {
		site.Scheme = "http"
	}
	// Ensure it's set to the login page works.
	if site.LoginURL == "" {
		site.LoginURL = "admin"
	}
}

// New returns a writable and readable site object. Returns an error if the
// object cannot be initially read.
func New(ds Datastorer) (*Storage, error) {
	ttlString := os.Getenv(cacheTTLEnv)
	ttl := defaultCacheTTL
	if ttlString != "" {
		if newTTL, err := time.ParseDuration(ttlString); err == nil {
			ttl = newTTL
		}
	}
	s := &Storage{
		datastorer: ds,
	}
	s.Site = stalecache.New(
		func(context.Context) (*model.Site, error) {
			b, err := s.datastorer.Load()
			if err != nil {
				return nil, err
			}

			site := new(model.Site)
			err = json.Unmarshal(b, site)
			if err != nil {
				return nil, err
			}

			validateSite(site)
			return site, nil
		},
		stalecache.WithTTL[model.Site](ttl),
		stalecache.WithValidator(func(context.Context, *model.Site, time.Time) (fresh bool) {
			return !s.reload.CompareAndSwap(1, 0)
		}),
	)

	if _, err := s.Site.Load(context.Background()); err != nil {
		return nil, err
	}

	return s, nil
}

// Save writes the site object to the data storage and returns an error if it
// cannot be written.
func (s *Storage) Save(site *model.Site) error {
	var b []byte
	var err error

	if envdetect.RunningLocalDev() {
		// Indent so the data is easy to read.
		b, err = json.MarshalIndent(site, "", "    ")
	} else {
		b, err = json.Marshal(site)
	}

	if err != nil {
		return err
	}

	if err := s.datastorer.Save(b); err != nil {
		return err
	}
	s.Site.Update(site)

	return nil
}

// InvalidateSite invalidates the site cache and force a reload on next load.
func (s *Storage) InvalidateSite() {
	s.reload.Store(1)
}
