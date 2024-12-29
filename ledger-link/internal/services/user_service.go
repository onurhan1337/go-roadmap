package services

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"ledger-link/internal/models"
	"ledger-link/pkg/logger"
)

type UserService struct {
	repo          models.UserRepository
	balanceService models.BalanceService
	auditSvc      models.AuditService
	logger        *logger.Logger
}

func NewUserService(
	repo models.UserRepository,
	balanceService models.BalanceService,
	auditSvc models.AuditService,
	logger *logger.Logger,
) *UserService {
	return &UserService{
		repo:          repo,
		balanceService: balanceService,
		auditSvc:      auditSvc,
		logger:        logger,
	}
}

func (s *UserService) Create(ctx context.Context, user *models.User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// Create the user first
	if err := s.repo.Create(ctx, user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Create audit log after user is created
	details := fmt.Sprintf("User created with email: %s", user.Email)
	if err := s.auditSvc.LogAction(ctx, models.EntityTypeUser, user.ID, models.ActionCreate, details); err != nil {
		s.logger.Error("failed to log user creation", "error", err)
	}

	// Initialize user's balance
	if err := s.balanceService.UpdateBalance(ctx, user.ID, 0); err != nil {
		s.logger.Error("failed to initialize user balance", "error", err)
	}

	return nil
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return user, nil
}

func (s *UserService) Authenticate(ctx context.Context, email, password string) (*models.User, error) {
	user, err := s.GetByEmail(ctx, email)
	if err != nil {
		if err == models.ErrNotFound {
			return nil, models.ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, models.ErrInvalidCredentials
	}

	details := fmt.Sprintf("User authenticated with email: %s", email)
	if err := s.auditSvc.LogAction(ctx, models.EntityTypeUser, user.ID, models.ActionUpdate, details); err != nil {
		s.logger.Error("failed to log user authentication", "error", err)
	}

	return user, nil
}

func (s *UserService) ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error {
	user, err := s.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return models.ErrInvalidCredentials
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.PasswordHash = string(hashedPassword)
	user.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user password: %w", err)
	}

	details := "Password changed"
	if err := s.auditSvc.LogAction(ctx, models.EntityTypeUser, user.ID, models.ActionUpdate, details); err != nil {
		s.logger.Error("failed to log password change", "error", err)
	}

	return nil
}

func (s *UserService) CanAccessUser(requestingUser *models.User, targetUserID uint) bool {
	return requestingUser.Role == models.RoleAdmin || requestingUser.ID == targetUserID
}

func (s *UserService) GetByID(ctx context.Context, id uint) (*models.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (s *UserService) Update(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	details := fmt.Sprintf("User updated with email: %s", user.Email)
	if err := s.auditSvc.LogAction(ctx, models.EntityTypeUser, user.ID, models.ActionUpdate, details); err != nil {
		s.logger.Error("failed to log user update", "error", err)
	}

	return nil
}

func (s *UserService) Delete(ctx context.Context, id uint) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	details := "User deleted"
	if err := s.auditSvc.LogAction(ctx, models.EntityTypeUser, id, models.ActionDelete, details); err != nil {
		s.logger.Error("failed to log user deletion", "error", err)
	}

	return nil
}

func (s *UserService) GetUsers(ctx context.Context) ([]*models.User, error) {
	return s.repo.GetUsers(ctx)
}

func (s *UserService) IsAdmin(user *models.User) bool {
	return user != nil && user.Role == models.RoleAdmin
}

func (s *UserService) Register(ctx context.Context, user *models.User) (*models.User, error) {
	// Set default role if not specified
	if user.Role == "" {
		user.Role = models.RoleUser
	}

	if err := s.Create(ctx, user); err != nil {
		return nil, err
	}

	// Fetch the complete user object with all relations
	return s.GetByID(ctx, user.ID)
}

func (s *UserService) UpdateProfile(ctx context.Context, user *models.User) error {
	return s.Update(ctx, user)
}