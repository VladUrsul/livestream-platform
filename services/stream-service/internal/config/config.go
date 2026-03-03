package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	RTMP     RTMPConfig
	HLS      HLSConfig
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

type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

type RTMPConfig struct {
	Port    string
	BaseURL string
}

type HLSConfig struct {
	OutputDir   string
	SegmentTime int
	ListSize    int
	BaseURL     string
}

type RabbitMQConfig struct {
	URL            string
	StreamExchange string
}

type JWTConfig struct {
	AccessSecret string
}

func Load() (*Config, error) {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8083"),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		Database: DatabaseConfig{
			Host:     mustEnv("DB_HOST"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     mustEnv("DB_USER"),
			Password: mustEnv("DB_PASSWORD"),
			Name:     getEnv("DB_NAME", "stream_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
		},
		RTMP: RTMPConfig{
			Port:    getEnv("RTMP_PORT", "1935"),
			BaseURL: getEnv("RTMP_BASE_URL", "rtmp://localhost:1935/live"),
		},
		HLS: HLSConfig{
			OutputDir:   getEnv("HLS_OUTPUT_DIR", "/tmp/hls"),
			SegmentTime: 2,
			ListSize:    5,
			BaseURL:     getEnv("HLS_BASE_URL", "http://localhost:8083/hls"),
		},
		RabbitMQ: RabbitMQConfig{
			URL:            getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
			StreamExchange: getEnv("STREAM_EXCHANGE", "stream.events"),
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
