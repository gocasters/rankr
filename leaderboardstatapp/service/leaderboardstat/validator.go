package leaderboardstat

type ValidatorLeaderboardstatRepository interface {
}
type Validator struct {
	repo ValidatorLeaderboardstatRepository
}

func NewValidator(repo ValidatorLeaderboardstatRepository) Validator {
	return Validator{repo: repo}
}
