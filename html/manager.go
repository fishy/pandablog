package html

import (
	"context"
	"crypto/md5"
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"go.yhsif.com/pandablog/app/lib/datastorage"
	"go.yhsif.com/pandablog/app/lib/websession"
	"go.yhsif.com/pandablog/assets"
)

//go:embed *
var templates embed.FS

// TemplateManager -
type TemplateManager struct {
	storage *datastorage.Storage
	sess    *websession.Session
}

// NewTemplateManager -
func NewTemplateManager(storage *datastorage.Storage, sess *websession.Session) *TemplateManager {
	return &TemplateManager{
		storage: storage,
		sess:    sess,
	}
}

// PartialTemplate -
func (tm *TemplateManager) PartialTemplate(r *http.Request, mainTemplate string, partialTemplate string) (*template.Template, error) {
	// Functions available in the templates.
	fm, err := FuncMap(r, tm.storage, tm.sess)
	if err != nil {
		return nil, err
	}

	baseTemplate := fmt.Sprintf("%v.tmpl", mainTemplate)
	headerTemplate := "partial/head.tmpl"
	contentTemplate := fmt.Sprintf("partial/%v.tmpl", partialTemplate)

	// Parse the main template with the functions.
	t, err := template.New(path.Base(baseTemplate)).Funcs(fm).ParseFS(templates, baseTemplate,
		headerTemplate, contentTemplate)
	if err != nil {
		return nil, err
	}

	return t, nil
}

// PostTemplate -
func (tm *TemplateManager) PostTemplate(r *http.Request, mainTemplate string) (*template.Template, error) {
	// Functions available in the templates.
	fm, err := FuncMap(r, tm.storage, tm.sess)
	if err != nil {
		return nil, err
	}

	baseTemplate := fmt.Sprintf("%v.tmpl", mainTemplate)
	headerTemplate := "partial/head.tmpl"

	// Parse the main template with the functions.
	t, err := template.New(path.Base(baseTemplate)).Funcs(fm).ParseFS(templates, baseTemplate, headerTemplate)
	if err != nil {
		return nil, err
	}

	return t, nil
}

// FooterMarkdown returns the markdown version of the footer.
func (tm *TemplateManager) FooterMarkdown(ctx context.Context) (string, error) {
	site, err := tm.storage.Site.Load(ctx)
	if err != nil {
		return "", err
	}
	return site.FooterMarkdown(), nil
}

// assetTimePath returns a URL with a MD5 hash appended.
func assetTimePath(s string) string {
	// Use the root directory.
	fsys, err := fs.Sub(assets.CSS, ".")
	if err != nil {
		return s
	}

	// Get the requested file name.
	fname := strings.TrimPrefix(s, "/assets/")

	// Open the file.
	f, err := fsys.Open(fname)
	if err != nil {
		return s
	}
	defer f.Close()

	// Get all the content.s
	b, err := io.ReadAll(f)
	if err != nil {
		return s
	}

	// Create a hash.
	hsh := md5.New()
	hsh.Write(b)

	return fmt.Sprintf("%v?%x", s, hsh.Sum(nil))
}
