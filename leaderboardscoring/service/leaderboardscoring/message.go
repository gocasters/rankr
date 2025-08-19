package leaderboardscoring

import "errors"

var (
	ErrFailedToUpdateScores        = errors.New("failed to update scores in redis")
	ErrFailedToPersistContribution = errors.New("failed to persist contribution to database")
)

const MsgSuccessfullyProcessedEvent = "successfully processed score event"
