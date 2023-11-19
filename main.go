package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"

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
		slog.Debug(
			"Read build info",
			"string", bi.String(),
			"json", bi,
		)
	} else {
		slog.Warn("Unable to read build info")
	}

	handler, err := app.Boot(context.Background())
	if err != nil {
		slog.Error("Failed to boot", "err", err)
		os.Exit(1)
	}

	// Start the web server.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.Info("Web server running", "port", port)
	slog.Info("Web server exited", "err", http.ListenAndServe(":"+port, handler))
}
