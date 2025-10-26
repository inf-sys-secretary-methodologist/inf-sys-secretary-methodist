package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

const (
	bcryptCost = 14 // Increased from default 10 for better security
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotActive      = errors.New("user account is not active")
	ErrUserBlocked        = errors.New("user account is blocked")
	ErrInvalidToken       = errors.New("invalid token")
)

type AuthUseCase struct {
	userRepo      repositories.UserRepository
	jwtSecret     []byte
	refreshSecret []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
	securityLog   *logging.SecurityLogger
	auditLog      *logging.AuditLogger
}

// NewAuthUseCase creates a new auth use case
func NewAuthUseCase(
	userRepo repositories.UserRepository,
	jwtSecret, refreshSecret []byte,
	securityLog *logging.SecurityLogger,
	auditLog *logging.AuditLogger,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo:      userRepo,
		jwtSecret:     jwtSecret,
		refreshSecret: refreshSecret,
		accessExpiry:  time.Minute * 15,
		refreshExpiry: time.Hour * 24 * 7,
		securityLog:   securityLog,
		auditLog:      auditLog,
	}
}

// Login authenticates user and returns JWT tokens
func (u *AuthUseCase) Login(ctx context.Context, input dto.LoginInput) (accessToken string, refreshToken string, err error) {
	startTime := time.Now()

	user, err := u.userRepo.GetByEmail(ctx, input.Email)

	// Dummy hash for timing attack prevention
	dummyHash := "$2a$14$0000000000000000000000000000000000000000000000000000000"

	// Always perform password comparison to prevent timing attacks
	if err != nil || user == nil {
		// Perform dummy comparison to maintain constant time
		bcrypt.CompareHashAndPassword([]byte(dummyHash), []byte(input.Password))

		// Log failed login attempt
		u.securityLog.LogLoginAttempt(ctx, input.Email, false, "user not found or invalid email")

		return "", "", fmt.Errorf("authentication failed: %w", ErrInvalidCredentials)
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		// Log failed login - invalid password
		u.securityLog.LogLoginAttempt(ctx, input.Email, false, "invalid password")

		return "", "", fmt.Errorf("authentication failed: %w", ErrInvalidCredentials)
	}

	// Check if user can login (status checks)
	if err := user.CanLogin(); err != nil {
		// Log failed login - account status issue
		reason := "account not active"
		if user.Status == entities.UserStatusBlocked {
			reason = "account blocked"
		}
		u.securityLog.LogLoginAttempt(ctx, input.Email, false, reason)

		return "", "", fmt.Errorf("cannot login: %w", err)
	}

	// Generate tokens
	accessToken, refreshToken, err = u.generateTokens(ctx, user)
	if err != nil {
		u.securityLog.LogLoginAttempt(ctx, input.Email, false, "token generation failed")
		return "", "", fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Log successful login
	u.securityLog.LogLoginAttempt(ctx, input.Email, true, "login successful")

	// Log audit event
	u.auditLog.LogAuditEvent(ctx, "login", "auth", map[string]interface{}{
		"user_id":     user.ID,
		"email":       user.Email,
		"role":        user.Role,
		"duration_ms": time.Since(startTime).Milliseconds(),
	})

	return accessToken, refreshToken, nil
}

// Register creates a new user
func (u *AuthUseCase) Register(ctx context.Context, input dto.RegisterInput) error {
	startTime := time.Now()

	// Hash password with increased cost
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcryptCost)
	if err != nil {
		u.securityLog.LogRegistration(ctx, input.Email, input.Role, false, "password hashing failed")
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user := entities.NewUser(
		input.Email,
		string(hashedPassword),
		"", // Name can be set later
		entities.UserRole(input.Role),
	)

	if err := u.userRepo.Create(ctx, user); err != nil {
		u.securityLog.LogRegistration(ctx, input.Email, input.Role, false, "database error")
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Log successful registration
	u.securityLog.LogRegistration(ctx, input.Email, input.Role, true, "registration successful")

	// Log audit event
	u.auditLog.LogAuditEvent(ctx, "user_registered", "user", map[string]interface{}{
		"user_id":     user.ID,
		"email":       user.Email,
		"role":        user.Role,
		"duration_ms": time.Since(startTime).Milliseconds(),
	})

	return nil
}

// RefreshToken validates refresh token and returns new tokens
func (u *AuthUseCase) RefreshToken(ctx context.Context, refreshTokenString string) (string, string, error) {
	// Parse and validate refresh token
	token, err := jwt.Parse(refreshTokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return u.refreshSecret, nil
	})

	if err != nil || !token.Valid {
		u.securityLog.LogTokenOperation(ctx, "refresh", false, 0)
		return "", "", fmt.Errorf("invalid refresh token: %w", ErrInvalidToken)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		u.securityLog.LogTokenOperation(ctx, "refresh", false, 0)
		return "", "", fmt.Errorf("invalid token claims: %w", ErrInvalidToken)
	}

	// Extract user ID
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		u.securityLog.LogTokenOperation(ctx, "refresh", false, 0)
		return "", "", fmt.Errorf("missing user_id in token: %w", ErrInvalidToken)
	}
	userID := int64(userIDFloat)

	// Get user from database
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		u.securityLog.LogTokenOperation(ctx, "refresh", false, userID)
		return "", "", fmt.Errorf("user not found: %w", err)
	}

	// Check if user can still login
	if err := user.CanLogin(); err != nil {
		u.securityLog.LogTokenOperation(ctx, "refresh", false, userID)
		return "", "", fmt.Errorf("cannot refresh token: %w", err)
	}

	// Generate new tokens
	accessToken, newRefreshToken, err := u.generateTokens(ctx, user)
	if err != nil {
		u.securityLog.LogTokenOperation(ctx, "refresh", false, userID)
		return "", "", fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Log successful token refresh
	u.securityLog.LogTokenOperation(ctx, "refresh", true, userID)

	// Log audit event
	u.auditLog.LogAuditEvent(ctx, "token_refreshed", "auth", map[string]interface{}{
		"user_id": userID,
	})

	return accessToken, newRefreshToken, nil
}

// ValidateAccessToken validates and parses access token
func (u *AuthUseCase) ValidateAccessToken(ctx context.Context, tokenString string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return u.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid access token: %w", ErrInvalidToken)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims: %w", ErrInvalidToken)
	}

	return &claims, nil
}

// generateTokens creates access and refresh tokens with all security claims
func (u *AuthUseCase) generateTokens(ctx context.Context, user *entities.User) (string, string, error) {
	now := time.Now()

	// Access Token with all security claims
	atClaims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     now.Add(u.accessExpiry).Unix(),
		"iat":     now.Unix(),
		"nbf":     now.Unix(),
		"jti":     uuid.New().String(),
		"iss":     "inf-sys-auth",
		"aud":     "inf-sys-api",
	}
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	accessToken, err := at.SignedString(u.jwtSecret)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	// Refresh Token
	rtClaims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     now.Add(u.refreshExpiry).Unix(),
		"iat":     now.Unix(),
		"jti":     uuid.New().String(),
		"iss":     "inf-sys-auth",
	}
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	refreshToken, err := rt.SignedString(u.refreshSecret)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}
