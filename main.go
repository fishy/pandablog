package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"

	"go.yhsif.com/pandablog/app"
	"go.yhsif.com/pandablog/app/lib/envdetect"
	"go.yhsif.com/pandablog/app/lib/timezone"
	"go.yhsif.com/pandablog/app/logging"
)

func init() {
	// Set the time zone.
	timezone.Set()
}

func main() {
	var logLevel slog.Level
	flag.TextVar(&logLevel, "log-level", slog.LevelDebug, "minimal log level to keep")
	flag.Parse()

	if envdetect.RunningLocalDev() {
		logging.InitText(logLevel)
	} else {
		logging.InitJSON(logLevel)
	}

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
