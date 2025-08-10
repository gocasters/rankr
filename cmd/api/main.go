package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Port     int
	RedisUrl string
}

func main() {
	var config Config
	flag.IntVar(&config.Port, "port", 8080, "Server port")
	flag.StringVar(&config.RedisUrl, "redis-url", getEnv("REDIS_URL", "redis://localhost:6379"), "Redis URL")
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
