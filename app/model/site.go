package model

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// DefaultFooter is the default Footer.
const DefaultFooter = `Powered by [🐼](https://github.com/fishy/pandablog), theme by [🐻](https://bearblog.dev/)`

// Site -
type Site struct {
	Title             string          `json:"title"`
	Subtitle          string          `json:"subtitle"`
	Author            string          `json:"author"`
	Favicon           string          `json:"favicon"`
	Description       string          `json:"description"`
	Scheme            string          `json:"scheme"`
	URL               string          `json:"url"`
	HomeURL           string          `json:"homeurl"`
	LoginURL          string          `json:"loginurl"`
	GoogleAnalyticsID string          `json:"googleanalytics"`
	DisqusID          string          `json:"disqus"`
	CactusSiteName    string          `json:"cactus"`
	Created           time.Time       `json:"created"`
	Updated           time.Time       `json:"updated"`
	Content           string          `json:"content"` // Home content.
	Styles            string          `json:"styles"`
	StylesAppend      bool            `json:"stylesappend"`
	StackEdit         bool            `json:"stackedit"`
	Prism             bool            `json:"prism"`
	ISODate           bool            `json:"isodate"`
	Posts             map[string]Post `json:"posts"`

	Footer *string `json:"footer"`

	postsLock sync.RWMutex `json:"-"`
}

// SiteURL -
func (s *Site) SiteURL() string {
	return fmt.Sprintf("%v://%v", s.Scheme, s.URL)
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
	s.postsLock.RLock()
	arr := make(PostList, 0, len(s.Posts))
	for _, v := range s.Posts {
		if v.Published && !v.Page {
			arr = append(arr, v)
		}
	}
	s.postsLock.RUnlock()

	sort.Sort(sort.Reverse(arr))

	return arr
}

// PublishedPages -
func (s *Site) PublishedPages() []Post {
	s.postsLock.RLock()
	var arr PostList
	for _, v := range s.Posts {
		if v.Published && v.Page {
			arr = append(arr, v)
		}
	}
	s.postsLock.RUnlock()

	sort.Sort(sort.Reverse(arr))

	return arr
}

// PostsAndPages -
func (s *Site) PostsAndPages(onlyPublished bool) PostWithIDList {
	s.postsLock.RLock()
	arr := make(PostWithIDList, 0, len(s.Posts))
	for k, v := range s.Posts {
		if onlyPublished && !v.Published {
			continue
		}

		p := PostWithID{Post: v, ID: k}
		arr = append(arr, p)
	}
	s.postsLock.RUnlock()

	sort.Sort(sort.Reverse(arr))

	return arr
}

// Tags -
func (s *Site) Tags(onlyPublished bool) TagList {
	s.postsLock.RLock()
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
	s.postsLock.RUnlock()

	// Create unsorted tag list.
	arr := make(TagList, 0, len(m))
	for _, v := range m {
		arr = append(arr, v)
	}

	// Sort by name.
	sort.Sort(arr)

	return arr
}

// PostBySlug -
func (s *Site) PostBySlug(slug string) PostWithID {
	s.postsLock.RLock()
	defer s.postsLock.RUnlock()

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
	s.postsLock.RLock()
	defer s.postsLock.RUnlock()

	post, ok := s.Posts[id]
	return post, ok
}

// UpdatePost - use nil to delete the post, otherwise add/update it.
func (s *Site) UpdatePost(id string, post *Post) {
	s.postsLock.Lock()
	defer s.postsLock.Unlock()

	if post == nil {
		delete(s.Posts, id)
	} else {
		s.Posts[id] = *post
	}
}
