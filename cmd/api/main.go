package main

import (
    "context"
    "database/sql"
    "flag"
    "fmt"
    "log"
    "net/http"
    "os"

    "github.com/redis/go-redis/v9"
    "github.com/gocasters/rankr/domain/contributor"
    "github.com/gocasters/rankr/adapters/postgres"
    "github.com/gocasters/rankr/adapter/redis"
    "github.com/gocasters/rankr/pkg/http/handler"
    _ "github.com/lib/pq" // PostgreSQL driver
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

func main() {
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
    opt, err := redis.ParseURL(config.RedisUrl)
    if err != nil {
        log.Fatal("Redis URL parse failed:", err)
    }
    rdb := redis.NewClient(opt)
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

    // Initialize contributor domain components
    redisAdapter := redis.NewAdapter(rdb)
    contributorRepo := postgres.NewContributorRepository(db)
    contributorCache := redis.NewContributorCache(redisAdapter)
    contributorUseCase := contributor.NewUseCase(contributorRepo, contributorCache)
    contributorService := contributor.NewService(contributorUseCase)
    contributorHandler := handler.NewContributorHandler(contributorService)

    // Set up routes
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("OK"))
    })
    
    http.HandleFunc("/leaderboard", func(w http.ResponseWriter, r *http.Request) {
        // Get leaderboard from Redis
        leaderboard, err := rdb.ZRevRangeWithScores(ctx, "leaderboard", 0, 9).Result()
        if err != nil {
            http.Error(w, "Failed to get leaderboard", http.StatusInternalServerError)
            return
        }
        // Simple response format
        w.Header().Set("Content-Type", "text/plain")
        for _, member := range leaderboard {
            fmt.Fprintf(w, "%s: %.0f\n", member.Member, member.Score)
        }
    })

    // Add contributor routes
    http.HandleFunc("/contributors", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case "POST":
            // This would normally be handled by the contributorHandler.Create method
            // For simplicity, we're using a basic handler here
            w.Write([]byte("Create contributor endpoint"))
        default:
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        }
    })
    
    http.HandleFunc("/contributors/", func(w http.ResponseWriter, r *http.Request) {
        // Extract ID from URL path
        id := r.URL.Path[len("/contributors/"):]
        if id == "" {
            http.Error(w, "ID is required", http.StatusBadRequest)
            return
        }
        
        switch r.Method {
        case "GET":
            // This would normally be handled by the contributorHandler.GetByID method
            // For simplicity, we're using a basic handler here
            w.Write([]byte(fmt.Sprintf("Get contributor with ID: %s", id)))
        default:
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        }
    })

    // Start server
    addr := fmt.Sprintf(":%d", config.Port)
    log.Printf("Server starting on %s", addr)
    log.Fatal(http.ListenAndServe(addr, nil))
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}