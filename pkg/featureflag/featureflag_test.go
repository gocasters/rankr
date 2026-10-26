package featureflag

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gocasters/rankr/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// clearFeatureEnv removes every RANKR_FEATURE_* variable for the duration of
// the test and restores the originals afterwards, so ambient environment can
// never change what New resolves.
func clearFeatureEnv(t *testing.T) {
	t.Helper()

	type kv struct{ key, value string }

	var saved []kv
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, EnvPrefix) {
			continue
		}
		key, value, _ := strings.Cut(entry, "=")
		saved = append(saved, kv{key: key, value: value})
		require.NoError(t, os.Unsetenv(key))
	}
	t.Cleanup(func() {
		for _, e := range saved {
			_ = os.Setenv(e.key, e.value)
		}
	})
}

func TestParseState(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    State
		wantErr bool
	}{
		{name: "off", value: "off", want: Off},
		{name: "shadow", value: "shadow", want: Shadow},
		{name: "on", value: "on", want: On},
		{name: "case insensitive", value: " On ", want: On},
		{name: "empty means off", value: "", want: Off},
		{name: "anything else is invalid", value: "enabled", wantErr: true},
		{name: "booleans are not values", value: "true", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseState(tt.value)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStateString(t *testing.T) {
	assert.Equal(t, "off", Off.String())
	assert.Equal(t, "shadow", Shadow.String())
	assert.Equal(t, "on", On.String())
}

// TestFlagsFromServiceConfigFeatures walks the whole path a service walks:
// `features:` in a YAML config file, through config.Load, into New.
func TestFlagsFromServiceConfigFeatures(t *testing.T) {
	clearFeatureEnv(t)

	tests := []struct {
		value string
		want  State
	}{
		{value: "off", want: Off},
		{value: "shadow", want: Shadow},
		{value: "on", want: On},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			yamlPath := filepath.Join(t.TempDir(), "config.yml")
			require.NoError(t, os.WriteFile(yamlPath, []byte("features:\n  scoring_v1: "+tt.value+"\n"), 0o600))

			var cfg struct {
				Features map[string]string `koanf:"features"`
			}
			options := config.Options{
				YamlFilePath: yamlPath,
				Prefix:       "TESTFEATURE_",
				Delimiter:    ".",
				Separator:    "__",
			}
			require.NoError(t, config.Load(options, &cfg))
			require.Equal(t, map[string]string{"scoring_v1": tt.value}, cfg.Features)

			flags, err := New(cfg.Features)
			require.NoError(t, err)
			assert.Equal(t, tt.want, flags.State("scoring_v1"))
		})
	}
}

func TestEnvOverridesServiceConfig(t *testing.T) {
	clearFeatureEnv(t)

	tests := []struct {
		name   string
		config string // empty means the flag is absent from features:
		envSet bool
		env    string
		want   State
	}{
		{name: "config only", config: "shadow", want: Shadow},
		{name: "env wins over config", config: "shadow", envSet: true, env: "on", want: On},
		{name: "env can turn config off", config: "on", envSet: true, env: "off", want: Off},
		{name: "env without config", envSet: true, env: "shadow", want: Shadow},
		{name: "empty env value means off", config: "on", envSet: true, env: "", want: Off},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var features map[string]string
			if tt.config != "" {
				features = map[string]string{"scoring_v1": tt.config}
			}
			if tt.envSet {
				t.Setenv(EnvPrefix+"SCORING_V1", tt.env)
			}

			flags, err := New(features)
			require.NoError(t, err)
			assert.Equal(t, tt.want, flags.State("scoring_v1"))
		})
	}
}

func TestNewRejectsUnregisteredConfigName(t *testing.T) {
	clearFeatureEnv(t)

	flags, err := New(map[string]string{"ghost_flag": "on"})
	require.Error(t, err)
	assert.Nil(t, flags)
	assert.Contains(t, err.Error(), "ghost_flag")
}

func TestNewRejectsUnregisteredEnvVar(t *testing.T) {
	clearFeatureEnv(t)
	t.Setenv(EnvPrefix+"GHOST_FLAG", "on")

	flags, err := New(nil)
	require.Error(t, err)
	assert.Nil(t, flags)
	assert.Contains(t, err.Error(), "GHOST_FLAG")
}

func TestNewRejectsInvalidValues(t *testing.T) {
	clearFeatureEnv(t)

	t.Run("config", func(t *testing.T) {
		flags, err := New(map[string]string{"scoring_v1": "enabled"})
		require.Error(t, err)
		assert.Nil(t, flags)
		assert.Contains(t, err.Error(), "scoring_v1")
	})

	t.Run("env", func(t *testing.T) {
		t.Setenv(EnvPrefix+"SCORING_V1", "maybe")
		flags, err := New(nil)
		require.Error(t, err)
		assert.Nil(t, flags)
		assert.Contains(t, err.Error(), "SCORING_V1")
	})
}

func TestUnconfiguredAndNilFlagsAreOff(t *testing.T) {
	clearFeatureEnv(t)

	flags, err := New(nil)
	require.NoError(t, err)
	assert.Equal(t, Off, flags.State("scoring_v1"))
	assert.False(t, flags.Enabled("scoring_v1"))
	assert.False(t, flags.Shadow("scoring_v1"))

	var nilFlags *Flags
	assert.Equal(t, Off, nilFlags.State("scoring_v1"))
	assert.False(t, nilFlags.Enabled("scoring_v1"))
	assert.False(t, nilFlags.Shadow("scoring_v1"))
}

func TestRegistered(t *testing.T) {
	assert.Contains(t, Registered(), "scoring_v1")
	assert.True(t, IsRegistered("scoring_v1"))
	assert.False(t, IsRegistered("scoring.v1"))
	assert.False(t, IsRegistered("ghost_flag"))
}

// TestRegisteredNamesSurviveConfigAndEnv pins the flag-name charset: koanf
// uses "." as its delimiter, so a dotted name would be mangled by the YAML
// config, and the env override must map name -> RANKR_FEATURE_<NAME>
// losslessly.
func TestRegisteredNamesSurviveConfigAndEnv(t *testing.T) {
	names := Registered()
	require.NotEmpty(t, names)

	for _, name := range names {
		require.NotEmpty(t, name)
		for _, r := range name {
			isLowerAlphaNum := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
			require.True(t, isLowerAlphaNum || r == '_', "flag name %q must match [a-z0-9_]+", name)
		}
		require.Equal(t, name, strings.ToLower(strings.ToUpper(name)))
	}
}
