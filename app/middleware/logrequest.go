package middleware

import (
	"net/http"
	"net/netip"
	"strings"
	"time"

	"golang.org/x/exp/slog"
)

func realIP(r *http.Request) netip.Addr {
	// First, try X-Forwarded-For header
	// Note that cloud run appends the real ip to the end
	xForwardedFor := r.Header.Get("x-forwarded-for")
	split := strings.Split(xForwardedFor, ",")
	for i := len(split) - 1; i >= 0; i-- {
		ip := strings.TrimSpace(split[i])
		addr, err := netip.ParseAddr(ip)
		if err != nil {
			slog.Default().Debug(
				"Wrong forwarded ip",
				"x-forwarded-for", xForwardedFor,
				"ip", ip,
			)
			continue
		}
		if addr.IsPrivate() || addr.IsLoopback() {
			continue
		}
		return addr
	}

	// Next, use the one from r.RemoteIP
	if parsed, err := netip.ParseAddrPort(r.RemoteAddr); err != nil {
		slog.Default().Debug(
			"Cannot parse RemoteAddr",
			"remoteAddr", r.RemoteAddr,
		)
		return netip.Addr{}
	} else {
		return parsed.Addr()
	}
}

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
func (c *Handler) LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := realIP(r)
		rw := &responseWriterWrapper{ResponseWriter: w}
		defer func(start time.Time) {
			slog.Default().Info(
				"request",
				"duration", time.Since(start),
				"method", r.Method,
				"url", r.URL.String(),
				"ip", ip.String(),
				"userAgent", r.UserAgent(),
				"code", rw.getCode(),
			)
		}(time.Now())

		next.ServeHTTP(rw, r)
	})
}
