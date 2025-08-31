package services

import (
	"context"

	"api/domain/interfaces"
)

type UserService struct {
	userRepo interfaces.UserRepository
}

func NewUserService(userRepo interfaces.UserRepository) *UserService {
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