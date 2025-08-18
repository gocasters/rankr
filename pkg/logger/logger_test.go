package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// Test mapLevel function
func TestMapLevel(t *testing.T) {
	tests := []struct {
		name     string
		levelStr string
		expected slog.Level
	}{
		{"debug level", "debug", slog.LevelDebug},
		{"info level", "info", slog.LevelInfo},
		{"warn level", "warn", slog.LevelWarn},
		{"error level", "error", slog.LevelError},
		{"unknown level defaults to info", "unknown", slog.LevelInfo},
		{"empty string defaults to info", "", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapLevel(tt.levelStr)
			if got != tt.expected {
				t.Errorf("mapLevel(%q) = %v, want %v", tt.levelStr, got, tt.expected)
			}
		})
	}
}

// Test Init initializes the global logger
func TestInit(t *testing.T) {
	// Reset global state
	globalLogger = nil
	once = sync.Once{}

	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	cfg := Config{
		Level:            "debug",
		FilePath:         "test.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}

	if err := Init(cfg); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if globalLogger == nil {
		t.Error("Init() should initialize globalLogger")
	}

	// Test sync.Once: calling Init again should not overwrite
	oldLogger := globalLogger
	if err := Init(cfg); err != nil {
		t.Fatalf("Second Init failed: %v", err)
	}
	if globalLogger != oldLogger {
		t.Error("Init() should only initialize once")
	}
}

// Test L() returns global logger correctly
func TestL(t *testing.T) {
	// Reset global state
	globalLogger = nil
	once = sync.Once{}

	// L() should return error if not initialized
	l, err := L()
	if l != nil || err == nil {
		t.Error("L() should return nil and error when logger is not initialized")
	}

	// Initialize logger
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	cfg := Config{
		Level:            "info",
		FilePath:         "test.log",
		UseLocalTime:     false,
		FileMaxSizeInMB:  5,
		FileMaxAgeInDays: 3,
	}
	Init(cfg)

	l, err = L()
	if err != nil {
		t.Fatalf("L() returned error after Init: %v", err)
	}
	if l != globalLogger {
		t.Error("L() should return the same instance as globalLogger")
	}
}

// Test New() creates a fresh logger instance
func TestNew(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	cfg := Config{
		Level:            "warn",
		FilePath:         "service.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  20,
		FileMaxAgeInDays: 14,
	}

	l1, err := New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	if l1 == nil {
		t.Fatal("New() returned nil")
	}

	l2, err := New(cfg)
	if err != nil {
		t.Fatalf("Second New() failed: %v", err)
	}
	if l1 == l2 {
		t.Error("New() should return different logger instances")
	}
}

// Test that Config struct works as expected
func TestConfigStruct(t *testing.T) {
	cfg := Config{
		Level:            "debug",
		FilePath:         "/var/log/app.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 30,
	}

	if cfg.Level != "debug" {
		t.Errorf("Expected Level to be 'debug', got %s", cfg.Level)
	}
	if cfg.FilePath != "/var/log/app.log" {
		t.Errorf("Expected FilePath '/var/log/app.log', got %s", cfg.FilePath)
	}
	if !cfg.UseLocalTime {
		t.Error("Expected UseLocalTime to be true")
	}
	if cfg.FileMaxSizeInMB != 100 {
		t.Errorf("Expected FileMaxSizeInMB 100, got %d", cfg.FileMaxSizeInMB)
	}
	if cfg.FileMaxAgeInDays != 30 {
		t.Errorf("Expected FileMaxAgeInDays 30, got %d", cfg.FileMaxAgeInDays)
	}
}

// Integration test: ensure logs are written to file
func TestLoggerIntegration(t *testing.T) {
	// Reset global state
	globalLogger = nil
	once = sync.Once{}

	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	cfg := Config{
		Level:            "info",
		FilePath:         "integration.log",
		UseLocalTime:     false,
		FileMaxSizeInMB:  1,
		FileMaxAgeInDays: 1,
	}

	if err := Init(cfg); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	l, _ := L()
	l.Info("integration test", "key", "value")

	logPath := filepath.Join(tempDir, "integration.log")
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Fatal("Log file should exist")
	}

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	if len(content) == 0 {
		t.Error("Log file should contain log entries")
	}
}

// Helper: capture stdout (optional)
func captureOutput(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	out, _ := io.ReadAll(r)
	return string(out)
}
