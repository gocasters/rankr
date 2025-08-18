package main

type Config struct {
    JWTSecret     string
    TokenDuration int // in minutes
}

func LoadConfig() *Config {
    jwtSecret := os.Getenv("JWT_SECRET")
    if jwtSecret == "" {
        log.Fatal("JWT_SECRET environment variable is required")
    }
    
    tokenDuration, _ := strconv.Atoi(getEnvOrDefault("TOKEN_DURATION_MINUTES", "60"))
    
    return &Config{
        JWTSecret:     jwtSecret,
        TokenDuration: tokenDuration,
    }
}
