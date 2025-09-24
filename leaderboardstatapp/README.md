### Run leaderboard-stat App

```bash
  # run compose file
 docker compose -f deploy/leaderboardstat/development/docker-compose.no-service.yml up
  # run leaderboard-stat app
 go run ./cmd/leaderboardstat/main.go serve --migrate-up


 go run cmd/tui/main.go
 
 # select service and enter flags
 --migrate-up
```

### Stopping Service

```bash
  # stopping leaderboard-stat service
 ctrl + c
 
  # for down containers
 docker compose -f deploy/leaderboardstat/development/docker-compose.no-service.yml down
```

### Run Endpoints
```bash
 # check service healthy
 curl -X GET http://localhost:6011/v1/health-check
 
 # get a contributor's statistics
 curl -X GET http://localhost:6011/v1/contributors/1/stats
```

### Run Docker Containers
```bash
  ./deploy/docker-compose-dev-leaderboardstat-local.bash up
  
  docker build -f deploy/leaderboardstat/development/Dockerfile -t rankr-leaderboardstat:development . 
  
  # down containers
  ./deploy/docker-compose-dev-leaderboardstat-local.bash down
```