package route

import (
	"net/http"
	"time"

	"github.com/josephspurrier/polarbearblog/app/model"
)

// HomePost -
type HomePost struct {
	*Core
}

func registerHomePost(c *HomePost, homeURL string) {
	if homeURL == "" {
		c.Router.Get("/", c.show)
	}
	c.Router.Get("/dashboard", c.edit)
	c.Router.Post("/dashboard", c.update)
	c.Router.Get("/dashboard/reload", c.reload)
}

func (c *HomePost) show(w http.ResponseWriter, r *http.Request) (status int, err error) {
	site, err := c.Storage.Site.Load(r.Context())
	if err != nil {
		return http.StatusInternalServerError, err
	}

	p := model.Post{
		Content: site.Content,
		URL:     "/",
	}

	if p.Content == "" {
		p.Content = "*No content yet.*"
	}

	vars := make(map[string]interface{})
	return c.Render.Post(w, r, "base", p, vars)
}

func (c *HomePost) edit(w http.ResponseWriter, r *http.Request) (status int, err error) {
	site, err := c.Storage.Site.Load(r.Context())
	if err != nil {
		return http.StatusInternalServerError, err
	}

	vars := make(map[string]interface{})
	vars["title"] = "Edit site"
	vars["homeContent"] = site.Content
	vars["ptitle"] = site.Title
	vars["subtitle"] = site.Subtitle
	vars["token"] = c.Sess.SetCSRF(r)

	// Help the user set the domain based off the current URL.
	if site.URL == "" {
		vars["domain"] = r.Host
	} else {
		vars["domain"] = site.URL
	}

	vars["scheme"] = site.Scheme
	vars["pauthor"] = site.Author
	vars["pdescription"] = site.Description
	vars["loginurl"] = site.LoginURL
	vars["homeurl"] = site.HomeURL
	vars["googleanalytics"] = site.GoogleAnalyticsID
	vars["disqus"] = site.DisqusID
	vars["cactus"] = site.CactusSiteName
	vars["footer"] = site.Footer
	vars["isodate"] = site.ISODate

	return c.Render.Template(w, r, "dashboard", "home_edit", vars)
}

func (c *HomePost) update(w http.ResponseWriter, r *http.Request) (status int, err error) {
	r.ParseForm()

	// CSRF protection.
	success := c.Sess.CSRF(r)
	if !success {
		return http.StatusBadRequest, nil
	}

	site, err := c.Storage.Site.Load(r.Context())
	if err != nil {
		return http.StatusInternalServerError, err
	}

	site.Title = r.FormValue("title")
	site.Subtitle = r.FormValue("subtitle")
	site.URL = r.FormValue("domain")
	site.Content = r.FormValue("content")
	site.Scheme = r.FormValue("scheme")
	site.Author = r.FormValue("author")
	site.Description = r.FormValue("pdescription")
	site.LoginURL = r.FormValue("loginurl")
	site.HomeURL = r.FormValue("homeurl")
	site.GoogleAnalyticsID = r.FormValue("googleanalytics")
	site.DisqusID = r.FormValue("disqus")
	site.CactusSiteName = r.FormValue("cactus")
	site.Footer = r.FormValue("footer")
	site.ISODate = (r.FormValue("isodate") == "on")
	site.Updated = time.Now()

	err = c.Storage.Save(site)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	http.Redirect(w, r, "/dashboard", http.StatusFound)
	return
}

func (c *HomePost) reload(w http.ResponseWriter, r *http.Request) (status int, err error) {
	c.Storage.InvalidateSite()
	if _, err := c.Storage.Site.Load(r.Context()); err != nil {
		return http.StatusInternalServerError, err
	}

	http.Redirect(w, r, "/dashboard", http.StatusFound)
	return
}
