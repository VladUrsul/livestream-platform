package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	Server   ServerConfig
	Services ServicesConfig
	CORS     CORSConfig
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type ServicesConfig struct {
	AuthServiceURL         string
	UserServiceURL         string
	StreamServiceURL       string
	ChatServiceURL         string
	NotificationServiceURL string
	SubscriptionServiceURL string
}

type CORSConfig struct {
	AllowedOrigins []string
}

func Load() (*Config, error) {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		Services: ServicesConfig{
			AuthServiceURL:         getEnv("AUTH_SERVICE_URL", "http://localhost:8081"),
			UserServiceURL:         getEnv("USER_SERVICE_URL", "http://localhost:8082"),
			StreamServiceURL:       getEnv("STREAM_SERVICE_URL", "http://localhost:8083"),
			ChatServiceURL:         getEnv("CHAT_SERVICE_URL", "http://localhost:8084"),
			NotificationServiceURL: getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8085"),
			SubscriptionServiceURL: getEnv("SUBSCRIPTION_SERVICE_URL", "http://localhost:8086"),
		},
		CORS: CORSConfig{
			AllowedOrigins: []string{
				getEnv("CORS_ORIGIN", "http://localhost:3000"),
			},
		},
	}, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("required env var %q is not set", key))
	}
	return val
}
