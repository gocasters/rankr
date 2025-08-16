package githubwebhook

const (
	TopicGithubUserActivity = "github.user.activity"
)

type ActivityEvent struct {
	HookID   string `json:"hook_id"`
	Event    string `json:"event"`
	Delivery string `json:"delivery"`
	Body     []byte `json:"body"`
}
