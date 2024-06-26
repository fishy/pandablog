package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"go.yhsif.com/ctxslog"

	"go.yhsif.com/pandablog/app/lib/envdetect"
)

type responseWriterWrapper struct {
	http.ResponseWriter

	code *int
}

func (rww *responseWriterWrapper) WriteHeader(statusCode int) {
	if rww.code == nil {
		rww.code = &statusCode
	}
	rww.ResponseWriter.WriteHeader(statusCode)
}

func (rww responseWriterWrapper) getCode() int {
	if rww.code == nil {
		return 200
	}
	return *rww.code
}

// LogRequest will log the HTTP requests.
func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		realIPFunc := ctxslog.GCPRealIP
		if envdetect.RunningLocalDev() {
			realIPFunc = ctxslog.RemoteAddrIP
		}
		ctx := ctxslog.Attach(
			r.Context(),
			"httpRequest", ctxslog.HTTPRequest(r, realIPFunc),
			"goodHeaders", goodHeadersInRequest(r),
		)
		rw := &responseWriterWrapper{ResponseWriter: w}
		defer func(start time.Time) {
			slog.InfoContext(
				ctx,
				"request",
				"duration", time.Since(start),
				"code", rw.getCode(),
			)
		}(time.Now())

		next.ServeHTTP(rw, r.WithContext(ctx))
	})
}

var goodHeaders = []string{
	"Accept-Encoding",
	"If-Modified-Since",
}

func goodHeadersInRequest(r *http.Request) string {
	var headers []string
	for _, h := range goodHeaders {
		if _, ok := r.Header[h]; ok {
			headers = append(headers, h)
		}
	}
	if len(headers) == 0 {
		return "(none)"
	}
	return strings.Join(headers, ", ")
}
