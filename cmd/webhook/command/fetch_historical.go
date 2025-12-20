package command

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	projectadapter "github.com/gocasters/rankr/adapter/project"
	"github.com/gocasters/rankr/adapter/webhook/github"
	"github.com/gocasters/rankr/pkg/config"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/grpc"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/path"
	"github.com/gocasters/rankr/webhookapp"
	"github.com/gocasters/rankr/webhookapp/service/historical"
	nc "github.com/nats-io/nats.go"
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
that don't have webhook configured or need backfill of old data.`,
	Run: func(cmd *cobra.Command, args []string) {
		runFetchHistorical()
	},
	Example: "go run cmd/webhook/main.go fetch-historical --owner=gocasters --repo=rankr --token=$GITHUB_TOKEN --event-types=pr",
}

func init() {
	fetchHistoricalCmd.Flags().StringVar(&owner, "owner", "", "GitHub repository owner (required)")
	fetchHistoricalCmd.Flags().StringVar(&repo, "repo", "", "GitHub repository name (required)")
	fetchHistoricalCmd.Flags().StringVar(&token, "token", "", "GitHub PAT (required)")

	fetchHistoricalCmd.Flags().StringSliceVar(&eventTypes, "event-types", []string{"pr"}, "Event types to fetch: pr, issue")
	fetchHistoricalCmd.Flags().IntVar(&batchSize, "batch-size", 100, "GitHub API results per page")
	fetchHistoricalCmd.Flags().BoolVar(&includeReviews, "include-reviews", true, "Fetch PR reviews (more API calls)")

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

	publisher, err := nats.NewPublisher(
		nats.PublisherConfig{
			URL: cfg.NATSConfig.URL,
			JetStream: nats.JetStreamConfig{
				Disabled: !cfg.NATSConfig.JetStreamEnabled,
			},
			NatsOptions: []nc.Option{
				nc.Timeout(cfg.NATSConfig.ConnectTimeout),
				nc.ReconnectWait(cfg.NATSConfig.ReconnectWait),
			},
		},
		watermill.NewStdLogger(false, false),
	)
	if err != nil {
		lbLogger.Error("Failed to create NATS publisher", "error", err)
		return
	}
	defer publisher.Close()

	lbLogger.Info("NATS publisher created successfully", "url", cfg.NATSConfig.URL)

	githubClient := github.NewGitHubClient()

	lbLogger.Info("Fetching repository info from GitHub", "owner", owner, "repo", repo)
	ghRepo, err := githubClient.GetRepository(owner, repo, token)
	if err != nil {
		lbLogger.Error("Failed to fetch repository from GitHub", "error", err)
		return
	}
	lbLogger.Info("Repository found on GitHub", "repo_id", ghRepo.ID, "full_name", ghRepo.FullName)

	projectRPCClient, err := grpc.NewClient(cfg.ProjectGRPC, lbLogger)
	if err != nil {
		lbLogger.Error("Failed to create project gRPC client", "error", err)
		return
	}
	defer projectRPCClient.Close()

	projectClient, err := projectadapter.New(projectRPCClient)
	if err != nil {
		lbLogger.Error("Failed to create project adapter", "error", err)
		return
	}
	defer projectClient.Close()

	ctx := context.Background()
	repoIDStr := strconv.FormatUint(ghRepo.ID, 10)
	projectRes, err := projectClient.GetProjectByRepo(ctx, &projectadapter.GetProjectByRepoRequest{
		RepoProvider: "GITHUB",
		RepoID:       repoIDStr,
	})
	if err != nil {
		lbLogger.Error("Project not found in database. Please create a project first using POST /v1/projects",
			"owner", owner,
			"repo", repo,
			"repo_id", repoIDStr,
			"error", err)
		fmt.Printf("\nError: Project '%s/%s' (repo_id: %s) is not registered.\n", owner, repo, repoIDStr)
		fmt.Println("Please create the project first:")
		fmt.Printf(`
curl -X POST http://localhost:8084/v1/projects \
    -H "Content-Type: application/json" \
    -d '{
        "name": "%s",
        "slug": "%s-%s",
        "status": "ACTIVE",
        "repoProvider": "GITHUB",
        "owner": "%s",
        "repo": "%s",
        "vcsToken": "<your-github-token>"
    }'
`, repo, owner, repo, owner, repo)
		return
	}

	lbLogger.Info("Project found in database",
		"project_id", projectRes.ProjectID,
		"slug", projectRes.Slug,
		"git_repo_id", projectRes.GitRepoID)

	fetcherCfg := historical.Config{
		Owner:          owner,
		Repo:           repo,
		Token:          token,
		EventTypes:     eventTypes,
		BatchSize:      batchSize,
		IncludeReviews: includeReviews,
	}

	fetcher := historical.NewFetcher(fetcherCfg, githubClient, databaseConn.Pool, publisher)

	if err := fetcher.Run(ctx); err != nil {
		lbLogger.Error("Fetch historical failed", "error", err)
		return
	}

	lbLogger.Info("Fetch historical completed successfully")
}
