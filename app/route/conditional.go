package route

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var buildTime = sync.OnceValue(func() time.Time {
	str := os.Getenv("BUILD_TIMESTAMP")
	s, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		slog.Warn("Invalid BUILD_TIMESTAMP", "err", err, "value", str)
		return time.Time{}
	}
	return time.Unix(s, 0).UTC()
})

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
	if built := buildTime(); lastModified.Before(built) {
		lastModified = built
	}
	if ifModifiedSince(r).Before(lastModified) {
		w.Header().Set("last-modified", lastModified.UTC().Format(http.TimeFormat))
		return 0
	}
	status = http.StatusNotModified
	w.WriteHeader(status)
	return status
}
