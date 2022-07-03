package htmltemplate

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	blackfriday "github.com/russross/blackfriday/v2"
)

// SanitizedHTML returns a sanitized content html.
func (te *Engine) SanitizedHTML(content string) []byte {
	// Ensure unit line endings are used when pulling out of JSON.
	markdownWithUnixLineEndings := strings.Replace(content, "\r\n", "\n", -1)
	htmlCode := blackfriday.Run([]byte(markdownWithUnixLineEndings))

	// Sanitize by removing HTML if true.
	if !te.allowUnsafeHTML {
		htmlCode = bluemonday.UGCPolicy().SanitizeBytes(htmlCode)
	}
	return htmlCode
}

// sanitizedContent returns a sanitized content block or an error is one occurs.
func (te *Engine) sanitizedContent(t *template.Template, content string) (*template.Template, error) {
	htmlCode := te.SanitizedHTML(content)

	// Change delimiters temporarily so code samples can use Go blocks.
	safeContent := fmt.Sprintf(`[{[{define "content"}]}]%s[{[{end}]}]`, string(htmlCode))
	t = t.Delims("[{[{", "}]}]")
	var err error
	t, err = t.Parse(safeContent)
	if err != nil {
		return nil, err
	}
	// Reset delimiters
	t = t.Delims("{{", "}}")
	return t, nil
}
