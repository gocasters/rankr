/*
Package logger is responsible to log everything.
*/
package logger

import (

	"fmt"
	"io"

	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	globalLogger *slog.Logger
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
	})

	return initError
}

// L returns the global logger instance.
func L() (*slog.Logger, error) {
	if globalLogger == nil {
		return nil, fmt.Errorf("globalLogger is null")
	}
	return globalLogger, nil
}

// New creates a new logger instance for each service with specific settings.
func New(cfg Config) (*slog.Logger, error) {
	var newErr error
	var workingDir string
	workingDir, newErr = os.Getwd()
	if newErr != nil {
		return nil, fmt.Errorf("error getting current working directory: %w", newErr)

	}

	fileWriter := &lumberjack.Logger{
		Filename:  filepath.Join(workingDir, cfg.FilePath),
		LocalTime: cfg.UseLocalTime,
		MaxSize:   cfg.FileMaxSizeInMB,
		MaxAge:    cfg.FileMaxAgeInDays,
	}

	level := mapLevel(cfg.Level)
	return slog.New(
		slog.NewJSONHandler(io.MultiWriter(fileWriter), &slog.HandlerOptions{Level: level}),
	), nil

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
