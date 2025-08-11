package config_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gocasters/rankr/pkg/config"
)

// TestConfig represents a sample configuration structure for testing
type TestConfig struct {
	Database struct {
		Host     string `koanf:"host"`
		Port     int    `koanf:"port"`
		Username string `koanf:"username"`
		Password string `koanf:"password"`
	} `koanf:"database"`
	Server struct {
		Port int    `koanf:"port"`
		Host string `koanf:"host"`
	} `koanf:"server"`
	LogLevel string `koanf:"log_level"`
}

func TestLoadFromYamlFile(t *testing.T) {
	// Create a temporary YAML file
	yamlContent := `
database:
  host: localhost
  port: 5432
  username: testuser
  password: testpass
server:
  port: 8080
  host: 0.0.0.0
log_level: info
`
	tmpFile := createTempYAMLFile(t, yamlContent)
	defer os.Remove(tmpFile.Name())

	var cfg TestConfig
	options := config.Options{
		YamlFilePath: tmpFile.Name(),
	}

	err := config.Load(options, &cfg)
	require.NoError(t, err)

	// Verify default options are applied
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "info", cfg.LogLevel)
}

func TestLoadFromEnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("APP__DATABASE__HOST", "prod-db.example.com")
	os.Setenv("APP__DATABASE__PORT", "5433")
	os.Setenv("APP__SERVER__PORT", "9090")
	defer func() {
		os.Unsetenv("APP__DATABASE__HOST")
		os.Unsetenv("APP__DATABASE__PORT")
		os.Unsetenv("APP__SERVER__PORT")
	}()

	var cfg TestConfig
	options := config.Options{
		Prefix:    "APP",
		Delimiter: ".",
		Separator: "__",
	}

	err := config.Load(options, &cfg)
	require.NoError(t, err)

	fmt.Printf("config %+v", cfg)
	// Environment variables should override YAML values
	assert.Equal(t, "prod-db.example.com", cfg.Database.Host)
	assert.Equal(t, 5433, cfg.Database.Port)
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "info", cfg.LogLevel) // Not overridden by env var
}

// func TestLoadWithCustomTransformer(t *testing.T) {
// 	// Create a temporary YAML file
// 	yamlContent := `
// database:
//   host: localhost
//   port: 5432
// `
// 	tmpFile := createTempYAMLFile(t, yamlContent)
// 	defer os.Remove(tmpFile.Name())

// 	// Set environment variables
// 	os.Setenv("CUSTOM__DB__HOST", "custom-host")
// 	defer os.Unsetenv("CUSTOM__DB__HOST")

// 	var cfg TestConfig
// 	options := config.Options{
// 		YamlFilePath: tmpFile.Name(),
// 		Prefix:       "CUSTOM",
// 		Delimiter:    ".",
// 		Separator:    "__",
// 		Transformer: func(key, value string) (string, any) {
// 			// Custom transformer that maps "db" to "database"
// 			if strings.HasPrefix(key, "db.") {
// 				newKey := strings.Replace(key, "db.", "database.", 1)
// 				return newKey, value
// 			}
// 			return key, value
// 		},
// 	}

// 	err := config.Load(options, &cfg)
// 	require.NoError(t, err)

// 	// Custom transformer should map CUSTOM__DB__HOST to database.host
// 	assert.Equal(t, "custom-host", cfg.Database.Host)
// }

// func TestLoad_ErrorHandling(t *testing.T) {
// 	tests := []struct {
// 		name        string
// 		yamlContent string
// 		expectError bool
// 	}{
// 		{
// 			name:        "Invalid YAML",
// 			yamlContent: `database: host: localhost: invalid`,
// 			expectError: true,
// 		},
// 		{
// 			name:        "Empty YAML",
// 			yamlContent: ``,
// 			expectError: false,
// 		},
// 		{
// 			name:        "Valid YAML",
// 			yamlContent: `database: {host: localhost}`,
// 			expectError: false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tmpFile := createTempYAMLFile(t, tt.yamlContent)
// 			defer os.Remove(tmpFile.Name())

// 			var cfg TestConfig
// 			options := config.Options{
// 				YamlFilePath: tmpFile.Name(),
// 			}

// 			err := config.Load(options, &cfg)
// 			if tt.expectError {
// 				assert.Error(t, err)
// 			} else {
// 				assert.NoError(t, err)
// 			}
// 		})
// 	}
// }

// func TestLoad_NonExistentFile(t *testing.T) {
// 	var cfg TestConfig
// 	options := config.Options{
// 		YamlFilePath: "/non/existent/file.yaml",
// 	}

// 	err := config.Load(options, &cfg)
// 	assert.Error(t, err)
// }

// func TestLoad_EnvironmentVariablePriority(t *testing.T) {
// 	// Create YAML with some values
// 	yamlContent := `
// database:
//   host: yaml-host
//   port: 5432
// server:
//   port: 8080
// `
// 	tmpFile := createTempYAMLFile(t, yamlContent)
// 	defer os.Remove(tmpFile.Name())

// 	// Set environment variables (should override YAML)
// 	os.Setenv("TEST__DATABASE__HOST", "env-host")
// 	os.Setenv("TEST__DATABASE__PORT", "5433")
// 	defer func() {
// 		os.Unsetenv("TEST__DATABASE__HOST")
// 		os.Unsetenv("TEST__DATABASE__PORT")
// 	}()

// 	var cfg TestConfig
// 	options := config.Options{
// 		YamlFilePath: tmpFile.Name(),
// 		Prefix:       "TEST",
// 		Delimiter:    ".",
// 		Separator:    "__",
// 	}

// 	err := config.Load(options, &cfg)
// 	require.NoError(t, err)

// 	// Environment variables should take priority
// 	assert.Equal(t, "env-host", cfg.Database.Host)
// 	assert.Equal(t, 5433, cfg.Database.Port)
// 	assert.Equal(t, 8080, cfg.Server.Port) // Not overridden
// }

// func TestLoad_DifferentDelimiters(t *testing.T) {
// 	// Create YAML file
// 	yamlContent := `
// database:
//   host: localhost
// `
// 	tmpFile := createTempYAMLFile(t, yamlContent)
// 	defer os.Remove(tmpFile.Name())

// 	// Set environment variables with different delimiter
// 	os.Setenv("TEST_DATABASE_HOST", "custom-host")
// 	defer os.Unsetenv("TEST_DATABASE_HOST")

// 	var cfg TestConfig
// 	options := config.Options{
// 		YamlFilePath: tmpFile.Name(),
// 		Prefix:       "TEST",
// 		Delimiter:    "_",
// 		Separator:    "_",
// 	}

// 	err := config.Load(options, &cfg)
// 	require.NoError(t, err)

// 	// Should work with underscore delimiter
// 	assert.Equal(t, "custom-host", cfg.Database.Host)
// }

// func TestLoad_ComplexNestedStructure(t *testing.T) {
// 	// Test with a more complex config structure
// 	type ComplexConfig struct {
// 		App struct {
// 			Name     string `koanf:"name"`
// 			Version  string `koanf:"version"`
// 			Features struct {
// 				Auth  bool `koanf:"auth"`
// 				Cache bool `koanf:"cache"`
// 				API   bool `koanf:"api"`
// 			} `koanf:"features"`
// 		} `koanf:"app"`
// 		Services struct {
// 			Database struct {
// 				Type string `koanf:"type"`
// 				Host string `koanf:"host"`
// 				Port int    `koanf:"port"`
// 				SSL  bool   `koanf:"ssl"`
// 			} `koanf:"database"`
// 			Redis struct {
// 				Host string `koanf:"host"`
// 				Port int    `koanf:"port"`
// 			} `koanf:"redis"`
// 		} `koanf:"services"`
// 	}

// 	yamlContent := `
// app:
//   name: test-app
//   version: 1.0.0
//   features:
//     auth: true
//     cache: false
//     api: true
// services:
//   database:
//     type: postgres
//     host: localhost
//     port: 5432
//     ssl: false
//   redis:
//     host: localhost
//     port: 6379
// `
// 	tmpFile := createTempYAMLFile(t, yamlContent)
// 	defer os.Remove(tmpFile.Name())

// 	// Set some environment variables
// 	os.Setenv("COMPLEX__APP__NAME", "prod-app")
// 	os.Setenv("COMPLEX__SERVICES__DATABASE__HOST", "prod-db")
// 	os.Setenv("COMPLEX__SERVICES__DATABASE__SSL", "true")
// 	defer func() {
// 		os.Unsetenv("COMPLEX__APP__NAME")
// 		os.Unsetenv("COMPLEX__SERVICES__DATABASE__HOST")
// 		os.Unsetenv("COMPLEX__SERVICES__DATABASE__SSL")
// 	}()

// 	var cfg ComplexConfig
// 	options := config.Options{
// 		YamlFilePath: tmpFile.Name(),
// 		Prefix:       "COMPLEX",
// 		Delimiter:    ".",
// 		Separator:    "__",
// 	}

// 	err := config.Load(options, &cfg)
// 	require.NoError(t, err)

// 	// Verify complex structure is loaded correctly
// 	assert.Equal(t, "prod-app", cfg.App.Name) // Overridden by env
// 	assert.Equal(t, "1.0.0", cfg.App.Version) // From YAML
// 	assert.True(t, cfg.App.Features.Auth)
// 	assert.False(t, cfg.App.Features.Cache)
// 	assert.True(t, cfg.App.Features.API)
// 	assert.Equal(t, "postgres", cfg.Services.Database.Type)
// 	assert.Equal(t, "prod-db", cfg.Services.Database.Host) // Overridden by env
// 	assert.Equal(t, 5432, cfg.Services.Database.Port)
// 	assert.True(t, cfg.Services.Database.SSL) // Overridden by env
// 	assert.Equal(t, "localhost", cfg.Services.Redis.Host)
// 	assert.Equal(t, 6379, cfg.Services.Redis.Port)
// }

// // Helper function to create temporary YAML files for testing
// func createTempYAMLFile(t testing.TB, content string) *os.File {
// 	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
// 	require.NoError(t, err)

// 	_, err = tmpFile.WriteString(content)
// 	require.NoError(t, err)

// 	err = tmpFile.Close()
// 	require.NoError(t, err)

// 	return tmpFile
// }

// // Benchmark test for performance
// func BenchmarkLoad(b *testing.B) {
// 	yamlContent := `
// database:
//   host: localhost
//   port: 5432
// server:
//   port: 8080
// log_level: info
// `
// 	tmpFile := createTempYAMLFile(b, yamlContent)
// 	defer os.Remove(tmpFile.Name())

// 	var cfg TestConfig
// 	options := config.Options{
// 		YamlFilePath: tmpFile.Name(),
// 	}

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		config.Load(options, &cfg)
// 	}
// }
