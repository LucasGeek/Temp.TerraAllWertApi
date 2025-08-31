package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	API      APIConfig
	CORS     CORSConfig
	Cache    CacheConfig
	Minio    MinioConfig
}

type AppConfig struct {
	Name            string
	Version         string
	Environment     string
	Port            string
	Debug           bool
	GracefulTimeout time.Duration
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Host         string
	Port         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
}

type JWTConfig struct {
	Secret             string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}

type APIConfig struct {
	Key     string
	Secret  string
	BaseURL string
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
}

type CacheConfig struct {
	TTL     time.Duration
	Enabled bool
	Prefix  string
}

type MinioConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	UseSSL          bool
	BaseURL         string
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		App: AppConfig{
			Name:            getEnv("APP_NAME", "Terra-Allwert-API"),
			Version:         getEnv("APP_VERSION", "1.0.0"),
			Environment:     getEnv("APP_ENV", "development"),
			Port:            getEnv("APP_PORT", "3000"),
			Debug:           getEnvBool("APP_DEBUG", true),
			GracefulTimeout: getEnvDuration("APP_GRACEFUL_TIMEOUT", 30*time.Second),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "apiuser"),
			Password:        getEnv("DB_PASSWORD", "apipass"),
			Name:            getEnv("DB_NAME", "terraallwert"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         getEnv("REDIS_PORT", "6379"),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getEnvInt("REDIS_DB", 0),
			PoolSize:     getEnvInt("REDIS_POOL_SIZE", 10),
			MinIdleConns: getEnvInt("REDIS_MIN_IDLE_CONNS", 5),
			MaxRetries:   getEnvInt("REDIS_MAX_RETRIES", 3),
		},
		JWT: JWTConfig{
			Secret:             getEnv("JWT_SECRET", "your-super-secret-jwt-key"),
			AccessTokenExpiry:  getEnvDuration("JWT_ACCESS_TOKEN_EXPIRY", 15*time.Minute),
			RefreshTokenExpiry: getEnvDuration("JWT_REFRESH_TOKEN_EXPIRY", 7*24*time.Hour),
		},
		API: APIConfig{
			Key:     getEnv("API_KEY", ""),
			Secret:  getEnv("API_SECRET", ""),
			BaseURL: getEnv("API_BASE_URL", ""),
		},
		CORS: CORSConfig{
			AllowedOrigins:   getEnvSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
			AllowedMethods:   getEnvSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			AllowedHeaders:   getEnvSlice("CORS_ALLOWED_HEADERS", []string{"Content-Type", "Authorization"}),
			AllowCredentials: getEnvBool("CORS_ALLOW_CREDENTIALS", true),
		},
		Cache: CacheConfig{
			TTL:     getEnvDuration("CACHE_TTL", 5*time.Minute),
			Enabled: getEnvBool("CACHE_ENABLED", true),
			Prefix:  getEnv("CACHE_PREFIX", "terra:"),
		},
		Minio: MinioConfig{
			Endpoint:        getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKeyID:     getEnv("MINIO_ACCESS_KEY_ID", "minioadmin"),
			SecretAccessKey: getEnv("MINIO_SECRET_ACCESS_KEY", "minioadmin"),
			BucketName:      getEnv("MINIO_BUCKET_NAME", "terra-allwert"),
			UseSSL:          getEnvBool("MINIO_USE_SSL", false),
			BaseURL:         getEnv("MINIO_BASE_URL", "http://localhost:9000"),
		},
	}, nil
}

func (c *Config) IsDev() bool {
	return c.App.Environment == "development"
}

func (c *Config) IsPrd() bool {
	return c.App.Environment == "production"
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		return value == "true"
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value, exists := os.LookupEnv(key); exists {
		// Split by comma and trim whitespace
		parts := make([]string, 0)
		for _, part := range []string{value} {
			if trimmed := fmt.Sprintf("%s", part); trimmed != "" {
				// Simple split by comma
				for _, p := range []string{part} {
					if p != "" {
						parts = append(parts, p)
					}
				}
			}
		}
		// For now, just return the value as single item
		if value != "" {
			return []string{value}
		}
	}
	return defaultValue
}
