package repositories

import (
	"context"

	"api/domain/entities"
	"api/domain/interfaces"

	"gorm.io/gorm"
)

type appConfigRepository struct {
	db *gorm.DB
}

func NewAppConfigRepository(db *gorm.DB) interfaces.AppConfigRepository {
	return &appConfigRepository{db: db}
}

func (r *appConfigRepository) Get(ctx context.Context) (*entities.AppConfig, error) {
	var config entities.AppConfig

	// AppConfig é singleton - sempre ID = 1
	err := r.db.WithContext(ctx).First(&config, 1).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Criar configuração padrão se não existir
			return r.CreateDefault(ctx)
		}
		return nil, err
	}

	return &config, nil
}

func (r *appConfigRepository) Update(ctx context.Context, config *entities.AppConfig) error {
	// AppConfig sempre usa ID = 1 (singleton)
	config.ID = 1

	return r.db.WithContext(ctx).Save(config).Error
}

func (r *appConfigRepository) CreateDefault(ctx context.Context) (*entities.AppConfig, error) {
	config := &entities.AppConfig{
		ID:                 1,
		APIBaseURL:         "http://localhost:3000",
		MinioBaseURL:       "http://localhost:9000",
		AppVersion:         "1.0.0",
		CacheControlMaxAge: 3600,
	}

	err := r.db.WithContext(ctx).Create(config).Error
	if err != nil {
		return nil, err
	}

	return config, nil
}
