package historical

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gocasters/rankr/adapter/webhook/github"
	"github.com/gocasters/rankr/pkg/logger"
	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
	"github.com/gocasters/rankr/webhookapp/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Fetcher struct {
	config       Config
	githubClient *github.GitHubClient
	repo         *repository.WebhookRepository
	progress     *ProgressTracker
}

func NewFetcher(cfg Config, githubClient *github.GitHubClient, db *pgxpool.Pool) *Fetcher {
	repo := repository.NewWebhookRepository(db)
	return &Fetcher{
		config:       cfg,
		githubClient: githubClient,
		repo:         &repo,
		progress:     NewProgressTracker(),
	}
}

func (f *Fetcher) Run(ctx context.Context) error {
	log := logger.L()

	log.Info("Starting historical fetch",
		"owner", f.config.Owner,
		"repo", f.config.Repo,
		"event_types", f.config.EventTypes)

	f.progress.Start()

	for _, eventType := range f.config.EventTypes {
		switch eventType {
		case "pr":
			if err := f.fetchPullRequests(ctx); err != nil {
				return fmt.Errorf("failed to fetch PRs: %w", err)
			}
		case "issue":
			log.Warn("Issue fetching not implemented yet")
		default:
			log.Warn("Unknown event type", "type", eventType)
		}
	}

	f.progress.PrintFinalReport()

	return nil
}

func (f *Fetcher) fetchPullRequests(ctx context.Context) error {
	log := logger.L()
	log.Info("Fetching pull requests from GitHub API")

	page := 1
	totalPRs := 0

	for {
		log.Info("Fetching PR page", "page", page)

		prs, hasMore, err := f.githubClient.ListPullRequests(
			f.config.Owner,
			f.config.Repo,
			f.config.Token,
			page,
			f.config.BatchSize,
		)
		if err != nil {
			return fmt.Errorf("failed to fetch PRs page %d: %w", page, err)
		}

		log.Info("Fetched PRs", "count", len(prs), "page", page)
		totalPRs += len(prs)

		for _, pr := range prs {
			if err := f.processPR(ctx, pr); err != nil {
				log.Error("Failed to process PR",
					"pr_number", pr.Number,
					"error", err)
				f.progress.RecordFailure()
			} else {
				f.progress.RecordSuccess()
			}
		}

		if !hasMore {
			break
		}

		page++
	}

	log.Info("Finished fetching PRs", "total", totalPRs)
	return nil
}

func (f *Fetcher) processPR(ctx context.Context, pr *github.PullRequest) error {
	log := logger.L()

	events, err := TransformPRToEvents(pr)
	if err != nil {
		return fmt.Errorf("failed to transform PR: %w", err)
	}

	if f.config.IncludeReviews {
		reviews, err := f.githubClient.ListPRReviews(
			f.config.Owner,
			f.config.Repo,
			pr.Number,
			f.config.Token,
		)
		if err != nil {
			log.Warn("Failed to fetch reviews, skipping",
				"pr_number", pr.Number,
				"error", err)
		} else {
			for _, review := range reviews {
				reviewEvent, err := TransformReviewToEvent(review, pr)
				if err != nil {
					log.Warn("Failed to transform review", "error", err)
					continue
				}
				events = append(events, reviewEvent)
			}
		}
	}

	inputs := make([]repository.HistoricalEventInput, 0, len(events))
	for _, event := range events {
		resourceType := "pull_request"
		resourceID := int64(pr.Number)

		if event.EventName == eventpb.EventName_EVENT_NAME_PULL_REQUEST_REVIEW_SUBMITTED {
			resourceType = "pull_request_review"
			reviewID, err := extractReviewIDFromEventID(event.Id)
			if err != nil {
				return fmt.Errorf("failed to extract review ID from event %s: %w", event.Id, err)
			}
			resourceID = reviewID
		}

		inputs = append(inputs, repository.HistoricalEventInput{
			Event:        event,
			ResourceType: resourceType,
			ResourceID:   resourceID,
		})
	}

	return f.saveEventsBulk(ctx, inputs)
}

func extractReviewIDFromEventID(eventID string) (int64, error) {
	parts := strings.Split(eventID, "-")
	if len(parts) < 5 || parts[3] != "review" {
		return 0, fmt.Errorf("invalid review event ID format: %s", eventID)
	}
	return strconv.ParseInt(parts[4], 10, 64)
}

func (f *Fetcher) saveEventsBulk(ctx context.Context, inputs []repository.HistoricalEventInput) error {
	log := logger.L()

	if len(inputs) == 0 {
		return nil
	}

	result, err := f.repo.SaveHistoricalEventsBulk(ctx, inputs)
	if err != nil {
		return err
	}

	log.Debug("Bulk saved historical events",
		"inserted", result.Inserted,
		"duplicates", result.Duplicates)

	return nil
}
