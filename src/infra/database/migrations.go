package database

import (
	"api/domain/entities"
	"fmt"
	"log"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	log.Println("Starting database migrations...")

	// Criar extensão UUID se não existir
	if err := createUUIDExtension(db); err != nil {
		return fmt.Errorf("failed to create UUID extension: %w", err)
	}

	// Executar migrations em ordem específica devido às dependências de FK
	migrations := []interface{}{
		&entities.User{},
		&entities.Tower{},
		&entities.Floor{},
		&entities.Apartment{},
		&entities.ApartmentImage{},
		&entities.GalleryImage{},
		&entities.ImagePin{},
		&entities.AppConfig{},
	}

	for _, model := range migrations {
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate %T: %w", model, err)
		}
		log.Printf("Successfully migrated %T", model)
	}

	// Criar índices personalizados
	if err := createCustomIndexes(db); err != nil {
		return fmt.Errorf("failed to create custom indexes: %w", err)
	}

	// Criar constraints adicionais
	if err := createCustomConstraints(db); err != nil {
		return fmt.Errorf("failed to create custom constraints: %w", err)
	}

	log.Println("Database migrations completed successfully!")
	return nil
}

func createUUIDExtension(db *gorm.DB) error {
	return db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";").Error
}

func createCustomIndexes(db *gorm.DB) error {
	indexes := []string{
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email ON users(email);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_username ON users(username);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_role ON users(role);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_active ON users(active);",
		
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_towers_name ON towers(name);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_towers_created_at ON towers(created_at);",
		
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_floors_tower_id ON floors(tower_id);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_floors_number ON floors(number);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_floors_tower_number ON floors(tower_id, number);",
		
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_apartments_floor_id ON apartments(floor_id);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_apartments_number ON apartments(number);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_apartments_status ON apartments(status);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_apartments_available ON apartments(available);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_apartments_price ON apartments(price);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_apartments_bedrooms ON apartments(bedrooms);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_apartments_suites ON apartments(suites);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_apartments_solar_position ON apartments(solar_position);",
		
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_apartment_images_apartment_id ON apartment_images(apartment_id);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_apartment_images_order ON apartment_images(\"order\");",
		
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gallery_images_route ON gallery_images(route);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_gallery_images_display_order ON gallery_images(display_order);",
		
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_image_pins_gallery_image_id ON image_pins(gallery_image_id);",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_image_pins_apartment_id ON image_pins(apartment_id) WHERE apartment_id IS NOT NULL;",
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_image_pins_coordinates ON image_pins(x_coord, y_coord);",
	}

	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			// Log warning but don't fail - index might already exist
			log.Printf("Warning: Failed to create index: %v", err)
		}
	}

	return nil
}

func createCustomConstraints(db *gorm.DB) error {
	constraints := []string{
		// Unique constraint for apartment number per floor
		`ALTER TABLE apartments 
		 DROP CONSTRAINT IF EXISTS unique_apartment_number_per_floor;`,
		`ALTER TABLE apartments 
		 ADD CONSTRAINT unique_apartment_number_per_floor 
		 UNIQUE (floor_id, number);`,

		// Unique constraint for floor number per tower  
		`ALTER TABLE floors 
		 DROP CONSTRAINT IF EXISTS unique_floor_number_per_tower;`,
		`ALTER TABLE floors 
		 ADD CONSTRAINT unique_floor_number_per_tower 
		 UNIQUE (tower_id, number);`,

		// Check constraint for apartment price
		`ALTER TABLE apartments 
		 DROP CONSTRAINT IF EXISTS check_apartment_price_positive;`,
		`ALTER TABLE apartments 
		 ADD CONSTRAINT check_apartment_price_positive 
		 CHECK (price IS NULL OR price > 0);`,

		// Check constraint for bedrooms and suites
		`ALTER TABLE apartments 
		 DROP CONSTRAINT IF EXISTS check_apartment_bedrooms_positive;`,
		`ALTER TABLE apartments 
		 ADD CONSTRAINT check_apartment_bedrooms_positive 
		 CHECK (bedrooms IS NULL OR bedrooms >= 0);`,

		`ALTER TABLE apartments 
		 DROP CONSTRAINT IF EXISTS check_apartment_suites_positive;`,
		`ALTER TABLE apartments 
		 ADD CONSTRAINT check_apartment_suites_positive 
		 CHECK (suites IS NULL OR suites >= 0);`,

		// Check constraint for parking spots
		`ALTER TABLE apartments 
		 DROP CONSTRAINT IF EXISTS check_apartment_parking_positive;`,
		`ALTER TABLE apartments 
		 ADD CONSTRAINT check_apartment_parking_positive 
		 CHECK (parking_spots IS NULL OR parking_spots >= 0);`,

		// Check constraint for image coordinates (0-100 percentage)
		`ALTER TABLE image_pins 
		 DROP CONSTRAINT IF EXISTS check_pin_coordinates_range;`,
		`ALTER TABLE image_pins 
		 ADD CONSTRAINT check_pin_coordinates_range 
		 CHECK (x_coord >= 0 AND x_coord <= 100 AND y_coord >= 0 AND y_coord <= 100);`,
	}

	for _, constraintSQL := range constraints {
		if err := db.Exec(constraintSQL).Error; err != nil {
			// Log warning but don't fail - constraint might already exist
			log.Printf("Warning: Failed to create constraint: %v", err)
		}
	}

	return nil
}

// DropAllTables remove todas as tabelas (útil para testes)
func DropAllTables(db *gorm.DB) error {
	tables := []string{
		"image_pins",
		"gallery_images", 
		"apartment_images",
		"apartments",
		"floors",
		"towers",
		"users",
		"app_config",
	}

	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)).Error; err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	return nil
}

// Migration representa uma migração específica
type Migration struct {
	Version     string
	Description string
	Up          func(*gorm.DB) error
	Down        func(*gorm.DB) error
}

// GetMigrations retorna todas as migrations disponíveis
func GetMigrations() []Migration {
	return []Migration{
		{
			Version:     "001_initial_schema",
			Description: "Create initial database schema",
			Up: func(db *gorm.DB) error {
				return AutoMigrate(db)
			},
			Down: func(db *gorm.DB) error {
				return DropAllTables(db)
			},
		},
	}
}