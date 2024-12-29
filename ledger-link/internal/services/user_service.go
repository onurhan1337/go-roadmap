package services

import (
	"context"
	"errors"
	"fmt"

	"ledger-link/internal/models"
	"ledger-link/pkg/logger"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailExists       = errors.New("email already exists")
	ErrUsernameExists    = errors.New("username already exists")
	ErrUnauthorized      = errors.New("unauthorized")
)

type UserService struct {
	repo      models.UserRepository
	balanceSvc models.BalanceService
	auditSvc   models.AuditService
	logger     *logger.Logger
}

func NewUserService(
	repo models.UserRepository,
	balanceSvc models.BalanceService,
	auditSvc models.AuditService,
	logger *logger.Logger,
) *UserService {
	return &UserService{
		repo:      repo,
		balanceSvc: balanceSvc,
		auditSvc:   auditSvc,
		logger:     logger,
	}
}

func (s *UserService) Register(ctx context.Context, user *models.User) error {
	if existing, _ := s.repo.GetByEmail(ctx, user.Email); existing != nil {
		return ErrEmailExists
	}

	if existing, _ := s.repo.GetByUsername(ctx, user.Username); existing != nil {
		return ErrUsernameExists
	}

	if user.Role == "" {
		user.Role = models.RoleUser
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	if err := s.balanceSvc.UpdateBalance(ctx, user.ID, 0); err != nil {
		s.logger.Error("failed to initialize user balance", "error", err, "user_id", user.ID)
	}

	details := fmt.Sprintf("User registered: %s (%s)", user.Username, user.Email)
	if err := s.auditSvc.LogAction(ctx, models.EntityTypeUser, user.ID, models.ActionCreate, details); err != nil {
		s.logger.Error("failed to log user registration", "error", err)
	}

	return nil
}

func (s *UserService) Authenticate(ctx context.Context, email, password string) (*models.User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	details := fmt.Sprintf("User authenticated: %s", user.Email)
	if err := s.auditSvc.LogAction(ctx, models.EntityTypeUser, user.ID, "login", details); err != nil {
		s.logger.Error("failed to log user authentication", "error", err)
	}

	return user.SafeCopy(), nil
}

func (s *UserService) UpdateProfile(ctx context.Context, user *models.User) error {
	existing, err := s.repo.GetByID(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	existing.Username = user.Username
	existing.Email = user.Email

	if err := s.repo.Update(ctx, existing); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	details := fmt.Sprintf("User profile updated: %s", user.Username)
	if err := s.auditSvc.LogAction(ctx, models.EntityTypeUser, user.ID, models.ActionUpdate, details); err != nil {
		s.logger.Error("failed to log profile update", "error", err)
	}

	return nil
}

func (s *UserService) ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return ErrInvalidCredentials
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.PasswordHash = string(hashedPassword)
	if err := s.repo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	details := "Password changed"
	if err := s.auditSvc.LogAction(ctx, models.EntityTypeUser, user.ID, models.ActionUpdate, details); err != nil {
		s.logger.Error("failed to log password change", "error", err)
	}

	return nil
}

func (s *UserService) IsAdmin(user *models.User) bool {
	return user != nil && user.Role == models.RoleAdmin
}

func (s *UserService) CanAccessUser(requestingUser *models.User, targetUserID uint) bool {
	return requestingUser != nil && (requestingUser.ID == targetUserID || s.IsAdmin(requestingUser))
}

func (s *UserService) GetByID(ctx context.Context, id uint) (*models.User, error) {
	return s.repo.GetByID(ctx, id)
}