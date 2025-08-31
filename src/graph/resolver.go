package graph

import (
	"api/domain/interfaces"
	"api/infra/storage"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	TowerRepo     interfaces.TowerRepository
	FloorRepo     interfaces.FloorRepository  
	ApartmentRepo interfaces.ApartmentRepository
	GalleryRepo   interfaces.GalleryRepository
	UserRepo      interfaces.UserRepository
	StorageService interfaces.StorageService
	BulkDownloadService *storage.BulkDownloadService
}
