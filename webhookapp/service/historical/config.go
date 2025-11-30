package historical

type Config struct {
	Owner          string
	Repo           string
	Token          string
	EventTypes     []string
	BatchSize      int
	IncludeReviews bool
}
