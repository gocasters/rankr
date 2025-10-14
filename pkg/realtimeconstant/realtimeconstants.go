package realtimeconstant

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
