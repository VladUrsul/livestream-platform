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
