# Agentic workflow — one page

Small on purpose: four roles, one gate (`make check`), and one rule. **Agents propose, humans decide.**

```
 1. Issue ─────────────▶ [product]  acceptance criteria   (optional) ──▶ HUMAN approves the criteria
 2. Criteria ──────────▶ [tester]   failing tests (red), committed first
 3. Red tests ─────────▶ [developer] smallest change to green; make check after each step ──▶ PR
 4. PR ────────────────▶ [reviewer] fresh session, read-only review of the diff ──▶ author fixes or answers
 5. CI make check + buddy review + final approver ──▶ HUMAN squash-merges
 6. Deploy to stage (when the pipeline exists) ──▶ Friday demo: flip the flag
```

## Who acts when

| Step | Role | Required? | Output |
|---|---|---|---|
| Clarify | `product` | Only when the issue has no testable acceptance criteria | A comment on the issue: goal, criteria, out of scope, flag, demo |
| Specify | `tester` | Yes, for any behaviour change | Failing table or golden tests that fail for the right reason |
| Build | `developer` | Yes | Code that turns the tests green, `make check` green, a PR with disclosure |
| Check | `reviewer` | Yes, in a **new** session (fresh context) | Findings with severity and `file:line`, three "why?" questions, a verdict |
| Decide | humans | Always | The author answers every question; buddy and final approver review; a human merges |

## How much process

- **Tiny** (docs, typo, config value): developer only. `make check` and a human review.
- **Normal** (most tasks): tester → developer → reviewer.
- **Risky** (scoring rules, auth or OAuth, webhook verification, migrations, proto, CI): product → tester → developer → reviewer. A one-paragraph plan goes in the PR before the code, and a maintainer approves.

## Loop limits

- One task per session. Start a fresh session for each issue.
- After two failed attempts at the same step, `git reset` to the last green commit and re-plan. Don't argue with the agent.
- Autonomous "keep going until green" runs are allowed only when `make check` or a specific test is the judge, for at most 5 rounds, and never on risky tasks.

## What is deliberately not here

No hooks, no skills library, no orchestration and no extra tools. Add something only after the same problem has hit us twice, and prefer a test or a lint rule over a new paragraph. A repeated agent mistake becomes one line in `AGENTS.md`, added by PR.
