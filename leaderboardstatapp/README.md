### Run leaderboard-stat app

```bash
  # run compose file
 docker compose -f deploy/leaderboardstat/development/docker-compose.no-service.yml up
  # run leaderboard-stat app
 go run ./cmd/leaderboardstat/main.go serve --migrate-up

```

### Stopping service

```bash
  # stopping leaderboard-stat service
 ctrl + c
 
  # for down containers
 docker compose -f deploy/leaderboardstat/development/docker-compose.no-service.yml down
```