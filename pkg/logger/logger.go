package logger

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

type txContextKey struct{}

var (
	errNilBaseLogger = errors.New("base logger nil")
)

var (
	baseLogger *zerolog.Logger
)

func Get() zerolog.Logger {
	if baseLogger == nil {
		panic(errNilBaseLogger)
	}

	return *baseLogger
}

func Named(name string) zerolog.Logger {
	if baseLogger == nil {
		panic(errNilBaseLogger)
	}

	return Get().With().Str("name", name).Logger()
}

func WrapToContext(ctx context.Context, logger zerolog.Logger) context.Context {
	return context.WithValue(ctx, txContextKey{}, logger)
}

func FromContext(ctx context.Context) zerolog.Logger {
	logger, ok := ctx.Value(txContextKey{}).(zerolog.Logger)
	if !ok {
		panic(errNilBaseLogger)
	}

	return logger
}

func Init(cfg Config) error {
	cfg = validateConfig(cfg)

	if err := os.MkdirAll(cfg.Dir, 0o755); err != nil {
		return fmt.Errorf("mkdir '%s' failed to logs dir: %w", cfg.Dir, err)
	}

	path := filepath.Join(cfg.Dir, cfg.Filename)

	rot := &lumberjack.Logger{
		Filename:   path,
		MaxSize:    cfg.MaxSizeMB,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAgeDays,
		Compress:   cfg.Compress,
	}

	var w io.Writer = rot

	if cfg.DuplicateToStdout {
		w = io.MultiWriter(os.Stdout, rot)
	}

	zerolog.TimeFieldFormat = cfg.TimeFormat

	logger := zerolog.New(w).
		Level(parseLogLevel(cfg.Level)).
		With().
		Timestamp().
		Logger()

	baseLogger = &logger

	return nil
}
