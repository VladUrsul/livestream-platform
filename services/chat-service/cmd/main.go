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

	"github.com/VladUrsul/livestream-platform/services/chat-service/internal/config"
	"github.com/VladUrsul/livestream-platform/services/chat-service/internal/handler"
	"github.com/VladUrsul/livestream-platform/services/chat-service/internal/hub"
	"github.com/VladUrsul/livestream-platform/services/chat-service/internal/repository"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// DB
	db, err := pgxpool.New(ctx, cfg.Database.DSN())
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		log.Fatalf("db ping: %v", err)
	}
	log.Println("✓ chat_db connected")

	// Wire
	repo := repository.New(db)
	h := hub.NewHub(repo, cfg.Chat.HistoryLimit, cfg.Chat.MaxMsgLen)
	chatHandler := handler.New(h, cfg.JWT.AccessSecret)

	go h.Run(ctx)

	// HTTP + WS server
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	chatHandler.Register(r)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Printf("chat-service on :%s", cfg.Server.Port)
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
