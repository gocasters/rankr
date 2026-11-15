# Issue #238 review: F-09 score breakdown by event type

## Findings

### Major — required check fails to compile stats service tests

- **File/path:** `leaderboardstatapp/service/leaderboardstat/daily_scores_test.go:182` and `:207`
- **Relevant function:** `TestGetDailyContributorScores` and `TestGetDailyContributorScores_RowProjectIDIsAuthoritative`
- **What is wrong:** Both tests still call `NewService` with six arguments, but Issue #238 added the required `ContributorClient` seventh argument. `make check` stops at vet with “not enough arguments in call to NewService.”
- **Why it matters:** The required Rankr gate cannot pass, so this branch is not ready to proceed through the project workflow.
- **Recommendation:** Update both test setups to provide a test `ContributorClient`, then rerun `make check`.

### Minor — new breakdown behavior has no focused tests

- **File/path:** `leaderboardscoringapp/repository/database/database_repository.go:256`, `leaderboardscoringapp/service/leaderboardscoring/service.go:464`, `leaderboardscoringapp/delivery/grpc/handler.go:84`, and `leaderboardstatapp/service/leaderboardstat/service.go:386`
- **Relevant functions:** `GetVCSUserScoreBreakdown` and `GetContributorStats`
- **What is wrong:** No tests exercise the new SQL category aggregation, scoring service and RPC mappings, contributor ID to GitHub ID lookup and decimal conversion, or mapping of the returned breakdown into the public stats response. Existing stats client tests also do not assert `ScoreBreakdown` or preservation of `ScoreHistory` with the new field.
- **Why it matters:** The behavior is spread across repository, service, and RPC boundaries. Without focused coverage, category classification or identifier/response mapping errors can return plausible but incorrect totals. The Rankr workflow requires tests first for behavior changes.
- **Recommendation:** Add focused tests for PR/review/issue aggregation (including empty and unrelated event types), repository and downstream errors, the ID conversion and RPC request, and public response mapping while asserting the existing score history remains intact.

## Review summary

The current implementation follows the requested ownership path: `leaderboardstat` resolves the internal contributor ID through Contributor RPC, converts its GitHub ID to a decimal string, and calls `leaderboardscoring`; the scoring repository sums `score_delta` from `processed_score_events` by `event_type`. The current category lists correspond to the event names in the scoring domain. The new protobuf fields and scoring RPC are additive, the public stats response includes `score_breakdown`, and the existing score history mapping remains in place. Userprofile propagation is out of scope and is not a finding.

The prior scheduler dependency issue is fixed in the current tree: `runScheduler` creates the Contributor RPC client, constructs the Contributor adapter, passes it to `leaderboardstat.NewService`, and defers closing the RPC client. The normal HTTP/gRPC application setup also owns and closes the Contributor RPC client on setup errors and shutdown. No obvious N+1 query or expensive processing is introduced; the breakdown uses one aggregate query.

## Validation

- `make check`: **failed**. Initial sandboxed attempt could not write Go caches. Re-run with cache-write access reached `go build ./...` and gofmt, then failed in vet at `leaderboardstatapp/service/leaderboardstat/daily_scores_test.go:182` because `NewService` now requires a seventh `ContributorClient` argument. This is caused by Issue #238's constructor change, not an identified pre-existing baseline failure. `make check` did not reach the remaining tests/build gate stages.
- `git diff --check`: **failed**. Added trailing whitespace at `leaderboardscoringapp/repository/database/database_repository.go:267` and `:274`.
- Other read-only checks: reviewed the root instructions, all `.claude/agents/*` role guidance, workflow and architecture docs, supplied Issue #238 requirements, full current diff, event names and processed-event schema, tests, client/config wiring, scheduler and server resource lifecycle, and git status. No nested `AGENTS.md` files were present. The required check may have run gofmt; a subsequent status check showed no changes outside the pre-existing Issue #238 diff and this review report.

## Verdict

Needs changes before tests

# Iteration 2

## Findings

No implementation findings.

## Review summary

The score breakdown is aggregated in `leaderboardscoring` from `processed_score_events` using `score_delta` and the persisted `event_type`, and `leaderboardstat` passes the GitHub/VCS user ID returned by Contributor RPC to the scoring RPC. The PR, review, and issue categories map to event names defined by the scoring domain; the aggregate query returns zeroes when a user has no matching rows and excludes unrelated event types. Scoring rules are not recreated in `leaderboardstat`, and the existing score history mapping is preserved.

The new scoring RPC and stats response fields are additive protobuf changes. Request validation, error-to-status mapping, response mapping, and caller context propagation are present. Contributor and leaderboard scoring RPC clients are retained by the stats application, closed during setup failures and shutdown, and created/closed by the standalone scheduler command. Contributor RPC configuration is present in development, local development, and production configs. All current `leaderboardstat.NewService` call sites, including the previously failing calls in `daily_scores_test.go`, now supply the new argument.

## Validation

- `make check`: **passed** (`check: OK`), including build, gofmt, vet, and the scoped `go test -race -short` run.
- `git diff --check`: **failed** on two trailing spaces in the new SQL at `leaderboardscoringapp/repository/database/database_repository.go:266` and `:272`.
- Other read-only checks: reviewed the root and nested agent instructions (only the root `AGENTS.md` exists), `.claude/agents/*`, workflow and architecture docs, supplied Issue #238 requirements, current diff, event definitions/scoring rules, protobuf changes, configuration, service wiring, resource lifecycle, and current constructor usages. Working tree remains limited to the Issue #238 changes and this review document.

## Verdict

Ready for tests

# Final Review

## Findings

No findings.

## Review summary

The score breakdown is calculated in `leaderboardscoring` from `processed_score_events` using the persisted `user_id`, `event_type`, and `score_delta`. The PR, review, and issue classifications match the event constants and database constraint; the aggregate query sums multiple category events, excludes other users and unrelated event types, and returns zeroes when no rows match. `leaderboardstat` resolves the contributor ID through Contributor RPC, converts the GitHub ID to a decimal string, calls the scoring RPC, and maps the returned categories to its stats response. Existing score history and score fields are retained. The protobuf additions are additive, and the Contributor RPC client is wired and closed in application setup/shutdown and the scheduler command.

The repository and service tests cover the requested aggregation behavior and cross-service mapping, including history preservation. Config additions are consistent across development, local development, and production. I found no material performance or ownership issue in the new path.

## Test review

The repository integration test seeds multiple PR, review, and issue events, an unrelated commit event, and an event for another user, then checks the category sums and the empty-user zero result. The stats service test checks contributor ID lookup, decimal GitHub ID conversion, the scoring client request, category mapping, and preservation of score history. This is focused coverage for the behavior and boundary risks in scope.

## Validation

- `make check`: passed (`check: OK`), including build, gofmt, vet, and scoped `go test -race -short`.
- `git diff --check`: passed.
- Other read-only checks: reviewed the project instructions, all role guidance, workflow and architecture docs, Issue #238 requirements, full implementation diff, persisted event schema and event constants, tests, protobuf/API mapping, configs, client wiring, and shutdown paths. No nested `AGENTS.md` files are present. The required check did not introduce additional working-tree changes.

## Verdict

Ready for PR
