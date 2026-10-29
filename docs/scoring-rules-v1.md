# Scoring rules v1

Status: **draft**, for approval by the maintainers in the week-2 session (#229).
Implemented by #233 behind the `scoring.v1` flag. Until then the legacy formula in
`leaderboardscoringapp/service/leaderboardscoring/service.go` (`calculateScore`) stays live.

## Principles

1. **Score outcomes, not activity.** Points come from work the project accepted. Opening,
   commenting and pushing never score on their own.
2. **Deterministic.** A score depends only on the event payload and this rule set. No
   `math/rand`, no processing time: time windows use the event's `time` field.
3. **Idempotent.** Each event is scored at most once, keyed by its event `id`. Replaying the
   same events produces the same leaderboard.
4. **Append-only corrections.** A reverted or voided contribution is undone by a negative
   correction event, never by editing history.
5. **Versioned and public.** Changing a weight or a cap means a new version of this file.

## What scores

| Outcome | Points | Who gets them | Detected from |
|---|---|---|---|
| Merged pull request | 10 | PR author (`user_id`) | `pull_request_closed` with `merged = true` |
| Review on a pull request that is later merged | 3 | Reviewer (`reviewer_user_id`) | `pull_request_review` with state `APPROVED` or `CHANGES_REQUESTED`, scored when that PR merges |
| Accepted issue closed by a merged PR | 2 | Issue author (`issue_author_id`) | `issue_closed` with reason `COMPLETED` and a linked merged PR (see data gaps) |

## What never scores

- Opening a PR or an issue.
- Comments (`issue_comment`), whatever their length or content.
- Pushes and commits (`push`).
- A PR closed without merge, and an issue closed as `NOT_PLANNED`.
- Reviews with state `COMMENTED`, and reviews on a PR that is closed without merge.
- Merging someone else's PR (`merger_user_id`): the author gets the points, the merger does not.

## Caps and anti-gaming

| Rule | Value |
|---|---|
| Reviews scored per reviewer per PR | 1 (the last qualifying review counts) |
| Reviews scored per reviewer per day (UTC, by event `time`) | 5 |
| Self-review (`reviewer_user_id == pr_author_user_id`) | 0 |
| Self-filed issue closed by own PR | 0 for the issue (the PR still scores) |
| Bot accounts (e.g. dependabot, CodeRabbit) | 0, from an explicit excluded-account list in config |
| Reverted PR | The original merge is undone with a −10 correction; the revert PR itself scores 0 |
| Voided by an admin (F-20, #249) | Negative correction equal to the points awarded |

## Data gaps to close before v1 can go live

These rules depend on data the pipeline does not deliver correctly today.

1. **Issue → closing PR link is wrong.** `webhookapp/service/delivery/issues.go` fills
   `closing_pr_number` from `issue.pull_request`, which GitHub sets only when the issue *is* a
   PR. The issues webhook does not carry the closing PR. Options: resolve `closingIssuesReferences`
   of a merged PR via the GitHub API, or parse `Closes #N` from the PR body (needs a new, additive
   payload field). Until fixed, the "accepted issue" rule scores 0.
2. **Reviews arrive before the merge.** Scoring a review "on a merged PR" means holding qualifying
   reviews per `pr_id` and awarding them when `pull_request_closed` with `merged = true` arrives.
   This is a #233 engine concern; the rule itself does not change.
3. **No revert signal.** A revert is a normal merged PR. Detection proposal: GitHub's revert PRs use
   the branch `revert-<number>-…` and the title `Revert "<original title>"`, both available in
   `pull_request_opened`. Anything not matched automatically goes through the admin void tool (#249).
4. **No bot flag in payloads.** Payloads carry only user IDs, so bots are excluded by a config list
   of account IDs.
5. **Merged-PR points go to the merger (#306).** The live webhook path sets
   `PullRequestClosedPayload.user_id` to `sender.id` (`webhookapp/service/delivery/pull_request.go`),
   which is whoever clicked merge. The history path already uses the PR author. Until #306 is fixed,
   the "merged pull request" rule credits the wrong person whenever someone else merges.

## Decisions for the maintainers

- [ ] Weights: 10 / 3 / 2 as above.
- [ ] Daily review cap: 5.
- [ ] Count `CHANGES_REQUESTED` reviews, or only `APPROVED`.
- [ ] Revert detection: branch/title convention plus the admin tool, or the admin tool only for v1.
- [ ] Accepted issues: fix the closing-PR link in v1, or ship v1 without that rule.
- [ ] Backfill: rescore history with v1 in shadow mode (#233) before switching the flag (#240).