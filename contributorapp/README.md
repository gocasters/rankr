# contributorapp

Service for contributor profiles (read/update), contributor creation (auth-facing), the upsert bridge, and the GitHub CSV/XLSX import pipeline (`service/job` parses files; `delivery/` serves HTTP/gRPC and the import worker).

## Testing

All tests in this service are unit tests: no Docker, no database, no network. **Read this before changing anything under `contributorapp/`** and keep the suite green.

```bash
# full unit suite of the service — the fast loop, must always be green
go test -short -count=1 ./contributorapp/...

# with the race detector (what `make check` runs on the packages your branch changes)
go test -short -race -count=1 ./contributorapp/...

# the repo gate: gofmt, vet, -race -short tests and build on changed packages
make check

# while iterating: one package or one scenario
go test -short -count=1 -v -run TestValidate ./contributorapp/service/contributor/
go test -short -count=1 -v -run TestCSV      ./contributorapp/service/job/
```

### What the suite covers (issue #274, TD-23)

| File | Scenarios | Covers |
|---|---|---|
| `service/contributor/validator_test.go` | A1–A14 | Validation contract: privacy-mode enum, username/password length bounds, GitHubID rules, update DTO shape, error-message constants |
| `service/contributor/service_test.go` | B1–B16 | Profile mapping without secret leakage, create (bcrypt hashing), update ordering, upsert new/existing — via a testify mock of `Repository` |
| `service/contributor/password_test.go` | — | `UpdatePassword` (old-password check, bcrypt of the new one), `VerifyPassword` (authapp login contract, ID-over-username lookup), `passwordMatches` |
| `*/knownbug_test.go` | — | `knownBug` helper and the issue ids of the known bugs |
| `service/job/process_test.go` | C1–C13 | CSV parser: header validation, per-row vs whole-file failures, quoted fields, concurrent 500-row run, row numbering |
| `service/job/testdata/*.csv` | — | Fixtures: `valid.csv`, `short_row.csv`, `bad_id.csv`, `missing_col.csv` |

### Rules when touching these tests

- They are **characterization tests**: they lock today's behaviour. Never change an expected value to make a test pass (AGENTS.md hard rule 2). If behaviour must change, change code and test in the same PR and explain why.
- **Known bugs are `TestKnownBug_*` tests** (helper: `knownbug_test.go` in each package). They assert the *correct* behaviour; while the bug exists they report `SKIP` with the issue id, and once the code is fixed they **fail** with "appears FIXED". Then remove the `knownBug(...)` wrapper and the issue constant and keep the assertions as a normal test. Never make one pass by changing its assertions to match the code. List the open bugs with `go test -short -v -run KnownBug ./contributorapp/... | grep SKIP`.
- One surprising behaviour is agreed and locked as a normal test: a short or long CSV row aborts the whole file (`TestCSV_ShortRow_FailsWholeFile`, `TestCSV_LongRow_FailsWholeFile`).
- `-short` is mandatory (nothing may require Docker or network), no sleeps for synchronisation (assert on sets/counts instead), and `-race` must stay green.
- Tests sit in the same directory and package as the code (`package contributor`, `package job`) so they can reach unexported helpers; mock new repository methods by hand with `testify/mock` (see `mockRepository` in `service_test.go`) — no extra mocking dependency.
- `service/contributor` initialises the logger once in `TestMain`; if you add tests anywhere in this service that reach `logger.L()`, do the same — it panics when uninitialised.
