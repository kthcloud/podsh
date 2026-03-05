package profiles

import (
	"fmt"
	"log/slog"
	"strings"
)

func ParseLevel(s string) (slog.Level, error) {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("invalid slog level: %s", s)
	}
}

func MustParseLevel(s string) slog.Level {
	lvl, err := ParseLevel(s)
	if err != nil {
		panic(err)
	}
	return lvl
}
