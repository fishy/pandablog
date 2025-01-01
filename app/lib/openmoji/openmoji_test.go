package openmoji_test

import (
	"strings"
	"testing"

	"go.yhsif.com/pandablog/app/lib/openmoji"
)

func TestLoad(t *testing.T) {
	const svgBase = "svg"
	for _, c := range []struct {
		emoji string
		png   bool
	}{
		{
			emoji: "foo", // not an emoji
			png:   false,
		},
		{
			emoji: "🐼",
			png:   true,
		},
		{
			emoji: "🐼🏿",
			png:   true,
		},
		{
			emoji: "🏿🐼",
			png:   true,
		},
		{
			emoji: "\uf8ff", // This is the last entry in files.txt
			png:   true,
		},
		{
			emoji: "",
			png:   false,
		},
	} {
		t.Run(c.emoji, func(t *testing.T) {
			res := openmoji.Load(c.emoji, svgBase)
			if c.png {
				if want := "image/png"; res.MimeType != want {
					t.Errorf("MimeType got %q want %q", res.MimeType, want)
				}
				wantURL := "openmoji"
				if !strings.Contains(res.URLLarge, wantURL) {
					t.Errorf("URLLarge %q does not contain %q", res.URLLarge, wantURL)
				}
				if !strings.Contains(res.URLSmall, wantURL) {
					t.Errorf("URLSmall %q does not contain %q", res.URLSmall, wantURL)
				}
			} else {
				if want := "image/svg+xml"; res.MimeType != want {
					t.Errorf("MimeType got %q want %q", res.MimeType, want)
				}
				if !strings.HasPrefix(res.URLLarge, svgBase) {
					t.Errorf("URLLarge %q does not start with %q", res.URLLarge, svgBase)
				}
				if !strings.HasPrefix(res.URLSmall, svgBase) {
					t.Errorf("URLSmall %q does not start with %q", res.URLSmall, svgBase)
				}
			}
		})
	}
}
