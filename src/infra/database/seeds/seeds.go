package seeds

import (
	"fmt"
	"log"
	"time"

	"terra-allwert/domain/entities"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Seeder struct {
	db *gorm.DB
}

func NewSeeder(db *gorm.DB) *Seeder {
	return &Seeder{db: db}
}

func (s *Seeder) SeedAll() error {
	log.Println("🌱 Starting database seeding...")

	// Auto-migrate tables first
	if err := s.AutoMigrate(); err != nil {
		return fmt.Errorf("failed to auto-migrate tables: %w", err)
	}

	if err := s.SeedEnterprises(); err != nil {
		return fmt.Errorf("failed to seed enterprises: %w", err)
	}

	if err := s.SeedUsers(); err != nil {
		return fmt.Errorf("failed to seed users: %w", err)
	}

	log.Println("✅ Database seeding completed successfully!")
	return nil
}

func (s *Seeder) AutoMigrate() error {
	log.Println("🔄 Running auto-migrations...")

	// First, migrate tables without foreign key references to break circular dependencies
	baseEntities := []any{
		&entities.File{},
		&entities.Enterprise{},
		&entities.User{},
	}

	log.Println("🔧 Migrating base entities...")
	for _, entity := range baseEntities {
		if err := s.db.AutoMigrate(entity); err != nil {
			log.Printf("⚠️  Warning: Failed to migrate %T individually: %v", entity, err)
			// Continue with batch migration as fallback
			break
		}
	}

	// Then migrate all remaining entities
	allEntities := []any{
		&entities.File{},
		&entities.FileVariant{},
		&entities.Enterprise{},
		&entities.User{},
		&entities.Menu{},
		&entities.MenuFloorPlan{},
		&entities.Tower{},
		&entities.Floor{},
		&entities.Suite{},
		&entities.MenuCarousel{},
		&entities.CarouselItem{},
		&entities.CarouselTextOverlay{},
		&entities.MenuPins{},
		&entities.PinMarker{},
		&entities.PinMarkerImage{},
		&entities.AuditLog{},
		&entities.PropertyView{},
	}

	log.Println("🔧 Running full auto-migration...")
	if err := s.db.AutoMigrate(allEntities...); err != nil {
		return fmt.Errorf("auto-migration failed: %w", err)
	}

	log.Println("✅ Auto-migrations completed successfully!")
	return nil
}

func (s *Seeder) SeedEnterprises() error {
	log.Println("🏢 Seeding enterprises...")

	enterprises := []entities.Enterprise{
		{
			Title:               "Terra Allwert",
			Description:         stringPtr("Empreendimento"),
			Slug:                "allwert",
			AddressStreet:       stringPtr("Avenida"),
			AddressNumber:       stringPtr("1500"),
			AddressNeighborhood: stringPtr("Centro"),
			AddressCity:         "Pelotas",
			AddressState:        "SC",
			AddressZipCode:      stringPtr("88330-000"),
			Latitude:            float64Ptr(-26.9900),
			Longitude:           float64Ptr(-48.6350),
			Status:              entities.EnterpriseStatusConstruction,
			CreatedAt:           time.Now(),
		},
	}

	for _, enterprise := range enterprises {
		// Check if enterprise already exists
		var existingEnterprise entities.Enterprise
		if err := s.db.Where("slug = ?", enterprise.Slug).First(&existingEnterprise).Error; err == nil {
			log.Printf("📝 Enterprise '%s' already exists, skipping...", enterprise.Title)
			continue
		}

		if err := s.db.Create(&enterprise).Error; err != nil {
			return fmt.Errorf("failed to create enterprise '%s': %w", enterprise.Title, err)
		}
		log.Printf("✅ Created enterprise: %s", enterprise.Title)
	}

	return nil
}

func (s *Seeder) SeedUsers() error {
	log.Println("👥 Seeding users...")

	// Get enterprises for user assignment
	var enterprises []entities.Enterprise
	if err := s.db.Find(&enterprises).Error; err != nil {
		return fmt.Errorf("failed to fetch enterprises: %w", err)
	}

	if len(enterprises) == 0 {
		return fmt.Errorf("no enterprises found, please seed enterprises first")
	}

	// Create users for each enterprise
	for i, enterprise := range enterprises {
		users := []entities.User{
			{
				EnterpriseID:    enterprise.ID,
				Name:            "Admin " + enterprise.Title,
				Email:           fmt.Sprintf("admin@%s", enterprise.Slug),
				PasswordHash:    hashPassword("senha123"),
				Role:            entities.UserRoleAdmin,
				Phone:           stringPtr(fmt.Sprintf("+55 48 9999-%04d", 1000+i)),
				IsActive:        true,
				EmailVerifiedAt: timePtr(time.Now()),
				CreatedAt:       time.Now(),
			},
			{
				EnterpriseID:    enterprise.ID,
				Name:            "Manager " + enterprise.Title,
				Email:           fmt.Sprintf("manager@%s", enterprise.Slug),
				PasswordHash:    hashPassword("senha123"),
				Role:            entities.UserRoleManager,
				Phone:           stringPtr(fmt.Sprintf("+55 48 9999-%04d", 2000+i)),
				IsActive:        true,
				EmailVerifiedAt: timePtr(time.Now()),
				CreatedAt:       time.Now(),
			},
			{
				EnterpriseID: enterprise.ID,
				Name:         "Visitor " + enterprise.Title,
				Email:        fmt.Sprintf("visitor@%s", enterprise.Slug),
				PasswordHash: hashPassword("senha123"),
				Role:         entities.UserRoleVisitor,
				Phone:        stringPtr(fmt.Sprintf("+55 48 9999-%04d", 3000+i)),
				IsActive:     true,
				CreatedAt:    time.Now(),
			},
		}

		for _, user := range users {
			// Check if user already exists
			var existingUser entities.User
			if err := s.db.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
				log.Printf("📝 User '%s' already exists, skipping...", user.Email)
				continue
			}

			if err := s.db.Create(&user).Error; err != nil {
				return fmt.Errorf("failed to create user '%s': %w", user.Email, err)
			}
			log.Printf("✅ Created user: %s (%s) for %s", user.Name, user.Role, enterprise.Title)
		}
	}

	// Create some additional general users
	generalUsers := []entities.User{
		{
			EnterpriseID:    enterprises[0].ID, // Assign to first enterprise
			Name:            "Super Admin",
			Email:           "admin@terra.com",
			PasswordHash:    hashPassword("senha123"),
			Role:            entities.UserRoleAdmin,
			Phone:           stringPtr("+55 48 99999-0001"),
			IsActive:        true,
			EmailVerifiedAt: timePtr(time.Now()),
			CreatedAt:       time.Now(),
		},
		{
			EnterpriseID: enterprises[0].ID, // Assign to first enterprise
			Name:         "Demo User",
			Email:        "demo@terraallwert.com",
			PasswordHash: hashPassword("senha123"),
			Role:         entities.UserRoleVisitor,
			Phone:        stringPtr("+55 48 99999-0002"),
			IsActive:     true,
			CreatedAt:    time.Now(),
		},
	}

	for _, user := range generalUsers {
		// Check if user already exists
		var existingUser entities.User
		if err := s.db.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
			log.Printf("📝 User '%s' already exists, skipping...", user.Email)
			continue
		}

		if err := s.db.Create(&user).Error; err != nil {
			return fmt.Errorf("failed to create user '%s': %w", user.Email, err)
		}
		log.Printf("✅ Created general user: %s (%s)", user.Name, user.Role)
	}

	return nil
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func hashPassword(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}
	return string(hashedPassword)
}
