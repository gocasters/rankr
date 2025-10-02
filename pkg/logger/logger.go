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
	var logPath string
	once.Do(func() {
		if cfg.FilePath != "" {

			if filepath.IsAbs(cfg.FilePath) {
				initError = fmt.Errorf("absolute paths are not allowed for log file path")
				return
			}

			cleanPath := filepath.Clean(cfg.FilePath)

			workingDir, err := os.Getwd()
			if err != nil {
				initError = fmt.Errorf("error getting current working directory: %w", err)
				return
			}

			logPath = filepath.Join(workingDir, cleanPath)

			relPath, err := filepath.Rel(workingDir, logPath)
			if err != nil {
				initError = fmt.Errorf("invalid log file path: %w", err)
				return
			}

			if len(relPath) >= 2 && relPath[0] == '.' && relPath[1] == '.' {
				initError = fmt.Errorf("log file path cannot traverse outside working directory")
				return
			}
		} else {
			exePath, err := os.Executable()
			if err != nil {
				initError = fmt.Errorf("error getting executable path: %w", err)
				return
			}
			logPath = filepath.Join(filepath.Dir(exePath), "logs", "app.log")
		}

		fileWriter := &lumberjack.Logger{
			Filename:  logPath,
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

// L returns the global logger instance.
func L() *slog.Logger {
	if globalLogger == nil {
		panic("logger not initialized. Call logger.Init first")
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
	var logPath string
	if cfg.FilePath != "" {
		// Reject absolute paths
		if filepath.IsAbs(cfg.FilePath) {
			return nil, nil, fmt.Errorf("absolute paths are not allowed for log file path")
		}

		// Clean the path
		cleanPath := filepath.Clean(cfg.FilePath)

		workingDir, err := os.Getwd()
		if err != nil {
			return nil, nil, fmt.Errorf("error getting current working directory: %w", err)
		}

		// Join working directory with cleaned path
		logPath = filepath.Join(workingDir, cleanPath)

		// Verify the result does not escape the working directory
		relPath, err := filepath.Rel(workingDir, logPath)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid log file path: %w", err)
		}
		// Check for parent directory traversal
		if len(relPath) >= 2 && relPath[0] == '.' && relPath[1] == '.' {
			return nil, nil, fmt.Errorf("log file path cannot traverse outside working directory")
		}
	} else {
		exePath, err := os.Executable()
		if err != nil {
			return nil, nil, fmt.Errorf("error getting executable path: %w", err)
		}
		logPath = filepath.Join(filepath.Dir(exePath), "logs", "app.log")
	}

	fileWriter := &lumberjack.Logger{
		Filename:  logPath,
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
