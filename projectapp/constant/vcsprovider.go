package constant

type VcsProvider string

const (
	VcsProviderGitHub    VcsProvider = "GITHUB"
	VcsProviderGitLab    VcsProvider = "GITLAB"
	VcsProviderBitbucket VcsProvider = "BITBUCKET"
)
