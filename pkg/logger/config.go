package logger

import (
	"time"

	"github.com/rs/zerolog"
)

const (
	defaultDir         = "/app/data/logs"
	defaultFilename    = "app.log"
	defaultServiceName = "main"
	defaultTimeFormat  = time.RFC3339
	defaultMaxSizeMB   = 50
	defaultMaxBackups  = 10
	defaultMaxAgeDays  = 30
)

const (
	LevelDebug = "DEBUG"
	LevelInfo  = "INFO"
	LevelWarn  = "WARNING"
	LevelError = "ERROR"
	LevelFatal = "FATAL"
	LevelPanic = "PANIC"
)

type Config struct {
	Dir               string
	Filename          string
	Level             string
	MaxSizeMB         int
	MaxBackups        int
	MaxAgeDays        int
	Compress          bool
	DuplicateToStdout bool
	TimeFormat        string
	ServiceName       string
}

func validateConfig(cfg Config) Config {
	if cfg.Dir == "" {
		cfg.Dir = defaultDir
	}

	if cfg.Filename == "" {
		cfg.Filename = defaultFilename
	}

	if cfg.TimeFormat == "" {
		cfg.TimeFormat = defaultTimeFormat
	}

	if cfg.MaxSizeMB == 0 {
		cfg.MaxSizeMB = defaultMaxSizeMB
	}

	if cfg.MaxBackups == 0 {
		cfg.MaxBackups = defaultMaxBackups
	}

	if cfg.MaxAgeDays == 0 {
		cfg.MaxAgeDays = defaultMaxAgeDays
	}

	if cfg.ServiceName == "" {
		cfg.ServiceName = defaultServiceName
	}

	return cfg
}

func parseLogLevel(s string) zerolog.Level {
	switch s {
	case LevelDebug:
		return zerolog.DebugLevel
	case LevelInfo:
		return zerolog.InfoLevel
	case LevelWarn:
		return zerolog.WarnLevel
	case LevelError:
		return zerolog.ErrorLevel
	case LevelFatal:
		return zerolog.FatalLevel
	case LevelPanic:
		return zerolog.PanicLevel
	default:
		return zerolog.DebugLevel
	}
}
