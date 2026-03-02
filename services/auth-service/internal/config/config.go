package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
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
	DB       int
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

type JWTConfig struct {
	AccessSecret      string
	RefreshSecret     string
	AccessExpiryMins  int
	RefreshExpiryDays int
}

func (j JWTConfig) AccessExpiry() time.Duration {
	return time.Duration(j.AccessExpiryMins) * time.Minute
}

func (j JWTConfig) RefreshExpiry() time.Duration {
	return time.Duration(j.RefreshExpiryDays) * 24 * time.Hour
}

func Load() (*Config, error) {
	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_DB: %w", err)
	}
	accessExpiry, err := strconv.Atoi(getEnv("JWT_ACCESS_EXPIRY_MINS", "15"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_EXPIRY_MINS: %w", err)
	}
	refreshExpiry, err := strconv.Atoi(getEnv("JWT_REFRESH_EXPIRY_DAYS", "7"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_EXPIRY_DAYS: %w", err)
	}

	return &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8081"),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		Database: DatabaseConfig{
			Host:     mustEnv("DB_HOST"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     mustEnv("DB_USER"),
			Password: mustEnv("DB_PASSWORD"),
			Name:     getEnv("DB_NAME", "auth_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       redisDB,
		},
		JWT: JWTConfig{
			AccessSecret:      mustEnv("JWT_ACCESS_SECRET"),
			RefreshSecret:     mustEnv("JWT_REFRESH_SECRET"),
			AccessExpiryMins:  accessExpiry,
			RefreshExpiryDays: refreshExpiry,
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
