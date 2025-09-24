package consistency

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/webhookapp/repository/lostevent"
	"github.com/gocasters/rankr/webhookapp/repository/rawevent"
	"github.com/gocasters/rankr/webhookapp/service"
	"log"
	"time"
)

type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

const (
	maxPages = 100
	perPage  = 100
	apiDelay = 100 * time.Millisecond
)

type Service struct {
	rawEventRepo  *rawevent.RawWebhookRepository
	lostEventRepo *lostevent.LostWebhookRepository
	//logger        *slog.Logger
	client *GitHubClient
}

func New(rawEventRepo *rawevent.RawWebhookRepository,
	lostEventRepo *lostevent.LostWebhookRepository,
	client *GitHubClient) *Service {
	return &Service{
		rawEventRepo:  rawEventRepo,
		lostEventRepo: lostEventRepo,
		client:        client,
	}
}

// FindLostEvents find the lost events and save all lost event delivery IDs.
func (s *Service) FindLostEvents(ctx context.Context, provider int32, owner, repoName string, hookID int64, startTime, endTime time.Time) error {
	//get events from github api
	timeRange := TimeRange{Start: startTime, End: endTime}
	allEvents, fErr := s.fetchGitHubEvents(ctx, owner, repoName, hookID, timeRange)
	if fErr != nil {
		return fErr
	}
	deliveryIDs := make([]string, 0, len(allEvents))
	for _, gEvent := range allEvents {
		deliveryIDs = append(deliveryIDs, gEvent.Guid)
	}

	//get stored events delivery ids
	existingDeliveryIDs, dErr := s.rawEventRepo.FindExistingDeliveryIDs(ctx, deliveryIDs)
	if dErr != nil {
		return dErr
	}

	//detect lost events
	lostEventIDs := make([]string, 0, len(allEvents))
	for _, gEvent := range allEvents {
		if _, exists := existingDeliveryIDs[gEvent.Guid]; !exists {
			lostEventIDs = append(lostEventIDs, gEvent.Guid)
		}
	}

	//save lost event IDs
	err := s.lostEventRepo.SaveBatch(ctx, provider, lostEventIDs)
	if err != nil {
		return err
	}

	return nil
}

// fetchGitHubEvents fetches events from GitHub API within the specified time range.
func (s *Service) fetchGitHubEvents(ctx context.Context, owner, repo string, hookID int64, timeRange TimeRange) ([]service.DeliveryEvent, error) {
	var allEvents []service.DeliveryEvent
	var hitPageLimit bool

	for page := 1; page <= maxPages; page++ {
		pageEvents, hasMore, err := s.client.GetRepositoryWebhookEvents(ctx, owner, repo, hookID, page, perPage)
		if err != nil {
			// Return partial results if we have any, otherwise return the error
			if len(allEvents) > 0 {
				log.Printf("Warning: Error on page %d for repository %s/%s, returning partial results: %v", page, owner, repo, err)
				return allEvents, nil
			}
			return nil, fmt.Errorf("failed to fetch events on page %d: %w", page, err)
		}

		if len(pageEvents) == 0 {
			break
		}

		filteredEvents, shouldStop := s.filterEventsByTimeRange(pageEvents, timeRange)
		allEvents = append(allEvents, filteredEvents...)

		if shouldStop || !hasMore {
			break
		}

		// Check if we're about to hit the page limit
		if page == maxPages {
			hitPageLimit = true
			break
		}

		// Rate limit protection
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(apiDelay):
		}
	}

	if hitPageLimit {
		log.Printf("Warning: Hit page limit (%d) for repository %s/%s", maxPages, owner, repo)
	}

	return allEvents, nil
}

// filterEventsByTimeRange filters events by time range and returns whether to stop pagination.
func (s *Service) filterEventsByTimeRange(events []service.DeliveryEvent, timeRange TimeRange) ([]service.DeliveryEvent, bool) {
	var filtered []service.DeliveryEvent

	for _, event := range events {
		if event.DeliveredAt.Before(timeRange.Start) {
			// Events are ordered by created_at desc, so we can stop here
			return filtered, true
		}

		if event.DeliveredAt.After(timeRange.End) {
			continue
		}

		filtered = append(filtered, event)
	}

	return filtered, false
}

// RedeliverLostEvents gets a deliveryID and redeliver it
func (s *Service) RedeliverLostEvents(ctx context.Context, provider int32, owner, repoName string, hookID int64) map[string]error {
	errorCollections := make(map[string]error)

	storedLostIDs, err := s.lostEventRepo.GetAllDeliveryIDs(ctx, provider)
	if err != nil {
		errorCollections["db_err"] = err
		return errorCollections
	}

	for _, storedLostID := range storedLostIDs {
		if rErr := s.client.RedeliverLostEvent(ctx, owner, repoName, hookID, storedLostID); rErr != nil {
			errorCollections[storedLostID] = rErr
		}
		// remove from db the id
		dErr := s.lostEventRepo.DeleteByID(ctx, provider, storedLostID)
		if dErr != nil {
			errorCollections[storedLostID] = dErr
		}

	}
	return errorCollections
}
