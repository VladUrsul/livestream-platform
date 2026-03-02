package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	"github.com/VladUrsul/livestream-platform/services/auth-service/internal/cache"
	"github.com/VladUrsul/livestream-platform/services/auth-service/internal/config"
	"github.com/VladUrsul/livestream-platform/services/auth-service/internal/handler"
	"github.com/VladUrsul/livestream-platform/services/auth-service/internal/repository"
	"github.com/VladUrsul/livestream-platform/services/auth-service/internal/service"
	"github.com/VladUrsul/livestream-platform/services/auth-service/internal/token"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx := context.Background()

	// ── Database ─────────────────────────────────────────────────────────────
	dbPool, err := pgxpool.New(ctx, cfg.Database.DSN())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		log.Fatalf("database not reachable: %v", err)
	}
	log.Println("✓ connected to auth_db")

	// ── Redis ─────────────────────────────────────────────────────────────────
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis not reachable: %v", err)
	}
	defer redisClient.Close()
	log.Println("✓ connected to redis")

	// ── Dependency Wiring ─────────────────────────────────────────────────────
	tokenProvider := token.NewProvider(
		cfg.JWT.AccessSecret,
		cfg.JWT.RefreshSecret,
		cfg.JWT.AccessExpiry(),
		cfg.JWT.RefreshExpiry(),
	)
	userRepo := repository.NewPostgresUserRepository(dbPool)
	authCache := cache.NewRedisAuthCache(redisClient)
	authSvc := service.NewAuthService(userRepo, tokenProvider, authCache)
	authHandler := handler.NewAuthHandler(authSvc)

	// ── HTTP Server ───────────────────────────────────────────────────────────
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "auth-service"})
	})

	api := r.Group("/api/v1")
	authHandler.RegisterRoutes(api.Group("/auth"))

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// ── Graceful Shutdown ─────────────────────────────────────────────────────
	go func() {
		log.Printf("auth-service listening on :%s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down auth-service...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("auth-service stopped cleanly")
}
