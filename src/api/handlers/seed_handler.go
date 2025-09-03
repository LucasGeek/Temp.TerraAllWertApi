package handlers

import (
	"terra-allwert/infra/config"
	"terra-allwert/infra/database/seeds"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type SeedHandler struct {
	config *config.Config
}

func NewSeedHandler(cfg *config.Config) *SeedHandler {
	return &SeedHandler{
		config: cfg,
	}
}

type SeedResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Seeds executed successfully"`
}

type ErrorResponse struct {
	Error   string `json:"error" example:"Internal server error"`
	Message string `json:"message" example:"Detailed error message"`
}

// RunSeeds godoc
// @Summary Run Database Seeds
// @Description Execute database seeds to populate initial data (enterprises and users)
// @Tags seeds
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SeedResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /seeds/run [post]
func (h *SeedHandler) RunSeeds(c *fiber.Ctx) error {
	// Build database connection string
	dsn := "host=" + h.config.DBHost +
		" port=" + h.config.DBPort +
		" user=" + h.config.DBUser +
		" password=" + h.config.DBPassword +
		" dbname=" + h.config.DBName +
		" sslmode=" + h.config.DBSSLMode +
		" TimeZone=America/Sao_Paulo"

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Database connection failed",
			Message: err.Error(),
		})
	}

	// Create seeder and run seeds
	seeder := seeds.NewSeeder(db)
	if err := seeder.SeedAll(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Seeding failed",
			Message: err.Error(),
		})
	}

	return c.JSON(SeedResponse{
		Success: true,
		Message: "Seeds executed successfully",
	})
}

// RunEnterpriseSeeds godoc
// @Summary Run Enterprise Seeds
// @Description Execute only enterprise seeds
// @Tags seeds
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SeedResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /seeds/enterprises [post]
func (h *SeedHandler) RunEnterpriseSeeds(c *fiber.Ctx) error {
	// Build database connection string
	dsn := "host=" + h.config.DBHost +
		" port=" + h.config.DBPort +
		" user=" + h.config.DBUser +
		" password=" + h.config.DBPassword +
		" dbname=" + h.config.DBName +
		" sslmode=" + h.config.DBSSLMode +
		" TimeZone=America/Sao_Paulo"

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Database connection failed",
			Message: err.Error(),
		})
	}

	// Create seeder and run enterprise seeds only
	seeder := seeds.NewSeeder(db)
	if err := seeder.SeedEnterprises(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Enterprise seeding failed",
			Message: err.Error(),
		})
	}

	return c.JSON(SeedResponse{
		Success: true,
		Message: "Enterprise seeds executed successfully",
	})
}

// RunUserSeeds godoc
// @Summary Run User Seeds
// @Description Execute only user seeds (requires enterprises to exist)
// @Tags seeds
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SeedResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /seeds/users [post]
func (h *SeedHandler) RunUserSeeds(c *fiber.Ctx) error {
	// Build database connection string
	dsn := "host=" + h.config.DBHost +
		" port=" + h.config.DBPort +
		" user=" + h.config.DBUser +
		" password=" + h.config.DBPassword +
		" dbname=" + h.config.DBName +
		" sslmode=" + h.config.DBSSLMode +
		" TimeZone=America/Sao_Paulo"

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Database connection failed",
			Message: err.Error(),
		})
	}

	// Create seeder and run user seeds only
	seeder := seeds.NewSeeder(db)
	if err := seeder.SeedUsers(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "User seeding failed",
			Message: err.Error(),
		})
	}

	return c.JSON(SeedResponse{
		Success: true,
		Message: "User seeds executed successfully",
	})
}
