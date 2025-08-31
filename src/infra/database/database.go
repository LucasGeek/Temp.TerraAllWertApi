package database

import (
	"fmt"
	"log"

	"api/infra/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectPostgres(cfg config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	var logLevel logger.LogLevel
	if cfg.IsDev() {
		logLevel = logger.Info
	} else {
		logLevel = logger.Silent
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Database connection established: %s:%s", cfg.Database.Host, cfg.Database.Port)
	return db, nil
}

func RunMigrations(db *gorm.DB) error {
	log.Println("Running database migrations...")
	
	// Add your models here to auto-migrate
	// err := db.AutoMigrate(
	//     &models.User{},
	//     &models.Product{},
	// )
	// 
	// if err != nil {
	//     return fmt.Errorf("failed to run migrations: %w", err)
	// }

	log.Println("Migrations completed successfully")
	return nil
}

func RunSeeds(db *gorm.DB) error {
	log.Println("Running database seeds...")
	
	// Add your seed logic here
	// Example:
	// if err := seedUsers(db); err != nil {
	//     return fmt.Errorf("failed to seed users: %w", err)
	// }

	log.Println("Seeds completed successfully")
	return nil
}

func CloseConnection(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}
	
	return sqlDB.Close()
}