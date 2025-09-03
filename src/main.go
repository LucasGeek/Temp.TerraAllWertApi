package main

import (
	"context"
	"log"
	"strconv"
	"time"

	"terra-allwert/api/routes"
	_ "terra-allwert/docs"
	"terra-allwert/infra/auth"
	"terra-allwert/infra/config"
	"terra-allwert/infra/database"
	"terra-allwert/infra/middleware"
	"terra-allwert/infra/repositories"
	"terra-allwert/infra/websocket"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
)

// @title Terra Allwert API
// @version 1.0
// @description API REST para sistema de gerenciamento de torres
// @host localhost:3000
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
// @example Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.New(cfg)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db.GetDB())

	// Initialize JWT service
	accessTokenHours, _ := strconv.Atoi("24")  // Default 24 hours
	refreshTokenHours, _ := strconv.Atoi("168") // Default 7 days (168 hours)
	jwtService := auth.NewJWTService(cfg.JWTSecret, accessTokenHours, refreshTokenHours)

	// Initialize progress hub for WebSocket connections
	progressHub := websocket.NewProgressHub()
	go progressHub.Run() // Start in background

	// Initialize rate limiter with production-ready config
	rateLimiter := middleware.NewUploadRateLimiter(middleware.DefaultRateLimitConfig())

	app := fiber.New(fiber.Config{
		AppName: "Terra Allwert API v1.0",
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "*",
		AllowCredentials: false,
	}))

	// Upload rate limiting middleware for file uploads
	app.Use("/api/v1/files", middleware.UploadRateLimitMiddleware(rateLimiter))

	// Swagger
	app.Get("/swagger/*", swagger.HandlerDefault)

	// WebSocket endpoint for progress tracking
	app.Get("/ws/progress", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusUpgradeRequired).JSON(fiber.Map{
			"error": "WebSocket upgrade required",
			"note":  "Use WebSocket client to connect to this endpoint",
		})
	})

	// API Routes
	api := app.Group("/api/v1")

	// Health endpoint
	api.Get("/health", func(c *fiber.Ctx) error {
		return healthCheck(c, cfg)
	})

	// Setup optimized upload routes (commented until FileRepository is available)
	// routes.SetupOptimizedUploadRoutes(app, fileRepo, storageService, redisClient, progressHub, rateLimiter, circuitBreaker)

	// Initialize auth middleware with actual services
	authMiddleware := middleware.NewAuthMiddleware(jwtService, userRepo)

	// Seed routes (for development)
	if cfg.Environment == "development" {
		routes.SetupSeedRoutes(app, cfg, authMiddleware)
	}

	// Setup auth routes
	routes.SetupAuthRoutes(api, userRepo, jwtService, authMiddleware)

	// Setup main API routes
	// TODO: Initialize handlers and setup main routes when repositories are available
	// handlers := &routes.Handlers{...}
	// routes.SetupAllRoutes(app, handlers, authMiddleware)

	// Start server
	log.Printf("🚀 Server starting on port %s", cfg.Port)
	log.Printf("📈 Progress tracking available at ws://localhost:%s/ws/progress", cfg.Port)
	log.Printf("📚 API documentation at http://localhost:%s/swagger/", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}

// ServiceHealth represents the status of a service
type ServiceHealth struct {
	Status  string `json:"status" example:"ok"`
	Message string `json:"message,omitempty" example:"Connected"`
	Error   string `json:"error,omitempty" example:"Connection failed"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status   string                   `json:"status" example:"ok"`
	Message  string                   `json:"message" example:"API is running"`
	Version  string                   `json:"version" example:"1.0.0"`
	Services map[string]ServiceHealth `json:"services"`
}

// healthCheck godoc
// @Summary Health Check
// @Description Verifica se a API está funcionando e testa conexões com MinIO e Redis
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func healthCheck(c *fiber.Ctx, cfg *config.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	services := make(map[string]ServiceHealth)
	overallStatus := "ok"

	// Test Redis connection
	redisHealth := checkRedis(ctx, cfg)
	services["redis"] = redisHealth
	if redisHealth.Status != "ok" {
		overallStatus = "degraded"
	}

	// Test MinIO connection
	minioHealth := checkMinIO(ctx, cfg)
	services["minio"] = minioHealth
	if minioHealth.Status != "ok" {
		overallStatus = "degraded"
	}

	return c.JSON(HealthResponse{
		Status:   overallStatus,
		Message:  "API is running",
		Version:  "1.0.0",
		Services: services,
	})
}

func checkRedis(ctx context.Context, cfg *config.Config) ServiceHealth {
	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisHost + ":" + cfg.RedisPort,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer rdb.Close()

	// Test connection with ping
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return ServiceHealth{
			Status: "error",
			Error:  "Connection failed: " + err.Error(),
		}
	}

	return ServiceHealth{
		Status:  "ok",
		Message: "Connected to " + cfg.RedisHost + ":" + cfg.RedisPort,
	}
}

func checkMinIO(ctx context.Context, cfg *config.Config) ServiceHealth {
	// Create MinIO client
	minioClient, err := minio.New(cfg.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
		Secure: cfg.MinIOUseSSL,
	})
	if err != nil {
		return ServiceHealth{
			Status: "error",
			Error:  "Failed to create client: " + err.Error(),
		}
	}

	// Test connection by listing buckets
	_, err = minioClient.ListBuckets(ctx)
	if err != nil {
		return ServiceHealth{
			Status: "error",
			Error:  "Connection failed: " + err.Error(),
		}
	}

	return ServiceHealth{
		Status:  "ok",
		Message: "Connected to " + cfg.MinIOEndpoint,
	}
}
