package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	RabbitMQ RabbitMQConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

type RabbitMQConfig struct {
	URL            string
	AuthExchange   string
	StreamExchange string
	QueueName      string
}

type JWTConfig struct {
	AccessSecret string
}

func Load() (*Config, error) {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8082"),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		Database: DatabaseConfig{
			Host:     mustEnv("DB_HOST"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     mustEnv("DB_USER"),
			Password: mustEnv("DB_PASSWORD"),
			Name:     getEnv("DB_NAME", "user_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		RabbitMQ: RabbitMQConfig{
			URL:            getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
			AuthExchange:   getEnv("AUTH_EXCHANGE", "auth.events"),
			StreamExchange: getEnv("STREAM_EXCHANGE", "stream.events"),
			QueueName:      getEnv("USER_QUEUE_NAME", "user-service.events"),
		},
		JWT: JWTConfig{
			AccessSecret: mustEnv("JWT_ACCESS_SECRET"),
		},
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required env var %q is not set", key))
	}
	return v
}
