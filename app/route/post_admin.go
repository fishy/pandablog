package route

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/matryer/way"

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

	return c.Render.Template(w, r, "dashboard", "post_create", vars)
}

func (c *AdminPost) store(w http.ResponseWriter, r *http.Request) (status int, err error) {
	site, err := c.Storage.Site.Load(r.Context())
	if err != nil {
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

	http.Redirect(w, r, "/dashboard/posts/"+id.String(), http.StatusFound)
	return http.StatusFound, nil
}

func (c *AdminPost) edit(w http.ResponseWriter, r *http.Request) (status int, err error) {
	site, err := c.Storage.Site.Load(r.Context())
	if err != nil {
		return http.StatusInternalServerError, err
	}

	vars := make(map[string]any)
	vars["title"] = "Edit post"
	vars["token"] = c.Sess.SetCSRF(r)

	ID := way.Param(r.Context(), "id")

	var p model.Post
	var ok bool
	if p, ok = site.PostByID(ID); !ok {
		return http.StatusNotFound, nil
	}

	vars["id"] = ID
	vars["ptitle"] = p.Title
	vars["url"] = p.URL
	vars["canonical"] = p.Canonical
	vars["timestamp"] = p.Timestamp
	vars["lang"] = p.Lang
	vars["body"] = p.Content
	vars["tags"] = p.Tags.String()
	vars["page"] = p.Page
	vars["published"] = p.Published

	return c.Render.Template(w, r, "dashboard", "post_edit", vars)
}

func (c *AdminPost) update(w http.ResponseWriter, r *http.Request) (status int, err error) {
	site, err := c.Storage.Site.Load(r.Context())
	if err != nil {
		return http.StatusInternalServerError, err
	}

	ID := way.Param(r.Context(), "id")

	var p model.Post
	var ok bool
	if p, ok = site.PostByID(ID); !ok {
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

	site.UpdatePost(ID, &p)

	if err := c.Storage.Save(site); err != nil {
		return http.StatusInternalServerError, err
	}

	http.Redirect(w, r, "/dashboard/posts/"+ID, http.StatusFound)
	return http.StatusFound, nil
}

func (c *AdminPost) destroy(w http.ResponseWriter, r *http.Request) (status int, err error) {
	site, err := c.Storage.Site.Load(r.Context())
	if err != nil {
		return http.StatusInternalServerError, err
	}

	ID := way.Param(r.Context(), "id")

	var ok bool
	if _, ok = site.PostByID(ID); !ok {
		return http.StatusNotFound, nil
	}

	site.UpdatePost(ID, nil)

	if err := c.Storage.Save(site); err != nil {
		return http.StatusInternalServerError, err
	}

	http.Redirect(w, r, "/dashboard/posts", http.StatusFound)
	return http.StatusFound, nil
}
