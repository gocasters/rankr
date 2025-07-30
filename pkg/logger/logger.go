/*
Package logger is responsible to log everything.
*/
package logger

import (
	"io"
	"log"
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
func Init(cfg Config) {
	once.Do(func() {
		workingDir, err := os.Getwd()
		if err != nil {
			log.Fatalf("Error getting current working directory: %v", err)
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
}

// L returns the global logger instance.
func L() *slog.Logger {
	return globalLogger
}

// New creates a new logger instance for each service with specific settings.
func New(cfg Config) *slog.Logger {
	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
	}

	fileWriter := &lumberjack.Logger{
		Filename:  filepath.Join(workingDir, cfg.FilePath),
		LocalTime: cfg.UseLocalTime,
		MaxSize:   cfg.FileMaxSizeInMB,
		MaxAge:    cfg.FileMaxAgeInDays,
	}
	return slog.New(
		slog.NewJSONHandler(io.MultiWriter(fileWriter), &slog.HandlerOptions{}),
	)
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
