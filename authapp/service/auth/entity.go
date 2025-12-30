package auth

type LoginRequest struct {
	ContributorName string `json:"contributor_name"`
	Password        string `json:"password"`
}
