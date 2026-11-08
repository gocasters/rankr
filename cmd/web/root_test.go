package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestServeReturnsMissingConfigError(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "missing.yml")
	t.Setenv("CONFIG_PATH", configPath)

	root := newRootCommand()
	root.SetArgs([]string{"serve"})
	root.SilenceUsage = true
	root.SilenceErrors = true

	err := root.ExecuteContext(t.Context())
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected missing config error, got: %v", err)
	}
}
