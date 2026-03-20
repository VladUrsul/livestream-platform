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
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/VladUrsul/livestream-platform/services/user-service/internal/config"
	"github.com/VladUrsul/livestream-platform/services/user-service/internal/consumer"
	"github.com/VladUrsul/livestream-platform/services/user-service/internal/handler"
	"github.com/VladUrsul/livestream-platform/services/user-service/internal/publisher"
	"github.com/VladUrsul/livestream-platform/services/user-service/internal/repository"
	"github.com/VladUrsul/livestream-platform/services/user-service/internal/service"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ── DB ────────────────────────────────────────────────────────────
	db, err := pgxpool.New(ctx, cfg.Database.DSN())
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		log.Fatalf("db ping: %v", err)
	}
	log.Println("✓ user_db connected")

	// ── Database Migrations ───────────────────────────────────────────────────
	m, err := migrate.New("file://../migrations", fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	))
	if err == nil {
		defer m.Close()
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Printf("⚠ migration error: %v", err)
		} else if err == migrate.ErrNoChange {
			log.Println("✓ migrations already applied")
		} else {
			log.Println("✓ migrations applied")
		}
	} else {
		log.Printf("⚠ migration setup failed: %v (continuing anyway)", err)
	}

	// ── RabbitMQ ──────────────────────────────────────────────────────
	var rabbitConn *amqp.Connection
	for i := 0; i < 10; i++ {
		rabbitConn, err = amqp.Dial(cfg.RabbitMQ.URL)
		if err == nil {
			break
		}
		log.Printf("⚠ rabbitmq not ready, retrying in 3s... (%d/10)", i+1)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Fatalf("rabbitmq unavailable: %v", err)
	}
	defer rabbitConn.Close()
	log.Println("✓ rabbitmq connected")

	// ── Publisher ─────────────────────────────────────────────────────
	var userPub *publisher.Publisher
	userPub, err = publisher.New(rabbitConn, cfg.RabbitMQ.UserExchange)
	if err != nil {
		log.Printf("⚠ publisher init failed: %v — running without follow events", err)
		userPub = nil
	} else {
		defer userPub.Close()
		log.Println("✓ user events publisher ready")
	}

	// ── Wiring ────────────────────────────────────────────────────────
	repo := repository.New(db)
	userSvc := service.NewUserService(repo, userPub)
	h := handler.New(userSvc)
	c := consumer.New(
		rabbitConn, userSvc,
		cfg.RabbitMQ.AuthExchange,
		cfg.RabbitMQ.StreamExchange,
		cfg.RabbitMQ.QueueName,
	)
	c.Start(ctx)

	// ── HTTP ──────────────────────────────────────────────────────────
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "user-service"})
	})

	api := r.Group("/api/v1/users")
	h.Register(api, cfg.JWT.AccessSecret)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Printf("user-service on :%s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	cancel()

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutCancel()
	srv.Shutdown(shutCtx)
	log.Println("user-service stopped")
}
