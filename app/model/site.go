package model

import (
	"fmt"
	"net/url"
	"slices"
	"sync"
	"time"

	"go.yhsif.com/pandablog/app/lib/openmoji"
)

// DefaultFooter is the default Footer.
const DefaultFooter = `Powered by [🐼](https://github.com/fishy/pandablog), theme by [🐻](https://bearblog.dev/), favicon by [🙋](https://openmoji.org)`

// Site -
type Site struct {
	Title             string    `json:"title"`
	Subtitle          string    `json:"subtitle"`
	Author            string    `json:"author"`
	FediCreator       string    `json:"fedicreator"`
	Favicon           string    `json:"favicon"`
	Description       string    `json:"description"`
	Scheme            string    `json:"scheme"`
	URL               string    `json:"url"`
	HomeURL           string    `json:"homeurl"`
	LoginURL          string    `json:"loginurl"`
	GoogleAnalyticsID string    `json:"googleanalytics"`
	DisqusID          string    `json:"disqus"`
	CactusSiteName    string    `json:"cactus"`
	Created           time.Time `json:"created"`
	Updated           time.Time `json:"updated"`
	Content           string    `json:"content"` // Home content.
	Styles            string    `json:"styles"`
	StylesAppend      bool      `json:"stylesappend"`
	StackEdit         bool      `json:"stackedit"`
	Prism             bool      `json:"prism"`
	ISODate           bool      `json:"isodate"`
	Lang              string    `json:"lang"`

	BridgyFedDomain string `json:"bridgyFedDomain"`
	BridgyFedWeb    string `json:"bridgyFedWeb"`

	WebmentionDomain string `json:"webmentionDomain,omitempty"`
	IndieLoginURI    string `json:"indieLoginURI,omitempty"`

	Footer *string `json:"footer"`

	lock           sync.RWMutex             `json:"-"`
	Posts          map[string]Post          `json:"posts"`
	emojiResources *openmoji.EmojiResources `json:"-"`
}

// SiteURL -
func (s *Site) SiteURL(post *Post) string {
	url := fmt.Sprintf("%v://%v", s.Scheme, s.URL)
	if post != nil {
		url += "/" + post.URL
	}
	return url
}

// SiteTitle -
func (s *Site) SiteTitle() string {
	return s.Title
}

// SiteSubtitle -
func (s *Site) SiteSubtitle() string {
	return s.Subtitle
}

// FooterMarkdown returns the markdown content of the footer.
func (s *Site) FooterMarkdown() string {
	if s.Footer == nil {
		return DefaultFooter
	}
	return *s.Footer
}

// PublishedPosts -
func (s *Site) PublishedPosts() []Post {
	s.lock.RLock()
	arr := make([]Post, 0, len(s.Posts))
	for _, v := range s.Posts {
		if v.Published && !v.Page {
			arr = append(arr, v)
		}
	}
	s.lock.RUnlock()

	slices.SortFunc(arr, func(left, right Post) int {
		return left.Compare(right)
	})

	return arr
}

// PublishedPages -
func (s *Site) PublishedPages() []Post {
	s.lock.RLock()
	var arr []Post
	for _, v := range s.Posts {
		if v.Published && v.Page {
			arr = append(arr, v)
		}
	}
	s.lock.RUnlock()

	slices.SortFunc(arr, func(left, right Post) int {
		return left.Compare(right)
	})

	return arr
}

// PostsAndPages -
func (s *Site) PostsAndPages(onlyPublished bool) []PostWithID {
	s.lock.RLock()
	arr := make([]PostWithID, 0, len(s.Posts))
	for k, v := range s.Posts {
		if onlyPublished && !v.Published {
			continue
		}

		p := PostWithID{Post: v, ID: k}
		arr = append(arr, p)
	}
	s.lock.RUnlock()

	slices.SortFunc(arr, func(left, right PostWithID) int {
		return left.Compare(right.Post)
	})

	return arr
}

// Tags -
func (s *Site) Tags(onlyPublished bool) TagList {
	s.lock.RLock()
	// Get unique values.
	m := make(map[string]Tag)
	for _, v := range s.Posts {
		if onlyPublished && !v.Published {
			continue
		}

		for _, t := range v.Tags {
			m[t.Name] = t
		}
	}
	s.lock.RUnlock()

	// Create unsorted tag list.
	arr := make(TagList, 0, len(m))
	for _, v := range m {
		arr = append(arr, v)
	}

	// Sort by name.
	slices.SortFunc(arr, func(left, right Tag) int {
		return left.Compare(right)
	})

	return arr
}

// PostBySlug -
func (s *Site) PostBySlug(slug string) PostWithID {
	s.lock.RLock()
	defer s.lock.RUnlock()

	// FIXME: This needs to be optimized.
	var p PostWithID
	for k, v := range s.Posts {
		if v.URL == slug {
			p = PostWithID{
				Post: v,
				ID:   k,
			}
			break
		}
	}

	return p
}

// PostByID -
func (s *Site) PostByID(id string) (Post, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	post, ok := s.Posts[id]
	return post, ok
}

// UpdatePost - use nil to delete the post, otherwise add/update it.
func (s *Site) UpdatePost(id string, post *Post) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if post == nil {
		delete(s.Posts, id)
	} else {
		s.Posts[id] = *post
	}
}

// BridgyFedURL constructs bridgy fed url from BridgyFedDomain, for example
// `"https://fed.brid.gy/"`, or `""` if BridgyFedDomain is unset.
func (s *Site) BridgyFedURL(path, query string) string {
	if s.BridgyFedDomain == "" {
		return ""
	}
	if path == "" {
		path = "/"
	}
	return (&url.URL{
		Scheme:   "https",
		Host:     s.BridgyFedDomain,
		Path:     path,
		RawQuery: query,
	}).String()
}

// LastUpdate returns the last update time for the site itself or any of the
// posts, in UTC.
func (s *Site) LastModified() time.Time {
	s.lock.RLock()
	defer s.lock.RUnlock()

	modified := s.Updated
	for _, p := range s.Posts {
		if updated := p.Updated; updated.After(modified) {
			modified = updated
		}
	}
	return modified.UTC()
}

func (s *Site) EmojiResources() openmoji.EmojiResources {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.emojiResources == nil {
		resources := openmoji.Load(s.Favicon, fmt.Sprintf("%s/icon.svg", s.SiteURL(nil)))
		s.emojiResources = &resources
	}
	return *s.emojiResources
}

func (s *Site) Update() {
	s.lock.Lock()
	defer s.lock.Unlock()

	resources := openmoji.Load(s.Favicon, fmt.Sprintf("%s/icon.svg", s.SiteURL(nil)))
	s.emojiResources = &resources
	s.Updated = time.Now()
}
