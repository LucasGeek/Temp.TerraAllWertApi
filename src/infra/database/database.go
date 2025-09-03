package database

import (
	"fmt"
	"log"
	"time"

	"terra-allwert/domain/entities"
	"terra-allwert/infra/config"
	"terra-allwert/infra/database/seeds"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database holds the database connection
type Database struct {
	DB *gorm.DB
}

// New creates a new database connection
func New(cfg *config.Config) (*Database, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBPort,
		cfg.DBSSLMode,
	)

	// Configure GORM logger
	gormLogger := logger.Default.LogMode(logger.Info)
	if cfg.Environment == "production" {
		gormLogger = logger.Default.LogMode(logger.Error)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(30)
	sqlDB.SetConnMaxLifetime(time.Hour)

	database := &Database{DB: db}

	// Auto-migrate tables
	if err := database.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Check if database is empty and run seeds if needed
	if err := database.checkAndSeed(); err != nil {
		log.Printf("⚠️  Warning: Failed to check/seed database: %v", err)
		// Don't fail startup, just log the warning
	}

	log.Println("✅ Database connected and migrated successfully")
	return database, nil
}

// migrate runs auto-migration for all entities
func (d *Database) migrate() error {
	return d.DB.AutoMigrate(
		&entities.User{},
		&entities.Enterprise{},
		&entities.File{},
		&entities.FileVariant{},
		&entities.Menu{},
		&entities.MenuCarousel{},
		&entities.CarouselItem{},
		&entities.CarouselTextOverlay{},
		&entities.Tower{},
		&entities.Floor{},
		&entities.Suite{},
		&entities.MenuFloorPlan{},
		&entities.MenuPins{},
		&entities.PinMarker{},
		&entities.PinMarkerImage{},
	)
}

// checkAndSeed verifica se o banco está vazio e executa seeds se necessário
func (d *Database) checkAndSeed() error {
	// Verifica se já existem dados críticos (enterprises e users)
	var enterpriseCount, userCount int64
	
	if err := d.DB.Model(&entities.Enterprise{}).Count(&enterpriseCount).Error; err != nil {
		return fmt.Errorf("failed to count enterprises: %w", err)
	}
	
	if err := d.DB.Model(&entities.User{}).Count(&userCount).Error; err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}

	// Se não há dados essenciais, executar seeding
	if enterpriseCount == 0 && userCount == 0 {
		log.Println("🌱 Database appears to be empty, running automatic seeding...")
		
		seeder := seeds.NewSeeder(d.DB)
		if err := seeder.SeedAll(); err != nil {
			return fmt.Errorf("failed to run automatic seeding: %w", err)
		}
		
		log.Println("✅ Automatic database seeding completed successfully!")
	} else {
		log.Printf("📊 Database contains %d enterprises and %d users - skipping seeding", enterpriseCount, userCount)
	}

	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GetDB returns the GORM database instance
func (d *Database) GetDB() *gorm.DB {
	return d.DB
}