package topicsname

const (
	// NATS Topics
	TopicContributorCreated = "contributor.created"
	TopicContributorUpdated = "contributor.updated"
	TopicTaskCreated        = "task.created"
	TopicTaskUpdated        = "task.updated"
	TopicTaskCompleted      = "task.completed"
	TopicLeaderboardScored  = "leaderboard.scored"
	TopicLeaderboardUpdated = "leaderboard.updated"
	TopicProjectCreated     = "project.created"
	TopicProjectUpdated     = "project.updated"

	// WebSocket Message Types
	MessageTypeEvent       = "event"
	MessageTypeSubscribe   = "subscribe"
	MessageTypeUnsubscribe = "unsubscribe"
	MessageTypeError       = "error"
	MessageTypeAck         = "ack"
)

const (
	StreamNameRawEvents                         = "rankr_raw_events"
	StreamNameLeaderboardscoringProcessedEvents = "leaderboardscoring_processed_events"

	TopicProcessedScoreEvents    = "leaderboardscoring.processed.score.events"
	TopicProcessedScoreEventsDLQ = "leaderboardscoring.processed.score.events.dlq"
)
