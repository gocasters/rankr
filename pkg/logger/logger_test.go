package logger

import (
	"fmt"
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
		{"case insensitive - DEBUG", "DEBUG", slog.LevelDebug},
		{"case insensitive - Info", "Info", slog.LevelInfo},
		{"case insensitive - WARN", "WARN", slog.LevelWarn},
		{"case insensitive - ERROR", "ERROR", slog.LevelError},
		{"mixed case - dEbUg", "dEbUg", slog.LevelDebug},
		{"whitespace padded - 

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

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerErrorHandling(t *testing.T) {
	t.Run("nil logger handling", func(t *testing.T) {
		resetGlobalState()
		
		// L() should return nil when not initialized
		logger := L()
		if logger != nil {
			t.Error("L() should return nil when not initialized")
		}
	})
	
	t.Run("multiple resets and initializations", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)
		
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		
		for i := 0; i < 5; i++ {
			resetGlobalState()
			
			cfg := Config{
				Level:            "info",
				FilePath:         fmt.Sprintf("reset_%d.log", i),
				UseLocalTime:     true,
				FileMaxSizeInMB:  10,
				FileMaxAgeInDays: 7,
			}
			
			Init(cfg)
			logger := L()
			
			if logger == nil {
				t.Errorf("Logger should be initialized on iteration %d", i)
			}
			
			logger.Info("test message", "iteration", i)
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestStructuredLogging(t *testing.T) {
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
		Level:            "debug",
		FilePath:         "structured.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	// Test various structured logging patterns
	logger.Info("user action", 
		"user_id", 12345, 
		"action", "login", 
		"ip_address", "192.168.1.1", 
		"timestamp", "2024-01-01T00:00:00Z")
	
	logger.Warn("performance warning", 
		"operation", "database_query", 
		"duration_ms", 5000, 
		"table", "users", 
		"slow_query", true)
	
	logger.Error("system error", 
		"error_code", 500, 
		"component", "auth_service", 
		"details", map[string]interface{}{
			"nested": "value",
			"count":  42,
		})
	
	// Test with various data types
	logger.Debug("data types test", 
		"string", "value", 
		"int", 123, 
		"float", 45.67, 
		"bool", true, 
		"nil", nil)
	
	// Check that log file contains structured data
	logPath := filepath.Join(tempDir, "structured.log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	if len(content) == 0 {
		t.Error("Log file should contain structured log entries")
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerErrorHandling(t *testing.T) {
	t.Run("nil logger handling", func(t *testing.T) {
		resetGlobalState()
		
		// L() should return nil when not initialized
		logger := L()
		if logger != nil {
			t.Error("L() should return nil when not initialized")
		}
	})
	
	t.Run("multiple resets and initializations", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)
		
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		
		for i := 0; i < 5; i++ {
			resetGlobalState()
			
			cfg := Config{
				Level:            "info",
				FilePath:         fmt.Sprintf("reset_%d.log", i),
				UseLocalTime:     true,
				FileMaxSizeInMB:  10,
				FileMaxAgeInDays: 7,
			}
			
			Init(cfg)
			logger := L()
			
			if logger == nil {
				t.Errorf("Logger should be initialized on iteration %d", i)
			}
			
			logger.Info("test message", "iteration", i)
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerTimeFormats(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	t.Run("local time format", func(t *testing.T) {
		cfg := Config{
			Level:            "info",
			FilePath:         "local_time.log",
			UseLocalTime:     true,
			FileMaxSizeInMB:  10,
			FileMaxAgeInDays: 7,
		}
		
		logger := New(cfg)
		logger.Info("message with local time")
		
		// Check that log file was created
		logPath := filepath.Join(tempDir, "local_time.log")
		content, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}
		
		if len(content) == 0 {
			t.Error("Log file should contain entries with local time")
		}
	})
	
	t.Run("UTC time format", func(t *testing.T) {
		cfg := Config{
			Level:            "info",
			FilePath:         "utc_time.log",
			UseLocalTime:     false,
			FileMaxSizeInMB:  10,
			FileMaxAgeInDays: 7,
		}
		
		logger := New(cfg)
		logger.Info("message with UTC time")
		
		// Check that log file was created
		logPath := filepath.Join(tempDir, "utc_time.log")
		content, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}
		
		if len(content) == 0 {
			t.Error("Log file should contain entries with UTC time")
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerErrorHandling(t *testing.T) {
	t.Run("nil logger handling", func(t *testing.T) {
		resetGlobalState()
		
		// L() should return nil when not initialized
		logger := L()
		if logger != nil {
			t.Error("L() should return nil when not initialized")
		}
	})
	
	t.Run("multiple resets and initializations", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)
		
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		
		for i := 0; i < 5; i++ {
			resetGlobalState()
			
			cfg := Config{
				Level:            "info",
				FilePath:         fmt.Sprintf("reset_%d.log", i),
				UseLocalTime:     true,
				FileMaxSizeInMB:  10,
				FileMaxAgeInDays: 7,
			}
			
			Init(cfg)
			logger := L()
			
			if logger == nil {
				t.Errorf("Logger should be initialized on iteration %d", i)
			}
			
			logger.Info("test message", "iteration", i)
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestStructuredLogging(t *testing.T) {
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
		Level:            "debug",
		FilePath:         "structured.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	// Test various structured logging patterns
	logger.Info("user action", 
		"user_id", 12345, 
		"action", "login", 
		"ip_address", "192.168.1.1", 
		"timestamp", "2024-01-01T00:00:00Z")
	
	logger.Warn("performance warning", 
		"operation", "database_query", 
		"duration_ms", 5000, 
		"table", "users", 
		"slow_query", true)
	
	logger.Error("system error", 
		"error_code", 500, 
		"component", "auth_service", 
		"details", map[string]interface{}{
			"nested": "value",
			"count":  42,
		})
	
	// Test with various data types
	logger.Debug("data types test", 
		"string", "value", 
		"int", 123, 
		"float", 45.67, 
		"bool", true, 
		"nil", nil)
	
	// Check that log file contains structured data
	logPath := filepath.Join(tempDir, "structured.log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	if len(content) == 0 {
		t.Error("Log file should contain structured log entries")
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerErrorHandling(t *testing.T) {
	t.Run("nil logger handling", func(t *testing.T) {
		resetGlobalState()
		
		// L() should return nil when not initialized
		logger := L()
		if logger != nil {
			t.Error("L() should return nil when not initialized")
		}
	})
	
	t.Run("multiple resets and initializations", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)
		
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		
		for i := 0; i < 5; i++ {
			resetGlobalState()
			
			cfg := Config{
				Level:            "info",
				FilePath:         fmt.Sprintf("reset_%d.log", i),
				UseLocalTime:     true,
				FileMaxSizeInMB:  10,
				FileMaxAgeInDays: 7,
			}
			
			Init(cfg)
			logger := L()
			
			if logger == nil {
				t.Errorf("Logger should be initialized on iteration %d", i)
			}
			
			logger.Info("test message", "iteration", i)
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerFileHandling(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	t.Run("large log file creation", func(t *testing.T) {
		cfg := Config{
			Level:            "debug",
			FilePath:         "large.log",
			UseLocalTime:     true,
			FileMaxSizeInMB:  1,
			FileMaxAgeInDays: 1,
		}
		
		logger := New(cfg)
		
		// Write many log entries to potentially trigger rotation
		for i := 0; i < 1000; i++ {
			logger.Info("Large log entry with lots of data to fill up the file quickly", 
				"iteration", i, 
				"data", string(make([]byte, 100)))
		}
		
		// Check that log file exists
		logPath := filepath.Join(tempDir, "large.log")
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			t.Error("Large log file should be created")
		}
	})
	
	t.Run("nested directory creation", func(t *testing.T) {
		cfg := Config{
			Level:            "info",
			FilePath:         "nested/deep/path/test.log",
			UseLocalTime:     false,
			FileMaxSizeInMB:  10,
			FileMaxAgeInDays: 7,
		}
		
		logger := New(cfg)
		logger.Info("test message in nested directory")
		
		// Check that nested directories and log file were created
		logPath := filepath.Join(tempDir, "nested/deep/path/test.log")
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			t.Error("Nested log file should be created along with directories")
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerErrorHandling(t *testing.T) {
	t.Run("nil logger handling", func(t *testing.T) {
		resetGlobalState()
		
		// L() should return nil when not initialized
		logger := L()
		if logger != nil {
			t.Error("L() should return nil when not initialized")
		}
	})
	
	t.Run("multiple resets and initializations", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)
		
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		
		for i := 0; i < 5; i++ {
			resetGlobalState()
			
			cfg := Config{
				Level:            "info",
				FilePath:         fmt.Sprintf("reset_%d.log", i),
				UseLocalTime:     true,
				FileMaxSizeInMB:  10,
				FileMaxAgeInDays: 7,
			}
			
			Init(cfg)
			logger := L()
			
			if logger == nil {
				t.Errorf("Logger should be initialized on iteration %d", i)
			}
			
			logger.Info("test message", "iteration", i)
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestStructuredLogging(t *testing.T) {
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
		Level:            "debug",
		FilePath:         "structured.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	// Test various structured logging patterns
	logger.Info("user action", 
		"user_id", 12345, 
		"action", "login", 
		"ip_address", "192.168.1.1", 
		"timestamp", "2024-01-01T00:00:00Z")
	
	logger.Warn("performance warning", 
		"operation", "database_query", 
		"duration_ms", 5000, 
		"table", "users", 
		"slow_query", true)
	
	logger.Error("system error", 
		"error_code", 500, 
		"component", "auth_service", 
		"details", map[string]interface{}{
			"nested": "value",
			"count":  42,
		})
	
	// Test with various data types
	logger.Debug("data types test", 
		"string", "value", 
		"int", 123, 
		"float", 45.67, 
		"bool", true, 
		"nil", nil)
	
	// Check that log file contains structured data
	logPath := filepath.Join(tempDir, "structured.log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	if len(content) == 0 {
		t.Error("Log file should contain structured log entries")
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerErrorHandling(t *testing.T) {
	t.Run("nil logger handling", func(t *testing.T) {
		resetGlobalState()
		
		// L() should return nil when not initialized
		logger := L()
		if logger != nil {
			t.Error("L() should return nil when not initialized")
		}
	})
	
	t.Run("multiple resets and initializations", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)
		
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		
		for i := 0; i < 5; i++ {
			resetGlobalState()
			
			cfg := Config{
				Level:            "info",
				FilePath:         fmt.Sprintf("reset_%d.log", i),
				UseLocalTime:     true,
				FileMaxSizeInMB:  10,
				FileMaxAgeInDays: 7,
			}
			
			Init(cfg)
			logger := L()
			
			if logger == nil {
				t.Errorf("Logger should be initialized on iteration %d", i)
			}
			
			logger.Info("test message", "iteration", i)
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerTimeFormats(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	t.Run("local time format", func(t *testing.T) {
		cfg := Config{
			Level:            "info",
			FilePath:         "local_time.log",
			UseLocalTime:     true,
			FileMaxSizeInMB:  10,
			FileMaxAgeInDays: 7,
		}
		
		logger := New(cfg)
		logger.Info("message with local time")
		
		// Check that log file was created
		logPath := filepath.Join(tempDir, "local_time.log")
		content, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}
		
		if len(content) == 0 {
			t.Error("Log file should contain entries with local time")
		}
	})
	
	t.Run("UTC time format", func(t *testing.T) {
		cfg := Config{
			Level:            "info",
			FilePath:         "utc_time.log",
			UseLocalTime:     false,
			FileMaxSizeInMB:  10,
			FileMaxAgeInDays: 7,
		}
		
		logger := New(cfg)
		logger.Info("message with UTC time")
		
		// Check that log file was created
		logPath := filepath.Join(tempDir, "utc_time.log")
		content, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}
		
		if len(content) == 0 {
			t.Error("Log file should contain entries with UTC time")
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerErrorHandling(t *testing.T) {
	t.Run("nil logger handling", func(t *testing.T) {
		resetGlobalState()
		
		// L() should return nil when not initialized
		logger := L()
		if logger != nil {
			t.Error("L() should return nil when not initialized")
		}
	})
	
	t.Run("multiple resets and initializations", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)
		
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		
		for i := 0; i < 5; i++ {
			resetGlobalState()
			
			cfg := Config{
				Level:            "info",
				FilePath:         fmt.Sprintf("reset_%d.log", i),
				UseLocalTime:     true,
				FileMaxSizeInMB:  10,
				FileMaxAgeInDays: 7,
			}
			
			Init(cfg)
			logger := L()
			
			if logger == nil {
				t.Errorf("Logger should be initialized on iteration %d", i)
			}
			
			logger.Info("test message", "iteration", i)
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestStructuredLogging(t *testing.T) {
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
		Level:            "debug",
		FilePath:         "structured.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	// Test various structured logging patterns
	logger.Info("user action", 
		"user_id", 12345, 
		"action", "login", 
		"ip_address", "192.168.1.1", 
		"timestamp", "2024-01-01T00:00:00Z")
	
	logger.Warn("performance warning", 
		"operation", "database_query", 
		"duration_ms", 5000, 
		"table", "users", 
		"slow_query", true)
	
	logger.Error("system error", 
		"error_code", 500, 
		"component", "auth_service", 
		"details", map[string]interface{}{
			"nested": "value",
			"count":  42,
		})
	
	// Test with various data types
	logger.Debug("data types test", 
		"string", "value", 
		"int", 123, 
		"float", 45.67, 
		"bool", true, 
		"nil", nil)
	
	// Check that log file contains structured data
	logPath := filepath.Join(tempDir, "structured.log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	if len(content) == 0 {
		t.Error("Log file should contain structured log entries")
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerErrorHandling(t *testing.T) {
	t.Run("nil logger handling", func(t *testing.T) {
		resetGlobalState()
		
		// L() should return nil when not initialized
		logger := L()
		if logger != nil {
			t.Error("L() should return nil when not initialized")
		}
	})
	
	t.Run("multiple resets and initializations", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)
		
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		
		for i := 0; i < 5; i++ {
			resetGlobalState()
			
			cfg := Config{
				Level:            "info",
				FilePath:         fmt.Sprintf("reset_%d.log", i),
				UseLocalTime:     true,
				FileMaxSizeInMB:  10,
				FileMaxAgeInDays: 7,
			}
			
			Init(cfg)
			logger := L()
			
			if logger == nil {
				t.Errorf("Logger should be initialized on iteration %d", i)
			}
			
			logger.Info("test message", "iteration", i)
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerLevels(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	levels := []string{"debug", "info", "warn", "error"}
	
	for _, level := range levels {
		t.Run("level_"+level, func(t *testing.T) {
			cfg := Config{
				Level:            level,
				FilePath:         level + ".log",
				UseLocalTime:     true,
				FileMaxSizeInMB:  10,
				FileMaxAgeInDays: 7,
			}
			
			logger := New(cfg)
			
			// Test all logging methods
			logger.Debug("debug message", "level", level)
			logger.Info("info message", "level", level)
			logger.Warn("warn message", "level", level)
			logger.Error("error message", "level", level)
			
			// Check that log file was created
			logPath := filepath.Join(tempDir, level+".log")
			if _, err := os.Stat(logPath); os.IsNotExist(err) {
				t.Errorf("Log file should be created for level %s", level)
			}
		})
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerErrorHandling(t *testing.T) {
	t.Run("nil logger handling", func(t *testing.T) {
		resetGlobalState()
		
		// L() should return nil when not initialized
		logger := L()
		if logger != nil {
			t.Error("L() should return nil when not initialized")
		}
	})
	
	t.Run("multiple resets and initializations", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)
		
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		
		for i := 0; i < 5; i++ {
			resetGlobalState()
			
			cfg := Config{
				Level:            "info",
				FilePath:         fmt.Sprintf("reset_%d.log", i),
				UseLocalTime:     true,
				FileMaxSizeInMB:  10,
				FileMaxAgeInDays: 7,
			}
			
			Init(cfg)
			logger := L()
			
			if logger == nil {
				t.Errorf("Logger should be initialized on iteration %d", i)
			}
			
			logger.Info("test message", "iteration", i)
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestStructuredLogging(t *testing.T) {
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
		Level:            "debug",
		FilePath:         "structured.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	// Test various structured logging patterns
	logger.Info("user action", 
		"user_id", 12345, 
		"action", "login", 
		"ip_address", "192.168.1.1", 
		"timestamp", "2024-01-01T00:00:00Z")
	
	logger.Warn("performance warning", 
		"operation", "database_query", 
		"duration_ms", 5000, 
		"table", "users", 
		"slow_query", true)
	
	logger.Error("system error", 
		"error_code", 500, 
		"component", "auth_service", 
		"details", map[string]interface{}{
			"nested": "value",
			"count":  42,
		})
	
	// Test with various data types
	logger.Debug("data types test", 
		"string", "value", 
		"int", 123, 
		"float", 45.67, 
		"bool", true, 
		"nil", nil)
	
	// Check that log file contains structured data
	logPath := filepath.Join(tempDir, "structured.log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	if len(content) == 0 {
		t.Error("Log file should contain structured log entries")
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerErrorHandling(t *testing.T) {
	t.Run("nil logger handling", func(t *testing.T) {
		resetGlobalState()
		
		// L() should return nil when not initialized
		logger := L()
		if logger != nil {
			t.Error("L() should return nil when not initialized")
		}
	})
	
	t.Run("multiple resets and initializations", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)
		
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		
		for i := 0; i < 5; i++ {
			resetGlobalState()
			
			cfg := Config{
				Level:            "info",
				FilePath:         fmt.Sprintf("reset_%d.log", i),
				UseLocalTime:     true,
				FileMaxSizeInMB:  10,
				FileMaxAgeInDays: 7,
			}
			
			Init(cfg)
			logger := L()
			
			if logger == nil {
				t.Errorf("Logger should be initialized on iteration %d", i)
			}
			
			logger.Info("test message", "iteration", i)
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerTimeFormats(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	t.Run("local time format", func(t *testing.T) {
		cfg := Config{
			Level:            "info",
			FilePath:         "local_time.log",
			UseLocalTime:     true,
			FileMaxSizeInMB:  10,
			FileMaxAgeInDays: 7,
		}
		
		logger := New(cfg)
		logger.Info("message with local time")
		
		// Check that log file was created
		logPath := filepath.Join(tempDir, "local_time.log")
		content, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}
		
		if len(content) == 0 {
			t.Error("Log file should contain entries with local time")
		}
	})
	
	t.Run("UTC time format", func(t *testing.T) {
		cfg := Config{
			Level:            "info",
			FilePath:         "utc_time.log",
			UseLocalTime:     false,
			FileMaxSizeInMB:  10,
			FileMaxAgeInDays: 7,
		}
		
		logger := New(cfg)
		logger.Info("message with UTC time")
		
		// Check that log file was created
		logPath := filepath.Join(tempDir, "utc_time.log")
		content, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}
		
		if len(content) == 0 {
			t.Error("Log file should contain entries with UTC time")
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerErrorHandling(t *testing.T) {
	t.Run("nil logger handling", func(t *testing.T) {
		resetGlobalState()
		
		// L() should return nil when not initialized
		logger := L()
		if logger != nil {
			t.Error("L() should return nil when not initialized")
		}
	})
	
	t.Run("multiple resets and initializations", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)
		
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		
		for i := 0; i < 5; i++ {
			resetGlobalState()
			
			cfg := Config{
				Level:            "info",
				FilePath:         fmt.Sprintf("reset_%d.log", i),
				UseLocalTime:     true,
				FileMaxSizeInMB:  10,
				FileMaxAgeInDays: 7,
			}
			
			Init(cfg)
			logger := L()
			
			if logger == nil {
				t.Errorf("Logger should be initialized on iteration %d", i)
			}
			
			logger.Info("test message", "iteration", i)
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestStructuredLogging(t *testing.T) {
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
		Level:            "debug",
		FilePath:         "structured.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	// Test various structured logging patterns
	logger.Info("user action", 
		"user_id", 12345, 
		"action", "login", 
		"ip_address", "192.168.1.1", 
		"timestamp", "2024-01-01T00:00:00Z")
	
	logger.Warn("performance warning", 
		"operation", "database_query", 
		"duration_ms", 5000, 
		"table", "users", 
		"slow_query", true)
	
	logger.Error("system error", 
		"error_code", 500, 
		"component", "auth_service", 
		"details", map[string]interface{}{
			"nested": "value",
			"count":  42,
		})
	
	// Test with various data types
	logger.Debug("data types test", 
		"string", "value", 
		"int", 123, 
		"float", 45.67, 
		"bool", true, 
		"nil", nil)
	
	// Check that log file contains structured data
	logPath := filepath.Join(tempDir, "structured.log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	if len(content) == 0 {
		t.Error("Log file should contain structured log entries")
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerErrorHandling(t *testing.T) {
	t.Run("nil logger handling", func(t *testing.T) {
		resetGlobalState()
		
		// L() should return nil when not initialized
		logger := L()
		if logger != nil {
			t.Error("L() should return nil when not initialized")
		}
	})
	
	t.Run("multiple resets and initializations", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)
		
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		
		for i := 0; i < 5; i++ {
			resetGlobalState()
			
			cfg := Config{
				Level:            "info",
				FilePath:         fmt.Sprintf("reset_%d.log", i),
				UseLocalTime:     true,
				FileMaxSizeInMB:  10,
				FileMaxAgeInDays: 7,
			}
			
			Init(cfg)
			logger := L()
			
			if logger == nil {
				t.Errorf("Logger should be initialized on iteration %d", i)
			}
			
			logger.Info("test message", "iteration", i)
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerFileHandling(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	t.Run("large log file creation", func(t *testing.T) {
		cfg := Config{
			Level:            "debug",
			FilePath:         "large.log",
			UseLocalTime:     true,
			FileMaxSizeInMB:  1,
			FileMaxAgeInDays: 1,
		}
		
		logger := New(cfg)
		
		// Write many log entries to potentially trigger rotation
		for i := 0; i < 1000; i++ {
			logger.Info("Large log entry with lots of data to fill up the file quickly", 
				"iteration", i, 
				"data", string(make([]byte, 100)))
		}
		
		// Check that log file exists
		logPath := filepath.Join(tempDir, "large.log")
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			t.Error("Large log file should be created")
		}
	})
	
	t.Run("nested directory creation", func(t *testing.T) {
		cfg := Config{
			Level:            "info",
			FilePath:         "nested/deep/path/test.log",
			UseLocalTime:     false,
			FileMaxSizeInMB:  10,
			FileMaxAgeInDays: 7,
		}
		
		logger := New(cfg)
		logger.Info("test message in nested directory")
		
		// Check that nested directories and log file were created
		logPath := filepath.Join(tempDir, "nested/deep/path/test.log")
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			t.Error("Nested log file should be created along with directories")
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerErrorHandling(t *testing.T) {
	t.Run("nil logger handling", func(t *testing.T) {
		resetGlobalState()
		
		// L() should return nil when not initialized
		logger := L()
		if logger != nil {
			t.Error("L() should return nil when not initialized")
		}
	})
	
	t.Run("multiple resets and initializations", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)
		
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		
		for i := 0; i < 5; i++ {
			resetGlobalState()
			
			cfg := Config{
				Level:            "info",
				FilePath:         fmt.Sprintf("reset_%d.log", i),
				UseLocalTime:     true,
				FileMaxSizeInMB:  10,
				FileMaxAgeInDays: 7,
			}
			
			Init(cfg)
			logger := L()
			
			if logger == nil {
				t.Errorf("Logger should be initialized on iteration %d", i)
			}
			
			logger.Info("test message", "iteration", i)
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestStructuredLogging(t *testing.T) {
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
		Level:            "debug",
		FilePath:         "structured.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	// Test various structured logging patterns
	logger.Info("user action", 
		"user_id", 12345, 
		"action", "login", 
		"ip_address", "192.168.1.1", 
		"timestamp", "2024-01-01T00:00:00Z")
	
	logger.Warn("performance warning", 
		"operation", "database_query", 
		"duration_ms", 5000, 
		"table", "users", 
		"slow_query", true)
	
	logger.Error("system error", 
		"error_code", 500, 
		"component", "auth_service", 
		"details", map[string]interface{}{
			"nested": "value",
			"count":  42,
		})
	
	// Test with various data types
	logger.Debug("data types test", 
		"string", "value", 
		"int", 123, 
		"float", 45.67, 
		"bool", true, 
		"nil", nil)
	
	// Check that log file contains structured data
	logPath := filepath.Join(tempDir, "structured.log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	if len(content) == 0 {
		t.Error("Log file should contain structured log entries")
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerErrorHandling(t *testing.T) {
	t.Run("nil logger handling", func(t *testing.T) {
		resetGlobalState()
		
		// L() should return nil when not initialized
		logger := L()
		if logger != nil {
			t.Error("L() should return nil when not initialized")
		}
	})
	
	t.Run("multiple resets and initializations", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)
		
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		
		for i := 0; i < 5; i++ {
			resetGlobalState()
			
			cfg := Config{
				Level:            "info",
				FilePath:         fmt.Sprintf("reset_%d.log", i),
				UseLocalTime:     true,
				FileMaxSizeInMB:  10,
				FileMaxAgeInDays: 7,
			}
			
			Init(cfg)
			logger := L()
			
			if logger == nil {
				t.Errorf("Logger should be initialized on iteration %d", i)
			}
			
			logger.Info("test message", "iteration", i)
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerTimeFormats(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	t.Run("local time format", func(t *testing.T) {
		cfg := Config{
			Level:            "info",
			FilePath:         "local_time.log",
			UseLocalTime:     true,
			FileMaxSizeInMB:  10,
			FileMaxAgeInDays: 7,
		}
		
		logger := New(cfg)
		logger.Info("message with local time")
		
		// Check that log file was created
		logPath := filepath.Join(tempDir, "local_time.log")
		content, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}
		
		if len(content) == 0 {
			t.Error("Log file should contain entries with local time")
		}
	})
	
	t.Run("UTC time format", func(t *testing.T) {
		cfg := Config{
			Level:            "info",
			FilePath:         "utc_time.log",
			UseLocalTime:     false,
			FileMaxSizeInMB:  10,
			FileMaxAgeInDays: 7,
		}
		
		logger := New(cfg)
		logger.Info("message with UTC time")
		
		// Check that log file was created
		logPath := filepath.Join(tempDir, "utc_time.log")
		content, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}
		
		if len(content) == 0 {
			t.Error("Log file should contain entries with UTC time")
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerErrorHandling(t *testing.T) {
	t.Run("nil logger handling", func(t *testing.T) {
		resetGlobalState()
		
		// L() should return nil when not initialized
		logger := L()
		if logger != nil {
			t.Error("L() should return nil when not initialized")
		}
	})
	
	t.Run("multiple resets and initializations", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)
		
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		
		for i := 0; i < 5; i++ {
			resetGlobalState()
			
			cfg := Config{
				Level:            "info",
				FilePath:         fmt.Sprintf("reset_%d.log", i),
				UseLocalTime:     true,
				FileMaxSizeInMB:  10,
				FileMaxAgeInDays: 7,
			}
			
			Init(cfg)
			logger := L()
			
			if logger == nil {
				t.Errorf("Logger should be initialized on iteration %d", i)
			}
			
			logger.Info("test message", "iteration", i)
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestStructuredLogging(t *testing.T) {
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
		Level:            "debug",
		FilePath:         "structured.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	// Test various structured logging patterns
	logger.Info("user action", 
		"user_id", 12345, 
		"action", "login", 
		"ip_address", "192.168.1.1", 
		"timestamp", "2024-01-01T00:00:00Z")
	
	logger.Warn("performance warning", 
		"operation", "database_query", 
		"duration_ms", 5000, 
		"table", "users", 
		"slow_query", true)
	
	logger.Error("system error", 
		"error_code", 500, 
		"component", "auth_service", 
		"details", map[string]interface{}{
			"nested": "value",
			"count":  42,
		})
	
	// Test with various data types
	logger.Debug("data types test", 
		"string", "value", 
		"int", 123, 
		"float", 45.67, 
		"bool", true, 
		"nil", nil)
	
	// Check that log file contains structured data
	logPath := filepath.Join(tempDir, "structured.log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	if len(content) == 0 {
		t.Error("Log file should contain structured log entries")
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerErrorHandling(t *testing.T) {
	t.Run("nil logger handling", func(t *testing.T) {
		resetGlobalState()
		
		// L() should return nil when not initialized
		logger := L()
		if logger != nil {
			t.Error("L() should return nil when not initialized")
		}
	})
	
	t.Run("multiple resets and initializations", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)
		
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		
		for i := 0; i < 5; i++ {
			resetGlobalState()
			
			cfg := Config{
				Level:            "info",
				FilePath:         fmt.Sprintf("reset_%d.log", i),
				UseLocalTime:     true,
				FileMaxSizeInMB:  10,
				FileMaxAgeInDays: 7,
			}
			
			Init(cfg)
			logger := L()
			
			if logger == nil {
				t.Errorf("Logger should be initialized on iteration %d", i)
			}
			
			logger.Info("test message", "iteration", i)
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestNewWithDifferentConfigs(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	configs := []Config{
		{
			Level:            "debug",
			FilePath:         "debug.log",
			UseLocalTime:     true,
			FileMaxSizeInMB:  1,
			FileMaxAgeInDays: 1,
		},
		{
			Level:            "info",
			FilePath:         "info.log",
			UseLocalTime:     false,
			FileMaxSizeInMB:  50,
			FileMaxAgeInDays: 10,
		},
		{
			Level:            "warn",
			FilePath:         "warn.log",
			UseLocalTime:     true,
			FileMaxSizeInMB:  100,
			FileMaxAgeInDays: 30,
		},
		{
			Level:            "error",
			FilePath:         "error.log",
			UseLocalTime:     false,
			FileMaxSizeInMB:  200,
			FileMaxAgeInDays: 60,
		},
	}
	
	loggers := make([]*slog.Logger, len(configs))
	
	for i, cfg := range configs {
		logger := New(cfg)
		if logger == nil {
			t.Errorf("New() with config %d should return a logger instance", i)
		}
		loggers[i] = logger
	}
	
	// All loggers should be different instances
	for i := 0; i < len(loggers); i++ {
		for j := i + 1; j < len(loggers); j++ {
			if loggers[i] == loggers[j] {
				t.Errorf("New() should return different instances for different configs")
			}
		}
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerErrorHandling(t *testing.T) {
	t.Run("nil logger handling", func(t *testing.T) {
		resetGlobalState()
		
		// L() should return nil when not initialized
		logger := L()
		if logger != nil {
			t.Error("L() should return nil when not initialized")
		}
	})
	
	t.Run("multiple resets and initializations", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)
		
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		
		for i := 0; i < 5; i++ {
			resetGlobalState()
			
			cfg := Config{
				Level:            "info",
				FilePath:         fmt.Sprintf("reset_%d.log", i),
				UseLocalTime:     true,
				FileMaxSizeInMB:  10,
				FileMaxAgeInDays: 7,
			}
			
			Init(cfg)
			logger := L()
			
			if logger == nil {
				t.Errorf("Logger should be initialized on iteration %d", i)
			}
			
			logger.Info("test message", "iteration", i)
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestStructuredLogging(t *testing.T) {
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
		Level:            "debug",
		FilePath:         "structured.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	// Test various structured logging patterns
	logger.Info("user action", 
		"user_id", 12345, 
		"action", "login", 
		"ip_address", "192.168.1.1", 
		"timestamp", "2024-01-01T00:00:00Z")
	
	logger.Warn("performance warning", 
		"operation", "database_query", 
		"duration_ms", 5000, 
		"table", "users", 
		"slow_query", true)
	
	logger.Error("system error", 
		"error_code", 500, 
		"component", "auth_service", 
		"details", map[string]interface{}{
			"nested": "value",
			"count":  42,
		})
	
	// Test with various data types
	logger.Debug("data types test", 
		"string", "value", 
		"int", 123, 
		"float", 45.67, 
		"bool", true, 
		"nil", nil)
	
	// Check that log file contains structured data
	logPath := filepath.Join(tempDir, "structured.log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	if len(content) == 0 {
		t.Error("Log file should contain structured log entries")
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerErrorHandling(t *testing.T) {
	t.Run("nil logger handling", func(t *testing.T) {
		resetGlobalState()
		
		// L() should return nil when not initialized
		logger := L()
		if logger != nil {
			t.Error("L() should return nil when not initialized")
		}
	})
	
	t.Run("multiple resets and initializations", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)
		
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		
		for i := 0; i < 5; i++ {
			resetGlobalState()
			
			cfg := Config{
				Level:            "info",
				FilePath:         fmt.Sprintf("reset_%d.log", i),
				UseLocalTime:     true,
				FileMaxSizeInMB:  10,
				FileMaxAgeInDays: 7,
			}
			
			Init(cfg)
			logger := L()
			
			if logger == nil {
				t.Errorf("Logger should be initialized on iteration %d", i)
			}
			
			logger.Info("test message", "iteration", i)
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerTimeFormats(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	t.Run("local time format", func(t *testing.T) {
		cfg := Config{
			Level:            "info",
			FilePath:         "local_time.log",
			UseLocalTime:     true,
			FileMaxSizeInMB:  10,
			FileMaxAgeInDays: 7,
		}
		
		logger := New(cfg)
		logger.Info("message with local time")
		
		// Check that log file was created
		logPath := filepath.Join(tempDir, "local_time.log")
		content, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}
		
		if len(content) == 0 {
			t.Error("Log file should contain entries with local time")
		}
	})
	
	t.Run("UTC time format", func(t *testing.T) {
		cfg := Config{
			Level:            "info",
			FilePath:         "utc_time.log",
			UseLocalTime:     false,
			FileMaxSizeInMB:  10,
			FileMaxAgeInDays: 7,
		}
		
		logger := New(cfg)
		logger.Info("message with UTC time")
		
		// Check that log file was created
		logPath := filepath.Join(tempDir, "utc_time.log")
		content, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}
		
		if len(content) == 0 {
			t.Error("Log file should contain entries with UTC time")
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerErrorHandling(t *testing.T) {
	t.Run("nil logger handling", func(t *testing.T) {
		resetGlobalState()
		
		// L() should return nil when not initialized
		logger := L()
		if logger != nil {
			t.Error("L() should return nil when not initialized")
		}
	})
	
	t.Run("multiple resets and initializations", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)
		
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		
		for i := 0; i < 5; i++ {
			resetGlobalState()
			
			cfg := Config{
				Level:            "info",
				FilePath:         fmt.Sprintf("reset_%d.log", i),
				UseLocalTime:     true,
				FileMaxSizeInMB:  10,
				FileMaxAgeInDays: 7,
			}
			
			Init(cfg)
			logger := L()
			
			if logger == nil {
				t.Errorf("Logger should be initialized on iteration %d", i)
			}
			
			logger.Info("test message", "iteration", i)
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestStructuredLogging(t *testing.T) {
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
		Level:            "debug",
		FilePath:         "structured.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	// Test various structured logging patterns
	logger.Info("user action", 
		"user_id", 12345, 
		"action", "login", 
		"ip_address", "192.168.1.1", 
		"timestamp", "2024-01-01T00:00:00Z")
	
	logger.Warn("performance warning", 
		"operation", "database_query", 
		"duration_ms", 5000, 
		"table", "users", 
		"slow_query", true)
	
	logger.Error("system error", 
		"error_code", 500, 
		"component", "auth_service", 
		"details", map[string]interface{}{
			"nested": "value",
			"count":  42,
		})
	
	// Test with various data types
	logger.Debug("data types test", 
		"string", "value", 
		"int", 123, 
		"float", 45.67, 
		"bool", true, 
		"nil", nil)
	
	// Check that log file contains structured data
	logPath := filepath.Join(tempDir, "structured.log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	if len(content) == 0 {
		t.Error("Log file should contain structured log entries")
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerErrorHandling(t *testing.T) {
	t.Run("nil logger handling", func(t *testing.T) {
		resetGlobalState()
		
		// L() should return nil when not initialized
		logger := L()
		if logger != nil {
			t.Error("L() should return nil when not initialized")
		}
	})
	
	t.Run("multiple resets and initializations", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)
		
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		
		for i := 0; i < 5; i++ {
			resetGlobalState()
			
			cfg := Config{
				Level:            "info",
				FilePath:         fmt.Sprintf("reset_%d.log", i),
				UseLocalTime:     true,
				FileMaxSizeInMB:  10,
				FileMaxAgeInDays: 7,
			}
			
			Init(cfg)
			logger := L()
			
			if logger == nil {
				t.Errorf("Logger should be initialized on iteration %d", i)
			}
			
			logger.Info("test message", "iteration", i)
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerFileHandling(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	t.Run("large log file creation", func(t *testing.T) {
		cfg := Config{
			Level:            "debug",
			FilePath:         "large.log",
			UseLocalTime:     true,
			FileMaxSizeInMB:  1,
			FileMaxAgeInDays: 1,
		}
		
		logger := New(cfg)
		
		// Write many log entries to potentially trigger rotation
		for i := 0; i < 1000; i++ {
			logger.Info("Large log entry with lots of data to fill up the file quickly", 
				"iteration", i, 
				"data", string(make([]byte, 100)))
		}
		
		// Check that log file exists
		logPath := filepath.Join(tempDir, "large.log")
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			t.Error("Large log file should be created")
		}
	})
	
	t.Run("nested directory creation", func(t *testing.T) {
		cfg := Config{
			Level:            "info",
			FilePath:         "nested/deep/path/test.log",
			UseLocalTime:     false,
			FileMaxSizeInMB:  10,
			FileMaxAgeInDays: 7,
		}
		
		logger := New(cfg)
		logger.Info("test message in nested directory")
		
		// Check that nested directories and log file were created
		logPath := filepath.Join(tempDir, "nested/deep/path/test.log")
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			t.Error("Nested log file should be created along with directories")
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerErrorHandling(t *testing.T) {
	t.Run("nil logger handling", func(t *testing.T) {
		resetGlobalState()
		
		// L() should return nil when not initialized
		logger := L()
		if logger != nil {
			t.Error("L() should return nil when not initialized")
		}
	})
	
	t.Run("multiple resets and initializations", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)
		
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		
		for i := 0; i < 5; i++ {
			resetGlobalState()
			
			cfg := Config{
				Level:            "info",
				FilePath:         fmt.Sprintf("reset_%d.log", i),
				UseLocalTime:     true,
				FileMaxSizeInMB:  10,
				FileMaxAgeInDays: 7,
			}
			
			Init(cfg)
			logger := L()
			
			if logger == nil {
				t.Errorf("Logger should be initialized on iteration %d", i)
			}
			
			logger.Info("test message", "iteration", i)
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestStructuredLogging(t *testing.T) {
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
		Level:            "debug",
		FilePath:         "structured.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	// Test various structured logging patterns
	logger.Info("user action", 
		"user_id", 12345, 
		"action", "login", 
		"ip_address", "192.168.1.1", 
		"timestamp", "2024-01-01T00:00:00Z")
	
	logger.Warn("performance warning", 
		"operation", "database_query", 
		"duration_ms", 5000, 
		"table", "users", 
		"slow_query", true)
	
	logger.Error("system error", 
		"error_code", 500, 
		"component", "auth_service", 
		"details", map[string]interface{}{
			"nested": "value",
			"count":  42,
		})
	
	// Test with various data types
	logger.Debug("data types test", 
		"string", "value", 
		"int", 123, 
		"float", 45.67, 
		"bool", true, 
		"nil", nil)
	
	// Check that log file contains structured data
	logPath := filepath.Join(tempDir, "structured.log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	if len(content) == 0 {
		t.Error("Log file should contain structured log entries")
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerErrorHandling(t *testing.T) {
	t.Run("nil logger handling", func(t *testing.T) {
		resetGlobalState()
		
		// L() should return nil when not initialized
		logger := L()
		if logger != nil {
			t.Error("L() should return nil when not initialized")
		}
	})
	
	t.Run("multiple resets and initializations", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)
		
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		
		for i := 0; i < 5; i++ {
			resetGlobalState()
			
			cfg := Config{
				Level:            "info",
				FilePath:         fmt.Sprintf("reset_%d.log", i),
				UseLocalTime:     true,
				FileMaxSizeInMB:  10,
				FileMaxAgeInDays: 7,
			}
			
			Init(cfg)
			logger := L()
			
			if logger == nil {
				t.Errorf("Logger should be initialized on iteration %d", i)
			}
			
			logger.Info("test message", "iteration", i)
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerTimeFormats(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	t.Run("local time format", func(t *testing.T) {
		cfg := Config{
			Level:            "info",
			FilePath:         "local_time.log",
			UseLocalTime:     true,
			FileMaxSizeInMB:  10,
			FileMaxAgeInDays: 7,
		}
		
		logger := New(cfg)
		logger.Info("message with local time")
		
		// Check that log file was created
		logPath := filepath.Join(tempDir, "local_time.log")
		content, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}
		
		if len(content) == 0 {
			t.Error("Log file should contain entries with local time")
		}
	})
	
	t.Run("UTC time format", func(t *testing.T) {
		cfg := Config{
			Level:            "info",
			FilePath:         "utc_time.log",
			UseLocalTime:     false,
			FileMaxSizeInMB:  10,
			FileMaxAgeInDays: 7,
		}
		
		logger := New(cfg)
		logger.Info("message with UTC time")
		
		// Check that log file was created
		logPath := filepath.Join(tempDir, "utc_time.log")
		content, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}
		
		if len(content) == 0 {
			t.Error("Log file should contain entries with UTC time")
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerErrorHandling(t *testing.T) {
	t.Run("nil logger handling", func(t *testing.T) {
		resetGlobalState()
		
		// L() should return nil when not initialized
		logger := L()
		if logger != nil {
			t.Error("L() should return nil when not initialized")
		}
	})
	
	t.Run("multiple resets and initializations", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)
		
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		
		for i := 0; i < 5; i++ {
			resetGlobalState()
			
			cfg := Config{
				Level:            "info",
				FilePath:         fmt.Sprintf("reset_%d.log", i),
				UseLocalTime:     true,
				FileMaxSizeInMB:  10,
				FileMaxAgeInDays: 7,
			}
			
			Init(cfg)
			logger := L()
			
			if logger == nil {
				t.Errorf("Logger should be initialized on iteration %d", i)
			}
			
			logger.Info("test message", "iteration", i)
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestStructuredLogging(t *testing.T) {
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
		Level:            "debug",
		FilePath:         "structured.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	// Test various structured logging patterns
	logger.Info("user action", 
		"user_id", 12345, 
		"action", "login", 
		"ip_address", "192.168.1.1", 
		"timestamp", "2024-01-01T00:00:00Z")
	
	logger.Warn("performance warning", 
		"operation", "database_query", 
		"duration_ms", 5000, 
		"table", "users", 
		"slow_query", true)
	
	logger.Error("system error", 
		"error_code", 500, 
		"component", "auth_service", 
		"details", map[string]interface{}{
			"nested": "value",
			"count":  42,
		})
	
	// Test with various data types
	logger.Debug("data types test", 
		"string", "value", 
		"int", 123, 
		"float", 45.67, 
		"bool", true, 
		"nil", nil)
	
	// Check that log file contains structured data
	logPath := filepath.Join(tempDir, "structured.log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	if len(content) == 0 {
		t.Error("Log file should contain structured log entries")
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerErrorHandling(t *testing.T) {
	t.Run("nil logger handling", func(t *testing.T) {
		resetGlobalState()
		
		// L() should return nil when not initialized
		logger := L()
		if logger != nil {
			t.Error("L() should return nil when not initialized")
		}
	})
	
	t.Run("multiple resets and initializations", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)
		
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		
		for i := 0; i < 5; i++ {
			resetGlobalState()
			
			cfg := Config{
				Level:            "info",
				FilePath:         fmt.Sprintf("reset_%d.log", i),
				UseLocalTime:     true,
				FileMaxSizeInMB:  10,
				FileMaxAgeInDays: 7,
			}
			
			Init(cfg)
			logger := L()
			
			if logger == nil {
				t.Errorf("Logger should be initialized on iteration %d", i)
			}
			
			logger.Info("test message", "iteration", i)
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerLevels(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	levels := []string{"debug", "info", "warn", "error"}
	
	for _, level := range levels {
		t.Run("level_"+level, func(t *testing.T) {
			cfg := Config{
				Level:            level,
				FilePath:         level + ".log",
				UseLocalTime:     true,
				FileMaxSizeInMB:  10,
				FileMaxAgeInDays: 7,
			}
			
			logger := New(cfg)
			
			// Test all logging methods
			logger.Debug("debug message", "level", level)
			logger.Info("info message", "level", level)
			logger.Warn("warn message", "level", level)
			logger.Error("error message", "level", level)
			
			// Check that log file was created
			logPath := filepath.Join(tempDir, level+".log")
			if _, err := os.Stat(logPath); os.IsNotExist(err) {
				t.Errorf("Log file should be created for level %s", level)
			}
		})
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerErrorHandling(t *testing.T) {
	t.Run("nil logger handling", func(t *testing.T) {
		resetGlobalState()
		
		// L() should return nil when not initialized
		logger := L()
		if logger != nil {
			t.Error("L() should return nil when not initialized")
		}
	})
	
	t.Run("multiple resets and initializations", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)
		
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		
		for i := 0; i < 5; i++ {
			resetGlobalState()
			
			cfg := Config{
				Level:            "info",
				FilePath:         fmt.Sprintf("reset_%d.log", i),
				UseLocalTime:     true,
				FileMaxSizeInMB:  10,
				FileMaxAgeInDays: 7,
			}
			
			Init(cfg)
			logger := L()
			
			if logger == nil {
				t.Errorf("Logger should be initialized on iteration %d", i)
			}
			
			logger.Info("test message", "iteration", i)
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestStructuredLogging(t *testing.T) {
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
		Level:            "debug",
		FilePath:         "structured.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	// Test various structured logging patterns
	logger.Info("user action", 
		"user_id", 12345, 
		"action", "login", 
		"ip_address", "192.168.1.1", 
		"timestamp", "2024-01-01T00:00:00Z")
	
	logger.Warn("performance warning", 
		"operation", "database_query", 
		"duration_ms", 5000, 
		"table", "users", 
		"slow_query", true)
	
	logger.Error("system error", 
		"error_code", 500, 
		"component", "auth_service", 
		"details", map[string]interface{}{
			"nested": "value",
			"count":  42,
		})
	
	// Test with various data types
	logger.Debug("data types test", 
		"string", "value", 
		"int", 123, 
		"float", 45.67, 
		"bool", true, 
		"nil", nil)
	
	// Check that log file contains structured data
	logPath := filepath.Join(tempDir, "structured.log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	if len(content) == 0 {
		t.Error("Log file should contain structured log entries")
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerErrorHandling(t *testing.T) {
	t.Run("nil logger handling", func(t *testing.T) {
		resetGlobalState()
		
		// L() should return nil when not initialized
		logger := L()
		if logger != nil {
			t.Error("L() should return nil when not initialized")
		}
	})
	
	t.Run("multiple resets and initializations", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)
		
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		
		for i := 0; i < 5; i++ {
			resetGlobalState()
			
			cfg := Config{
				Level:            "info",
				FilePath:         fmt.Sprintf("reset_%d.log", i),
				UseLocalTime:     true,
				FileMaxSizeInMB:  10,
				FileMaxAgeInDays: 7,
			}
			
			Init(cfg)
			logger := L()
			
			if logger == nil {
				t.Errorf("Logger should be initialized on iteration %d", i)
			}
			
			logger.Info("test message", "iteration", i)
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerTimeFormats(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	t.Run("local time format", func(t *testing.T) {
		cfg := Config{
			Level:            "info",
			FilePath:         "local_time.log",
			UseLocalTime:     true,
			FileMaxSizeInMB:  10,
			FileMaxAgeInDays: 7,
		}
		
		logger := New(cfg)
		logger.Info("message with local time")
		
		// Check that log file was created
		logPath := filepath.Join(tempDir, "local_time.log")
		content, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}
		
		if len(content) == 0 {
			t.Error("Log file should contain entries with local time")
		}
	})
	
	t.Run("UTC time format", func(t *testing.T) {
		cfg := Config{
			Level:            "info",
			FilePath:         "utc_time.log",
			UseLocalTime:     false,
			FileMaxSizeInMB:  10,
			FileMaxAgeInDays: 7,
		}
		
		logger := New(cfg)
		logger.Info("message with UTC time")
		
		// Check that log file was created
		logPath := filepath.Join(tempDir, "utc_time.log")
		content, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}
		
		if len(content) == 0 {
			t.Error("Log file should contain entries with UTC time")
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerErrorHandling(t *testing.T) {
	t.Run("nil logger handling", func(t *testing.T) {
		resetGlobalState()
		
		// L() should return nil when not initialized
		logger := L()
		if logger != nil {
			t.Error("L() should return nil when not initialized")
		}
	})
	
	t.Run("multiple resets and initializations", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)
		
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		
		for i := 0; i < 5; i++ {
			resetGlobalState()
			
			cfg := Config{
				Level:            "info",
				FilePath:         fmt.Sprintf("reset_%d.log", i),
				UseLocalTime:     true,
				FileMaxSizeInMB:  10,
				FileMaxAgeInDays: 7,
			}
			
			Init(cfg)
			logger := L()
			
			if logger == nil {
				t.Errorf("Logger should be initialized on iteration %d", i)
			}
			
			logger.Info("test message", "iteration", i)
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestStructuredLogging(t *testing.T) {
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
		Level:            "debug",
		FilePath:         "structured.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	// Test various structured logging patterns
	logger.Info("user action", 
		"user_id", 12345, 
		"action", "login", 
		"ip_address", "192.168.1.1", 
		"timestamp", "2024-01-01T00:00:00Z")
	
	logger.Warn("performance warning", 
		"operation", "database_query", 
		"duration_ms", 5000, 
		"table", "users", 
		"slow_query", true)
	
	logger.Error("system error", 
		"error_code", 500, 
		"component", "auth_service", 
		"details", map[string]interface{}{
			"nested": "value",
			"count":  42,
		})
	
	// Test with various data types
	logger.Debug("data types test", 
		"string", "value", 
		"int", 123, 
		"float", 45.67, 
		"bool", true, 
		"nil", nil)
	
	// Check that log file contains structured data
	logPath := filepath.Join(tempDir, "structured.log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	if len(content) == 0 {
		t.Error("Log file should contain structured log entries")
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerErrorHandling(t *testing.T) {
	t.Run("nil logger handling", func(t *testing.T) {
		resetGlobalState()
		
		// L() should return nil when not initialized
		logger := L()
		if logger != nil {
			t.Error("L() should return nil when not initialized")
		}
	})
	
	t.Run("multiple resets and initializations", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)
		
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		
		for i := 0; i < 5; i++ {
			resetGlobalState()
			
			cfg := Config{
				Level:            "info",
				FilePath:         fmt.Sprintf("reset_%d.log", i),
				UseLocalTime:     true,
				FileMaxSizeInMB:  10,
				FileMaxAgeInDays: 7,
			}
			
			Init(cfg)
			logger := L()
			
			if logger == nil {
				t.Errorf("Logger should be initialized on iteration %d", i)
			}
			
			logger.Info("test message", "iteration", i)
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerFileHandling(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	t.Run("large log file creation", func(t *testing.T) {
		cfg := Config{
			Level:            "debug",
			FilePath:         "large.log",
			UseLocalTime:     true,
			FileMaxSizeInMB:  1,
			FileMaxAgeInDays: 1,
		}
		
		logger := New(cfg)
		
		// Write many log entries to potentially trigger rotation
		for i := 0; i < 1000; i++ {
			logger.Info("Large log entry with lots of data to fill up the file quickly", 
				"iteration", i, 
				"data", string(make([]byte, 100)))
		}
		
		// Check that log file exists
		logPath := filepath.Join(tempDir, "large.log")
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			t.Error("Large log file should be created")
		}
	})
	
	t.Run("nested directory creation", func(t *testing.T) {
		cfg := Config{
			Level:            "info",
			FilePath:         "nested/deep/path/test.log",
			UseLocalTime:     false,
			FileMaxSizeInMB:  10,
			FileMaxAgeInDays: 7,
		}
		
		logger := New(cfg)
		logger.Info("test message in nested directory")
		
		// Check that nested directories and log file were created
		logPath := filepath.Join(tempDir, "nested/deep/path/test.log")
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			t.Error("Nested log file should be created along with directories")
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerErrorHandling(t *testing.T) {
	t.Run("nil logger handling", func(t *testing.T) {
		resetGlobalState()
		
		// L() should return nil when not initialized
		logger := L()
		if logger != nil {
			t.Error("L() should return nil when not initialized")
		}
	})
	
	t.Run("multiple resets and initializations", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)
		
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		
		for i := 0; i < 5; i++ {
			resetGlobalState()
			
			cfg := Config{
				Level:            "info",
				FilePath:         fmt.Sprintf("reset_%d.log", i),
				UseLocalTime:     true,
				FileMaxSizeInMB:  10,
				FileMaxAgeInDays: 7,
			}
			
			Init(cfg)
			logger := L()
			
			if logger == nil {
				t.Errorf("Logger should be initialized on iteration %d", i)
			}
			
			logger.Info("test message", "iteration", i)
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestStructuredLogging(t *testing.T) {
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
		Level:            "debug",
		FilePath:         "structured.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	// Test various structured logging patterns
	logger.Info("user action", 
		"user_id", 12345, 
		"action", "login", 
		"ip_address", "192.168.1.1", 
		"timestamp", "2024-01-01T00:00:00Z")
	
	logger.Warn("performance warning", 
		"operation", "database_query", 
		"duration_ms", 5000, 
		"table", "users", 
		"slow_query", true)
	
	logger.Error("system error", 
		"error_code", 500, 
		"component", "auth_service", 
		"details", map[string]interface{}{
			"nested": "value",
			"count":  42,
		})
	
	// Test with various data types
	logger.Debug("data types test", 
		"string", "value", 
		"int", 123, 
		"float", 45.67, 
		"bool", true, 
		"nil", nil)
	
	// Check that log file contains structured data
	logPath := filepath.Join(tempDir, "structured.log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	if len(content) == 0 {
		t.Error("Log file should contain structured log entries")
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerErrorHandling(t *testing.T) {
	t.Run("nil logger handling", func(t *testing.T) {
		resetGlobalState()
		
		// L() should return nil when not initialized
		logger := L()
		if logger != nil {
			t.Error("L() should return nil when not initialized")
		}
	})
	
	t.Run("multiple resets and initializations", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)
		
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		
		for i := 0; i < 5; i++ {
			resetGlobalState()
			
			cfg := Config{
				Level:            "info",
				FilePath:         fmt.Sprintf("reset_%d.log", i),
				UseLocalTime:     true,
				FileMaxSizeInMB:  10,
				FileMaxAgeInDays: 7,
			}
			
			Init(cfg)
			logger := L()
			
			if logger == nil {
				t.Errorf("Logger should be initialized on iteration %d", i)
			}
			
			logger.Info("test message", "iteration", i)
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerTimeFormats(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	t.Run("local time format", func(t *testing.T) {
		cfg := Config{
			Level:            "info",
			FilePath:         "local_time.log",
			UseLocalTime:     true,
			FileMaxSizeInMB:  10,
			FileMaxAgeInDays: 7,
		}
		
		logger := New(cfg)
		logger.Info("message with local time")
		
		// Check that log file was created
		logPath := filepath.Join(tempDir, "local_time.log")
		content, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}
		
		if len(content) == 0 {
			t.Error("Log file should contain entries with local time")
		}
	})
	
	t.Run("UTC time format", func(t *testing.T) {
		cfg := Config{
			Level:            "info",
			FilePath:         "utc_time.log",
			UseLocalTime:     false,
			FileMaxSizeInMB:  10,
			FileMaxAgeInDays: 7,
		}
		
		logger := New(cfg)
		logger.Info("message with UTC time")
		
		// Check that log file was created
		logPath := filepath.Join(tempDir, "utc_time.log")
		content, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}
		
		if len(content) == 0 {
			t.Error("Log file should contain entries with UTC time")
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func TestLoggerErrorHandling(t *testing.T) {
	t.Run("nil logger handling", func(t *testing.T) {
		resetGlobalState()
		
		// L() should return nil when not initialized
		logger := L()
		if logger != nil {
			t.Error("L() should return nil when not initialized")
		}
	})
	
	t.Run("multiple resets and initializations", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd)
		
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		
		for i := 0; i < 5; i++ {
			resetGlobalState()
			
			cfg := Config{
				Level:            "info",
				FilePath:         fmt.Sprintf("reset_%d.log", i),
				UseLocalTime:     true,
				FileMaxSizeInMB:  10,
				FileMaxAgeInDays: 7,
			}
			
			Init(cfg)
			logger := L()
			
			if logger == nil {
				t.Errorf("Logger should be initialized on iteration %d", i)
			}
			
			logger.Info("test message", "iteration", i)
		}
	})
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("benchmark debug", "iteration", i)
		}
	})
	
	b.Run("Error", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("benchmark error", "iteration", i, "error", "test_error")
		}
	})
}

func BenchmarkLoggerInit(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "bench.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetGlobalState()
		Init(cfg)
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	cfg := Config{
		Level:            "info",
		FilePath:         "benchmark.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  100,
		FileMaxAgeInDays: 7,
	}
	
	logger := New(cfg)
	
	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message", "iteration", i, "data", "test")
		}
	})
	
	b.Run("Debug", func(b *testing.B) {
		b.ResetTimer()
		f