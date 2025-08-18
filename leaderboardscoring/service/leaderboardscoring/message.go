package leaderboardscoring

import "errors"

var (
	ErrFailedToUpdateScores        = errors.New("failed to update scores in redis")
	ErrFailedToPersistContribution = errors.New("failed to persist contribution to database")
	ErrSuccessfullyProcessedEvent  = errors.New("successfully processed score event")
)
