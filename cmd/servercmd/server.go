package servercmd

import (
    "context"
    "database/sql"
    "flag"
    "fmt"
    "log"
    "net/http"
    "os"
    "strings"

    "github.com/gocasters/rankr/domain/contributor"
    "github.com/gocasters/rankr/adapters/postgres"
    "github.com/gocasters/rankr/adapter/redis"
    "github.com/gocasters/rankr/pkg/http/handler"

    _ "github.com/lib/pq" // PostgreSQL driver
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    "github.com/spf13/cobra"
    redisClient "github.com/redis/go-redis/v9"
)

type Config struct {
    Port     int
    RedisUrl string
    DBHost   string
    DBPort   int
    DBUser   string
    DBPass   string
    DBName   string
}

var ServerCmd = &cobra.Command{
    Use:   "server",
    Short: "Start the HTTP server",
    Run: func(cmd *cobra.Command, args []string) {
        var config Config
        flag.IntVar(&config.Port, "port", 8080, "Server port")
        flag.StringVar(&config.RedisUrl, "redis-url", getEnv("REDIS_URL", "redis://localhost:6379"), "Redis URL")
        flag.StringVar(&config.DBHost, "db-host", getEnv("DB_HOST", "localhost"), "Database host")
        flag.IntVar(&config.DBPort, "db-port", 5432, "Database port")
        flag.StringVar(&config.DBUser, "db-user", getEnv("DB_USER", "postgres"), "Database user")
        flag.StringVar(&config.DBPass, "db-pass", getEnv("DB_PASSWORD", ""), "Database password")
        flag.StringVar(&config.DBName, "db-name", getEnv("DB_NAME", "rankr"), "Database name")
        flag.Parse()

        // Connect to Redis
        redisAddr := strings.TrimPrefix(config.RedisUrl, "redis://")
        rdb := redisClient.NewClient(&redisClient.Options{
            Addr: redisAddr,
        })
        ctx := context.Background()
        
        // Test Redis connection
        if err := rdb.Ping(ctx).Err(); err != nil {
            log.Fatal("Redis connection failed:", err)
        }
        log.Println("Connected to Redis successfully")

        // Connect to PostgreSQL
        dbConnStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
            config.DBHost, config.DBPort, config.DBUser, config.DBPass, config.DBName)
        db, err := sql.Open("postgres", dbConnStr)
        if err != nil {
            log.Fatal("Failed to connect to database:", err)
        }
        defer db.Close()
        
        // Test database connection
        if err := db.Ping(); err != nil {
            log.Fatal("Database connection failed:", err)
        }
        log.Println("Connected to PostgreSQL successfully")

        // Ensure the contributors table exists
        ensureContributorsTableExists(db)

        // Initialize contributor domain components
        redisAdapt := redis.NewAdapter(rdb)
        contributorRepo := postgres.NewContributorRepository(db)
        contributorCache := redis.NewContributorCache(redisAdapt)
        contributorUseCase := contributor.NewUseCase(contributorRepo, contributorCache)
        contributorService := contributor.NewService(contributorUseCase)
        contributorHandler := handler.NewContributorHandler(contributorService)

        // Set up Echo server
        e := echo.New()
        e.HideBanner = true
        
        // Middleware
        e.Use(middleware.Logger())
        e.Use(middleware.Recover())
        
        // Routes
        e.GET("/health", func(c echo.Context) error {
            return c.String(http.StatusOK, "OK")
        })
        
        // Leaderboard route
        e.GET("/leaderboard", func(c echo.Context) error {
            // Get leaderboard from Redis
            leaderboard, err := rdb.ZRevRangeWithScores(ctx, "leaderboard", 0, 9).Result()
            if err != nil {
                return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get leaderboard"})
            }
            
            // Simple response format
            return c.JSON(http.StatusOK, leaderboard)
        })
        
        // Contributor routes
        contributorGroup := e.Group("/contributors")
        contributorGroup.POST("", contributorHandler.Create)
        contributorGroup.GET("/:id", contributorHandler.GetByID)
        contributorGroup.PUT("/:id", contributorHandler.Update)
        contributorGroup.DELETE("/:id", contributorHandler.Delete)
        contributorGroup.GET("", contributorHandler.List)

        // Start server
        addr := fmt.Sprintf(":%d", config.Port)
        log.Printf("Server starting on %s", addr)
        log.Fatal(e.Start(addr))
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

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
