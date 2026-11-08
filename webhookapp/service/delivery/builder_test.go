package delivery

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	payloadDir     = "testdata/github"
	goldenFile     = "testdata/events.golden.json"
	goldenFileMode = 0o644
	randomIDMarker = "random-uuid"
	testProvider   = eventpb.EventProvider_EVENT_PROVIDER_GITHUB
)

var update = flag.Bool("update", false, "rewrite golden files in testdata")

type eventBuilder func(provider eventpb.EventProvider, action string, body []byte, deliveryUID string) (*eventpb.Event, error)

var builders = map[EventType]eventBuilder{
	EventTypePullRequest:       buildPullRequestEvent,
	EventTypePullRequestReview: buildPullRequestReviewEvent,
	EventTypeIssues:            buildIssuesEvent,
	EventTypeIssueComment:      buildIssueCommentEvent,
	EventTypePush: func(provider eventpb.EventProvider, _ string, body []byte, deliveryUID string) (*eventpb.Event, error) {
		return buildPushEvent(provider, body, deliveryUID)
	},
}

func parsePayloadName(name string) (EventType, string) {
	parts := strings.Split(strings.TrimSuffix(name, ".json"), ".")
	if len(parts) == 1 {
		return EventType(parts[0]), ""
	}
	return EventType(parts[0]), parts[1]
}

func buildFromFile(t *testing.T, name string) *eventpb.Event {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(payloadDir, name))
	require.NoError(t, err)

	eventType, action := parsePayloadName(name)
	build, ok := builders[eventType]
	require.Truef(t, ok, "no builder for event type %q", eventType)

	deliveryUID := "delivery-" + strings.TrimSuffix(name, ".json")
	ev, err := build(testProvider, action, body, deliveryUID)
	require.NoError(t, err, name)
	if ev.GetId() != deliveryUID {
		ev.Id = randomIDMarker
	}
	return ev
}

func TestBuildEventGolden(t *testing.T) {
	entries, err := os.ReadDir(payloadDir)
	require.NoError(t, err)
	require.NotEmpty(t, entries)

	got := make(map[string]json.RawMessage, len(entries))
	for _, entry := range entries {
		encoded, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(buildFromFile(t, entry.Name()))
		require.NoError(t, err)
		var compact bytes.Buffer
		require.NoError(t, json.Compact(&compact, encoded))
		got[entry.Name()] = compact.Bytes()
	}

	encoded, err := json.MarshalIndent(got, "", "  ")
	require.NoError(t, err)
	encoded = append(encoded, '\n')

	if *update {
		require.NoError(t, os.WriteFile(goldenFile, encoded, goldenFileMode))
	}

	want, err := os.ReadFile(goldenFile)
	require.NoError(t, err, "run go test -run TestBuildEventGolden -update to create the golden file")
	require.Equal(t, string(want), string(encoded))
}

func TestBuildEventRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name      string
		eventType EventType
		action    string
		body      string
	}{
		{name: "malformed pull_request", eventType: EventTypePullRequest, action: "opened", body: "{"},
		{name: "unhandled pull_request action", eventType: EventTypePullRequest, action: "labeled", body: "{}"},
		{name: "unhandled review action", eventType: EventTypePullRequestReview, action: "dismissed", body: "{}"},
		{name: "unhandled issues action", eventType: EventTypeIssues, action: "edited", body: "{}"},
		{name: "issue closed without closed_at", eventType: EventTypeIssues, action: "closed", body: `{"issue":{"user":{"id":1}}}`},
		{name: "unhandled issue_comment action", eventType: EventTypeIssueComment, action: "deleted", body: "{}"},
		{name: "malformed push", eventType: EventTypePush, body: "["},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev, err := builders[tt.eventType](testProvider, tt.action, []byte(tt.body), "delivery-invalid")
			require.Error(t, err)
			require.Nil(t, ev)
		})
	}
}
