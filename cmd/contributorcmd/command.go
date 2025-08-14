package contributorcmd

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "os"

    "github.com/spf13/cobra"
    "github.com/gocasters/rankr/domain/contributor"
    "github.com/gocasters/rankr/adapters/postgres"
    "github.com/gocasters/rankr/adapter/redis"

    _ "github.com/lib/pq"
    redisClient "github.com/redis/go-redis/v9"
)

var ContributorCmd = &cobra.Command{
    Use:   "contributor",
    Short: "Manage contributors",
    Long:  `Commands for managing contributors in the rankr application.`,
}

var createCmd = &cobra.Command{
    Use:   "create [username] [email] [display-name]",
    Short: "Create a new contributor",
    Args:  cobra.ExactArgs(3),
    Run: func(cmd *cobra.Command, args []string) {
        username := args[0]
        email := args[1]
        displayName := args[2]
        
        // Initialize database connection
        db, err := sql.Open("postgres", getDBConnectionString())
        if err != nil {
            log.Fatalf("Failed to connect to database: %v", err)
        }
        defer db.Close()
        
        // Test database connection
        if err := db.Ping(); err != nil {
            log.Fatalf("Database connection failed: %v", err)
        }
        
        // Ensure the contributors table exists
        ensureContributorsTableExists(db)
        
        // Initialize Redis connection
        redisURL := getEnv("REDIS_URL", "redis://localhost:6379")
        opt, err := redisClient.ParseURL(redisURL)
        if err != nil {
            log.Fatalf("Failed to parse Redis URL: %v", err)
        }
        rdb := redisClient.NewClient(opt)
        defer rdb.Close()
        
        // Test Redis connection
        ctx := context.Background()
        if err := rdb.Ping(ctx).Err(); err != nil {
            log.Fatalf("Redis connection failed: %v", err)
        }
        
        // Initialize contributor domain components
        redisAdapt := redis.NewAdapter(rdb)
        contributorRepo := postgres.NewContributorRepository(db)
        contributorCache := redis.NewContributorCache(redisAdapt)
        contributorUseCase := contributor.NewUseCase(contributorRepo, contributorCache)
        
        // Create contributor
        contrib, err := contributorUseCase.CreateContributor(context.Background(), username, email, displayName)
        if err != nil {
            log.Fatalf("Failed to create contributor: %v", err)
        }
        
        fmt.Printf("Created contributor: %s (%s) - %s\n", username, email, displayName)
        fmt.Printf("ID: %s\n", contrib.ID)
    },
}

func ensureContributorsTableExists(db *sql.DB) {
    // Check if table exists
    var exists bool
    err := db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'contributors')").Scan(&exists)
    if err != nil {
        log.Printf("Warning: Failed to check if contributors table exists: %v", err)
        return
    }
    
    if !exists {
        log.Println("Contributors table does not exist, creating it...")
        
        // Create the table
        _, err = db.Exec(`
            CREATE TABLE contributors (
                id VARCHAR(36) PRIMARY KEY,
                username VARCHAR(50) UNIQUE NOT NULL,
                email VARCHAR(255) UNIQUE NOT NULL,
                display_name VARCHAR(100) NOT NULL,
                avatar_url TEXT,
                github_id VARCHAR(50),
                is_active BOOLEAN DEFAULT true,
                created_at TIMESTAMP DEFAULT NOW(),
                updated_at TIMESTAMP DEFAULT NOW()
            );
        `)
        if err != nil {
            log.Fatalf("Failed to create contributors table: %v", err)
        }
        
        // Create indexes
        _, err = db.Exec("CREATE INDEX idx_contributors_username ON contributors(username);")
        if err != nil {
            log.Printf("Warning: Failed to create username index: %v", err)
        }
        
        _, err = db.Exec("CREATE INDEX idx_contributors_email ON contributors(email);")
        if err != nil {
            log.Printf("Warning: Failed to create email index: %v", err)
        }
        
        log.Println("✅ Contributors table created successfully!")
    }
}

var getCmd = &cobra.Command{
    Use:   "get [id]",
    Short: "Get a contributor by ID",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        id := args[0]
        
        // Initialize database connection
        db, err := sql.Open("postgres", getDBConnectionString())
        if err != nil {
            log.Fatalf("Failed to connect to database: %v", err)
        }
        defer db.Close()
        
        // Ensure the contributors table exists
        ensureContributorsTableExists(db)
        
        // Initialize Redis connection
        redisURL := getEnv("REDIS_URL", "redis://localhost:6379")
        opt, err := redisClient.ParseURL(redisURL)
        if err != nil {
            log.Fatalf("Failed to parse Redis URL: %v", err)
        }
        rdb := redisClient.NewClient(opt)
        defer rdb.Close()
        
        // Initialize contributor domain components
        redisAdapt := redis.NewAdapter(rdb)
        contributorRepo := postgres.NewContributorRepository(db)
        contributorCache := redis.NewContributorCache(redisAdapt)
        contributorUseCase := contributor.NewUseCase(contributorRepo, contributorCache)
        
        // Get contributor
        contrib, err := contributorUseCase.GetContributor(context.Background(), id)
        if err != nil {
            log.Fatalf("Failed to get contributor: %v", err)
        }
        
        fmt.Printf("ID: %s\n", contrib.ID)
        fmt.Printf("Username: %s\n", contrib.Username)
        fmt.Printf("Email: %s\n", contrib.Email)
        fmt.Printf("Display Name: %s\n", contrib.DisplayName)
        fmt.Printf("Avatar URL: %s\n", contrib.AvatarURL)
        fmt.Printf("GitHub ID: %s\n", contrib.GitHubID)
        fmt.Printf("Is Active: %t\n", contrib.IsActive)
        fmt.Printf("Created At: %s\n", contrib.CreatedAt.Format("2006-01-02 15:04:05"))
        fmt.Printf("Updated At: %s\n", contrib.UpdatedAt.Format("2006-01-02 15:04:05"))
    },
}

var listCmd = &cobra.Command{
    Use:   "list",
    Short: "List all contributors",
    Run: func(cmd *cobra.Command, args []string) {
        // Parse limit and offset from flags
        limit, _ := cmd.Flags().GetInt("limit")
        offset, _ := cmd.Flags().GetInt("offset")
        
        // Initialize database connection
        db, err := sql.Open("postgres", getDBConnectionString())
        if err != nil {
            log.Fatalf("Failed to connect to database: %v", err)
        }
        defer db.Close()
        
        // Ensure the contributors table exists
        ensureContributorsTableExists(db)
        
        // Initialize Redis connection
        redisURL := getEnv("REDIS_URL", "redis://localhost:6379")
        opt, err := redisClient.ParseURL(redisURL)
        if err != nil {
            log.Fatalf("Failed to parse Redis URL: %v", err)
        }
        rdb := redisClient.NewClient(opt)
        defer rdb.Close()
        
        // Initialize contributor domain components
        redisAdapt := redis.NewAdapter(rdb)
        contributorRepo := postgres.NewContributorRepository(db)
        contributorCache := redis.NewContributorCache(redisAdapt)
        contributorUseCase := contributor.NewUseCase(contributorRepo, contributorCache)
        
        // List contributors
        contributors, err := contributorUseCase.ListContributors(context.Background(), limit, offset)
        if err != nil {
            log.Fatalf("Failed to list contributors: %v", err)
        }
        
        fmt.Printf("Found %d contributors:\n", len(contributors))
        fmt.Println("=====================================")
        for _, contrib := range contributors {
            fmt.Printf("ID: %s\n", contrib.ID)
            fmt.Printf("Username: %s\n", contrib.Username)
            fmt.Printf("Email: %s\n", contrib.Email)
            fmt.Printf("Display Name: %s\n", contrib.DisplayName)
            fmt.Println("-------------------------------------")
        }
    },
}

var deleteCmd = &cobra.Command{
    Use:   "delete [id]",
    Short: "Delete a contributor by ID",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        id := args[0]
        
        // Initialize database connection
        db, err := sql.Open("postgres", getDBConnectionString())
        if err != nil {
            log.Fatalf("Failed to connect to database: %v", err)
        }
        defer db.Close()
        
        // Ensure the contributors table exists
        ensureContributorsTableExists(db)
        
        // Initialize Redis connection
        redisURL := getEnv("REDIS_URL", "redis://localhost:6379")
        opt, err := redisClient.ParseURL(redisURL)
        if err != nil {
            log.Fatalf("Failed to parse Redis URL: %v", err)
        }
        rdb := redisClient.NewClient(opt)
        defer rdb.Close()
        
        // Initialize contributor domain components
        redisAdapt := redis.NewAdapter(rdb)
        contributorRepo := postgres.NewContributorRepository(db)
        contributorCache := redis.NewContributorCache(redisAdapt)
        contributorUseCase := contributor.NewUseCase(contributorRepo, contributorCache)
        
        // Delete contributor
        err = contributorUseCase.DeleteContributor(context.Background(), id)
        if err != nil {
            log.Fatalf("Failed to delete contributor: %v", err)
        }
        
        fmt.Printf("Deleted contributor with ID: %s\n", id)
    },
}

func init() {
    // Add flags to list command
    listCmd.Flags().Int("limit", 10, "Maximum number of contributors to return")
    listCmd.Flags().Int("offset", 0, "Number of contributors to skip")
    
    ContributorCmd.AddCommand(createCmd)
    ContributorCmd.AddCommand(getCmd)
    ContributorCmd.AddCommand(listCmd)
    ContributorCmd.AddCommand(deleteCmd)
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getDBConnectionString() string {
    host := getEnv("DB_HOST", "localhost")
    port := getEnv("DB_PORT", "5432")
    user := getEnv("DB_USER", "postgres")
    password := getEnv("DB_PASSWORD", "")
    dbname := getEnv("DB_NAME", "rankr")
    
    // If password is empty, try to connect without password (trust authentication)
    if password == "" {
        return fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable",
            host, port, user, dbname)
    }
    
    return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        host, port, user, password, dbname)
}
