package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/webhookapp/repository/lostevent"
	"github.com/gocasters/rankr/webhookapp/repository/rawevent"
	"github.com/gocasters/rankr/webhookapp/service/consistency"
)

// --- Configuration ---
const (
	lookbackDuration = 4 * time.Hour   // How far back to check for lost events.
	overlapDuration  = 5 * time.Minute // A small overlap to prevent missing events at the boundary.
	githubProviderID = 1               // Example provider ID for GitHub.
	targetOwner      = "your-github-owner"
	targetRepo       = "your-github-repo"
	gitHubToken      = "your-github-token"
	targetHookID     = 123456789 // The ID of the webhook to check.
)

func main() {
	log.Println("Starting consistency checker job...")

	// 1. Setup all dependencies (database connections, GitHub client, etc.)
	// This is a placeholder for your application's specific setup logic.
	svc, err := Setup()
	if err != nil {
		log.Fatalf("Failed to set up dependencies: %v", err)
	}

	// 2. Create a new gocron scheduler. Using UTC is a best practice for servers.
	s := gocron.NewScheduler(time.UTC)

	// 3. Schedule the consistency check function to run every 4 hours.
	// The arguments for the function are passed directly to the .Do() method.
	job, err := s.Every(4).Hours().Do(runConsistencyCheck, svc)
	if err != nil {
		log.Fatalf("Failed to schedule job: %v", err)
	}

	// 4. Start the scheduler in the background.
	s.StartAsync()
	log.Printf("Job '%s' scheduled. Next run at: %v", job.GetName(), job.NextRun())

	// 5. Wait for a shutdown signal to gracefully stop the application.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown signal received, stopping scheduler...")

	s.Stop() // Gracefully stop the scheduler, waiting for running jobs to complete.
	log.Println("Scheduler stopped gracefully.")
}

// runConsistencyCheck executes the two-step consistency process.
// This function remains unchanged from the previous version.
func runConsistencyCheck(svc *consistency.Service) {
	log.Println("🚀 Starting a new consistency check run...")
	ctx := context.Background()

	// --- Step 1: Find Lost Events in the last 4 hours ---
	endTime := time.Now()
	startTime := endTime.Add(-lookbackDuration).Add(-overlapDuration)
	log.Printf("Checking for lost events between %v and %v", startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))

	err := svc.FindLostEvents(ctx, githubProviderID, targetOwner, targetRepo, targetHookID, startTime, endTime)
	if err != nil {
		log.Printf("❌ ERROR during FindLostEvents: %v", err)
		// Depending on the error, you might want to stop here.
		// For this example, we'll continue to the redelivery step.
	} else {
		log.Println("✅ Successfully completed search for lost events.")
	}

	// --- Step 2: Redeliver Events that were previously found to be lost ---
	log.Println("Attempting to redeliver any stored lost events...")
	errorCollection := svc.RedeliverLostEvents(ctx, githubProviderID)

	if len(errorCollection) > 0 {
		log.Println("⚠️ Encountered errors during redelivery:")
		for deliveryID, redeliveryErr := range errorCollection {
			log.Printf("  - Failed for Delivery ID '%s': %v", deliveryID, redeliveryErr)
		}
		log.Println("✅ Successfully processed redelivery queue.")
	}

	log.Println("Consistency check run finished.")
}

// Setup is a placeholder for your application's initialization logic.
// This function remains unchanged from the previous version.
func Setup() (*consistency.Service, error) {
	log.Println("⚙️ Initializing dependencies (DB, GitHub Client)...")

	conn, cnErr := database.Connect(database.Config{
		Host:              "127.0.0.1",
		Port:              5432,
		Username:          "root",
		Password:          "",
		DBName:            "dbname",
		SSLMode:           "disable",
		MaxConns:          10,   // Default value
		MinConns:          2,    // Default value
		MaxConnLifetime:   3600, // Default: 1 hour in seconds
		MaxConnIdleTime:   600,  // Default: 10 minutes in seconds
		HealthCheckPeriod: 60,   // Default: 1 minute in seconds
		PathOfMigrations:  "./migrations",
	})

	if cnErr != nil {
		slog.Error("fatal error occurred", "reason", "failed to connect to database", slog.Any("error", cnErr))
		panic(cnErr)
	}
	defer conn.Close()

	rawEventRepo := rawevent.NewRawWebhookRepository(conn.Pool)
	lostEventRepo := lostevent.NewLostWebhookRepository(conn.Pool)

	githubToken := os.Getenv(gitHubToken)
	client := consistency.NewGitHubClient(githubToken)

	service := consistency.New(rawEventRepo, lostEventRepo, client)

	log.Println("Dependencies initialized.")
	return service, nil // In a real scenario, you'd check for nil pointers.
}
