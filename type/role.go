package types

type Role string

var (
	Admin       Role = "admin"
	Contributor Role = "contributor"
)

func (r Role) String() string {
	switch r {
	case Admin:
		return "admin"
	case Contributor:
		return "contributor"
	}

	return ""
}
