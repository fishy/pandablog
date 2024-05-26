package model_test

import (
	"testing"
	"time"

	"go.yhsif.com/pandablog/app/model"
)

func TestPostCompare(t *testing.T) {
	for _, c := range []struct {
		label string
		want  int
		left  model.Post
		right model.Post
	}{
		{
			label: "equal-zero",
		},
		{
			label: "equal",
			want:  0,
			left: model.Post{
				Title:     "foo",
				Timestamp: time.Date(2024, 5, 26, 0, 0, 0, 0, time.UTC),
			},
			right: model.Post{
				Title:     "foo",
				Timestamp: time.Date(2024, 5, 26, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			label: "title<",
			want:  -1,
			left: model.Post{
				Title:     "1",
				Timestamp: time.Date(2024, 5, 26, 0, 0, 0, 0, time.UTC),
			},
			right: model.Post{
				Title:     "2",
				Timestamp: time.Date(2024, 5, 26, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			label: "title>",
			want:  1,
			left: model.Post{
				Title:     "2",
				Timestamp: time.Date(2024, 5, 26, 0, 0, 0, 0, time.UTC),
			},
			right: model.Post{
				Title:     "1",
				Timestamp: time.Date(2024, 5, 26, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			label: "timestamp<",
			want:  1,
			left: model.Post{
				Title:     "1",
				Timestamp: time.Date(2024, 5, 25, 0, 0, 0, 0, time.UTC),
			},
			right: model.Post{
				Title:     "2",
				Timestamp: time.Date(2024, 5, 26, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			label: "timestamp>",
			want:  -1,
			left: model.Post{
				Title:     "2",
				Timestamp: time.Date(2024, 5, 26, 0, 0, 0, 0, time.UTC),
			},
			right: model.Post{
				Title:     "1",
				Timestamp: time.Date(2024, 5, 25, 0, 0, 0, 0, time.UTC),
			},
		},
	} {
		t.Run(c.label, func(t *testing.T) {
			got := c.left.Compare(c.right)
			switch {
			default:
				// == 0
				if got == 0 {
					return
				}
			case c.want < 0:
				if got < 0 {
					return
				}
			case c.want > 0:
				if got > 0 {
					return
				}
			}
			t.Errorf("%+v.Compare(%+v) got %d want %d", c.left, c.right, got, c.want)
		})
	}
}
