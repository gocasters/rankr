package task

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskJSONHidesLinkedPR(t *testing.T) {
	without := Task{ID: 1}
	with := without
	linkedPR := int64(7)
	with.LinkedPR = &linkedPR

	withoutJSON, err := json.Marshal(without)
	require.NoError(t, err)

	withJSON, err := json.Marshal(with)
	require.NoError(t, err)

	assert.JSONEq(t, string(withoutJSON), string(withJSON))
	assert.NotContains(t, string(withJSON), "linked_pr")
	assert.NotContains(t, string(withJSON), "LinkedPR")
}
