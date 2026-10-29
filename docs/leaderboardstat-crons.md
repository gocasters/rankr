# leaderboardstat crons

`leaderboardstatapp` runs two scheduled jobs. They decide how fresh the public leaderboard and the contributor stats are.

Code: [`leaderboardstatapp/delivery/scheduler/scheduler.go`](../leaderboardstatapp/delivery/scheduler/scheduler.go). Config: the `scheduler_cfg` block in `deploy/leaderboardstat/{development,production}/config*.yml`.

## The jobs

| Job name (in logs) | Config key | Shipped schedule | Code default | Feeds |
|---|---|---|---|---|
| `public-leaderboard-update` | `public_leaderboard_cron` | `* * * * *` (every minute) | `*/3 * * * *` | `GET /v1/leaderboard/public/:project_id` |
| `daily-score-calculation` | `daily_score_calculation_cron` | `*/10 * * * *` (every 10 minutes) | `0 2 * * *` | `GET /v1/contributors/:id/stats` and gRPC `GetContributorStats` |

Both jobs are tagged `leaderboardstat-service`. The shipped schedules are the same in `development` and `production`. The code default is used only when the key is missing or empty.

### `public-leaderboard-update`

Calls `Service.SetPublicLeaderboard`:

1. It lists every project from projectapp over gRPC (`ListProjects`, 100 per page).
2. Each project's **daily** leaderboard comes from leaderboardscoring over gRPC (`GetLeaderboard`, `timeframe: daily`, `project_id: <git_repo_id>`). That leaderboard holds today's scores only and resets when leaderboardscoring's local date changes.
3. It replaces the Redis sorted set `public_leaderboard:project:<git_repo_id>` (DEL, ZADD, EXPIRE) and sets a **3-minute TTL**. The TTL is hard-coded in `SetPublicLeaderboard`.

Projects without a `git_repo_id` are skipped. The `:project_id` in the public URL is matched against that Redis key, so in practice it is the GitHub repository ID.

### `daily-score-calculation`

Calls `Service.GetDailyContributorScores`:

1. It reads the global **daily** leaderboard from leaderboardscoring over gRPC, 10 rows per page.
2. Each row is inserted into the Postgres table `scores` with `status = 0` (pending).
3. Pending rows are rolled up into `user_project_scores` and the global scores, then marked processed.

"Daily" is the data window it reads. It is not how often the job runs.

## Config keys

```yaml
scheduler_cfg:
  public_leaderboard_cron: "* * * * *"         # crontab for public-leaderboard-update
  job_context_timeout: "1m"                    # max duration of one run, either job
  daily_score_calculation_cron: "*/10 * * * *" # crontab for daily-score-calculation
```

| Key | Type | Code default | Env override |
|---|---|---|---|
| `public_leaderboard_cron` | 5-field crontab | `*/3 * * * *` | `STAT_SCHEDULER_CFG__PUBLIC_LEADERBOARD_CRON` |
| `daily_score_calculation_cron` | 5-field crontab | `0 2 * * *` | `STAT_SCHEDULER_CFG__DAILY_SCORE_CALCULATION_CRON` |
| `job_context_timeout` | Go duration (`90s`, `1m`, `15m`) | `3m` (also used when ≤ 0) | `STAT_SCHEDULER_CFG__JOB_CONTEXT_TIMEOUT` |

- **Crontab format:** five fields (`minute hour day-of-month month day-of-week`), with no seconds field.
- **Time zone:** the scheduler uses the process's local time zone (`time.Local`). In a container that is usually UTC unless `TZ` is set.
- **Invalid crontab:** the service still starts, but the job is not registered. Startup logs `failed to create … job`.
- **Env overrides:** environment variables take precedence over the YAML file. The prefix is `STAT_`, and `__` separates nesting levels.

## How fresh is the data?

| What you read | Refreshed | Worst-case age in normal operation |
|---|---|---|
| Public leaderboard | Every minute | About 1 minute plus one run (a run is capped by `job_context_timeout`, 1m) |
| Contributor stats | Every 10 minutes | About 10 minutes plus one run |

Behaviour worth knowing:

- **Startup:** neither job runs at startup. The first run happens at the next crontab match. On a fresh start the public leaderboard is empty for up to a minute.
- **No overlap:** jobs run in singleton mode (`LimitModeWait`). A slow run delays the next one and never runs concurrently with it.
- **Failure means empty, not stale:** if `public-leaderboard-update` keeps failing, the Redis key expires 3 minutes after the last successful run. The endpoint then returns an empty list.
- **Code default schedule:** with `*/3 * * * *`, the refresh interval equals the 3-minute TTL. A key can expire just before it is rebuilt, which causes short empty responses. Keep the schedule shorter than the TTL.
- **Day boundary:** both jobs read the **daily** timeframe. The public leaderboard therefore shows today's scores only and starts again from zero each day.

## Checking it is running

Every run logs `<job>Task started` and then either `... completed successfully` or `failed to run ...`. At startup each job logs its name, UUID and `crontab`:

```bash
# from the service's stdout / log file
grep -E 'publicLeaderboard job created|dailyScoreCalculation job created' app.log
grep -E 'publicLeaderboardTask (started|completed)|failed to run publicLeaderboardTask' app.log
```

## Known issues (not fixed here)

- **Over-counting in stats:** `daily-score-calculation` inserts today's *running* total into `scores` on every run, and `GetContributorTotalScore` sums all rows. At `*/10 * * * *` a contributor's total therefore grows with every run, not with every event. This needs its own issue before the schedule or the query changes.
- **Random project ID:** `GetDailyContributorScores` stores a random `project_id` (`math/rand`, `TODO add project_id to row response`). Per-project stats from this job are not meaningful yet.
- **Production schedules are the development ones:** `production/config.yml` ships the same every-minute and every-10-minutes schedules. Choosing production values is a maintainer decision.
