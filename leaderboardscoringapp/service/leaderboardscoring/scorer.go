package leaderboardscoring

type Scorer interface {
	Score(req *EventRequest) int64
}

type legacyScorer struct{}

func NewLegacyScorer() Scorer {
	return legacyScorer{}
}

func (legacyScorer) Score(req *EventRequest) int64 {
	return calculateScore(req.EventName)
}

const mergedPullRequestPoints = 10

type rulesV1Scorer struct {
	excluded map[uint64]struct{}
}

func NewRulesV1Scorer(excludedAccountIDs []uint64) Scorer {
	excluded := make(map[uint64]struct{}, len(excludedAccountIDs))
	for _, id := range excludedAccountIDs {
		excluded[id] = struct{}{}
	}
	return rulesV1Scorer{excluded: excluded}
}

func (s rulesV1Scorer) Score(req *EventRequest) int64 {
	pr, ok := req.Payload.(PullRequestClosedPayload)
	if !ok || !pr.Merged {
		return 0
	}
	if _, isExcluded := s.excluded[pr.UserID]; isExcluded {
		return 0
	}
	return mergedPullRequestPoints
}
