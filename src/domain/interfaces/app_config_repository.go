package interfaces

import (
	"context"
	
	"api/domain/entities"
)

type AppConfigRepository interface {
	Get(ctx context.Context) (*entities.AppConfig, error)
	Update(ctx context.Context, config *entities.AppConfig) error
	CreateDefault(ctx context.Context) (*entities.AppConfig, error)
}