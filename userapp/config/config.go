package config
 
import (
    "log"
    "os"
    "strconv"
)

type Config struct {
    JWTSecret     string
    TokenDuration int // in minutes
}

// getEnvOrDefault returns the env var if set (and non-empty), otherwise returns def.
func getEnvOrDefault(key, def string) string {
    if v, ok := os.LookupEnv(key); ok && v != "" {
        return v
    }
    return def
}


func LoadConfig() *Config {
jwtSecret := getEnvOrDefault("JWT_SECRET", "")
    if jwtSecret == "" {
        log.Fatal("JWT_SECRET environment variable is required")
    }

    durStr := getEnvOrDefault("TOKEN_DURATION_MINUTES", "60")
    tokenDuration, err := strconv.Atoi(durStr)
    if err != nil || tokenDuration <= 0 {
        log.Printf("Invalid TOKEN_DURATION_MINUTES=%q; defaulting to 60", durStr)
        tokenDuration = 60
    }
    
    return &Config{
        JWTSecret:     jwtSecret,
        TokenDuration: tokenDuration,
    }
}
