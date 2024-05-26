package route

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/matryer/way"

	"go.yhsif.com/pandablog/app/lib/envdetect"
	"go.yhsif.com/pandablog/app/model"
)

// AdminPost -
type AdminPost struct {
	*Core
}

func registerAdminPost(c *AdminPost) {
	c.Router.Get("/dashboard/posts", c.index)
	c.Router.Get("/dashboard/posts/new", c.create)
	c.Router.Post("/dashboard/posts/new", c.store)
	c.Router.Get("/dashboard/posts/:id", c.edit)
	c.Router.Post("/dashboard/posts/:id", c.update)
	c.Router.Get("/dashboard/posts/:id/delete", c.destroy)
}

func (c *AdminPost) index(w http.ResponseWriter, r *http.Request) (status int, err error) {
	site, err := c.Storage.Site.Load(r.Context())
	if err != nil {
		return http.StatusInternalServerError, err
	}

	vars := make(map[string]any)
	vars["title"] = "Posts"
	vars["posts"] = site.PostsAndPages(false)

	return c.Render.Template(w, r, "dashboard", "bloglist_edit", vars)
}

func (c *AdminPost) create(w http.ResponseWriter, r *http.Request) (status int, err error) {
	vars := make(map[string]any)
	vars["title"] = "New post"
	vars["token"] = c.Sess.SetCSRF(r)

	site, err := c.Core.Storage.Site.Load(r.Context())
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to load site", "err", err)
		return http.StatusInternalServerError, err
	}
	vars["bridgyFed"] = site.BridgyFedDomain != ""

	return c.Render.Template(w, r, "dashboard", "post_create", vars)
}

func (c *AdminPost) store(w http.ResponseWriter, r *http.Request) (status int, err error) {
	site, err := c.Storage.Site.Load(r.Context())
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to load site", "err", err)
		return http.StatusInternalServerError, err
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return http.StatusInternalServerError, err
	}

	r.ParseForm()

	// CSRF protection.
	success := c.Sess.CSRF(r)
	if !success {
		return http.StatusBadRequest, nil
	}

	now := time.Now()

	var p model.Post
	p.Title = r.FormValue("title")
	p.URL = r.FormValue("slug")
	p.Canonical = r.FormValue("canonical_url")
	p.Created = now
	p.Updated = now
	pubDate := r.FormValue("published_date")
	if pubDate == "" {
		pubDate = now.Format("2006-01-02")
	}
	ts, err := time.Parse("2006-01-02", pubDate)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	p.Timestamp = ts
	p.Lang = r.FormValue("lang")
	p.Content = r.FormValue("content")
	p.Tags = p.Tags.Split(r.FormValue("tags"))
	p.Page = r.FormValue("is_page") == "on"
	p.Published = r.FormValue("publish") == "on"

	// Save to storage.
	site.UpdatePost(id.String(), &p)
	if err := c.Storage.Save(site); err != nil {
		return http.StatusInternalServerError, err
	}

	if p.Published && site.BridgyFedDomain != "" && r.FormValue("skip_webmention") != "on" {
		sendBridgyFedWebmention(r.Context(), p, site)
	}

	http.Redirect(w, r, "/dashboard/posts/"+id.String(), http.StatusFound)
	return http.StatusFound, nil
}

func (c *AdminPost) edit(w http.ResponseWriter, r *http.Request) (status int, err error) {
	site, err := c.Storage.Site.Load(r.Context())
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to load site", "err", err)
		return http.StatusInternalServerError, err
	}

	vars := make(map[string]any)
	vars["title"] = "Edit post"
	vars["token"] = c.Sess.SetCSRF(r)

	id := way.Param(r.Context(), "id")
	p, ok := site.PostByID(id)
	if !ok {
		return http.StatusNotFound, nil
	}

	vars["id"] = id
	vars["ptitle"] = p.Title
	vars["url"] = p.URL
	vars["canonical"] = p.Canonical
	vars["timestamp"] = p.Timestamp
	vars["lang"] = p.Lang
	vars["body"] = p.Content
	vars["tags"] = p.Tags.String()
	vars["page"] = p.Page
	vars["published"] = p.Published
	vars["bridgyFed"] = site.BridgyFedDomain != ""

	return c.Render.Template(w, r, "dashboard", "post_edit", vars)
}

func (c *AdminPost) update(w http.ResponseWriter, r *http.Request) (status int, err error) {
	site, err := c.Storage.Site.Load(r.Context())
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to load site", "err", err)
		return http.StatusInternalServerError, err
	}

	id := way.Param(r.Context(), "id")
	p, ok := site.PostByID(id)
	if !ok {
		return http.StatusNotFound, nil
	}

	// Save the site.
	r.ParseForm()

	// CSRF protection.
	success := c.Sess.CSRF(r)
	if !success {
		return http.StatusBadRequest, nil
	}

	now := time.Now()

	p.Title = r.FormValue("title")
	p.URL = r.FormValue("slug")
	p.Canonical = r.FormValue("canonical_url")
	p.Updated = now
	pubDate := r.FormValue("published_date")
	ts, err := time.Parse("2006-01-02", pubDate)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	p.Timestamp = ts
	p.Lang = r.FormValue("lang")
	p.Content = r.FormValue("content")
	p.Tags = p.Tags.Split(r.FormValue("tags"))
	p.Page = r.FormValue("is_page") == "on"
	p.Published = r.FormValue("publish") == "on"

	site.UpdatePost(id, &p)

	if err := c.Storage.Save(site); err != nil {
		return http.StatusInternalServerError, err
	}

	if p.Published && site.BridgyFedDomain != "" && r.FormValue("skip_webmention") != "on" {
		sendBridgyFedWebmention(r.Context(), p, site)
	}

	http.Redirect(w, r, "/dashboard/posts/"+id, http.StatusFound)
	return http.StatusFound, nil
}

func (c *AdminPost) destroy(w http.ResponseWriter, r *http.Request) (status int, err error) {
	site, err := c.Storage.Site.Load(r.Context())
	if err != nil {
		return http.StatusInternalServerError, err
	}

	id := way.Param(r.Context(), "id")
	if _, ok := site.PostByID(id); !ok {
		return http.StatusNotFound, nil
	}

	site.UpdatePost(id, nil)

	if err := c.Storage.Save(site); err != nil {
		return http.StatusInternalServerError, err
	}

	http.Redirect(w, r, "/dashboard/posts", http.StatusFound)
	return http.StatusFound, nil
}

var httpClient http.Client

func sendBridgyFedWebmention(ctx context.Context, post model.Post, site *model.Site) {
	const (
		timeout   = time.Second
		readLimit = 1024

		postFormContentType = "application/x-www-form-urlencoded"
	)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	form := make(url.Values)
	form.Set("source", site.SiteURL(&post))
	form.Set("target", site.BridgyFedURL("/" /* path */, "" /* query */))
	encodedForm := form.Encode()
	endpoint := site.BridgyFedURL("/webmention", "" /* query */)
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		strings.NewReader(encodedForm),
	)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"Failed to create Bridgy Fed webmention request",
			"err", err,
			"url", endpoint,
			"form", encodedForm,
		)
	}
	req.Header.Set("Content-Type", postFormContentType)

	if envdetect.RunningLocalDev() {
		slog.InfoContext(
			ctx,
			"Skip sending Bridgy Fed webmention request in local dev mode",
			"url", endpoint,
			"form", encodedForm,
			"req", req,
		)
		return
	}

	start := time.Now()
	resp, err := httpClient.Do(req)
	took := time.Since(start)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"Bridgy Fed webmention request failed",
			"err", err,
			"url", endpoint,
			"req", req,
			"took", took,
		)
		return
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, readLimit))
	if status := resp.StatusCode; status >= 400 {
		slog.WarnContext(
			ctx,
			"Bridgy Fed webmention request returned status >= 400",
			"url", endpoint,
			"status", status,
			"body", string(body),
			"response headers", resp.Header,
			"took", took,
		)
	} else {
		slog.InfoContext(
			ctx,
			"Bridgy Fed webmention request returned",
			"url", endpoint,
			"status", status,
			"body", string(body),
			"response headers", resp.Header,
			"took", took,
		)
	}
}
