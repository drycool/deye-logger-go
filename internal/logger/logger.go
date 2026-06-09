package logger

import (
	"log/slog"
	"os"
	"sync"
)

var (
	once   sync.Once
	global *slog.Logger
)

// Init initializes the global logger. Call once at startup.
func Init(debug bool) {
	once.Do(func() {
		level := slog.LevelInfo
		if debug {
			level = slog.LevelDebug
		}
		global = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: level,
		}))
	})
}

// Get returns the global logger for a given module/component.
func Get(module string) *slog.Logger {
	if global == nil {
		Init(false)
	}
	return global.With("module", module)
}
