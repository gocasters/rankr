package role

// Role represents a user role value.
type Role string

const (
	Admin Role = "admin"
	User  Role = "user"
)

func Parse(value string) (Role, bool) {
	switch Role(value) {
	case Admin, User:
		return Role(value), true
	default:
		return "", false
	}
}
