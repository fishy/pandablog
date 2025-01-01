package route

import (
	"net/http"
)

// Styles -
type Styles struct {
	*Core
}

func registerStyles(c *Styles) {
	c.Router.Get("/dashboard/styles", c.edit)
	c.Router.Post("/dashboard/styles", c.update)
}

func (c *Styles) edit(w http.ResponseWriter, r *http.Request) (status int, err error) {
	site, err := c.Storage.Site.Load(r.Context())
	if err != nil {
		return http.StatusInternalServerError, err
	}

	vars := make(map[string]any)
	vars["title"] = "Site styles"
	vars["token"] = c.Sess.SetCSRF(r)
	vars["favicon"] = site.Favicon
	vars["styles"] = site.Styles
	vars["stylesappend"] = site.StylesAppend
	vars["stackedit"] = site.StackEdit
	vars["prism"] = site.Prism

	return c.Render.Template(w, r, "dashboard", "styles_edit", vars)
}

func (c *Styles) update(w http.ResponseWriter, r *http.Request) (status int, err error) {
	site, err := c.Storage.Site.Load(r.Context())
	if err != nil {
		return http.StatusInternalServerError, err
	}

	r.ParseForm()

	// CSRF protection.
	success := c.Sess.CSRF(r)
	if !success {
		return http.StatusBadRequest, nil
	}

	site.Favicon = r.FormValue("favicon")
	site.Styles = r.FormValue("styles")
	site.StylesAppend = (r.FormValue("stylesappend") == "on")
	site.StackEdit = (r.FormValue("stackedit") == "on")
	site.Prism = (r.FormValue("prism") == "on")

	site.Update()

	if err := c.Storage.Save(site); err != nil {
		return http.StatusInternalServerError, err
	}

	http.Redirect(w, r, "/dashboard/styles", http.StatusFound)
	return http.StatusFound, nil
}
