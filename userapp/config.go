package main

type Config struct {
    JWTSecret     string
    TokenDuration int // in minutes
}

func LoadConfig() *Config {
    return &Config{
        JWTSecret:     "super-secret-key", // TODO در عمل باید از env بیاد
        TokenDuration: 60,
    }
}

