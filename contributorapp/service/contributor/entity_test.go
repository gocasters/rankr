package contributor

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContributorJSONHidesTwoFactor(t *testing.T) {
	c := Contributor{
		ID:             1,
		GitHubID:       42,
		GitHubUsername: "octocat",
		TwoFactor:      true,
	}

	data, err := json.Marshal(c)
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(data, &got))

	assert.NotContains(t, got, "two_factor_enabled")
}
