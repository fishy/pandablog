package route

import (
	"net/http"
	"text/template"
)

// Image handler
type Image struct {
	*Core
}

func registerImage(img *Image) {
	img.Router.Get("/icon.svg", img.image)
}

var svgTmpl = template.Must(template.New("svg").Parse(`
<svg
		xmlns="http://www.w3.org/2000/svg"
		viewBox="0 0 100 100"
		{{if .Size -}}
		width="{{.Size}}"
		height="{{.Size}}"
		{{- end}}
	>
	<text y=".9em" font-size="90">{{.Favicon}}</text>
</svg>`))

type svgArgs struct {
	Size    string
	Favicon string
}

func (img *Image) image(w http.ResponseWriter, r *http.Request) (status int, err error) {
	site, err := img.Storage.Site.Load(r.Context())
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if status := handleConditionalGet(w, r, site.Updated); status > 0 {
		return status, nil
	}

	w.Header().Set("content-type", "image/svg+xml")
	if err := svgTmpl.Execute(w, svgArgs{
		Size:    r.FormValue("size"),
		Favicon: site.Favicon,
	}); err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}
