package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ledger-link/internal/models"
	"ledger-link/pkg/auth"
	"ledger-link/pkg/logger"
)

type AuthService struct {
	userSvc    *UserService
	tokenMaker auth.TokenMaker
	logger     *logger.Logger
}

func NewAuthService(
	userSvc *UserService,
	tokenMaker auth.TokenMaker,
	logger *logger.Logger,
) *AuthService {
	return &AuthService{
		userSvc:    userSvc,
		tokenMaker: tokenMaker,
		logger:     logger,
	}
}

func (s *AuthService) Register(ctx context.Context, input models.RegisterInput) (*models.AuthResponse, error) {
	user := &models.User{
		Username: input.Username,
		Email:    input.Email,
		Role:     models.RoleUser,
	}

	if err := user.SetPassword(input.Password); err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.userSvc.Register(ctx, user); err != nil {
		return nil, err
	}

	token, refreshToken, err := s.generateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &models.AuthResponse{
		User:         user.SafeCopy(),
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, input models.LoginInput) (*models.AuthResponse, error) {
	user, err := s.userSvc.Authenticate(ctx, input.Email, input.Password)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return nil, models.ErrInvalidCredentials
		}
		return nil, err
	}

	token, refreshToken, err := s.generateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &models.AuthResponse{
		User:         user.SafeCopy(),
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*models.AuthResponse, error) {
	claims, err := s.tokenMaker.VerifyToken(refreshToken)
	if err != nil {
		return nil, models.ErrInvalidToken
	}

	user, err := s.userSvc.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	token, newRefreshToken, err := s.generateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &models.AuthResponse{
		User:         user.SafeCopy(),
		Token:        token,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *AuthService) generateTokenPair(user *models.User) (token, refreshToken string, err error) {
	token, err = s.tokenMaker.CreateToken(user.ID, user.Role, 15*time.Minute)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = s.tokenMaker.CreateToken(user.ID, user.Role, 7*24*time.Hour)
	if err != nil {
		return "", "", err
	}

	return token, refreshToken, nil
}