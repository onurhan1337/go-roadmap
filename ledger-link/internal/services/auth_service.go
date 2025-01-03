package services

import (
	"context"
	"encoding/json"
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
	userSvc    models.UserService
	tokenMaker auth.TokenMaker
	logger     *logger.Logger
}

func NewAuthService(
	userSvc models.UserService,
	tokenMaker auth.TokenMaker,
	logger *logger.Logger,
) *AuthService {
	return &AuthService{
		userSvc:    userSvc,
		tokenMaker: tokenMaker,
		logger:     logger,
	}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	s.logger.Info("Login attempt", "email", email)

	user, err := s.userSvc.Authenticate(ctx, email, password)
	if err != nil {
		s.logger.Error("Authentication failed", "error", err)
		return "", fmt.Errorf("failed to authenticate user: %w", err)
	}

	userJSON, _ := json.Marshal(user)
	s.logger.Info("User authenticated", "user", string(userJSON))

	token, err := s.tokenMaker.CreateToken(user.ID, user.Role, defaultTokenDuration)
	if err != nil {
		s.logger.Error("Token creation failed", "error", err)
		return "", fmt.Errorf("failed to create token: %w", err)
	}

	s.logger.Info("Token created successfully", "user_id", user.ID)
	return token, nil
}

func (s *AuthService) Register(ctx context.Context, email, password, username string) (string, error) {
	s.logger.Info("Registration attempt", "email", email, "username", username)

	user := &models.User{
		Email:    email,
		Username: username,
		Role:     models.RoleUser, // Explicitly set role
	}

	if err := user.SetPassword(password); err != nil {
		s.logger.Error("Password hashing failed", "error", err)
		return "", fmt.Errorf("failed to set password: %w", err)
	}

	user, err := s.userSvc.Register(ctx, user)
	if err != nil {
		s.logger.Error("Registration failed", "error", err)
		return "", fmt.Errorf("failed to register user: %w", err)
	}

	userJSON, _ := json.Marshal(user)
	s.logger.Info("User registered", "user", string(userJSON))

	token, err := s.tokenMaker.CreateToken(user.ID, user.Role, defaultTokenDuration)
	if err != nil {
		s.logger.Error("Token creation failed", "error", err)
		return "", fmt.Errorf("failed to create token: %w", err)
	}

	s.logger.Info("Token created successfully", "user_id", user.ID)
	return token, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, token string) (*models.User, error) {
	s.logger.Info("Validating token")

	claims, err := s.tokenMaker.VerifyToken(token)
	if err != nil {
		s.logger.Error("Token verification failed", "error", err)
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	s.logger.Info("Token claims", "user_id", claims.UserID, "role", claims.Role)

	user, err := s.userSvc.GetByID(ctx, claims.UserID)
	if err != nil {
		s.logger.Error("Failed to get user", "error", err)
		return nil, err
	}

	// Ensure role from token matches user's role
	if user.Role != claims.Role {
		s.logger.Error("Role mismatch", "token_role", claims.Role, "user_role", user.Role)
		user.Role = claims.Role // Use role from token as it's more up to date
	}

	userJSON, _ := json.Marshal(user)
	s.logger.Info("Token validated", "user", string(userJSON))

	return user, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, oldToken string) (string, error) {
	s.logger.Info("Token refresh attempt")

	user, err := s.ValidateToken(ctx, oldToken)
	if err != nil {
		s.logger.Error("Token validation failed", "error", err)
		return "", fmt.Errorf("failed to validate token: %w", err)
	}

	token, err := s.tokenMaker.CreateToken(user.ID, user.Role, defaultTokenDuration)
	if err != nil {
		s.logger.Error("Token creation failed", "error", err)
		return "", fmt.Errorf("failed to create token: %w", err)
	}

	s.logger.Info("Token refreshed successfully", "user_id", user.ID)
	return token, nil
}
