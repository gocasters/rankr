package historical

import (
	"context"
	"fmt"

	"github.com/gocasters/rankr/adapter/webhook/github"
	"github.com/gocasters/rankr/pkg/logger"
	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
	"github.com/gocasters/rankr/webhookapp/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Fetcher struct {
	config       Config
	githubClient *github.GitHubClient
	repo         repository.WebhookRepository
	progress     *ProgressTracker
}

func NewFetcher(cfg Config, githubClient *github.GitHubClient, db *pgxpool.Pool) *Fetcher {
	return &Fetcher{
		config:       cfg,
		githubClient: githubClient,
		repo:         repository.NewWebhookRepository(db),
		progress:     NewProgressTracker(),
	}
}

func (f *Fetcher) Run(ctx context.Context) error {
	log := logger.L()

	log.Info("Starting historical fetch",
		"owner", f.config.Owner,
		"repo", f.config.Repo,
		"event_types", f.config.EventTypes,
		"dry_run", f.config.DryRun)

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

	for _, event := range events {
		if err := f.saveEvent(ctx, event, pr.Number); err != nil {
			return fmt.Errorf("failed to save event %s: %w", event.Id, err)
		}
	}

	return nil
}

func (f *Fetcher) saveEvent(ctx context.Context, event *eventpb.Event, prNumber int32) error {
	log := logger.L()

	if f.config.DryRun {
		log.Info("[DRY RUN] Would save event",
			"event_id", event.Id,
			"event_name", event.EventName,
			"pr_number", prNumber)
		return nil
	}

	if err := f.repo.SaveHistoricalEvent(ctx, event, "pull_request", int64(prNumber)); err != nil {
		if err == repository.ErrDuplicateEvent {
			log.Debug("Event already exists, skipping",
				"event_id", event.Id,
				"pr_number", prNumber)
			return nil
		}
		return err
	}

	log.Debug("Saved historical event",
		"event_id", event.Id,
		"event_name", event.EventName,
		"pr_number", prNumber)

	return nil
}
