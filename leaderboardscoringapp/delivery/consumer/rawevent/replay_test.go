package rawevent

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"maps"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/gocasters/rankr/pkg/logger"
	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	eventsFile      = "testdata/events.json"
	goldenFile      = "testdata/score_events.golden.json"
	processedTopic  = "test.processed"
	goldenFileMode  = 0o644
	testLogMaxSize  = 1
	testLogMaxAgeDy = 1
)

var update = flag.Bool("update", false, "rewrite golden files in testdata")

var errStoreDown = errors.New("store down")

func TestMain(m *testing.M) {
	if err := logger.Init(logger.Config{
		Level:            "error",
		FileMaxSizeInMB:  testLogMaxSize,
		FileMaxAgeInDays: testLogMaxAgeDy,
	}); err != nil {
		panic(err)
	}
	code := m.Run()
	_ = logger.Close()
	os.Exit(code)
}

type memoryStore struct {
	values    map[string]string
	existsErr error
}

func newMemoryStore() *memoryStore {
	return &memoryStore{values: map[string]string{}}
}

func (m *memoryStore) Exists(_ context.Context, key string) (bool, error) {
	if m.existsErr != nil {
		return false, m.existsErr
	}
	_, ok := m.values[key]
	return ok, nil
}

func (m *memoryStore) SetNX(_ context.Context, key, value string, _ time.Duration) (bool, error) {
	if _, ok := m.values[key]; ok {
		return false, nil
	}
	m.values[key] = value
	return true, nil
}

func (m *memoryStore) Set(_ context.Context, key, value string, _ time.Duration) error {
	m.values[key] = value
	return nil
}

func (m *memoryStore) DeleteIfEqual(_ context.Context, key, value string) error {
	if m.values[key] == value {
		delete(m.values, key)
	}
	return nil
}

type memoryLeaderboard struct {
	scores map[string]map[string]int64
}

func (m *memoryLeaderboard) UpsertScores(_ context.Context, score *leaderboardscoring.UpsertScore, _ leaderboardscoring.Timeframe) error {
	for _, key := range score.Keys {
		if m.scores[key] == nil {
			m.scores[key] = map[string]int64{}
		}
		m.scores[key][score.UserID] += score.Score
	}
	return nil
}

func (m *memoryLeaderboard) GetLeaderboard(context.Context, *leaderboardscoring.LeaderboardQuery) (leaderboardscoring.LeaderboardQueryResult, error) {
	return leaderboardscoring.LeaderboardQueryResult{}, nil
}

func (m *memoryLeaderboard) allTime() map[string]map[string]int64 {
	suffix := ":" + leaderboardscoring.AllTime.String()
	out := map[string]map[string]int64{}
	for key, users := range m.scores {
		if !strings.HasSuffix(key, suffix) {
			continue
		}
		out[key] = maps.Clone(users)
	}
	return out
}

type recordingPublisher struct {
	messages [][]byte
}

func (p *recordingPublisher) Publish(_ context.Context, _ string, data []byte) error {
	p.messages = append(p.messages, data)
	return nil
}

type pipeline struct {
	handler     Handler
	store       *memoryStore
	leaderboard *memoryLeaderboard
	publisher   *recordingPublisher
}

func newPipeline() *pipeline {
	p := &pipeline{
		store:       newMemoryStore(),
		leaderboard: &memoryLeaderboard{scores: map[string]map[string]int64{}},
		publisher:   &recordingPublisher{},
	}
	svc := leaderboardscoring.NewService(nil, p.leaderboard, p.publisher, processedTopic, leaderboardscoring.NewValidator())
	checker := &IdempotencyChecker{
		store:  p.store,
		config: Config{PrefixProcessedKey: "processed_events", PrefixLockKey: "lock"},
	}
	p.handler = NewHandler(svc, checker)
	return p
}

type scoreEvent struct {
	EventID   string `json:"event_id"`
	EventName string `json:"event_name"`
	UserID    string `json:"user_id"`
	Score     int64  `json:"score"`
}

type goldenResult struct {
	ScoreEvents        []scoreEvent                `json:"score_events"`
	AllTimeLeaderboard map[string]map[string]int64 `json:"all_time_leaderboard"`
}

type recordedEvent struct {
	id      string
	payload []byte
}

func loadEvents(t *testing.T) []recordedEvent {
	t.Helper()
	raw, err := os.ReadFile(eventsFile)
	require.NoError(t, err)

	var items []json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &items))

	events := make([]recordedEvent, 0, len(items))
	for _, item := range items {
		var ev eventpb.Event
		require.NoError(t, protojson.Unmarshal(item, &ev))
		payload, err := proto.Marshal(&ev)
		require.NoError(t, err)
		events = append(events, recordedEvent{id: ev.GetId(), payload: payload})
	}
	return events
}

func (p *pipeline) replay(t *testing.T, events []recordedEvent) []scoreEvent {
	t.Helper()
	scored := make([]scoreEvent, 0, len(events))
	for _, ev := range events {
		before := len(p.publisher.messages)
		require.NoError(t, p.handler.HandleEvent(message.NewMessage(ev.id, ev.payload)))
		if len(p.publisher.messages) == before {
			continue
		}
		var pse leaderboardscoring.ProcessedScoreEvent
		require.NoError(t, json.Unmarshal(p.publisher.messages[len(p.publisher.messages)-1], &pse))
		scored = append(scored, scoreEvent{
			EventID:   ev.id,
			EventName: string(pse.EventName),
			UserID:    pse.UserID,
			Score:     pse.Score,
		})
	}
	return scored
}

func TestScoreEventsGolden(t *testing.T) {
	p := newPipeline()
	got := goldenResult{
		ScoreEvents:        p.replay(t, loadEvents(t)),
		AllTimeLeaderboard: p.leaderboard.allTime(),
	}

	encoded, err := json.MarshalIndent(got, "", "  ")
	require.NoError(t, err)
	encoded = append(encoded, '\n')

	if *update {
		require.NoError(t, os.WriteFile(goldenFile, encoded, goldenFileMode))
	}

	want, err := os.ReadFile(goldenFile)
	require.NoError(t, err, "run go test -run TestScoreEventsGolden -update to create the golden file")
	require.Equal(t, string(want), string(encoded))
}

func TestReplayTwiceYieldsSameLeaderboard(t *testing.T) {
	p := newPipeline()
	events := loadEvents(t)

	first := p.replay(t, events)
	require.NotEmpty(t, first)
	afterFirst := p.leaderboard.allTime()
	publishedFirst := len(p.publisher.messages)

	second := p.replay(t, events)

	require.Empty(t, second)
	require.Equal(t, afterFirst, p.leaderboard.allTime())
	require.Equal(t, publishedFirst, len(p.publisher.messages))
}

func TestReplayDuplicateEventIDScoresOnce(t *testing.T) {
	p := newPipeline()
	events := loadEvents(t)

	seen := map[string]int{}
	for _, ev := range p.replay(t, events) {
		seen[ev.EventID]++
	}

	for id, count := range seen {
		require.Equalf(t, 1, count, "event %s scored %d times", id, count)
	}
	require.Less(t, len(seen), len(events))
}

func TestReplayStoreErrorIsNackedAndNotScored(t *testing.T) {
	p := newPipeline()
	p.store.existsErr = errStoreDown
	ev := loadEvents(t)[0]

	err := p.handler.HandleEvent(message.NewMessage(ev.id, ev.payload))

	require.ErrorIs(t, err, errStoreDown)
	require.Empty(t, p.publisher.messages)
	require.Empty(t, p.leaderboard.scores)
}

func TestReplayLockedEventIsAckedAndNotScored(t *testing.T) {
	p := newPipeline()
	ev := loadEvents(t)[0]
	p.store.values[p.handler.idempotencyChecker.lockKey(ev.id)] = "other-worker"

	err := p.handler.HandleEvent(message.NewMessage(ev.id, ev.payload))

	require.NoError(t, err)
	require.Empty(t, p.publisher.messages)
	require.Empty(t, p.leaderboard.scores)
}
