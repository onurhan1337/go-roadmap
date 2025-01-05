package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ledger-link/internal/models"
	"ledger-link/pkg/auth"
	"ledger-link/pkg/logger"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	defaultTokenDuration = 24 * time.Hour
)

var (
	authAttempts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ledger_auth_attempts_total",
			Help: "Total number of authentication attempts",
		},
		[]string{"method", "status"},
	)

	activeUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "ledger_active_sessions",
			Help: "Number of currently active user sessions",
		},
	)

	authDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ledger_auth_duration_seconds",
			Help:    "Authentication operation duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	authErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ledger_auth_errors_total",
			Help: "Total number of authentication errors by type",
		},
		[]string{"operation", "error_type"},
	)
)

type AuthService struct {
	userSvc    models.UserService
	tokenMaker auth.TokenMaker
	logger     *logger.Logger
	balanceSvc *BalanceService
}

func NewAuthService(
	userSvc models.UserService,
	tokenMaker auth.TokenMaker,
	logger *logger.Logger,
	balanceSvc *BalanceService,
) *AuthService {
	svc := &AuthService{
		userSvc:    userSvc,
		tokenMaker: tokenMaker,
		logger:     logger,
		balanceSvc: balanceSvc,
	}

	// Initialize active users count
	svc.InitializeActiveUsers(context.Background())

	return svc
}

func (s *AuthService) InitializeActiveUsers(ctx context.Context) error {
	users, err := s.userSvc.GetUsers(ctx)
	if err != nil {
		return err
	}

	// Reset the counter
	activeUsers.Set(0)

	// Count active users (you might want to adjust this logic based on your definition of "active")
	activeCount := float64(len(users))
	activeUsers.Add(activeCount)

	return nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	timer := prometheus.NewTimer(authDuration.WithLabelValues("login"))
	defer timer.ObserveDuration()

	s.logger.Info("Login attempt", "email", email)

	user, err := s.userSvc.Authenticate(ctx, email, password)
	if err != nil {
		if err == models.ErrInvalidCredentials {
			authErrors.WithLabelValues("login", "invalid_credentials").Inc()
		} else {
			authErrors.WithLabelValues("login", "internal_error").Inc()
		}
		authAttempts.WithLabelValues("login", "failure").Inc()
		return "", err
	}

	userJSON, _ := json.Marshal(user)
	s.logger.Info("User authenticated", "user", string(userJSON))

	token, err := s.tokenMaker.CreateToken(user.ID, user.Role, defaultTokenDuration)
	if err != nil {
		authErrors.WithLabelValues("login", "token_creation").Inc()
		return "", err
	}

	s.logger.Info("Token created successfully", "user_id", user.ID)
	activeUsers.Inc()
	authAttempts.WithLabelValues("login", "success").Inc()
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

	// Create initial balance for the user
	balance := &models.Balance{
		UserID:        user.ID,
		Amount:        0,
		LastUpdatedAt: time.Now(),
	}

	if err := s.balanceSvc.CreateInitialBalance(ctx, balance); err != nil {
		s.logger.Error("failed to create initial balance", "error", err, "userID", user.ID)
	}

	s.logger.Info("Token created successfully", "user_id", user.ID)
	return token, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, token string) (*models.User, error) {
	timer := prometheus.NewTimer(authDuration.WithLabelValues("validate"))
	defer timer.ObserveDuration()

	s.logger.Info("Validating token")

	claims, err := s.tokenMaker.VerifyToken(token)
	if err != nil {
		authErrors.WithLabelValues("validate", "invalid_token").Inc()
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

func (s *AuthService) Logout(ctx context.Context) error {
	timer := prometheus.NewTimer(authDuration.WithLabelValues("logout"))
	defer timer.ObserveDuration()

	if err := s.invalidateSession(ctx); err != nil {
		authErrors.WithLabelValues("logout", "session_invalidation").Inc()
		return err
	}

	activeUsers.Dec()
	return nil
}

func (s *AuthService) invalidateSession(ctx context.Context) error {
	userID := auth.GetUserIDFromContext(ctx)
	if userID == 0 {
		return fmt.Errorf("no user in context")
	}

	return nil
}
