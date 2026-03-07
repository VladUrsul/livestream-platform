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
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/VladUrsul/livestream-platform/services/notification-service/internal/config"
	"github.com/VladUrsul/livestream-platform/services/notification-service/internal/consumer"
	"github.com/VladUrsul/livestream-platform/services/notification-service/internal/handler"
	"github.com/VladUrsul/livestream-platform/services/notification-service/internal/hub"
	"github.com/VladUrsul/livestream-platform/services/notification-service/internal/repository"
	"github.com/VladUrsul/livestream-platform/services/notification-service/internal/scheduler"
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
	log.Println("✓ notification_db connected")

	// ── RabbitMQ ──────────────────────────────────────────────────────
	var rabbitConn *amqp.Connection
	for i := 0; i < 10; i++ {
		rabbitConn, err = amqp.Dial(cfg.RabbitMQ.URL)
		if err == nil {
			break
		}
		log.Printf("⚠ rabbitmq not ready, retrying (%d/10)...", i+1)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Fatalf("rabbitmq unavailable: %v", err)
	}
	defer rabbitConn.Close()
	log.Println("✓ rabbitmq connected")

	// ── Wire ──────────────────────────────────────────────────────────
	repo := repository.New(db)
	notifHub := hub.New()
	cons := consumer.New(
		rabbitConn, repo, notifHub,
		cfg.RabbitMQ.UserExchange,
		cfg.RabbitMQ.StreamExchange,
		cfg.RabbitMQ.QueueName,
	)
	sched := scheduler.New(repo)
	h := handler.New(notifHub, repo, cfg.JWT.AccessSecret)

	cons.Start(ctx)
	go sched.Run(ctx)

	// ── HTTP ──────────────────────────────────────────────────────────
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	h.Register(r)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Printf("notification-service on :%s", cfg.Server.Port)
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
	log.Println("notification-service stopped")
}
