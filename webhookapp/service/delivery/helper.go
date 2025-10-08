package delivery

import (
	"strings"
	"time"

	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
)

func getReviewState(state string) eventpb.ReviewState {
	lowerCasedState := strings.ToLower(state)
	var reviewState eventpb.ReviewState

	switch lowerCasedState {
	case "approved":
		reviewState = eventpb.ReviewState_REVIEW_STATE_APPROVED
	case "changes_requested":
		reviewState = eventpb.ReviewState_REVIEW_STATE_CHANGES_REQUESTED
	case "commented":
		reviewState = eventpb.ReviewState_REVIEW_STATE_COMMENTED
	default:
		reviewState = eventpb.ReviewState_REVIEW_STATE_UNSPECIFIED
	}

	return reviewState
}

func parseTime(stringTime string) time.Time {
	if stringTime == "" {
		return time.Time{}
	}
	//layout := "2019-11-17T17:43:43Z" // time.RFC3339
	parseTimed, err := time.Parse(time.RFC3339, stringTime)
	if err != nil {
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
			return eventpb.PrCloseReason_PR_CLOSE_REASON_MERGED
		}
	}
	return eventpb.PrCloseReason_PR_CLOSE_REASON_CLOSED_WITHOUT_MERGE
}

func containsCode(text string) bool {
	return strings.Contains(text, "```") || strings.Contains(text, "`")
}
