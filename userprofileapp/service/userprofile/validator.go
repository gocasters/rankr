package userprofile

type ValidatorUserProfileRepository interface{}

type Validator struct{}

func New(repo ValidatorUserProfileRepository) Validator {
	return Validator{}
}
