package route

import (
	"log/slog"
	"net/http"
	"time"
)

func ifModifiedSince(r *http.Request) time.Time {
	v := r.Header.Get("if-modified-since")
	if v == "" {
		return time.Time{}
	}
	t, err := http.ParseTime(v)
	if err != nil {
		slog.WarnContext(
			r.Context(),
			"invalid if-modified-since header",
			"if-modified-since", v,
		)
		return time.Time{}
	}
	return t
}

func handleConditionalGet(w http.ResponseWriter, r *http.Request, lastModified time.Time) (status int) {
	lastModified = lastModified.Round(time.Second)
	if ifModifiedSince(r).Before(lastModified) {
		w.Header().Set("last-modified", lastModified.Format(http.TimeFormat))
		return 0
	}
	status = http.StatusNotModified
	w.WriteHeader(status)
	return status
}
