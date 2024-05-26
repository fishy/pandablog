package model_test

import (
	"testing"

	"go.yhsif.com/pandablog/app/model"
)

func TestTagCompare(t *testing.T) {
	for _, c := range []struct {
		label string
		want  int
		left  model.Tag
		right model.Tag
	}{
		{
			label: "equal-zero",
		},
		{
			label: "equal",
			want:  0,
			left: model.Tag{
				Name: "foo",
			},
			right: model.Tag{
				Name: "foo",
			},
		},
		{
			label: "name<",
			want:  -1,
			left: model.Tag{
				Name: "A",
			},
			right: model.Tag{
				Name: "b",
			},
		},
		{
			label: "name>",
			want:  1,
			left: model.Tag{
				Name: "b",
			},
			right: model.Tag{
				Name: "A",
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
