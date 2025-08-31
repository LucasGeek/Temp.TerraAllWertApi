package services

import (
	"api/data/repositories"
)

type OrderService struct {
	orderRepo   *repositories.OrderRepository
	productRepo *repositories.ProductRepository
}

func NewOrderService(orderRepo *repositories.OrderRepository, productRepo *repositories.ProductRepository) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		productRepo: productRepo,
	}
}

// Add order service methods here