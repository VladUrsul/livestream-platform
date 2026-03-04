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

	"github.com/VladUrsul/livestream-platform/services/user-service/internal/config"
	"github.com/VladUrsul/livestream-platform/services/user-service/internal/consumer"
	"github.com/VladUrsul/livestream-platform/services/user-service/internal/handler"
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

	db, err := pgxpool.New(ctx, cfg.Database.DSN())
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		log.Fatalf("db ping: %v", err)
	}
	log.Println("✓ user_db connected")

	rabbitConn, err := amqp.Dial(cfg.RabbitMQ.URL)
	if err != nil {
		log.Fatalf("rabbitmq: %v", err)
	}
	defer rabbitConn.Close()
	log.Println("✓ rabbitmq connected")

	repo := repository.New(db)
	userSvc := service.New(repo)
	h := handler.New(userSvc)

	c := consumer.New(rabbitConn, userSvc,
		cfg.RabbitMQ.AuthExchange,
		cfg.RabbitMQ.StreamExchange,
		cfg.RabbitMQ.QueueName,
	)
	c.Start(ctx)

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
}
