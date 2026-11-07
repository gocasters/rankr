package task

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskJSONHidesLinkedPR(t *testing.T) {
	linkedPR := int64(7)
	task := Task{
		ID:       1,
		LinkedPR: &linkedPR,
	}

	data, err := json.Marshal(task)
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(data, &got))

	assert.NotContains(t, got, "linked_pr")
}
