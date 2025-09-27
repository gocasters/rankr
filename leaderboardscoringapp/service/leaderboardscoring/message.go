package leaderboardscoring

import "errors"

var (
	ErrInvalidEventRequest  = errors.New("invalid event request for upsert score")
	ErrFailedToUpdateScores = errors.New("failed to update scores in redis")
	ErrInvalidArguments     = errors.New("invalid arguments provided for the request")
	ErrNotImplemented       = errors.New("repository method not implemented")
	ErrLeaderboardNotFound  = errors.New("leaderboard data not found for the given criteria")
)

const MsgSuccessfullyProcessedEvent = "successfully processed score event"
