package route

import (
	"net/http"
	"strings"

	"github.com/matryer/way"

	"go.yhsif.com/pandablog/app/lib/htmltemplate"
	"go.yhsif.com/pandablog/app/model"
)

// Post -
type Post struct {
	*Core
}

func registerPost(c *Post, homeURL string) {
	if homeURL != "" {
		c.Router.Get("/", c.index)
	}
	c.Router.Get("/blog", c.index)
	c.Router.Get("/:slug", c.show)
}

func (c *Post) index(w http.ResponseWriter, r *http.Request) (status int, err error) {
	site, err := c.Storage.Site.Load(r.Context())
	if err != nil {
		return http.StatusInternalServerError, err
	}

	vars := make(map[string]any)
	vars["tags"] = site.Tags(true)

	// Determine if there is query.
	if q := r.URL.Query().Get("q"); len(q) > 0 {
		vars["query"] = q
		// Don't show tags when there is a filter.
		delete(vars, "tags")

		posts := make([]model.PostWithID, 0)
		for _, v := range site.PostsAndPages(true) {
			match := false
			for _, tag := range v.Tags {
				if tag.Name == q {
					match = true
					break
				}
			}

			if match {
				posts = append(posts, v)
			}
		}

		vars["posts"] = posts
	} else {
		vars["posts"] = site.PublishedPosts()
	}

	vars["siteLang"] = site.Lang

	return c.Render.Template(w, r, "base", "bloglist_index", vars)
}

func (c *Post) show(w http.ResponseWriter, r *http.Request) (status int, err error) {
	site, err := c.Storage.Site.Load(r.Context())
	if err != nil {
		return http.StatusInternalServerError, err
	}

	slug := way.Param(r.Context(), "slug")
	p := site.PostBySlug(slug)

	// Determine if in preview mode.
	preview := false
	if q := r.URL.Query().Get("preview"); len(q) > 0 && strings.ToLower(q) == "true" {
		preview = true
	}

	// Show 404 if not published and not in preview mode.
	if !p.Published && !preview {
		return http.StatusNotFound, nil
	}

	vars := make(map[string]any)
	// Don't show certain items on pages.
	if !p.Page {
		vars["title"] = p.Title
		vars["pubdate"] = p.Timestamp
	}

	vars["tags"] = p.Tags
	vars["canonical"] = p.Canonical
	vars["id"] = p.ID
	vars["posturl"] = p.URL
	vars["metadescription"] = htmltemplate.PlaintextBlurb(p.Content)
	vars["siteLang"] = site.Lang
	vars["postLang"] = p.Lang

	return c.Render.Post(w, r, "base", p.Post, vars)
}
