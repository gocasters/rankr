package userprofile

type ValidatorUserProfileRepository interface{}

type Validator struct{}

func NewValidator(repo ValidatorUserProfileRepository) Validator {
	return Validator{}
}
