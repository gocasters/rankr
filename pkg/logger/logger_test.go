package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

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
			result := mapLevel(tt.levelStr)
			if result != tt.expected {
				t.Errorf("mapLevel(%q) = %v, want %v", tt.levelStr, result, tt.expected)
			}
		})
	}
}

func TestInit(t *testing.T) {
	// Reset global state for testing
	globalLogger = nil
	once = sync.Once{}

	// Create temporary directory for test logs
	tempDir := t.TempDir()

	// Change to temp directory for test
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	cfg := Config{
		Level:            "debug",
		FilePath:         "test.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}

	Init(cfg)

	if globalLogger == nil {
		t.Error("Init() should initialize globalLogger")
	}

	// Test that calling Init again doesn't reinitialize (sync.Once behavior)
	originalLogger := globalLogger
	Init(cfg)
	if globalLogger != originalLogger {
		t.Error("Init() should only initialize once due to sync.Once")
	}
}

func TestL(t *testing.T) {
	// Reset global state for testing
	globalLogger = nil
	once = sync.Once{}

	// Test L() returns nil when not initialized
	if L() != nil {
		t.Error("L() should return nil when globalLogger is not initialized")
	}

	// Initialize logger
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	cfg := Config{
		Level:            "info",
		FilePath:         "test.log",
		UseLocalTime:     false,
		FileMaxSizeInMB:  5,
		FileMaxAgeInDays: 3,
	}

	Init(cfg)
	logger := L()

	if logger == nil {
		t.Error("L() should return initialized logger")
	}

	if logger != globalLogger {
		t.Error("L() should return the same instance as globalLogger")
	}
}

func TestNew(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	cfg := Config{
		Level:            "warn",
		FilePath:         "service.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  20,
		FileMaxAgeInDays: 14,
	}

	logger := New(cfg)

	if logger == nil {
		t.Error("New() should return a logger instance")
	}

	// Test that each call to New() returns a different instance
	logger2 := New(cfg)
	if logger == logger2 {
		t.Error("New() should return different instances on each call")
	}
}

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
		t.Errorf("Expected FilePath to be '/var/log/app.log', got %s", cfg.FilePath)
	}
	if !cfg.UseLocalTime {
		t.Error("Expected UseLocalTime to be true")
	}
	if cfg.FileMaxSizeInMB != 100 {
		t.Errorf("Expected FileMaxSizeInMB to be 100, got %d", cfg.FileMaxSizeInMB)
	}
	if cfg.FileMaxAgeInDays != 30 {
		t.Errorf("Expected FileMaxAgeInDays to be 30, got %d", cfg.FileMaxAgeInDays)
	}
}

func TestLoggerIntegration(t *testing.T) {
	// Reset global state
	globalLogger = nil
	once = sync.Once{}

	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	cfg := Config{
		Level:            "info",
		FilePath:         "integration.log",
		UseLocalTime:     false,
		FileMaxSizeInMB:  1,
		FileMaxAgeInDays: 1,
	}

	Init(cfg)
	logger := L()

	// Test that we can actually log something
	logger.Info("test message", "key", "value")

	// Check that log file was created
	logPath := filepath.Join(tempDir, "integration.log")
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("Log file should be created after logging")
	}

	// Check file content
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if len(content) == 0 {
		t.Error("Log file should contain log entries")
	}
}

// Cleanup helper to reset global state
func resetGlobalState() {
	globalLogger = nil
	once = sync.Once{}
}

// Test helper to capture output
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