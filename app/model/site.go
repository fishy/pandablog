package model

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// Site -
type Site struct {
	Title             string          `json:"title"`
	Subtitle          string          `json:"subtitle"`
	Author            string          `json:"author"`
	Favicon           string          `json:"favicon"`
	Description       string          `json:"description"`
	Footer            string          `json:"footer"`
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
	Posts             map[string]Post `json:"posts"`

	postsLock sync.Mutex `json:"-"`
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

// PublishedPosts -
func (s *Site) PublishedPosts() []Post {
	s.postsLock.Lock()
	arr := make(PostList, 0, len(s.Posts))
	for _, v := range s.Posts {
		if v.Published && !v.Page {
			arr = append(arr, v)
		}
	}
	s.postsLock.Unlock()

	sort.Sort(sort.Reverse(arr))

	return arr
}

// PublishedPages -
func (s *Site) PublishedPages() []Post {
	s.postsLock.Lock()
	var arr PostList
	for _, v := range s.Posts {
		if v.Published && v.Page {
			arr = append(arr, v)
		}
	}
	s.postsLock.Unlock()

	sort.Sort(sort.Reverse(arr))

	return arr
}

// PostsAndPages -
func (s *Site) PostsAndPages(onlyPublished bool) PostWithIDList {
	s.postsLock.Lock()
	arr := make(PostWithIDList, 0, len(s.Posts))
	for k, v := range s.Posts {
		if onlyPublished && !v.Published {
			continue
		}

		p := PostWithID{Post: v, ID: k}
		arr = append(arr, p)
	}
	s.postsLock.Unlock()

	sort.Sort(sort.Reverse(arr))

	return arr
}

// Tags -
func (s *Site) Tags(onlyPublished bool) TagList {
	s.postsLock.Lock()
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
	s.postsLock.Unlock()

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
	s.postsLock.Lock()
	defer s.postsLock.Unlock()

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
	s.postsLock.Lock()
	defer s.postsLock.Unlock()

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
