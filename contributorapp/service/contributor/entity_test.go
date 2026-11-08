package contributor

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContributorJSONHidesTwoFactor(t *testing.T) {
	without := Contributor{
		ID:             1,
		GitHubID:       42,
		GitHubUsername: "octocat",
	}
	with := without
	with.TwoFactor = true

	withoutJSON, err := json.Marshal(without)
	require.NoError(t, err)

	withJSON, err := json.Marshal(with)
	require.NoError(t, err)

	assert.JSONEq(t, string(withoutJSON), string(withJSON))
	assert.NotContains(t, string(withJSON), "two_factor_enabled")
	assert.NotContains(t, string(withJSON), "TwoFactor")
}
