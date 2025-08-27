package task

type ValidatorTaskRepository interface {
}

type Validator struct {
	repo ValidatorTaskRepository
}

func NewValidator(repo ValidatorTaskRepository) Validator {
	return Validator{repo: repo}
}
