package slogx

import (
	"log/slog"
	"strings"
)

// LevelNames maps log levels to their custom string representations.
type LevelNames map[slog.Level]string

const (
	// LevelTrace represents the trace level of logging.
	LevelTrace = slog.Level(-8)
	// LevelFatal represents the fatal level of logging.
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
