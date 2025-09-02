package main

import (
    "github.com/gocasters/rankr/authapp/auth"
    "github.com/gocasters/rankr/authapp/delivery/http"
    "github.com/gocasters/rankr/authapp/repository"
    "github.com/gocasters/rankr/authapp/service"
    "github.com/gocasters/rankr/authapp/config"
    "github.com/gocasters/rankr/pkg/database"
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    "time"
    "log"
    "os"
    "strconv"
)

func main() {
    cfg := config.LoadConfig()

        host := os.Getenv("DB_HOST")
        if host == "" {
            host = "127.0.0.1"
        }
        port := 5432
        if p := os.Getenv("DB_PORT"); p != "" {
            if v, err := strconv.Atoi(p); err == nil && v > 0 {
                port = v
            }
        }
        user := os.Getenv("DB_USER")
        pass := os.Getenv("DB_PASSWORD")
        dbname := os.Getenv("DB_NAME")
        sslmode := os.Getenv("DB_SSLMODE")
        if sslmode == "" {
            sslmode = "disable"
        }
    
        dbConfig := database.NewConfig(
            database.WithHost(host),
            database.WithPort(port),
            database.WithUsername(user),
            database.WithPassword(pass),
            database.WithDBName(dbname),
            database.WithSSLMode(sslmode),
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

