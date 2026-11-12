package usersfx

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type Service struct {
	users Repository

	logger *zap.Logger
}

func New(
	users Repository,
	logger *zap.Logger,
) *Service {
	return &Service{
		users: users,

		logger: logger,
	}
}

func (s *Service) RegisterOrLogin(ctx context.Context, ident Identity) (*User, error) {
	if err := ident.Validate(); err != nil {
		return nil, err
	}

	u, err := s.users.RegisterOrLogin(ctx, ident)
	if err != nil {
		return nil, fmt.Errorf("failed to register or login: %w", err)
	}

	return u, nil
}

func (s *Service) Login(ctx context.Context, ident Identity) (*User, error) {
	if err := ident.Validate(); err != nil {
		return nil, err
	}

	u, err := s.users.Login(ctx, ident)
	if err != nil {
		return nil, fmt.Errorf("failed to login: %w", err)
	}

	return u, nil
}

func (s *Service) GetUser(ctx context.Context, userID string) (*User, error) {
	u, err := s.users.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return u, nil
}
