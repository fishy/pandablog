package main

import (
	"net/http"
	"os"
	"runtime/debug"

	"golang.org/x/exp/slog"

	"go.yhsif.com/pandablog/app"
	"go.yhsif.com/pandablog/app/lib/timezone"
	"go.yhsif.com/pandablog/app/logging"
)

func init() {
	logging.InitJSON()
	// Set the time zone.
	timezone.Set()
}

func main() {
	if bi, ok := debug.ReadBuildInfo(); ok {
		slog.Default().Debug(
			"Read build info",
			"string", bi.String(),
			"json", bi,
		)
	} else {
		slog.Default().Warn("Unable to read build info")
	}

	handler, err := app.Boot()
	if err != nil {
		slog.Default().Error("Failed to boot", "err", err)
		os.Exit(1)
	}

	// Start the web server.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.Default().Info("Web server running", "port", port)
	slog.Default().Info("Web server exited", "err", http.ListenAndServe(":"+port, handler))
}
