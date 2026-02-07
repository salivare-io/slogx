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
	LevelTrace:      "TRACE",
	slog.LevelDebug: "DEBUG",
	slog.LevelInfo:  "INFO",
	slog.LevelWarn:  "WARN",
	slog.LevelError: "ERROR",
	LevelFatal:      "FATAL",
}

// getLevelName восстанавливает имя: ищет в кастомных, иначе берет стандартное
func getLevelName(l slog.Level, customNames LevelNames) string {
	if name, ok := customNames[l]; ok {
		return name
	}

	return strings.ToUpper(l.String())
}
