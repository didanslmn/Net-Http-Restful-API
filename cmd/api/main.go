package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"postgresDB/config"
	"postgresDB/internal/infrastruktur/cache"
	"postgresDB/internal/infrastruktur/postgres"

	"github.com/joho/godotenv"
)

func main() {
	// Use structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Load .env file. It's better to run the app from the root directory.
	if err := godotenv.Load("./../../.env"); err != nil {
		logger.Warn("No .env file found, reading environment variables")
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}
	ctx := context.Background()

	// Initialize database connection
	db, err := postgres.NewConnection(ctx, cfg.DB)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("Database connection pool initialized successfully")

	// Initialize redis
	redisClient, err := cache.NewRedisClient(ctx, cfg.Redis)
	if err != nil {
		logger.Error("Failed to connect to redis", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()
	logger.Info("Redis connection pool initialized successfully")

	// // Initialize JWT Manager
	// jwtManager, err := jwt.NewManager(&cfg.JWT)
	// if err != nil {
	// 	logger.Error("Failed to initialize JWT manager", "error", err)
	// 	os.Exit(1)
	// }

	// // Initialize Validator
	// val := validator.NewValidator()

	// // Dependency Injection
	// // Repository
	// userRepository := userRepo.NewUserRepository(db)

	// // Service
	// authSvc := service.NewAuthService(userRepository, jwtManager)
	// userSvc := service.NewUserService(userRepository)

	// // Handler
	// authHandler := handler.NewAuthHandler(authSvc, val)
	// userHandler := handler.NewUserHandler(userSvc, val)

	// // Router
	// router := handler.NewRouter(authHandler, userHandler, jwtManager)

	// Start server
	server := &http.Server{
		Addr: fmt.Sprintf(":%s", cfg.App.Port),
		//Handler: router,
	}

	logger.Info(fmt.Sprintf("Server starting on port %s", cfg.App.Port))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("Could not start server", "error", err)
		os.Exit(1)
	}
}
