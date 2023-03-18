package route

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"time"

	"go.yhsif.com/pandablog/app/model"
)

// XMLUtil -
type XMLUtil struct {
	*Core
}

func registerXMLUtil(c *XMLUtil) {
	c.Router.Get("/robots.txt", c.Robots)
	c.Router.Get("/sitemap.xml", c.sitemap)
	c.Router.Get("/rss.xml", c.rss)
}

// Robots returns a page for web crawlers.
func (c *XMLUtil) Robots(w http.ResponseWriter, r *http.Request) (status int, err error) {
	w.Header().Set("Content-Type", "text/plain")
	text :=
		`User-agent: *
Allow: /`
	fmt.Fprint(w, text)
	return
}

func (c *XMLUtil) sitemap(w http.ResponseWriter, r *http.Request) (status int, err error) {
	// Resource: https://www.sitemaps.org/protocol.html
	// Resource: https://golang.org/src/encoding/xml/example_test.go

	site, err := c.Storage.Site.Load(r.Context())
	if err != nil {
		return http.StatusInternalServerError, err
	}

	type URL struct {
		Location     string `xml:"loc"`
		LastModified string `xml:"lastmod"`
	}

	type Sitemap struct {
		XMLName xml.Name `xml:"urlset"`
		XMLNS   string   `xml:"xmlns,attr"`
		XHTML   string   `xml:"xmlns:xhtml,attr"`
		URL     []URL    `xml:"url"`
	}

	m := &Sitemap{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		XHTML: "http://www.w3.org/1999/xhtml",
	}

	// Home page
	m.URL = append(m.URL, URL{
		Location:     site.SiteURL(),
		LastModified: site.Updated.Format("2006-01-02"),
	})

	// Posts and pages
	for _, v := range site.PostsAndPages(true) {
		m.URL = append(m.URL, URL{
			Location:     site.SiteURL() + "/" + v.FullURL(),
			LastModified: v.Timestamp.Format("2006-01-02"),
		})
	}

	// Tags
	for _, v := range site.Tags(true) {
		m.URL = append(m.URL, URL{
			Location:     site.SiteURL() + "/blog?q=" + v.Name,
			LastModified: v.Timestamp.Format("2006-01-02"),
		})
	}

	output, err := xml.MarshalIndent(m, "  ", "    ")
	if err != nil {
		return http.StatusInternalServerError, err
	}

	header := []byte(xml.Header)
	output = append(header[:], output[:]...)

	w.Header().Set("Content-Type", "application/xml")
	fmt.Fprint(w, string(output))
	return
}

func (c *XMLUtil) rss(w http.ResponseWriter, r *http.Request) (status int, err error) {
	// Resource: https://www.rssboard.org/rss-specification
	// Rsource: https://validator.w3.org/feed/check.cgi

	site, err := c.Storage.Site.Load(r.Context())
	if err != nil {
		return http.StatusInternalServerError, err
	}

	type Cdata struct {
		Content string `xml:",cdata"`
	}

	type Item struct {
		Title       string `xml:"title"`
		Link        string `xml:"link"`
		PubDate     string `xml:"pubDate"`
		GUID        string `xml:"guid"`
		Description Cdata  `xml:"description"`
	}

	type AtomLink struct {
		Href string `xml:"href,attr"`
		Rel  string `xml:"rel,attr"`
		Type string `xml:"type,attr"`
	}

	type Sitemap struct {
		XMLName       xml.Name `xml:"rss"`
		Version       string   `xml:"version,attr"`
		Atom          string   `xml:"xmlns:atom,attr"`
		Title         string   `xml:"channel>title"`
		Link          string   `xml:"channel>link"`
		Description   string   `xml:"channel>description"`
		Generator     string   `xml:"channel>generator"`
		Language      string   `xml:"channel>language"`
		LastBuildDate string   `xml:"channel>lastBuildDate"`
		AtomLink      AtomLink `xml:"channel>atom:link"`
		Items         []Item   `xml:"channel>item"`
	}

	m := &Sitemap{
		Version:       "2.0",
		Atom:          "http://www.w3.org/2005/Atom",
		Title:         site.SiteTitle(),
		Link:          site.SiteURL(),
		Description:   site.Description,
		Generator:     "Polar Bear Blog - Selfhost Edition",
		Language:      "en-us",
		LastBuildDate: time.Now().Format(time.RFC1123Z),
		AtomLink: AtomLink{
			Href: site.SiteURL() + "/rss.xml",
			Rel:  "self",
			Type: "application/rss+xml",
		},
	}

	allPosts := site.PublishedPostsWithID() // Exclude Pages for RSS
	var posts []model.PostWithID

	// Determine if there is query.
	if q := r.URL.Query().Get("q"); len(q) > 0 {
		for _, v := range allPosts {
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
	} else {
		posts = allPosts
	}

	for _, v := range posts {
		html := c.Render.SanitizedHTML(v.Post.Content)
		m.Items = append(m.Items, Item{
			Title:   v.Title,
			Link:    site.SiteURL() + "/" + v.FullURL(),
			PubDate: v.Timestamp.Format(time.RFC1123Z),
			GUID:    site.SiteURL() + "/" + v.FullURL(),
			Description: Cdata{
				Content: string(html),
			},
		})
	}

	output, err := xml.MarshalIndent(m, "  ", "    ")
	if err != nil {
		return http.StatusInternalServerError, err
	}

	header := []byte(xml.Header)
	output = append(header[:], output[:]...)

	w.Header().Set("Content-Type", "application/xml")
	fmt.Fprint(w, string(output))
	return
}
