package command

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/gocasters/rankr/adapter/webhook/github"
	"github.com/gocasters/rankr/pkg/config"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/path"
	"github.com/gocasters/rankr/webhookapp"
	"github.com/gocasters/rankr/webhookapp/service/historical"
	"github.com/spf13/cobra"
)

var (
	owner          string
	repo           string
	token          string
	eventTypes     []string
	batchSize      int
	includeReviews bool
)

var fetchHistoricalCmd = &cobra.Command{
	Use:   "fetch-historical",
	Short: "Fetch historical PRs/Issues from GitHub API",
	Long: `Fetch historical events (PRs, Issues) from GitHub API for repositories
that don't have webhook configured or need backfill of old data.
Example:
  go run cmd/webhook/main.go fetch-historical \
    --owner=gocasters \
    --repo=rankr \
    --token=$GITHUB_TOKEN \
    --event-types=pr`,
	Run: func(cmd *cobra.Command, args []string) {
		runFetchHistorical()
	},
}

func init() {
	fetchHistoricalCmd.Flags().StringVar(&owner, "owner", "", "GitHub repository owner (required)")
	fetchHistoricalCmd.Flags().StringVar(&repo, "repo", "", "GitHub repository name (required)")
	fetchHistoricalCmd.Flags().StringVar(&token, "token", "", "GitHub PAT (or set GITHUB_TOKEN env) (required)")
	fetchHistoricalCmd.Flags().StringSliceVar(&eventTypes, "event-types", []string{"pr"}, "Event types to fetch: pr, issue")
	fetchHistoricalCmd.Flags().IntVar(&batchSize, "batch-size", 100, "GitHub API results per page")
	fetchHistoricalCmd.Flags().BoolVar(&includeReviews, "include-reviews", false, "Fetch PR reviews (more API calls)")

	fetchHistoricalCmd.MarkFlagRequired("owner")
	fetchHistoricalCmd.MarkFlagRequired("repo")
	fetchHistoricalCmd.MarkFlagRequired("token")

	RootCmd.AddCommand(fetchHistoricalCmd)
}

func runFetchHistorical() {
	var cfg webhookapp.Config

	projectRoot, err := path.PathProjectRoot()
	if err != nil {
		log.Fatalf("Error finding project root: %v", err)
	}

	yamlPath := os.Getenv("CONFIG_PATH")
	if yamlPath == "" {
		yamlPath = filepath.Join(projectRoot, "deploy", "webhook", "development", "config.local.yml")
	}

	options := config.Options{
		Prefix:       "WEBHOOK_",
		Delimiter:    ".",
		Separator:    "__",
		YamlFilePath: yamlPath,
	}

	if cErr := config.Load(options, &cfg); cErr != nil {
		log.Fatalf("Failed to load webhook config: %v", cErr)
	}

	if err := logger.Init(cfg.Logger); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer func() {
		if err := logger.Close(); err != nil {
			log.Printf("logger close error: %v", err)
		}
	}()

	lbLogger := logger.L()

	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
		if token == "" {
			lbLogger.Error("GitHub token required: use --token flag or GITHUB_TOKEN env")
			return
		}
	}

	databaseConn, cnErr := database.Connect(cfg.PostgresDB)
	if cnErr != nil {
		lbLogger.Error("Failed to connect to database", "error", cnErr)
		return
	}
	defer databaseConn.Close()

	githubClient := github.NewGitHubClient()

	fetcherCfg := historical.Config{
		Owner:          owner,
		Repo:           repo,
		Token:          token,
		EventTypes:     eventTypes,
		BatchSize:      batchSize,
		IncludeReviews: includeReviews,
	}

	fetcher := historical.NewFetcher(fetcherCfg, githubClient, databaseConn.Pool)

	ctx := context.Background()
	if err := fetcher.Run(ctx); err != nil {
		lbLogger.Error("Fetch historical failed", "error", err)
		return
	}

	lbLogger.Info("Fetch historical completed successfully")
}
