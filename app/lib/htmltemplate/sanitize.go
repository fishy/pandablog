package htmltemplate

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	blackfriday "github.com/russross/blackfriday/v2"
)

// RenderMarkdown renders markdown to html
func (te *Engine) RenderMarkdown(markdown string) template.HTML {
	// Ensure unit line endings are used when pulling out of JSON.
	markdownWithUnixLineEndings := strings.ReplaceAll(markdown, "\r\n", "\n")
	htmlCode := blackfriday.Run([]byte(markdownWithUnixLineEndings))

	// Sanitize by removing HTML if true.
	if !te.allowUnsafeHTML {
		htmlCode = bluemonday.UGCPolicy().SanitizeBytes(htmlCode)
	}
	return template.HTML(htmlCode)
}

// sanitizedContent returns a sanitized content block or an error is one occurs.
func (te *Engine) sanitizedContent(t *template.Template, markdown string) (*template.Template, error) {
	htmlCode := te.RenderMarkdown(markdown)

	// Change delimiters temporarily so code samples can use Go blocks.
	safeContent := fmt.Sprintf(`[{[{define "content"}]}]%s[{[{end}]}]`, htmlCode)
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
