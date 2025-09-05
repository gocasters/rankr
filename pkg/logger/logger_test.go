package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// --- mapLevel Tests ---
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
		{"unknown level defaults to info", "???", slog.LevelInfo},
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

// --- Init Tests ---
func TestInit(t *testing.T) {
	resetGlobals()

	tempDir := t.TempDir()
	// Change the working directory to the temporary one to avoid creating log files in the project root.
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Could not get working directory: %v", err)
	}
	os.Chdir(tempDir)
	defer os.Chdir(originalWd) // Change back to original directory after test.

	cfg := Config{
		Level:            "debug",
		FilePath:         "test.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  5,
		FileMaxAgeInDays: 3,
	}

	if err := Init(cfg); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	if globalLogger == nil {
		t.Fatal("Init() should initialize globalLogger")
	}

	// Test sync.Once: Init should not re-initialize the logger.
	oldLogger := globalLogger
	if err := Init(cfg); err != nil {
		t.Fatalf("Second Init() failed: %v", err)
	}
	if globalLogger != oldLogger {
		t.Error("Init() should only run once and not create a new logger instance")
	}
}

// TestInit_Concurrent tests that calling Init from multiple goroutines is safe
// and only initializes the logger once.
func TestInit_Concurrent(t *testing.T) {
	resetGlobals()

	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Could not get working directory: %v", err)
	}
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	cfg := Config{FilePath: "concurrent.log"}

	var wg sync.WaitGroup
	numGoroutines := 50
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			// Use MustInit for simplicity as handling errors in goroutines within tests is complex.
			MustInit(cfg)
		}()
	}
	wg.Wait()

	// Capture the first logger instance and verify that subsequent calls do not change it.
	firstInstance := L()
	MustInit(cfg) // Call again from the main goroutine
	secondInstance := L()

	if firstInstance != secondInstance {
		t.Error("Init() created a new logger instance on a subsequent call, sync.Once failed")
	}
}

// TestInit_WritesToStdout checks if the logger correctly writes to standard output,
// in addition to the log file.
func TestInit_WritesToStdout(t *testing.T) {
	resetGlobals()
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Could not get working directory: %v", err)
	}
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	cfg := Config{
		FilePath: "stdout_test.log",
		Level:    "info",
	}

	// Redirect os.Stdout to our pipe
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	MustInit(cfg)
	// At the end, restore stdout and close the global logger
	defer func() {
		os.Stdout = originalStdout
		Close()
	}()

	// Use a channel to signal when reading is complete
	outC := make(chan string)
	// Read all output from the pipe in a separate goroutine
	go func() {
		var b strings.Builder
		io.Copy(&b, r)
		r.Close()
		outC <- b.String()
	}()

	// Perform the log operation. This writes to the pipe.
	L().Info("hello stdout", "user", "test")

	// Close the write-end of the pipe. This will unblock the io.Copy in the goroutine.
	w.Close()

	// Wait for the reading goroutine to finish and get the captured output.
	output := <-outC

	if !strings.Contains(output, `"level":"INFO"`) {
		t.Errorf("Expected stdout log to contain level, but got: %s", output)
	}
	if !strings.Contains(output, `"msg":"hello stdout"`) {
		t.Errorf("Expected stdout log to contain message, but got: %s", output)
	}
	if !strings.Contains(output, `"user":"test"`) {
		t.Errorf("Expected stdout log to contain attribute, but got: %s", output)
	}
}

// --- MustInit Tests ---
func TestMustInit(t *testing.T) {
	resetGlobals()

	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Could not get working directory: %v", err)
	}
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	cfg := Config{
		Level:            "debug",
		FilePath:         "mustinit.log",
		FileMaxSizeInMB:  1,
		FileMaxAgeInDays: 1,
	}

	// Should not panic on a valid configuration.
	MustInit(cfg)
	if globalLogger == nil {
		t.Fatal("MustInit should initialize the global logger")
	}
}

// --- L Tests ---
func TestL_PanicBeforeInit(t *testing.T) {
	resetGlobals()

	// Calling L() before Init() should cause a panic.
	defer func() {
		if r := recover(); r == nil {
			t.Error("L() should panic when logger is not initialized, but it did not")
		}
	}()
	_ = L()
}

func TestL_AfterInit(t *testing.T) {
	resetGlobals()

	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Could not get working directory: %v", err)
	}
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	cfg := Config{
		Level:    "info",
		FilePath: "test_l_after_init.log",
	}

	if err := Init(cfg); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if L() != globalLogger {
		t.Error("L() should return the initialized globalLogger instance")
	}
}

// --- Close Tests ---
func TestClose(t *testing.T) {
	t.Run("close before init", func(t *testing.T) {
		resetGlobals()
		// Calling Close before Init should be a no-op and not return an error.
		err := Close()
		if err != nil {
			t.Errorf("Close() before Init() should not return an error, but got: %v", err)
		}
	})

	t.Run("close multiple times", func(t *testing.T) {
		resetGlobals()
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Could not get working directory: %v", err)
		}
		os.Chdir(tempDir)
		defer os.Chdir(originalWd)

		cfg := Config{FilePath: "test_close.log"}
		MustInit(cfg)

		// First call should succeed.
		err1 := Close()
		if err1 != nil {
			t.Errorf("First Close() should not return an error, but got: %v", err1)
		}
		if globalWriter != nil {
			t.Error("globalWriter should be nil after the first Close()")
		}

		// Second call should be a no-op and not return an error.
		err2 := Close()
		if err2 != nil {
			t.Errorf("Second Close() should not return an error, but got: %v", err2)
		}
	})
}

// --- New Tests ---
func TestNew_ReturnsSeparateInstances(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Could not get working directory: %v", err)
	}
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	cfg := Config{
		Level:    "warn",
		FilePath: "service.log",
	}

	l1, closer1, err := New(cfg)
	if err != nil {
		t.Fatalf("First New() failed: %v", err)
	}
	defer closer1.Close()

	l2, closer2, err := New(cfg)
	if err != nil {
		t.Fatalf("Second New() failed: %v", err)
	}
	defer closer2.Close()

	if l1 == nil || l2 == nil {
		t.Fatal("New() returned a nil logger")
	}
	if l1 == l2 {
		t.Error("New() should return different logger instances on each call")
	}
}

// TestNew_Integration provides an end-to-end test for a logger created with New().
// It checks if logging, closing, and file writing work as expected.
func TestNew_Integration(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Could not get working directory: %v", err)
	}
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	logFilePath := "new_logger_integration.log"
	cfg := Config{
		Level:    "debug",
		FilePath: logFilePath,
	}

	logger, closer, err := New(cfg)
	if err != nil {
		t.Fatalf("New() failed with error: %v", err)
	}
	if logger == nil || closer == nil {
		t.Fatal("New() returned nil logger or closer")
	}

	// Log a message and then close the writer to ensure it's flushed to disk.
	logger.Debug("message from New() logger", "id", 42)
	if err := closer.Close(); err != nil {
		t.Fatalf("Failed to close the logger writer: %v", err)
	}

	// Verify the content of the log file.
	content, err := os.ReadFile(filepath.Join(tempDir, logFilePath))
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, `"msg":"message from New() logger"`) {
		t.Error("Log file does not contain the expected message")
	}
	if !strings.Contains(logContent, `"level":"DEBUG"`) {
		t.Error("Log file does not contain the expected log level")
	}
	if !strings.Contains(logContent, `"id":42`) {
		t.Error("Log file does not contain the expected attribute")
	}
}

// --- Integration Test ---
func TestLoggerIntegration(t *testing.T) {
	resetGlobals()

	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Could not get working directory: %v", err)
	}
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	cfg := Config{
		Level:            "info",
		FilePath:         "integration.log",
		FileMaxSizeInMB:  1,
		FileMaxAgeInDays: 1,
	}

	// Initialize the global logger.
	if err := Init(cfg); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Use the global logger.
	l := L()
	l.Info("integration test", "key", "value")

	// Close the logger to flush and release resources.
	if err := Close(); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}

	// Verify the log file was created and contains data.
	logPath := filepath.Join(tempDir, "integration.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	if len(data) == 0 {
		t.Error("Log file is empty after writing and closing")
	}
	if !strings.Contains(string(data), `"msg":"integration test"`) {
		t.Error("Log file content is incorrect")
	}

	// Ensure the global writer is nil after Close().
	if globalWriter != nil {
		t.Error("globalWriter should be nil after Close()")
	}
}

// --- Helpers ---

// resetGlobals resets the global state of the logger package.
// This is crucial for ensuring tests are isolated from each other.
func resetGlobals() {
	// Close any existing global writer to release file handles.
	if globalWriter != nil {
		globalWriter.Close()
		globalWriter = nil
	}
	globalLogger = nil
	once = sync.Once{}
}
