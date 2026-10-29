# Issue 235 review

## Review scope

Reviewed the working-tree changes on branch `fix/235-userprofile-grpc-adapters` against the current base, focusing on the contributor and leaderboard-stat gRPC integrations, the user profile response path, contributor protobuf additions, configuration, and tests. The review covers the 19 changed files reported by `git diff --stat`.

## Final verdict

**Request changes.** The new profile path drops the GitHub username before it reaches the gRPC response, and the profile service can return HTTP 200 with an empty profile when its contributor lookup fails. Both issues affect core profile behavior.

## Findings

### P1 — Missing contributor lookup is reported as a successful empty profile

**Location:** `userprofileapp/service/userprofile/service.go:53-110`; visible at `userprofileapp/delivery/http/handler.go:36-53`.

The service records errors from all three data sources, logs them, and unconditionally returns a non-nil response with a nil error. In particular, a contributor-not-found or contributor-service failure becomes an HTTP 200 response containing zero-valued contributor data. The endpoint therefore makes an unavailable or nonexistent profile indistinguishable from a valid profile with empty fields, and clients cannot retry or show the appropriate not-found/error state. The new partial-profile tests exercise this behavior but do not verify that a failed required contributor lookup is surfaced.

**Recommendation:** Treat the contributor lookup as required and return its error (mapping a missing contributor to the appropriate not-found response). If partial responses are intended for stats or tasks, preserve that behavior only for those optional sections and make the partial status explicit to callers.

### P2 — GitHub username is lost in the new contributor gRPC endpoint

**Location:** `contributorapp/service/contributor/service.go:61-69` and `contributorapp/delivery/grpc/handler.go:99-107`.

`GetContributorInfo` copies `c.GitHubUsername` into the protobuf response, but `Service.GetProfile` never assigns `GitHubUsername` when constructing `GetProfileResponse`. As a result, `c.GitHubUsername` is always empty on this path. The new client and user profile adapter faithfully propagate that empty value, so every profile returned through this new integration lacks its GitHub username.

**Recommendation:** Populate `GitHubUsername` from the repository result in `Service.GetProfile`, and cover the new gRPC profile response with a test that asserts the username survives the service-to-response mapping.

## Tests and checks

- `git diff --check` — passed.
- `make check` — not completed. The first attempt could not write to the default Go module cache because it is read-only. A retry with module and build caches redirected to `/tmp` was blocked from downloading dependencies by the sandbox network policy. After network permission was granted, dependencies downloaded, but the `go build ./...` stage stopped producing output and no Go process remained; that check session was interrupted. A rerun with the now-cached dependencies also stopped producing output before reporting a result. Therefore build, vet, and race-test results are **inconclusive**, not passing.

## Scope decision confirmed

The intentionally unimplemented Task adapter V1 scope decision is **NOT a finding**. The stub behavior described in `docs/architecture.md` is accepted as intentional for this review.

# Iteration 2

## Changes since Iteration 1

Re-read the reviewer workflow, architecture guide, and full Iteration 1 review, then reviewed the current working-tree diff and the live text of Issue #235. The GitHub username mapping reported in Iteration 1 is fixed in the contributor service response path. The Task adapter remains an intentional V1 stub documented in the architecture guide and is accepted scope.

Issue #235's acceptance criteria say that ContributorInfo, ContributorStat, and Tasks come from their services over gRPC, and that timeouts and partial-failure behavior are tested (“one upstream down → partial profile, not 500”).

## P1 Re-evaluation

The original P1 is **not blocking for Issue #235**. Contributor is one of the upstreams covered by the partial-profile acceptance criterion, and the new service test explicitly verifies that a contributor lookup failure still returns the other available profile fields without an error. Returning an error for Contributor failure would change the issue contract and conflict with its stated expectation of a partial profile when one upstream is down.

**Non-blocking follow-up concern**

A successful HTTP response with zero-value contributor data can be ambiguous: clients cannot tell a failed Contributor lookup from a valid profile with empty data. That API-semantics concern should be addressed separately if callers need to distinguish those states; it is not a required behavior change for Issue #235.

## New findings

### Medium — An upstream unavailable at startup prevents the profile service from starting

**Location:** `userprofileapp/app.go:34-42`; `pkg/grpc/client.go:44-66`.

`Setup` creates both upstream clients before constructing the HTTP service. The shared client enables `grpc.WithBlock()` and bounds dialing with a context derived from `InitialBackoff` (30 seconds in the checked-in userprofile configs). If either upstream is unreachable during startup, `NewClient` returns an error and `Setup` exits, so the userprofile endpoint cannot serve a partial profile while that upstream is down.

**Recommendation:** Allow the profile service to start without a currently reachable upstream, then let request-time calls fail independently and return partial data. Add a focused test for startup with one unavailable upstream and for the timeout path required by the issue.

## Verification

- Read `AGENTS.md`, `docs/agentic/WORKFLOW.md`, `.claude/agents/reviewer.md`, `docs/architecture.md`, and the complete existing `docs/review/issue_235.md`.
- Opened Issue #235 and confirmed its goal and acceptance criteria, including the partial-profile and timeout-testing requirement.
- Reviewed the current working-tree diff and relevant userprofile service, adapter, HTTP handler, gRPC client, configuration, and tests.
- `git diff --check` — passed (exit code 0).
- No Go tests or `make check` were run; their results are unverified.

## Iteration 2 verdict

**Changes requested.** The original P1 is non-blocking and retained as an API-semantics follow-up concern. The startup availability finding above remains to be addressed against the partial-profile contract.

# Iteration 3

## Final review

The current implementation fetches ContributorInfo and ContributorStat through their gRPC adapters and calls the intentionally stubbed Task adapter. The profile service starts all three lookups independently, waits for each result, logs individual failures, and returns the values that succeeded. The Contributor, LeaderboardStat, and Task failure tests each assert a partial response and preserve the other successful sections. The service uses distinct result/error variables per goroutine, and the focused race test passes. The Task V1 stub is documented in `docs/architecture.md` and is accepted as intentionally scoped.

One issue requirement remains unverified and unimplemented: calls have no bounded request deadline, and no timeout test exercises a slow or non-returning upstream. `ContributorProfile` passes the incoming context directly to each adapter; the gRPC client retry policy configures retry backoff but does not impose an RPC deadline. Thus an upstream that stays connected without responding can keep the profile request waiting indefinitely, rather than producing a partial profile after a timeout.

## Previous findings status

- **P1:** Non-blocking follow-up concern, not a blocker for #235. Contributor failure returning zero-value contributor data in an otherwise successful partial profile follows the issue's “one upstream down → partial profile, not 500” requirement. The API-semantics ambiguity can be considered separately.
- **P2:** Resolved / not a valid remaining finding. The current `GetProfile` mapping includes `GitHubUsername`.
- **Startup availability:** Technically valid but out of scope for #235. It concerns service startup before profile retrieval requests, and shared gRPC client initialization is not changed as part of this issue.

## New findings

### Medium — Upstream calls have no bounded timeout, and timeout behavior is not tested

**Location:** `userprofileapp/service/userprofile/service.go:52-86`; `userprofileapp/service/userprofile/service_test.go:43-177`.

`ContributorProfile` passes the inbound context directly to each upstream without setting a call deadline. The checked-in tests cover immediate errors returned by mocks, but none simulate a deadline or an upstream that blocks past a timeout. The configured gRPC retry backoff is not a per-call timeout, so a connected upstream that stops responding can hold the handler open indefinitely instead of yielding the required partial profile.

**Recommendation:** Bound each upstream call with a request timeout and add a deterministic test where an upstream waits for its context deadline; assert the timed-out section is omitted while the other successful sections remain in the response.

## Verification

- Read `AGENTS.md`, `docs/agentic/WORKFLOW.md`, `.claude/agents/reviewer.md`, Issue #235, the complete prior review history, `docs/architecture.md`, the current diff, and relevant implementation/tests.
- `go test -race -short ./userprofileapp/service/userprofile/...` — passed (exit code 0; cached).
- `git diff --check` — passed (exit code 0), before appending this review section.
- `make check` — did not complete. The initial run failed because the default Go module/build cache paths were read-only. A retry using `/tmp/rankr-gomodcache` and `/tmp/rankr-gocache` was approved and downloaded dependencies, but produced no further output and had no Go build/test process remaining; I interrupted the still-active command (exit code 130). The repository gate is therefore inconclusive, not passing.

## Final verdict

**Changes requested**

# Iteration 4

## Review scope / summary

The fresh Reviewer independently reviewed the complete Issue #235 implementation, including Contributor and LeaderboardStat gRPC integration, UserProfile aggregation and HTTP behavior, timeout configuration and tests, additive protobuf changes, and configuration wiring. The Task adapter stub was explicitly considered and excluded from V1 according to the team scope decision documented in `docs/architecture.md`; it is not a finding for Issue #235.

## Final findings

There are **no remaining blocking findings**. The review confirmed that UserProfile runs upstream calls concurrently, returns a partial profile rather than HTTP 500 when an individual upstream fails, and preserves successful upstream results when another upstream fails. The configured 5-second upstream timeout is wired correctly, and deterministic service tests cover timeout behavior. The HTTP endpoint returns the partial profile with HTTP 200. Protobuf changes are additive, and configuration wiring is correct.

## Verification

- `go test -race -short ./userprofileapp/service/userprofile/...` — PASS.
- `git diff --check` — PASS.
- `make check` — PASS.

The reviewer reported that these checks had to be rerun outside the sandbox because gRPC tests could not bind a local TCP listener inside it. The checks passed after rerunning with the required environment access.

## Final verdict

**APPROVE**
