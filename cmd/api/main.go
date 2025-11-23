package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"review-service/internal/config"
	"review-service/internal/handlers"
	"review-service/internal/handlers/middleware"
	"review-service/internal/repository/postgres"
	"review-service/internal/service"
)

func main() {

	cfg := config.LoadConfig()

	db, err := connectDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Successfully connected to PostgreSQL")

	userRepo := postgres.NewUserPostgresRepository(db)
	teamRepo := postgres.NewTeamPostgresRepository(db)
	prRepo := postgres.NewPullRequestPostgresRepository(db)

	userService := service.NewUserService(userRepo, prRepo)
	teamService := service.NewTeamService(teamRepo, userRepo)
	prService := service.NewPullRequestService(prRepo, userRepo, teamRepo)

	router := handlers.NewRouter(userService, teamService, prService)

	mux := http.NewServeMux()
	router.SetupRoutes(mux)

	port := cfg.ServerPort
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      middleware.Logger(middleware.Recovery(mux)),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Starting server on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

func connectDB(cfg *config.Config) (*sql.DB, error) {
	connStr := cfg.GetPostgresConnectionString()

	var db *sql.DB
	var err error

	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			return nil, fmt.Errorf("failed to open database: %w", err)
		}

		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(5 * time.Minute)

		err = db.Ping()
		if err == nil {
			return db, nil
		}

		log.Printf("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries, err)
		db.Close()

		waitTime := time.Duration(i+1) * time.Second
		if waitTime > 10*time.Second {
			waitTime = 10 * time.Second
		}
		time.Sleep(waitTime)
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
}
