package contributor

type ValidatorContributorRepository interface {
}

type Validator struct {
	repo ValidatorContributorRepository
}

func NewValidator(repo ValidatorContributorRepository) Validator {
	return Validator{repo: repo}
}
