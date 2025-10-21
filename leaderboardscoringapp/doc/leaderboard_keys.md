## **Leaderboard Redis Key Structure & Snapshot Policy**

## **1. Redis Sorted Set Keys**

Each leaderboard is stored as a **Redis Sorted Set (ZSET)** using the following key patterns:

| Scope                           | Key Pattern                          | Description                              |
|---------------------------------|--------------------------------------|------------------------------------------|
| **Per Project (ProjectID = 1)** | `leaderboard:1:weekly:2025-w43`      | Weekly leaderboard for project 1         |
|                                 | `leaderboard:1:monthly:2025-10`      | Monthly leaderboard for project 1        |
|                                 | `leaderboard:1:yearly:2025`          | Yearly leaderboard for project 1         |
|                                 | `leaderboard:1:all_time`             | All-time leaderboard for project 1       |
| **Global**                      | `leaderboard:global:weekly:2025-w43` | Weekly leaderboard across all projects   |
|                                 | `leaderboard:global:monthly:2025-10` | Monthly leaderboard across all projects  |
|                                 | `leaderboard:global:yearly:2025`     | Yearly leaderboard across all projects   |
|                                 | `leaderboard:global:all_time`        | All-time leaderboard across all projects |

Each ZSET stores:

* **Member:** `user_id`
* **Score:** accumulated ranking score

---

## **2. Snapshot Selection**

| Key Type                            | Included in Snapshot? | Reason                                            |  
|-------------------------------------|-----------------------|---------------------------------------------------|
| `leaderboard:<project_id>:all_time` | ✅ Yes                 | Captures each project’s overall ranking over time |
| `leaderboard:global:all_time`       | ✅ Yes                 | Preserves system-wide ranking history             |    

---

## **3. Summary**

Snapshots are taken **only for `all_time` leaderboards** (both per-project and global) to:

* Reduce processing and storage cost
* Preserve long-term ranking data
* Avoid redundancy with short-lived leaderboards (weekly, monthly, yearly)