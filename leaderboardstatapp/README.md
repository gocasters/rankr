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
 curl -X GET http://localhost:6011/v1/contributors/8/stats
```

```bash
 # discover available grpc services
 grpcurl -plaintext localhost:8090 list
 grpcurl -plaintext localhost:8090 list leaderboardstat.LeaderboardStatService

 # describe a method
 grpcurl -plaintext localhost:8090 describe leaderboardstat.LeaderboardStatService.GetContributorStats
```

```bash 
 # call a gRPC method
  grpcurl -plaintext \
  -d '{"contributor_id": 8}' \
  localhost:8090 leaderboardstat.LeaderboardStatService.GetContributorStats
```

### Run Docker Containers
```bash
  ./deploy/docker-compose-dev-leaderboardstat-local.bash up
  
  docker build -f deploy/leaderboardstat/development/Dockerfile -t rankr-leaderboardstat:development . 
  
  # down containers
  ./deploy/docker-compose-dev-leaderboardstat-local.bash down
```