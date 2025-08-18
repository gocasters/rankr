package main

import (
    "github.com/gocasters/rankr/userapp/auth"
    "github.com/gocasters/rankr/userapp/delivery"
    "github.com/gocasters/rankr/userapp/repository"
    "github.com/gocasters/rankr/userapp/service"
    "github.com/labstack/echo/v4"
    "time"
)

func main() {
    cfg := LoadConfig()

    jwtManager := auth.NewJWTManager(cfg.JWTSecret, time.Duration(cfg.TokenDuration)*time.Minute)
    authService := service.NewAuthService(jwtManager)
    roleRepo := repository.NewRoleRepository()

    e := echo.New()
    // Add essential middleware
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())
    e.Use(middleware.CORS())
    e.Use(middleware.Secure())

    delivery.NewAuthHandler(e, authService, roleRepo)

    e.Logger.Fatal(e.Start(":5002")) // auth service port
}

