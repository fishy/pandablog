// Package openmoji provides image urls for emojis.
package openmoji

import (
	"fmt"
	"net/url"
	"strings"
)

// EmojiResources represents the image resources for an emoji.
type EmojiResources struct {
	MimeType string
	URLLarge string
	URLSmall string
}

var filesMap = func() map[string]struct{} {
	m := make(map[string]struct{}, len(files))
	for _, file := range files {
		m[strings.TrimSuffix(file, ".png")] = struct{}{}
	}
	return m
}()

const (
	pngMimeType      = "image/png"
	pngTemplateLarge = "https://raw.githubusercontent.com/hfg-gmuend/openmoji/15.1.0/color/618x618/%s.png"
	pngTemplateSmall = "https://raw.githubusercontent.com/hfg-gmuend/openmoji/15.1.0/color/72x72/%s.png"

	svgMimeType      = "image/svg+xml"
	svgTemplateLarge = "%s?emoji=%s&size=618px"
	svgTemplateSmall = "%s?emoji=%s&size=72px"
)

func Load(emoji string, svgBaseURL string) EmojiResources {
	var runes []string
	for _, r := range emoji {
		runes = append(runes, fmt.Sprintf("%04X", r))
	}
	// Find the longest matching in files from the beginning
	for len(runes) > 0 {
		filename := strings.Join(runes, "-")
		if _, ok := filesMap[filename]; ok {
			return EmojiResources{
				MimeType: pngMimeType,
				URLLarge: fmt.Sprintf(pngTemplateLarge, filename),
				URLSmall: fmt.Sprintf(pngTemplateSmall, filename),
			}
		}
		runes = runes[:len(runes)-1]
	}
	// Fallback to builtin svg
	emojiEncoded := url.QueryEscape(emoji)
	return EmojiResources{
		MimeType: svgMimeType,
		URLLarge: fmt.Sprintf(svgTemplateLarge, svgBaseURL, emojiEncoded),
		URLSmall: fmt.Sprintf(svgTemplateSmall, svgBaseURL, emojiEncoded),
	}
}
