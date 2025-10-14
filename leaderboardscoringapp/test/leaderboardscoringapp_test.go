package test_test

import (
	"context"
	"fmt"
	redisadapter "github.com/gocasters/rankr/adapter/redis"
	postgrerepository "github.com/gocasters/rankr/leaderboardscoringapp/repository/database"
	"github.com/gocasters/rankr/leaderboardscoringapp/repository/redisrepository"
	"github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"strconv"
	"testing"
	"time"
)

type IntegrationTestSuite struct {
	suite.Suite
	ctx               context.Context
	postgresContainer testcontainers.Container
	redisContainer    testcontainers.Container
	network           testcontainers.Network
	networkName       string

	postgresConn *database.Database
	redisClient  *redis.Client

	persistence leaderboardscoring.EventPersistence
	leaderboard leaderboardscoring.LeaderboardCache

	mockPublisher *MockNATSPublisher
	//app *leaderboardscoringapp.Application
}

func TestIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(IntegrationTestSuite))
}

// SetupSuite initializes all containers and dependencies
func (suite *IntegrationTestSuite) SetupSuite() {
	logger.Init(logger.Config{
		Level:            "debug",
		FilePath:         "logs/leaderboardscoringapp/integration_test.log",
		UseLocalTime:     true,
		FileMaxSizeInMB:  10,
		FileMaxAgeInDays: 7,
	})

	suite.ctx = context.Background()

	var networkName = "leaderboard-integration-net" + uuid.New().String()
	suite.networkName = networkName

	// Create network for container communication
	req := testcontainers.NetworkRequest{
		Name:           suite.networkName,
		CheckDuplicate: true,
	}
	nw, err := testcontainers.GenericNetwork(suite.ctx, testcontainers.GenericNetworkRequest{
		NetworkRequest: req,
	})
	suite.Require().NoError(err)
	suite.network = nw

	// Start PostgreSQL container
	suite.setupPostgres()

	// Start Redis container
	suite.setupRedis()

	// Initialize repositories
	suite.initializeRepositories()

	time.Sleep(3 * time.Second)
}

func (suite *IntegrationTestSuite) setupPostgres() {
	req := testcontainers.ContainerRequest{
		Image:    "postgres:15-alpine",
		Networks: []string{suite.networkName},
		Hostname: "leaderboardscoring-db",
		Env: map[string]string{
			"POSTGRES_USER":     "test_user",
			"POSTGRES_PASSWORD": "test_pass",
			"POSTGRES_DB":       "test_db",
		},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections"),
			wait.ForListeningPort("5432/tcp"),
		).WithStartupTimeoutDefault(120 * time.Second),
	}

	container, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(suite.T(), err)
	suite.postgresContainer = container

	host, err := container.Host(suite.ctx)
	require.NoError(suite.T(), err)

	port, err := container.MappedPort(suite.ctx, "5432")
	require.NoError(suite.T(), err)

	dbConfig := database.Config{
		Host:              host,
		Port:              port.Int(),
		Username:          "test_user",
		Password:          "test_pass",
		DBName:            "test_db",
		SSLMode:           "disable",
		MaxConns:          10,
		MinConns:          2,
		MaxConnLifetime:   300,
		MaxConnIdleTime:   60,
		HealthCheckPeriod: 60,
	}

	var connErr error
	suite.postgresConn, connErr = database.Connect(dbConfig)
	require.NoError(suite.T(), connErr)

	// Verify connection
	err = suite.postgresConn.Pool.Ping(suite.ctx)
	require.NoError(suite.T(), err)

	// Run migrations
	suite.runMigrations()
}

func (suite *IntegrationTestSuite) setupRedis() {
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		Networks:     []string{suite.networkName},
		Hostname:     "leaderboardscoring-redis",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(suite.T(), err)
	suite.redisContainer = container

	host, err := container.Host(suite.ctx)
	require.NoError(suite.T(), err)

	port, err := container.MappedPort(suite.ctx, "6379")
	require.NoError(suite.T(), err)

	redisAdapter, err := redisadapter.New(suite.ctx, redisadapter.Config{
		Host:     host,
		Port:     port.Int(),
		Password: "",
		DB:       0,
	})
	require.NoError(suite.T(), err)

	suite.redisClient = redisAdapter.Client()

	pErr := suite.redisClient.Ping(suite.ctx).Err()
	require.NoError(suite.T(), pErr)
}

func (suite *IntegrationTestSuite) runMigrations() {
	createTablesSQL := `
	CREATE TABLE IF NOT EXISTS processed_score_events (
		id BIGSERIAL PRIMARY KEY,
		user_id VARCHAR(100) NOT NULL,
		event_type VARCHAR(50) NOT NULL,
		event_timestamp TIMESTAMP NOT NULL,
		score_delta BIGINT NOT NULL,
		processed_at TIMESTAMP DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_score_events_user_id ON processed_score_events (user_id);
	CREATE INDEX IF NOT EXISTS idx_score_events_event_type ON processed_score_events (event_type);
	
	CREATE TABLE IF NOT EXISTS user_total_scores (
		id BIGSERIAL PRIMARY KEY,
		user_id VARCHAR(100) NOT NULL,
		total_score BIGINT NOT NULL,
		snapshot_timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
		CONSTRAINT uniq_user_total_scores_user_ts UNIQUE (user_id, snapshot_timestamp)
	);

	CREATE INDEX IF NOT EXISTS idx_user_total_scores_user_ts 
		ON user_total_scores (user_id, snapshot_timestamp DESC);
	`

	_, err := suite.postgresConn.Pool.Exec(suite.ctx, createTablesSQL)
	require.NoError(suite.T(), err)
}

func (suite *IntegrationTestSuite) initializeRepositories() {
	suite.persistence = postgrerepository.NewPostgreSQLRepository(
		suite.postgresConn,
		postgrerepository.RetryConfig{
			MaxRetries: 3,
			RetryDelay: 100 * time.Millisecond,
		},
	)

	suite.leaderboard = redisrepository.NewRedisLeaderboardRepository(suite.redisClient)

	suite.mockPublisher = &MockNATSPublisher{}
}

func (suite *IntegrationTestSuite) TearDownSuite() {
	if suite.postgresContainer != nil {
		_ = suite.postgresContainer.Terminate(suite.ctx)
	}
	if suite.redisContainer != nil {
		_ = suite.redisContainer.Terminate(suite.ctx)
	}
	if suite.network != nil {
		_ = suite.network.Remove(suite.ctx)
	}
}

func (suite *IntegrationTestSuite) SetupTest() {
	// Clear Redis before each test
	suite.redisClient.FlushAll(suite.ctx)

	// Clear database tables
	suite.postgresConn.Pool.Exec(suite.ctx, "TRUNCATE processed_score_events CASCADE")
	suite.postgresConn.Pool.Exec(suite.ctx, "TRUNCATE user_total_scores CASCADE")

	suite.mockPublisher.Reset()
}

// Integration Tests

// Process score one event
func (suite *IntegrationTestSuite) TestProcessScoreEvent_Success() {
	ctx := context.Background()

	service := leaderboardscoring.NewService(
		suite.persistence,
		suite.leaderboard,
		suite.mockPublisher,
		"processed_events",
		leaderboardscoring.NewValidator(),
	)

	userID := uint64(123)
	eventReq := &leaderboardscoring.EventRequest{
		ID:             uuid.New().String(),
		UserID:         strconv.FormatUint(userID, 10),
		EventName:      leaderboardscoring.PullRequestOpened.String(),
		RepositoryID:   1001,
		RepositoryName: "test-repo",
		Timestamp:      time.Now().UTC(),
		Payload: leaderboardscoring.PullRequestOpenedPayload{
			UserID:       userID,
			PrID:         1,
			PrNumber:     1,
			Title:        "Test PR",
			BranchName:   "feature/test",
			TargetBranch: "main",
			Labels:       []string{"feature"},
			Assignees:    []uint64{userID},
		},
	}
	err := service.ProcessScoreEvent(ctx, eventReq)
	suite.NoError(err)

	leaderboardKey := "leaderboard:1001:all_time"
	score, err := suite.redisClient.ZScore(ctx, leaderboardKey, strconv.FormatUint(userID, 10)).Result()
	suite.NoError(err)
	suite.Equal(float64(1), score, "Score should be 1 for PullRequestOpened")

	// Verify event was published to mock NATS
	suite.True(suite.mockPublisher.PublishCalled)
}

func (suite *IntegrationTestSuite) TestProcessScoreEvent_InvalidEvent() {
	ctx := context.Background()

	service := leaderboardscoring.NewService(
		suite.persistence,
		suite.leaderboard,
		suite.mockPublisher,
		"processed_events",
		leaderboardscoring.NewValidator(),
	)

	// Missing required fields
	invalidEvent := &leaderboardscoring.EventRequest{
		ID:        "",
		EventName: leaderboardscoring.PullRequestOpened.String(),
	}

	err := service.ProcessScoreEvent(ctx, invalidEvent)
	suite.Error(err)
}

// Multiple events increase score in Redis
func (suite *IntegrationTestSuite) TestMultipleEvents_IncreaseScore() {
	ctx := context.Background()

	err := suite.redisClient.FlushAll(ctx).Err()
	suite.NoError(err)

	time.Sleep(500 * time.Millisecond)

	keys, _ := suite.redisClient.Keys(ctx, "*").Result()
	fmt.Printf("Redis keys after flush: %v\n", keys)
	require.Len(suite.T(), keys, 0, "Redis should be empty before test")

	service := leaderboardscoring.NewService(
		suite.persistence,
		suite.leaderboard,
		suite.mockPublisher,
		"processed_events",
		leaderboardscoring.NewValidator(),
	)

	userID := uint64(456)
	userIDStr := strconv.FormatUint(userID, 10)

	// Send multiple events
	events := []struct {
		eventName     leaderboardscoring.EventName
		payload       leaderboardscoring.EventPayload
		expectedDelta int64
	}{
		{
			leaderboardscoring.PullRequestOpened,
			leaderboardscoring.PullRequestOpenedPayload{UserID: userID},
			1,
		},
		{
			leaderboardscoring.PullRequestClosed,
			leaderboardscoring.PullRequestClosedPayload{UserID: userID},
			2,
		},
		{
			leaderboardscoring.IssueOpened,
			leaderboardscoring.IssueOpenedPayload{UserID: userID},
			4,
		},
	}

	for _, evt := range events {
		eventReq := &leaderboardscoring.EventRequest{
			ID:             uuid.New().String(),
			UserID:         userIDStr,
			EventName:      evt.eventName.String(),
			RepositoryID:   1001,
			RepositoryName: "test-repo",
			Timestamp:      time.Now().UTC(),
			Payload:        evt.payload,
		}

		err := service.ProcessScoreEvent(ctx, eventReq)
		suite.NoError(err)
	}

	// Verify total score in Redis : 1 + 2 + 4 = 7
	leaderboardKey := "leaderboard:1001:all_time"
	score, err := suite.redisClient.ZScore(ctx, leaderboardKey, userIDStr).Result()

	suite.NoError(err)
	suite.Equal(float64(7), score)

}

// Leaderboard ranking order in Redis
func (suite *IntegrationTestSuite) TestLeaderboardRanking_Real() {
	ctx := context.Background()

	service := leaderboardscoring.NewService(
		suite.persistence,
		suite.leaderboard,
		suite.mockPublisher,
		"processed_events",
		leaderboardscoring.NewValidator(),
	)

	// Create users with different scores
	users := []struct {
		id    uint64
		score int64
	}{
		{100, 5},
		{200, 10},
		{300, 3},
		{400, 15},
	}

	for _, user := range users {
		for i := int64(0); i < user.score; i++ {
			eventReq := &leaderboardscoring.EventRequest{
				ID:             uuid.New().String(),
				UserID:         strconv.FormatUint(user.id, 10),
				EventName:      leaderboardscoring.PullRequestOpened.String(),
				RepositoryID:   1001,
				RepositoryName: "test-repo",
				Timestamp:      time.Now().UTC(),
				Payload: leaderboardscoring.PullRequestOpenedPayload{
					UserID:       user.id,
					PrID:         uint64(i),
					PrNumber:     int32(i),
					Title:        "Test",
					BranchName:   "main",
					TargetBranch: "main",
					Labels:       nil,
					Assignees:    nil,
				},
			}
			_ = service.ProcessScoreEvent(ctx, eventReq)
		}
	}

	// Get leaderboard from Redis
	leaderboardKey := "leaderboard:1001:all_time"
	rankings, err := suite.redisClient.ZRevRangeWithScores(ctx, leaderboardKey, 0, -1).Result()
	suite.NoError(err)

	// Verify ranking order (highest score first)
	expectedOrder := []uint64{400, 200, 100, 300}
	suite.Len(rankings, 4)

	for i, expected := range expectedOrder {
		suite.Equal(strconv.FormatUint(expected, 10), rankings[i].Member)
	}

	// Verify scores are descending
	for i := 1; i < len(rankings); i++ {
		suite.GreaterOrEqual(rankings[i-1].Score, rankings[i].Score)
	}
}

// Persistence to PostgreSQL
func (suite *IntegrationTestSuite) TestEventPersistence_Real() {
	ctx := context.Background()

	// Create events directly in PostgreSQL
	events := []leaderboardscoring.ProcessedScoreEvent{
		{
			ID:        1,
			UserID:    "user1",
			EventName: leaderboardscoring.PullRequestOpened,
			Score:     10,
			Timestamp: time.Now().UTC(),
		},
		{
			ID:        2,
			UserID:    "user2",
			EventName: leaderboardscoring.PullRequestClosed,
			Score:     20,
			Timestamp: time.Now().UTC(),
		},
	}

	err := suite.persistence.AddProcessedScoreEvents(ctx, events)
	suite.NoError(err)

	// Verify data was persisted in real PostgreSQL
	var count int
	err = suite.postgresConn.Pool.QueryRow(
		ctx,
		"SELECT COUNT(*) FROM processed_score_events WHERE user_id IN ('user1', 'user2')",
	).Scan(&count)
	suite.NoError(err)
	suite.Equal(2, count)

	// Verify specific event
	var eventType string
	var score int64
	err = suite.postgresConn.Pool.QueryRow(
		ctx,
		"SELECT event_type, score_delta FROM processed_score_events WHERE user_id = 'user1' LIMIT 1",
	).Scan(&eventType, &score)
	suite.NoError(err)
	suite.Equal(leaderboardscoring.PullRequestOpened.String(), eventType)
	suite.Equal(int64(10), score)
}

// Get leaderboard through service (reads from real Redis)
func (suite *IntegrationTestSuite) TestGetLeaderboard_Real() {
	ctx := context.Background()

	service := leaderboardscoring.NewService(
		suite.persistence,
		suite.leaderboard,
		suite.mockPublisher,
		"processed_events",
		leaderboardscoring.NewValidator(),
	)

	// Add users to Redis leaderboard
	userScores := map[string]int64{
		"mohammad": 100,
		"reza":     75,
		"mahdi":    90,
	}

	leaderboardKey := "leaderboard:1001:all_time"
	for user, score := range userScores {
		err := suite.redisClient.ZAdd(ctx, leaderboardKey, redis.Z{Score: float64(score), Member: user}).Err()
		suite.NoError(err)
	}

	var projectID = "1001"
	// Get leaderboard through service
	req := &leaderboardscoring.GetLeaderboardRequest{
		Timeframe: leaderboardscoring.AllTime.String(),
		ProjectID: &projectID,
		PageSize:  10,
		Offset:    0,
	}

	resp, err := service.GetLeaderboard(ctx, req)
	suite.NoError(err)

	// Verify results
	suite.Len(resp.LeaderboardRows, 3)
	suite.Equal("mohammad", resp.LeaderboardRows[0].UserID)
	suite.Equal(int64(100), resp.LeaderboardRows[0].Score)
}

// Concurrent events from multiple users
func (suite *IntegrationTestSuite) TestConcurrentEvents() {
	ctx := context.Background()

	service := leaderboardscoring.NewService(
		suite.persistence,
		suite.leaderboard,
		suite.mockPublisher,
		"processed_events",
		leaderboardscoring.NewValidator(),
	)

	// Simulate concurrent requests from different users
	done := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func(userID int) {
			eventReq := &leaderboardscoring.EventRequest{
				ID:             uuid.New().String(),
				UserID:         strconv.Itoa(userID),
				EventName:      leaderboardscoring.PullRequestOpened.String(),
				RepositoryID:   1001,
				RepositoryName: "test-repo",
				Timestamp:      time.Now().UTC(),
				Payload: leaderboardscoring.PullRequestOpenedPayload{
					UserID:       uint64(userID),
					PrID:         1,
					PrNumber:     1,
					Title:        "Concurrent",
					BranchName:   "main",
					TargetBranch: "main",
				},
			}

			err := service.ProcessScoreEvent(ctx, eventReq)
			done <- err
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		err := <-done
		suite.NoError(err)
	}

	// Verify all users are in Redis
	leaderboardKey := "leaderboard:1001:all_time"
	size, err := suite.redisClient.ZCard(ctx, leaderboardKey).Result()
	suite.NoError(err)
	suite.Equal(int64(10), size)
}

// Pagination works with real Redis
func (suite *IntegrationTestSuite) TestPagination_Real() {
	ctx := context.Background()

	service := leaderboardscoring.NewService(
		suite.persistence,
		suite.leaderboard,
		suite.mockPublisher,
		"processed_events",
		leaderboardscoring.NewValidator(),
	)

	// Add 25 users to Redis
	leaderboardKey := "leaderboard:1001:all_time"
	for i := 0; i < 25; i++ {
		user := fmt.Sprintf("user%d", i)
		score := float64(100 - i)
		suite.redisClient.ZAdd(ctx, leaderboardKey, redis.Z{Score: score, Member: user})
	}

	var projectID = "1001"

	// Get first page (10 items)
	req1 := &leaderboardscoring.GetLeaderboardRequest{
		Timeframe: leaderboardscoring.AllTime.String(),
		ProjectID: &projectID,
		PageSize:  10,
		Offset:    0,
	}
	resp1, err := service.GetLeaderboard(ctx, req1)
	suite.NoError(err)
	suite.Len(resp1.LeaderboardRows, 10)

	// Get second page (10 items)
	req2 := &leaderboardscoring.GetLeaderboardRequest{
		Timeframe: leaderboardscoring.AllTime.String(),
		ProjectID: &projectID,
		PageSize:  10,
		Offset:    10,
	}
	resp2, err := service.GetLeaderboard(ctx, req2)
	suite.NoError(err)
	suite.Len(resp2.LeaderboardRows, 10)

	// Verify different users in each page
	suite.NotEqual(resp1.LeaderboardRows[0].UserID, resp2.LeaderboardRows[0].UserID)
}

func (suite *IntegrationTestSuite) TestGetLeaderboard_InvalidOffset() {
	ctx := context.Background()

	service := leaderboardscoring.NewService(
		suite.persistence,
		suite.leaderboard,
		suite.mockPublisher,
		"processed_events",
		leaderboardscoring.NewValidator(),
	)

	var projectID = "1001"
	req := &leaderboardscoring.GetLeaderboardRequest{
		Timeframe: leaderboardscoring.Monthly.String(),
		ProjectID: &projectID,
		PageSize:  10,
		Offset:    -1,
	}

	_, err := service.GetLeaderboard(ctx, req)
	suite.Error(err)
}

// Mock implementations
type MockNATSPublisher struct {
	PublishCalled bool
	publishedData [][]byte
	lastSubject   string
}

func (m *MockNATSPublisher) Publish(ctx context.Context, subject string, data []byte) error {
	m.PublishCalled = true
	m.lastSubject = subject
	m.publishedData = append(m.publishedData, data)
	return nil
}

func (m *MockNATSPublisher) Reset() {
	m.PublishCalled = false
	m.publishedData = nil
	m.lastSubject = ""
}
