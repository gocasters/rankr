// Package featureflag is the registry and evaluator of rankr feature flags.
//
// A flag name is declared once in the registry (Registered) and gets a row in
// docs/flags.md. Each service resolves its flags from the `features:` map of
// its own config file, overridden per flag by RANKR_FEATURE_<NAME> in the
// environment. Values are off, shadow or on; anything unset is off.
package featureflag

import (
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
)

// State is how far a flag has been rolled out.
type State uint8

// Flag states, in increasing order of exposure.
const (
	// Off leaves the old behaviour in place. It is the zero value, so a flag
	// nobody configured is off (AGENTS.md hard rule 6).
	Off State = iota
	// Shadow runs the new behaviour but keeps serving the old result: use it
	// to compare outputs or collect metrics before anyone is affected.
	Shadow
	// On serves the new behaviour.
	On
)

// EnvPrefix is the prefix of the environment variable that overrides a
// single flag: RANKR_FEATURE_<NAME>, where NAME is the upper-cased flag name,
// e.g. RANKR_FEATURE_SCORING_V1.
const EnvPrefix = "RANKR_FEATURE_"

// String returns the configuration value of s: off, shadow or on.
func (s State) String() string {
	switch s {
	case Shadow:
		return "shadow"
	case On:
		return "on"
	default:
		return "off"
	}
}

// ParseState parses a flag value: "off", "shadow" or "on", compared
// case-insensitively. An empty value means off. Anything else is an error, so
// a typo in config or environment stops the service at boot instead of
// quietly running the wrong behaviour.
func ParseState(value string) (State, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "off":
		return Off, nil
	case "shadow":
		return Shadow, nil
	case "on":
		return On, nil
	default:
		return Off, fmt.Errorf("invalid value %q: must be off, shadow or on", value)
	}
}

// Registered returns every flag name that code is allowed to read. It builds
// a fresh slice on each call: the registry is read-only, never mutated at
// runtime. Every name here must also have a row in docs/flags.md, and must
// match [a-z0-9_]+ so it survives both the koanf "." delimiter in YAML and
// the RANKR_FEATURE_<NAME> environment mapping.
func Registered() []string {
	return []string{"scoring_v1"}
}

// IsRegistered reports whether name is a registered flag.
func IsRegistered(name string) bool {
	return slices.Contains(Registered(), name)
}

// Flags is the resolved flag state of one service process. The zero value and
// a nil *Flags read as every flag off.
type Flags struct {
	states map[string]State
}

// New resolves a service's flag states from its config `features:` map, with
// each RANKR_FEATURE_<NAME> environment variable overriding the config value.
// It returns an error when a name is not in the registry or a value is not
// off|shadow|on, for config and environment alike: a typo fails at boot
// instead of silently disabling (or enabling) the wrong thing.
func New(features map[string]string) (*Flags, error) {
	states := make(map[string]State, len(features))

	for _, name := range slices.Sorted(maps.Keys(features)) {
		if !IsRegistered(name) {
			return nil, fmt.Errorf("featureflag: features: %q is not in the registry: add it to pkg/featureflag and docs/flags.md", name)
		}
		state, err := ParseState(features[name])
		if err != nil {
			return nil, fmt.Errorf("featureflag: features.%s: %w", name, err)
		}
		states[name] = state
	}

	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, EnvPrefix) {
			continue
		}
		key, value, _ := strings.Cut(entry, "=")
		name := strings.ToLower(strings.TrimPrefix(key, EnvPrefix))
		if !IsRegistered(name) {
			return nil, fmt.Errorf("featureflag: %s: flag %q is not in the registry: add it to pkg/featureflag and docs/flags.md", key, name)
		}
		state, err := ParseState(value)
		if err != nil {
			return nil, fmt.Errorf("featureflag: %s: %w", key, err)
		}
		states[name] = state
	}

	return &Flags{states: states}, nil
}

// State returns the state of name. A flag that is unknown or was never
// configured is Off.
func (f *Flags) State(name string) State {
	if f == nil {
		return Off
	}
	return f.states[name]
}

// Enabled reports whether name is on: the new behaviour is live.
func (f *Flags) Enabled(name string) bool { return f.State(name) == On }

// Shadow reports whether name is in shadow: the new behaviour runs for
// comparison, the old behaviour stays live.
func (f *Flags) Shadow(name string) bool { return f.State(name) == Shadow }
