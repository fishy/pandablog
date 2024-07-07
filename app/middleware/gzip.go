package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"go.yhsif.com/ctxslog"
)

// Gzip Compression
type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// Gzip -
// Source: https://gist.github.com/bryfry/09a650eb8aac0fb76c24
func Gzip(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if header := r.Header.Get("Accept-Encoding"); !strings.Contains(header, "gzip") {
			ctx := ctxslog.Attach(
				r.Context(),
				"gzip", false,
			)
			if header != "" {
				ctx = ctxslog.Attach(
					ctx,
					"accept-encoding", header,
				)
			}
			handler.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		gzw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		r = r.WithContext(ctxslog.Attach(
			r.Context(),
			"gzip", true,
		))
		handler.ServeHTTP(gzw, r)
		gz.Close()
	})
}
