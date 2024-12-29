package services

import (
	"context"
	"fmt"
	"time"

	"ledger-link/internal/models"
	"ledger-link/pkg/auth"
	"ledger-link/pkg/logger"
)

const (
	defaultTokenDuration = 24 * time.Hour
)

type AuthService struct {
	userSvc models.UserService
	tokenMaker auth.TokenMaker
	logger *logger.Logger
}

func NewAuthService(
	userSvc models.UserService,
	tokenMaker auth.TokenMaker,
	logger *logger.Logger,
) *AuthService {
	return &AuthService{
		userSvc: userSvc,
		tokenMaker: tokenMaker,
		logger: logger,
	}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.userSvc.Authenticate(ctx, email, password)
	if err != nil {
		return "", fmt.Errorf("failed to authenticate user: %w", err)
	}

	token, err := s.tokenMaker.CreateToken(user.ID, user.Role, defaultTokenDuration)
	if err != nil {
		return "", fmt.Errorf("failed to create token: %w", err)
	}

	return token, nil
}

func (s *AuthService) Register(ctx context.Context, email, password, username string) (string, error) {
	user := &models.User{
		Email:    email,
		Username: username,
	}

	if err := user.SetPassword(password); err != nil {
		return "", fmt.Errorf("failed to set password: %w", err)
	}

	user, err := s.userSvc.Register(ctx, user)
	if err != nil {
		return "", fmt.Errorf("failed to register user: %w", err)
	}

	token, err := s.tokenMaker.CreateToken(user.ID, user.Role, defaultTokenDuration)
	if err != nil {
		return "", fmt.Errorf("failed to create token: %w", err)
	}

	return token, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, token string) (*models.User, error) {
	claims, err := s.tokenMaker.VerifyToken(token)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	user, err := s.userSvc.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, oldToken string) (string, error) {
	user, err := s.ValidateToken(ctx, oldToken)
	if err != nil {
		return "", fmt.Errorf("failed to validate token: %w", err)
	}

	token, err := s.tokenMaker.CreateToken(user.ID, user.Role, defaultTokenDuration)
	if err != nil {
		return "", fmt.Errorf("failed to create token: %w", err)
	}

	return token, nil
}