package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"api/domain/entities"
	"api/domain/interfaces"

	"gorm.io/gorm"
)

// SeedData coordena a execução de todas as seeds
func SeedData(db *gorm.DB, authService interfaces.AuthService) error {
	log.Println("🌱 Starting database seeding...")

	// Executar seeds em ordem de dependência
	if err := createInitialUsers(db, authService); err != nil {
		return fmt.Errorf("failed to create initial users: %w", err)
	}

	if err := createInitialAppConfig(db); err != nil {
		return fmt.Errorf("failed to create app config: %w", err)
	}

	if err := createSampleTowers(db); err != nil {
		return fmt.Errorf("failed to create sample towers: %w", err)
	}

	if err := createSampleGallery(db); err != nil {
		return fmt.Errorf("failed to create sample gallery: %w", err)
	}

	log.Println("🌱 Database seeding completed successfully!")
	return nil
}

// createInitialUsers cria usuários iniciais para desenvolvimento
func createInitialUsers(db *gorm.DB, authService interfaces.AuthService) error {
	ctx := context.Background()

	var count int64
	err := db.Model(&entities.User{}).Count(&count).Error
	if err != nil {
		return err
	}

	if count > 0 {
		log.Println("👤 Users already exist, skipping user seed")
		return nil
	}

	users := []struct {
		Username string
		Email    string
		Password string
		Role     entities.UserRole
	}{
		{"admin", "admin@terraallwert.com", "admin123", entities.RoleAdmin},
		{"viewer", "viewer@terraallwert.com", "viewer123", entities.RoleViewer},
		{"admin2", "admin2@terraallwert.com", "admin123", entities.RoleAdmin},
		{"demo", "demo@terraallwert.com", "demo123", entities.RoleViewer},
	}

	for _, userData := range users {
		hashedPassword, err := authService.HashPassword(userData.Password)
		if err != nil {
			return fmt.Errorf("failed to hash password for %s: %w", userData.Username, err)
		}

		user := &entities.User{
			Username: userData.Username,
			Email:    userData.Email,
			Password: hashedPassword,
			Role:     userData.Role,
			Active:   true,
		}

		if err := db.WithContext(ctx).Create(user).Error; err != nil {
			return fmt.Errorf("failed to create user %s: %w", userData.Username, err)
		}

		log.Printf("✅ Created user: %s (%s)", userData.Username, userData.Email)
	}

	log.Println("👤 Sample credentials (login with email):")
	log.Println("   - Admin: admin@terraallwert.com / admin123")
	log.Println("   - Viewer: viewer@terraallwert.com / viewer123")
	log.Println("   - Admin2: admin2@terraallwert.com / admin123")
	log.Println("   - Demo (Viewer): demo@terraallwert.com / demo123")

	return nil
}

// createInitialAppConfig cria configuração inicial da aplicação
func createInitialAppConfig(db *gorm.DB) error {
	var count int64
	err := db.Model(&entities.AppConfig{}).Count(&count).Error
	if err != nil {
		return err
	}

	if count > 0 {
		log.Println("⚙️  App config already exists, skipping config seed")
		return nil
	}

	config := &entities.AppConfig{
		APIBaseURL:         "http://localhost:3000",
		MinioBaseURL:       "http://localhost:9000",
		AppVersion:         "1.0.0-dev",
		CacheControlMaxAge: 3600,
	}

	if err := db.Create(config).Error; err != nil {
		return fmt.Errorf("failed to create app config: %w", err)
	}

	log.Println("⚙️  Initial app configuration created")
	return nil
}

// createSampleTowers cria torres de exemplo com apartamentos
func createSampleTowers(db *gorm.DB) error {
	var count int64
	err := db.Model(&entities.Tower{}).Count(&count).Error
	if err != nil {
		return err
	}

	if count > 0 {
		log.Println("🏢 Towers already exist, skipping tower seed")
		return nil
	}

	// Torre A - Residencial Premium
	towerA := &entities.Tower{
		Name:        "Torre Residencial A",
		Description: stringPtr("Torre residencial com apartamentos de alto padrão, vista para o mar e acabamentos premium."),
	}

	if err := db.Create(towerA).Error; err != nil {
		return fmt.Errorf("failed to create tower A: %w", err)
	}

	// Torre B - Residencial Executivo
	towerB := &entities.Tower{
		Name:        "Torre Executiva B",
		Description: stringPtr("Torre executiva com apartamentos compactos e funcionais, ideal para profissionais."),
	}

	if err := db.Create(towerB).Error; err != nil {
		return fmt.Errorf("failed to create tower B: %w", err)
	}

	// Torre C - Comercial
	towerC := &entities.Tower{
		Name:        "Torre Comercial C",
		Description: stringPtr("Torre comercial com salas e lajes corporativas em localização privilegiada."),
	}

	if err := db.Create(towerC).Error; err != nil {
		return fmt.Errorf("failed to create tower C: %w", err)
	}

	log.Println("🏢 Created sample towers: A, B, C")

	// Criar pavimentos e apartamentos para cada torre
	if err := createFloorsAndApartments(db, towerA, "A", 20, 4); err != nil {
		return fmt.Errorf("failed to create floors for tower A: %w", err)
	}

	if err := createFloorsAndApartments(db, towerB, "B", 15, 6); err != nil {
		return fmt.Errorf("failed to create floors for tower B: %w", err)
	}

	if err := createFloorsAndApartments(db, towerC, "C", 25, 2); err != nil {
		return fmt.Errorf("failed to create floors for tower C: %w", err)
	}

	return nil
}

// createFloorsAndApartments cria pavimentos e apartamentos para uma torre
func createFloorsAndApartments(db *gorm.DB, tower *entities.Tower, towerLetter string, numFloors, aptsPerFloor int) error {
	for floor := 1; floor <= numFloors; floor++ {
		floorEntity := &entities.Floor{
			Number:  fmt.Sprintf("%d", floor),
			TowerID: tower.ID,
		}

		if err := db.Create(floorEntity).Error; err != nil {
			return fmt.Errorf("failed to create floor %d: %w", floor, err)
		}

		// Criar apartamentos para este pavimento
		for apt := 1; apt <= aptsPerFloor; apt++ {
			apartmentNumber := fmt.Sprintf("%s%d%02d", towerLetter, floor, apt)

			var status entities.ApartmentStatus
			var price *float64
			var available bool

			// Distribuir status de forma realística
			switch apt % 4 {
			case 0:
				status = entities.ApartmentStatusSold
				available = false
			case 1:
				status = entities.ApartmentStatusReserved
				available = false
			case 2:
				status = entities.ApartmentStatusMaintenance
				available = false
			default:
				status = entities.ApartmentStatusAvailable
				available = true
			}

			// Preços variados baseados no andar e torre
			if towerLetter == "A" {
				price = float64Ptr(float64(450000 + (floor-1)*15000 + apt*5000))
			} else if towerLetter == "B" {
				price = float64Ptr(float64(320000 + (floor-1)*10000 + apt*3000))
			} else {
				price = float64Ptr(float64(280000 + (floor-1)*8000 + apt*2000))
			}

			apartment := &entities.Apartment{
				Number:        apartmentNumber,
				FloorID:       floorEntity.ID,
				Area:          stringPtr(getApartmentArea(towerLetter, apt)),
				Suites:        intPtr(getSuites(towerLetter, apt)),
				Bedrooms:      intPtr(getBedrooms(towerLetter, apt)),
				ParkingSpots:  intPtr(getParkingSpots(towerLetter)),
				Status:        status,
				SolarPosition: stringPtr(getSolarPosition(apt)),
				Price:         price,
				Available:     available,
			}

			if err := db.Create(apartment).Error; err != nil {
				return fmt.Errorf("failed to create apartment %s: %w", apartmentNumber, err)
			}
		}
	}

	log.Printf("🏠 Created %d floors with %d apartments each for Tower %s", numFloors, aptsPerFloor, towerLetter)
	return nil
}

// createSampleGallery cria galeria de exemplo
func createSampleGallery(db *gorm.DB) error {
	var count int64
	err := db.Model(&entities.GalleryImage{}).Count(&count).Error
	if err != nil {
		return err
	}

	if count > 0 {
		log.Println("🖼️  Gallery already exists, skipping gallery seed")
		return nil
	}

	galleryImages := []struct {
		Route        string
		Title        string
		Description  string
		DisplayOrder int
	}{
		{"home", "Fachada Principal", "Vista frontal do empreendimento com paisagismo", 1},
		{"home", "Área de Lazer", "Piscina e área recreativa completa", 2},
		{"home", "Academia", "Espaço fitness com equipamentos modernos", 3},
		{"apartments", "Apartamento Decorado A301", "Apartamento de 3 quartos torre A", 1},
		{"apartments", "Apartamento Decorado B201", "Apartamento de 2 quartos torre B", 2},
		{"apartments", "Apartamento Cobertura A2001", "Cobertura duplex com terraço", 3},
		{"amenities", "Salão de Festas", "Ambiente para eventos e comemorações", 1},
		{"amenities", "Playground", "Área infantil segura e divertida", 2},
		{"amenities", "Portaria", "Recepção 24 horas com segurança", 3},
	}

	for i, img := range galleryImages {
		galleryImage := &entities.GalleryImage{
			Route:        img.Route,
			ImageURL:     fmt.Sprintf("https://picsum.photos/800/600?random=%d", i+1),
			Title:        stringPtr(img.Title),
			Description:  stringPtr(img.Description),
			DisplayOrder: img.DisplayOrder,
			ImageMetadata: entities.FileMetadata{
				FileName:    fmt.Sprintf("gallery_%d.jpg", i+1),
				FileSize:    int64(150000 + i*10000),
				ContentType: "image/jpeg",
				UploadedAt:  time.Now(),
				Width:       intPtr(800),
				Height:      intPtr(600),
			},
		}

		if err := db.Create(galleryImage).Error; err != nil {
			return fmt.Errorf("failed to create gallery image %s: %w", img.Title, err)
		}
	}

	log.Println("🖼️  Created sample gallery with 9 images")
	return nil
}

// Funções auxiliares para gerar dados realísticos
func getApartmentArea(towerLetter string, aptNumber int) string {
	switch towerLetter {
	case "A":
		areas := []string{"85m²", "95m²", "105m²", "120m²"}
		return areas[(aptNumber-1)%len(areas)]
	case "B":
		areas := []string{"65m²", "75m²", "85m²", "65m²"}
		return areas[(aptNumber-1)%len(areas)]
	default:
		areas := []string{"45m²", "55m²", "45m²", "55m²"}
		return areas[(aptNumber-1)%len(areas)]
	}
}

func getSuites(towerLetter string, aptNumber int) int {
	switch towerLetter {
	case "A":
		return []int{1, 1, 2, 2}[(aptNumber-1)%4]
	case "B":
		return []int{0, 1, 1, 0}[(aptNumber-1)%4]
	default:
		return []int{0, 0, 1, 0}[(aptNumber-1)%4]
	}
}

func getBedrooms(towerLetter string, aptNumber int) int {
	switch towerLetter {
	case "A":
		return []int{2, 3, 3, 4}[(aptNumber-1)%4]
	case "B":
		return []int{1, 2, 2, 1}[(aptNumber-1)%4]
	default:
		return []int{0, 1, 1, 0}[(aptNumber-1)%4] // Commercial spaces
	}
}

func getParkingSpots(towerLetter string) int {
	switch towerLetter {
	case "A":
		return 2
	case "B":
		return 1
	default:
		return 0 // Commercial
	}
}

func getSolarPosition(aptNumber int) string {
	positions := []string{"Norte", "Sul", "Leste", "Oeste"}
	return positions[(aptNumber-1)%4]
}

// Funções utilitárias
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

// CreateInitialUser mantido para compatibilidade
func CreateInitialUser(db *gorm.DB, authService interfaces.AuthService) error {
	return createInitialUsers(db, authService)
}

// ClearAllData limpa todos os dados (útil para testes)
func ClearAllData(db *gorm.DB) error {
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
		if err := db.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error; err != nil {
			return fmt.Errorf("failed to clear table %s: %w", table, err)
		}
	}

	log.Println("🗑️  All data cleared from database")
	return nil
}

// GetSeedStats retorna estatísticas dos dados de seed
func GetSeedStats(db *gorm.DB) (map[string]int64, error) {
	stats := make(map[string]int64)

	entities := map[string]any{
		"users":            &entities.User{},
		"towers":           &entities.Tower{},
		"floors":           &entities.Floor{},
		"apartments":       &entities.Apartment{},
		"apartment_images": &entities.ApartmentImage{},
		"gallery_images":   &entities.GalleryImage{},
		"image_pins":       &entities.ImagePin{},
		"app_config":       &entities.AppConfig{},
	}

	for name, model := range entities {
		var count int64
		if err := db.Model(model).Count(&count).Error; err != nil {
			return nil, fmt.Errorf("failed to count %s: %w", name, err)
		}
		stats[name] = count
	}

	return stats, nil
}
