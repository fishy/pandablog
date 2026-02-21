package model

import (
	"strings"
	"time"
)

// Post -
type Post struct {
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	Canonical string    `json:"canonical"`
	Created   time.Time `json:"created"`
	Updated   time.Time `json:"updated"`
	Timestamp time.Time `json:"timestamp"`
	Lang      string    `json:"lang"`
	Content   string    `json:"content"`
	Published bool      `json:"published"`
	Page      bool      `json:"page"`
	Tags      TagList   `json:"tags"`
}

// PostWithID -
type PostWithID struct {
	Post

	ID string `json:"id"`
}

// FullURL -
func (p *Post) FullURL() string {
	return p.URL
}

func (p Post) Compare(right Post) int {
	if result := right.Timestamp.Compare(p.Timestamp); result != 0 {
		// Sort by timestamp DESC
		return result
	}
	// Otherwise, sort by title, ASC
	switch {
	default:
		return 0
	case p.Title < right.Title:
		return -1
	case p.Title > right.Title:
		return 1
	}
}

// TagList -
type TagList []Tag

func (t TagList) Len() int           { return len(t) }
func (t TagList) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t TagList) Less(i, j int) bool { return t[i].Compare(t[j]) < 0 }

// String -
func (t TagList) String() string {
	arr := make([]string, 0)
	for _, v := range t {
		arr = append(arr, v.Name)
	}

	return strings.Join(arr, ",")
}

// Split -
func (t TagList) Split(s string) TagList {
	trimmed := strings.TrimSpace(s)

	// Return an empty object since split returns 1 element when empty.
	if len(trimmed) == 0 {
		return TagList{}
	}

	ts := time.Now()

	arrTags := make([]Tag, 0)
	for v := range strings.SplitSeq(trimmed, ",") {
		arrTags = append(arrTags, Tag{
			Name:      strings.TrimSpace(v),
			Timestamp: ts,
		})
	}

	return arrTags
}
