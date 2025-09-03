package main

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"terra-allwert/infra/config"
	"terra-allwert/infra/database/seeds"
)

func main() {
	log.Println("🌱 Terra Allwert Database Seeder")
	log.Println("================================")

	// Load configuration
	cfg := config.Load()

	// Build database connection string
	dsn := buildDSN(cfg)

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("✅ Database connection established")

	// Create seeder
	seeder := seeds.NewSeeder(db)

	// Run seeds
	if err := seeder.SeedAll(); err != nil {
		log.Fatalf("❌ Seeding failed: %v", err)
	}

	log.Println("🎉 All seeds completed successfully!")
}

func buildDSN(cfg *config.Config) string {
	return "host=" + cfg.DBHost +
		" port=" + cfg.DBPort +
		" user=" + cfg.DBUser +
		" password=" + cfg.DBPassword +
		" dbname=" + cfg.DBName +
		" sslmode=" + cfg.DBSSLMode +
		" TimeZone=America/Sao_Paulo"
}