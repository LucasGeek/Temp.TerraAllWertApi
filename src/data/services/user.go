package services

import (
	"context"

	"api/data/repositories"
)

type UserService struct {
	userRepo *repositories.UserRepository
}

func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) HasUsers(ctx context.Context) (bool, error) {
	// Placeholder implementation
	return false, nil
}

func (s *UserService) LoadInitialData(ctx context.Context) error {
	// Placeholder implementation
	return nil
}