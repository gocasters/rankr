package service

import (
	"encoding/json"
	"fmt"
	"github.com/gocasters/rankr/protobuf/golang/eventpb"
	"strings"
	"time"
)

func extractWebhookAction(body []byte) (string, error) {
	var actionData struct {
		Action string `json:"action"`
	}
	if err := json.Unmarshal(body, &actionData); err != nil {
		return "", err
	}
	if actionData.Action == "" {
		return "", fmt.Errorf("missing action field")
	}
	return actionData.Action, nil
}

func validateGitHubHeaders(hookID, eventName, deliveryUID string) error {
	if hookID == "" {
		return fmt.Errorf("missing X-GitHub-Hook-ID header")
	}
	if eventName == "" {
		return fmt.Errorf("missing X-GitHub-Event header")
	}
	if deliveryUID == "" {
		return fmt.Errorf("missing X-GitHub-Delivery header")
	}
	return nil
}

func getReviewState(state string) eventpb.ReviewState {
	lowerCasedState := strings.ToLower(state)
	var reviewState eventpb.ReviewState

	switch lowerCasedState {
	case "approved":
		reviewState = eventpb.ReviewState_APPROVED
	case "changes_requested":
		reviewState = eventpb.ReviewState_CHANGES_REQUESTED
	case "commented":
		reviewState = eventpb.ReviewState_COMMENTED
	default:
		reviewState = eventpb.ReviewState_REVIEW_STATE_UNSPECIFIED
	}

	return reviewState
}

func parseTime(stringTime string) time.Time {
	layout := "2019-11-17T17:43:43Z"
	parseTimed, err := time.Parse(layout, stringTime)
	if err != nil {
		fmt.Println("Error parsing time:", err)
		return time.Time{}
	}
	return parseTimed
}

func extractLabelsNames(labels []Label) []string {
	numberOfLabels := len(labels)
	var labelsNames = make([]string, 0, numberOfLabels)

	for _, label := range labels {
		// Append the Name field of each object to the fileNames slice
		labelsNames = append(labelsNames, label.Name)
	}

	return labelsNames
}

func extractAssigneesIDs(assignees []*User) []uint64 {
	numberOfAssignees := len(assignees)
	var assigneesIDs = make([]uint64, 0, numberOfAssignees)

	for _, assignee := range assignees {
		// Append the Name field of each object to the fileNames slice
		assigneesIDs = append(assigneesIDs, assignee.ID)
	}

	return assigneesIDs
}

func determineCloseReason(merged *bool) eventpb.PrCloseReason {
	if merged != nil {
		if *merged {
			return eventpb.PrCloseReason_MERGED
		}
	}
	return eventpb.PrCloseReason_CLOSED_WITHOUT_MERGE
}

func containsCode(text string) bool {
	return strings.Contains(text, "```") || strings.Contains(text, "`")
}
