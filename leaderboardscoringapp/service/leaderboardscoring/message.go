package leaderboardscoring

import "errors"

var (
	ErrFailedToUpdateScores        = errors.New("failed to update scores in redis")
	ErrFailedToPersistContribution = errors.New("failed to persist contribution to database")
	ErrFailedToAddToRedisStream    = errors.New("failed to add event to redis stream")
)

const MsgSuccessfullyProcessedEvent = "successfully processed score event"
