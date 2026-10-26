# Feature flags

Unfinished work merges to `main` behind a flag. A flag has exactly one home in
the registry, `pkg/featureflag`, and exactly one row under [Flags](#flags)
below. The registry and this file must agree: `TestCodeOnlyReadsRegisteredFlags`
in `pkg/featureflag` fails when code reads a flag that is missing from the
registry.

## Values

| Value | Meaning |
|---|---|
| `off` | Old behaviour stays live. The default: a flag nobody configured is off (`AGENTS.md` hard rule 6). |
| `shadow` | The new behaviour runs for comparison — a shadow score, a log line, a metric — but the old behaviour stays live for users. |
| `on` | The new behaviour is live for everyone. |

Any other value is a boot-time error: the service refuses to start rather than
guess. An empty value counts as `off`.

## Where flags come from

Each service can set flags in its own config file, under `features:`:

```yaml
features:
  scoring_v1: shadow
```

```go
type Config struct {
	// ...
	Features map[string]string `koanf:"features"`
}

flags, err := featureflag.New(cfg.Features) // error stops the service at boot
if flags.Shadow("scoring_v1") {
	// run the new path, keep serving the old result
}
if flags.Enabled("scoring_v1") {
	// the new behaviour is live
}
```

One environment variable overrides one flag, taking precedence over the config
file. `RANKR_FEATURE_<NAME>` is the flag name upper-cased, for example:

```sh
RANKR_FEATURE_SCORING_V1=on ./bin/rankr
```

Both sources are strict. A name that is not in the registry, or a value
outside `off|shadow|on`, stops the service at boot instead of silently
running the wrong behaviour.

Flag names match `[a-z0-9_]+`: koanf uses `.` as its key delimiter, so a
dotted name would be mangled in YAML, and environment variable names cannot
carry dots portably.

## Adding a flag

1. Add the name to `Registered()` in `pkg/featureflag/featureflag.go`.
2. Add a row to the table below, and open the removal issue.
3. Read the flag only as `flags.State("name")` / `flags.Enabled("name")` /
   `flags.Shadow("name")` with a string literal. The scan test covers non-test
   Go files and fails on any literal missing from the registry.
4. Ship it `off`. Flip it to `shadow` on stage, then `on` at the Friday demo.

## Removing a flag

When the removal issue closes: delete the row here, delete the name from the
registry — the scan test then reports every leftover call site — and delete
those call sites in the same PR.

## Flags

Columns: **Name** is the registry key; **Owner** is the GitHub handle that
flips and removes it; **Week** is the Season week it is expected to flip;
**Status** is the current value (`off`, `shadow` or `on`); **Removal issue**
is the issue that flips it and then deletes it.

| Name | Owner | Week | Status | Removal issue |
|---|---|---|---|---|
| `scoring_v1` | TBD | TBD | off | #240 |

`scoring_v1` gates scoring rules v1 (`docs/scoring-rules-v1.md`), implemented
behind the flag by #233.
