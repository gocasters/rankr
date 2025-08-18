package main

import (
    "authapp/auth"
    "authapp/delivery"
    "authapp/repository"
    "authapp/service"
    "github.com/labstack/echo/v4"
    "time"
)

func main() {
    cfg := LoadConfig()

    jwtManager := auth.NewJWTManager(cfg.JWTSecret, time.Duration(cfg.TokenDuration)*time.Minute)
    authService := service.NewAuthService(jwtManager)
    roleRepo := repository.NewRoleRepository()

    e := echo.New()
    delivery.NewAuthHandler(e, authService, roleRepo)

    e.Logger.Fatal(e.Start(":5002")) // auth service port
}

