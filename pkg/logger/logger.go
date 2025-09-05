/*
Package logger is responsible to log everything.
*/
package logger

import (
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

var (
	globalLogger *slog.Logger
	globalWriter io.Closer
	once         sync.Once
)

type Config struct {
	Level            string `koanf:"level"`
	FilePath         string `koanf:"file_path"`
	UseLocalTime     bool   `koanf:"use_local_time"`
	FileMaxSizeInMB  int    `koanf:"file_max_size_in_mb"`
	FileMaxAgeInDays int    `koanf:"file_max_age_in_days"`
}

// Init initializes the global logger instance.
func Init(cfg Config) error {
	var initError error
	var workingDir string
	once.Do(func() {
		workingDir, initError = os.Getwd()
		if initError != nil {
			initError = fmt.Errorf("error getting current working directory: %w", initError)
			return

		}

		fileWriter := &lumberjack.Logger{
			Filename:  filepath.Join(workingDir, cfg.FilePath),
			LocalTime: cfg.UseLocalTime,
			MaxSize:   cfg.FileMaxSizeInMB,
			MaxAge:    cfg.FileMaxAgeInDays,
		}

		level := mapLevel(cfg.Level)
		globalLogger = slog.New(
			slog.NewJSONHandler(io.MultiWriter(fileWriter, os.Stdout), &slog.HandlerOptions{
				Level: level,
			}),
		)

		globalWriter = fileWriter
	})

	return initError
}

// MustInit initializes the global logger of panics if it fails.
func MustInit(cfg Config) {
	if err := Init(cfg); err != nil {
		panic(err)
	}
}

// L returns the global logger instance.
func L() *slog.Logger {
	if globalLogger == nil {
		panic("logger not initialized. Call logger.Init or logger.MustInit first")
	}
	return globalLogger
}

// Close closes the global logger file writer.
func Close() error {
	if globalWriter != nil {
		err := globalWriter.Close()
		globalWriter = nil // To prevent the writer from being closed twice.

		return err
	}

	return nil
}

// New creates a new independent logger (not singleton).
func New(cfg Config) (*slog.Logger, io.Closer, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, nil, fmt.Errorf("error getting current working directory: %w", err)

	}

	fileWriter := &lumberjack.Logger{
		Filename:  filepath.Join(workingDir, cfg.FilePath),
		LocalTime: cfg.UseLocalTime,
		MaxSize:   cfg.FileMaxSizeInMB,
		MaxAge:    cfg.FileMaxAgeInDays,
	}

	level := mapLevel(cfg.Level)
	logger := slog.New(
		slog.NewJSONHandler(io.MultiWriter(fileWriter), &slog.HandlerOptions{Level: level}),
	)

	return logger, fileWriter, nil
}

func mapLevel(levelStr string) slog.Level {
	switch levelStr {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
