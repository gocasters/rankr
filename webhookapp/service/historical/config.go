package historical

type Config struct {
	Owner          string
	Repo           string
	Token          string
	EventTypes     []string
	DryRun         bool
	BatchSize      int
	IncludeReviews bool
}