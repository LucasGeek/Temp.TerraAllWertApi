package graph

import (
	"api/domain/interfaces"
	"api/domain/services"
	"api/infra/storage"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	TowerRepo            interfaces.TowerRepository
	FloorRepo            interfaces.FloorRepository  
	ApartmentRepo        interfaces.ApartmentRepository
	GalleryRepo          interfaces.GalleryRepository
	ImagePinRepo         interfaces.ImagePinRepository
	ApartmentImageRepo   interfaces.ApartmentImageRepository
	AppConfigRepo        interfaces.AppConfigRepository
	UserRepo             interfaces.UserRepository
	AuthService          interfaces.AuthService
	StorageService       interfaces.StorageService
	BulkDownloadService  *storage.BulkDownloadService
	FileService          *services.FileService
}
