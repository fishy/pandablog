package blocklist

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"go.yhsif.com/ctxslog/slogtest"
)

func TestYAML(t *testing.T) {
	filename, err := filepath.Abs("../../../blocklist.yaml")
	if err != nil {
		t.Fatalf("Failed to resolve filename: %v", err)
	}
	f, err := os.Open(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			t.Skipf("%q does not exist, skipping...", filename)
		} else {
			t.Fatalf("Failed to open %q: %v", filename, err)
		}
	}
	t.Cleanup(func() {
		f.Close()
	})

	slogtest.BackupGlobalLogger(t)
	slog.SetDefault(slog.New(slogtest.Handler(t, slog.LevelWarn, slog.LevelWarn)))
	if _, err := ParseYAML(context.Background(), f); err != nil {
		t.Errorf("Failed to parse %q: %v", filename, err)
	}
}
