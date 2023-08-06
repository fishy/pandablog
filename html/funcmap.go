package html

import (
	"html/template"
	"net/http"
	"os"
	"time"

	"go.yhsif.com/pandablog/app/lib/datastorage"
	"go.yhsif.com/pandablog/app/lib/envdetect"
	"go.yhsif.com/pandablog/app/lib/websession"
	"go.yhsif.com/pandablog/app/model"
)

// FuncMap returns a map of template functions that can be used in templates.
func FuncMap(
	r *http.Request,
	storage *datastorage.Storage,
	sess *websession.Session,
) (template.FuncMap, error) {
	site, err := storage.Site.Load(r.Context())
	if err != nil {
		return nil, err
	}

	fm := make(template.FuncMap)
	fm["Stamp"] = func(t time.Time) string {
		return t.Format("2006-01-02")
	}
	fm["StampHuman"] = func(t time.Time) string {
		if site.ISODate {
			return t.Format("2006-01-02")
		}
		return t.Format("02 Jan, 2006")
	}
	fm["PublishedPages"] = func() []model.Post {
		return site.PublishedPages()
	}
	fm["HomeURL"] = func() string {
		site := site
		if site.HomeURL != "" {
			return site.HomeURL
		}
		return "/"
	}
	fm["SiteURL"] = func() string {
		return site.SiteURL()
	}
	fm["SiteTitle"] = func() string {
		return site.SiteTitle()
	}
	fm["SiteSubtitle"] = func() string {
		return site.SiteSubtitle()
	}
	fm["SiteDescription"] = func() string {
		return site.Description
	}
	fm["SiteAuthor"] = func() string {
		return site.Author
	}
	fm["SiteFavicon"] = func() string {
		return site.Favicon
	}
	fm["SiteLang"] = func() string {
		return site.Lang
	}
	fm["Authenticated"] = func() bool {
		// If user is not authenticated, don't allow them to access the page.
		_, loggedIn := sess.User(r)
		return loggedIn
	}
	fm["GoogleAnalyticsID"] = func() string {
		if envdetect.RunningLocalDev() {
			return ""
		}
		return site.GoogleAnalyticsID
	}
	fm["DisqusID"] = func() string {
		if envdetect.RunningLocalDev() {
			return ""
		}
		return site.DisqusID
	}
	fm["CactusSiteName"] = func() string {
		if envdetect.RunningLocalDev() {
			return ""
		}
		return site.CactusSiteName
	}
	fm["MFAEnabled"] = func() bool {
		return len(os.Getenv("PBB_MFA_KEY")) > 0
	}
	fm["AssetStamp"] = func(f string) string {
		return assetTimePath(f)
	}
	fm["SiteStyles"] = func() template.CSS {
		return template.CSS(site.Styles)
	}
	fm["StylesAppend"] = func() bool {
		site := site
		if len(site.Styles) == 0 {
			// If there are no style, then always append.
			return true
		} else if site.StylesAppend {
			// Else if there are style and it's append, then append.
			return true
		}
		return false
	}
	fm["EnableStackEdit"] = func() bool {
		return site.StackEdit
	}
	fm["EnablePrism"] = func() bool {
		return site.Prism
	}

	return fm, nil
}
