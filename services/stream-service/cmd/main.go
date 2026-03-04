package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"

	"github.com/VladUrsul/livestream-platform/services/stream-service/internal/cache"
	"github.com/VladUrsul/livestream-platform/services/stream-service/internal/config"
	"github.com/VladUrsul/livestream-platform/services/stream-service/internal/handler"
	"github.com/VladUrsul/livestream-platform/services/stream-service/internal/hls"
	"github.com/VladUrsul/livestream-platform/services/stream-service/internal/publisher"
	"github.com/VladUrsul/livestream-platform/services/stream-service/internal/repository"
	"github.com/VladUrsul/livestream-platform/services/stream-service/internal/rtmp"
	"github.com/VladUrsul/livestream-platform/services/stream-service/internal/service"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()

	// PostgreSQL
	db, err := pgxpool.New(ctx, cfg.Database.DSN())
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		log.Fatalf("db ping: %v", err)
	}
	log.Println("✓ connected to stream_db")

	// Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis ping: %v", err)
	}
	defer redisClient.Close()
	log.Println("✓ connected to redis")
	// RabbitMQ
	var pub *publisher.StreamPublisher
	if rabbitConn, err := amqp.Dial(cfg.RabbitMQ.URL); err != nil {
		log.Printf("⚠ rabbitmq unavailable: %v", err)
	} else {
		defer rabbitConn.Close()
		if pub, err = publisher.NewStreamPublisher(rabbitConn, cfg.RabbitMQ.StreamExchange); err != nil {
			log.Printf("⚠ publisher: %v", err)
		} else {
			defer pub.Close()
			log.Println("✓ connected to rabbitmq")
		}
	}

	// Wire
	streamRepo := repository.NewPostgresStreamRepository(db)
	streamCache := cache.NewRedisStreamCache(redisClient)
	transcoder := hls.NewTranscoder(cfg.HLS.OutputDir, cfg.HLS.SegmentTime, cfg.HLS.ListSize, cfg.RTMP.Port)
	streamSvc := service.NewStreamService(streamRepo, streamCache, transcoder, pub, cfg.HLS.BaseURL)
	streamHandler := handler.NewStreamHandler(streamSvc)

	// RTMP server
	rtmpSrv := rtmp.NewServer(&cfg.RTMP, &rtmpBridge{svc: streamSvc})
	go func() {
		if err := rtmpSrv.Listen(); err != nil {
			log.Printf("RTMP error: %v", err)
		}
	}()

	// HTTP server
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "stream-service"})
	})

	// Serve HLS segments
	r.Use(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/hls") {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Range")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Range")
			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(204)
				return
			}
		}
		c.Next()
	})
	r.GET("/hls/*filepath", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Range")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Range")
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")

		filePath := filepath.Join(cfg.HLS.OutputDir, c.Param("filepath"))
		c.File(filePath)
	})

	api := r.Group("/api/v1/streams")

	// Protected
	auth := api.Group("")
	auth.Use(handler.AuthMiddleware(cfg.JWT.AccessSecret))
	auth.GET("/key", streamHandler.GetStreamKey)
	auth.POST("/key/rotate", streamHandler.RotateStreamKey)
	auth.PUT("/settings", streamHandler.UpdateSettings)

	// Public
	api.GET("/live", streamHandler.GetLiveStreams)
	api.GET("/:username", streamHandler.GetStreamInfo)
	api.POST("/:username/join", streamHandler.JoinStream)
	api.POST("/:username/leave", streamHandler.LeaveStream)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}
	go func() {
		log.Printf("stream-service listening on :%s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(shutCtx)
	log.Println("stream-service stopped")
}

type rtmpBridge struct{ svc service.StreamService }

func (b *rtmpBridge) OnPublish(streamKey string) (io.WriteCloser, error) {
	return b.svc.HandleStreamStart(context.Background(), streamKey)
}

func (b *rtmpBridge) OnClose(streamKey string) {
	b.svc.HandleStreamEnd(context.Background(), streamKey)
}
