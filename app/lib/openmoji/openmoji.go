// Package openmoji provides image urls for emojis.
package openmoji

import (
	"bufio"
	_ "embed"
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

// The file list is coming from
// https://github.com/hfg-gmuend/openmoji/releases/download/15.1.0/openmoji-72x72-color.zip
//
// It can be updated like this:
//
//	curl -LO https://github.com/hfg-gmuend/openmoji/releases/download/15.1.0/openmoji-72x72-color.zip
//	mkdir tmp
//	unzip openmoji-72x72-color.zip -d tmp/
//	ls tmp/ > files.txt
//	rm -rf openmoji-72x72-color.zip tmp/
//
//go:embed files.txt
var files string

var filesMap = func() map[string]struct{} {
	scanner := bufio.NewScanner(strings.NewReader(files))
	m := make(map[string]struct{})
	for scanner.Scan() {
		m[strings.TrimSuffix(strings.TrimSpace(scanner.Text()), ".png")] = struct{}{}
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
