package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"postgresDB/config"
	"postgresDB/internal/delivery/handler"
	"postgresDB/internal/delivery/routers"
	"postgresDB/internal/infrastruktur/cache"
	"postgresDB/internal/infrastruktur/database"
	"postgresDB/internal/repository/postgres"
	"postgresDB/internal/repository/redis"
	"postgresDB/internal/service"
	"postgresDB/pkg/jwt"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("file .env tidak ditemukan")
	}

	// load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("load configurasi gagal: %v", err)
	}

	ctx := context.Background()
	// initial database
	dbPool, err := database.NewConnection(ctx, cfg.DB)
	if err != nil {
		log.Fatalf("koneksi ke database gagal: %v", err)
	}
	defer dbPool.Close()
	log.Println("koneksi ke database berhasil")

	// initial Redis
	redisClient, err := cache.NewRedisClient(ctx, cfg.Redis)
	if err != nil {
		log.Fatalf("koneksi ke Redis gagal: %v", err)
	}
	defer redisClient.Close()
	log.Println("koneksi ke Redis berhasil")

	// initial repository
	userRepo := postgres.NewUserRepository(dbPool)
	productRepo := postgres.NewProductRepository(dbPool)
	orderRepo := postgres.NewOrderRepository(dbPool)
	tokenRepo := redis.NewTokenRepository(redisClient)

	// initial JWT service with token repository
	jwtService, err := jwt.NewService(&cfg.JWT, tokenRepo)
	if err != nil {
		log.Fatalf("Failed to initialize jwt service: %v", err)
	}
	log.Println("JWT service initialized")

	// initialize service
	authService := service.NewAuthService(userRepo, jwtService)
	userService := service.NewUserService(userRepo)
	productService := service.NewProductService(productRepo)
	orderService := service.NewOrderService(orderRepo, productRepo)

	// initialize handler
	authHandler := handler.NewAuthHandler(authService, cfg.JWT.RefreshTokenTTL)
	userHandler := handler.NewUserHandler(userService)
	productHandler := handler.NewProductHandler(productService)
	orderHandler := handler.NewOrderHandler(orderService)

	// initialize router
	r := routers.NewRouter(
		authHandler,
		userHandler,
		productHandler,
		orderHandler,
		jwtService,
		cfg,
	)

	// setup server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      r.SetupRoutes(),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on %s:%s", cfg.Server.Host, cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
