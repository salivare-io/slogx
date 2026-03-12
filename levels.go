package slogx

import (
	"log/slog"
	"strings"
)

type LevelNames map[slog.Level]string

const (
	LevelTrace = slog.Level(-8)
	LevelFatal = slog.Level(12)
)

var defaultLevelNames = LevelNames{
	LevelTrace: "TRACE",
	LevelFatal: "FATAL",
}

// getLevelName resolves a level name using custom names first, then defaults.
func getLevelName(l slog.Level, customNames LevelNames) string {
	if name, ok := customNames[l]; ok {
		return name
	}

	return strings.ToUpper(l.String())
}
