package services

import (
	"api/data/repositories"
)

type ProductService struct {
	productRepo *repositories.ProductRepository
}

func NewProductService(productRepo *repositories.ProductRepository) *ProductService {
	return &ProductService{
		productRepo: productRepo,
	}
}

// Add product service methods here