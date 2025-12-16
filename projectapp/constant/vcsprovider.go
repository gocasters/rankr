package constant

type VcsProvider string

const (
	VcsProviderGitHub    VcsProvider = "GITHUB"
	VcsProviderGitLab    VcsProvider = "GITLAB"
	VcsProviderBitbucket VcsProvider = "BITBUCKET"
)

var validVcsProviders = map[VcsProvider]struct{}{
	VcsProviderGitHub:    {},
	VcsProviderGitLab:    {},
	VcsProviderBitbucket: {},
}

func IsValidVcsProvider(p string) bool {
	_, ok := validVcsProviders[VcsProvider(p)]
	return ok
}
