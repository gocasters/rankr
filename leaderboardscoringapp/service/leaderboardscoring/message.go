package leaderboardscoring

import "errors"

var (
	ErrFailedToUpdateScores        = errors.New("failed to update scores in redis")
	ErrFailedToPersistContribution = errors.New("failed to persist contribution to database")
	ErrNotImplemented              = errors.New("repository method not implemented")
)

const MsgSuccessfullyProcessedEvent = "successfully processed score event"
