package main

import (
    "github.com/gocasters/rankr/authApp/auth"
    "github.com/gocasters/rankr/authApp/delivery/http"
    "github.com/gocasters/rankr/authApp/repository"
    "github.com/gocasters/rankr/authApp/service"
    "github.com/gocasters/rankr/authApp/config"
    "github.com/gocasters/rankr/pkg/database"
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    "time"
    "log"
)

func main() {
    cfg := config.LoadConfig()

	dbConfig := database.NewConfig(
		database.WithHost("127.0.0.1"),
		database.WithPort(5432),
		database.WithUsername("leyla"),
		database.WithPassword("123456"),
		database.WithDBName("authdb"),
		database.WithSSLMode("disable"),
	)

	// 
	db, err := database.Connect(dbConfig)
	if err != nil {
		log.Fatal("❌ cannot connect to db: ", err)
	}
	defer db.Close()


    jwtManager := auth.NewJWTManager(cfg.JWTSecret, time.Duration(cfg.TokenDuration)*time.Minute)
    authService := service.NewAuthService(jwtManager)
    roleRepo := repository.NewRoleRepository(db)

    e := echo.New()
    // Add essential middleware
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())
    e.Use(middleware.CORS())
    e.Use(middleware.Secure())

    http.NewAuthHandler(e, authService, roleRepo)

    e.Logger.Fatal(e.Start(":5002")) // auth service port
}

