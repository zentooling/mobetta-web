package logging

import (
	"log/slog"
	"os"
)

// TODO make singleton?

var GlobalLoggingLevel *slog.LevelVar

func Initialize(logLevel string) *slog.LevelVar {

	// this logger can be used to change runtime setting through web
	loggingLevel := new(slog.LevelVar)
	GlobalLoggingLevel = loggingLevel // this value can be changed at runtime so we keep global access
	loggingLevel.Set(slog.LevelDebug) // default to debug
	level, err := StringToLevel(logLevel)
	if err == nil {
		loggingLevel.Set(level)
	}

	// set default logger level - TODO make this configurable
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   true,
		Level:       loggingLevel, // Set the default log level to DEBUG
		ReplaceAttr: nil,
	}))
	// Set the logger as the default
	slog.SetDefault(logger)

	return loggingLevel
}

func StringToLevel(s string) (slog.Level, error) {
	var level slog.Level
	var err = level.UnmarshalText([]byte(s))
	return level, err
}
func GetLoggingLevel() slog.Level {
	return GlobalLoggingLevel.Level()
}
func SetLoggingLevel(levelStr string) {
	level, err := StringToLevel(levelStr)
	if err != nil {
		panic(err)
	}

	// TODO make singleton?
	GlobalLoggingLevel.Set(level)
}
