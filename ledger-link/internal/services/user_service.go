package services

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"ledger-link/internal/models"
	"ledger-link/pkg/logger"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/shopspring/decimal"
)

var (
	userOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ledger_user_operations_total",
			Help: "Total number of user operations",
		},
		[]string{"operation", "status"}, // operation: create/update/delete, status: success/failure
	)

	userCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "ledger_total_users",
			Help: "Total number of registered users",
		},
	)

	userOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ledger_user_operation_duration_seconds",
			Help:    "User operation duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	userErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ledger_user_errors_total",
			Help: "Total number of user operation errors by type",
		},
		[]string{"operation", "error_type"},
	)

	userProfileUpdates = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ledger_user_profile_updates_total",
			Help: "Total number of user profile updates by field",
		},
		[]string{"field"},
	)
)

type UserService struct {
	repo           models.UserRepository
	balanceService models.BalanceService
	auditSvc       models.AuditService
	logger         *logger.Logger
}

func NewUserService(
	repo models.UserRepository,
	balanceService models.BalanceService,
	auditSvc models.AuditService,
	logger *logger.Logger,
) *UserService {
	return &UserService{
		repo:           repo,
		balanceService: balanceService,
		auditSvc:       auditSvc,
		logger:         logger,
	}
}

func (s *UserService) Create(ctx context.Context, user *models.User) error {
	timer := prometheus.NewTimer(userOperationDuration.WithLabelValues("create"))
	defer timer.ObserveDuration()

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	if err := s.repo.Create(ctx, user); err != nil {
		userErrors.WithLabelValues("create", "database").Inc()
		userOperations.WithLabelValues("create", "failure").Inc()
		return fmt.Errorf("failed to create user: %w", err)
	}

	userCount.Inc()
	userOperations.WithLabelValues("create", "success").Inc()

	initialBalance := &models.Balance{
		UserID:        user.ID,
		Amount:        decimal.NewFromInt(0),
		LastUpdatedAt: time.Now(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.balanceService.CreateInitialBalance(ctx, initialBalance); err != nil {
		s.logger.Error("failed to create initial balance", "error", err, "userID", user.ID)
	}

	details := fmt.Sprintf("User created with email: %s", user.Email)
	if err := s.auditSvc.LogAction(ctx, models.EntityTypeUser, user.ID, models.ActionCreate, details); err != nil {
		s.logger.Error("failed to log user creation", "error", err, "userID", user.ID)
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
	timer := prometheus.NewTimer(userOperationDuration.WithLabelValues("update"))
	defer timer.ObserveDuration()

	if err := user.Validate(); err != nil {
		userErrors.WithLabelValues("update", "validation").Inc()
		userOperations.WithLabelValues("update", "failure").Inc()
		return err
	}

	if err := s.repo.Update(ctx, user); err != nil {
		userErrors.WithLabelValues("update", "database").Inc()
		userOperations.WithLabelValues("update", "failure").Inc()
		return fmt.Errorf("failed to update user: %w", err)
	}

	userOperations.WithLabelValues("update", "success").Inc()

	details := fmt.Sprintf("User updated with email: %s", user.Email)
	if err := s.auditSvc.LogAction(ctx, models.EntityTypeUser, user.ID, models.ActionUpdate, details); err != nil {
		s.logger.Error("failed to log user update", "error", err)
	}

	return nil
}

func (s *UserService) Delete(ctx context.Context, id uint) error {
	timer := prometheus.NewTimer(userOperationDuration.WithLabelValues("delete"))
	defer timer.ObserveDuration()

	if err := s.repo.Delete(ctx, id); err != nil {
		userErrors.WithLabelValues("delete", "database").Inc()
		userOperations.WithLabelValues("delete", "failure").Inc()
		return fmt.Errorf("failed to delete user: %w", err)
	}

	userCount.Dec()
	userOperations.WithLabelValues("delete", "success").Inc()

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
	userProfileUpdates.WithLabelValues("profile").Inc()
	return s.Update(ctx, user)
}
