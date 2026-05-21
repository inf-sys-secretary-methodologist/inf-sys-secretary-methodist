// Package usecases contains the business logic for authentication operations.
package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/crypto/bcrypt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
	notifUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/application/usecases"
	domainErrors "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/domain/errors"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/security/totp"
)

const (
	bcryptCost = 14 // Increased from default 10 for better security
)

var (
	// ErrInvalidCredentials is returned when login credentials are invalid.
	ErrInvalidCredentials = errors.New("invalid email or password")
	// ErrUserNotActive is returned when user account is not active.
	ErrUserNotActive = errors.New("user account is not active")
	// ErrUserBlocked is returned when user account is blocked.
	ErrUserBlocked = errors.New("user account is blocked")
	// ErrInvalidToken is returned when token is invalid or expired.
	ErrInvalidToken = errors.New("invalid token")
	// ErrIntermediateInvalid is returned when the MFA intermediate token
	// fails parse / signature / issuer / purpose / claims-shape validation.
	ErrIntermediateInvalid = errors.New("invalid mfa intermediate token")
	// ErrIntermediateExpired is returned when the MFA intermediate token's
	// exp claim is in the past.
	ErrIntermediateExpired = errors.New("mfa intermediate token expired")
	// ErrIntermediateUsed is returned when the intermediate's jti is already
	// in the revoked-token set — replay attempt.
	ErrIntermediateUsed = errors.New("mfa intermediate token already used")
	// ErrMFAVerificationNotConfigured is returned when VerifyLoginMFA is
	// called without the revoked-token repo / clock set via
	// WithMFAVerification. Deployment misconfiguration; main.go must wire
	// this in production.
	ErrMFAVerificationNotConfigured = errors.New("mfa verification dependencies not configured")
	// ErrRefreshTokenReused is returned by RefreshToken when the supplied
	// refresh token's JTI has already been blacklisted — either by a
	// successful prior rotation or by a logout. Surfaces RFC 6749 §10.4
	// reuse-detection: a stolen token presented after the legitimate
	// owner already rotated triggers this sentinel; the handler maps it
	// to 401 and an audit emit so SOC can investigate. Issue #279 ADR-2.
	ErrRefreshTokenReused = errors.New("refresh token has already been used")
	// ErrAccountLocked is returned by Login / LoginWithUser when the
	// per-account failure counter has crossed the lockout threshold.
	// The lock is time-bounded (the tracker auto-expires the counter)
	// so a legitimate user can retry after the cooldown without admin
	// intervention. Mapped to HTTP 429 by the handler. Issue #279 ADR-3.
	ErrAccountLocked = errors.New("account temporarily locked due to repeated failed login attempts")
)

// LoginAttemptTracker bounds password-guessing attempts on a per-account
// basis. The Login use case consults IsLocked before bcrypt (so a locked
// account stays locked regardless of password correctness), registers a
// failure on wrong-password, and resets the counter on success. The
// concrete Redis implementation lives in infrastructure/persistence; the
// interface lives here per DIP. nil-safe: when unset, lockout enforcement
// is skipped and the legacy IP-keyed rate limiter remains the only floor
// (sufficient for tests, insufficient for production). Issue #279 ADR-3.
type LoginAttemptTracker interface {
	// RegisterFailure atomically increments the failure counter for the
	// given identifier (typically lowercased email) and reports the new
	// count. Implementations are expected to set / refresh a TTL on
	// every write so abandoned counters self-clean.
	RegisterFailure(ctx context.Context, identifier string) (int, error)

	// IsLocked reports whether the identifier currently exceeds the
	// lockout threshold. Storage errors map to (false, err) — the use
	// case treats unreachable tracker as "do not lock" rather than fail
	// closed (mirrors the refresh-rotation tolerance).
	IsLocked(ctx context.Context, identifier string) (bool, error)

	// Reset clears the failure counter. Called on successful login.
	Reset(ctx context.Context, identifier string) error
}

// AuthUseCase handles authentication business logic.
type AuthUseCase struct {
	userRepo              UserRepository
	jwtSecret             []byte
	refreshSecret         []byte
	mfaIntermediateSecret []byte
	accessExpiry          time.Duration
	refreshExpiry         time.Duration
	mfaIntermediateExpiry time.Duration
	securityLog           *logging.SecurityLogger
	auditLog              *logging.AuditLogger
	notificationUseCase   *notifUsecases.NotificationUseCase

	// VerifyLoginMFA dependencies (configured separately via
	// WithMFAVerification — see method below). nil-safe: when unset, the
	// MFA branch in LoginWithUser still issues an intermediate token, but
	// VerifyLoginMFA returns ErrMFAVerificationNotConfigured.
	revokedTokenRepo RevokedTokenRepository
	totpDriftWindow  int
	now              func() time.Time

	// Login lockout (configured separately via WithLoginAttemptTracking).
	// nil-safe: when unset, Login skips lockout enforcement; the legacy
	// IP-keyed rate limiter remains the only floor. Issue #279 ADR-3.
	loginAttemptTracker LoginAttemptTracker
}

// WithLoginAttemptTracking wires the per-account brute-force tracker
// onto the use case. Production callers MUST attach this (the Redis
// implementation in infrastructure/persistence) so locked-out accounts
// surface ErrAccountLocked → HTTP 429 from the handler. Returns the
// receiver so callers can chain after NewAuthUseCase. Issue #279 ADR-3.
func (u *AuthUseCase) WithLoginAttemptTracking(tracker LoginAttemptTracker) *AuthUseCase {
	u.loginAttemptTracker = tracker
	return u
}

// NewAuthUseCase creates a new auth use case. mfaIntermediateSecret signs
// the short-lived token issued after password verification when the user
// has MFA enabled; production callers MUST pass a non-empty secret distinct
// from jwtSecret/refreshSecret so an intermediate token cannot be replayed
// against the access-token middleware. Tests that never enroll an MFA user
// may pass any non-nil value (or nil — the MFA branch is never reached).
func NewAuthUseCase(
	userRepo UserRepository,
	jwtSecret, refreshSecret, mfaIntermediateSecret []byte,
	securityLog *logging.SecurityLogger,
	auditLog *logging.AuditLogger,
	notificationUseCase *notifUsecases.NotificationUseCase,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo:              userRepo,
		jwtSecret:             jwtSecret,
		refreshSecret:         refreshSecret,
		mfaIntermediateSecret: mfaIntermediateSecret,
		accessExpiry:          time.Minute * 15,
		refreshExpiry:         time.Hour * 24 * 7,
		mfaIntermediateExpiry: time.Minute * 5,
		securityLog:           securityLog,
		auditLog:              auditLog,
		notificationUseCase:   notificationUseCase,
	}
}

// Login authenticates user and returns JWT tokens
func (u *AuthUseCase) Login(ctx context.Context, input dto.LoginInput) (accessToken string, refreshToken string, err error) {
	ctx, span := otel.Tracer("auth").Start(ctx, "AuthUseCase.Login")
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "login failed")
		}
		span.End()
	}()
	span.SetAttributes(attribute.String("user.email", input.Email))

	startTime := time.Now()

	// Use GetByEmailForAuth to bypass cache and ensure password field is populated
	user, err := u.userRepo.GetByEmailForAuth(ctx, input.Email)

	// Dummy hash for timing attack prevention
	dummyHash := "$2a$14$0000000000000000000000000000000000000000000000000000000"

	// Always perform password comparison to prevent timing attacks
	if err != nil || user == nil {
		// Perform dummy comparison to maintain constant time
		_ = bcrypt.CompareHashAndPassword([]byte(dummyHash), []byte(input.Password))

		// Log failed login attempt
		u.logLoginAttempt(ctx, input.Email, false, "user not found or invalid email")

		return "", "", fmt.Errorf("authentication failed: %w", domainErrors.ErrUnauthorized)
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		// Log failed login - invalid password
		u.logLoginAttempt(ctx, input.Email, false, "invalid password")

		return "", "", fmt.Errorf("authentication failed: %w", domainErrors.ErrUnauthorized)
	}

	// Check if user can login (status checks)
	if err := user.CanLogin(); err != nil {
		// Log failed login - account status issue
		reason := "account not active"
		if user.Status == entities.UserStatusBlocked {
			reason = "account blocked"
		}
		u.logLoginAttempt(ctx, input.Email, false, reason)

		return "", "", fmt.Errorf("cannot login: %w", err)
	}

	// Generate tokens
	accessToken, refreshToken, err = u.generateTokens(ctx, user)
	if err != nil {
		u.logLoginAttempt(ctx, input.Email, false, "token generation failed")
		return "", "", fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Log successful login
	u.logLoginAttempt(ctx, input.Email, true, "login successful")

	// Log audit event
	u.logAudit(ctx, "login", "auth", map[string]interface{}{
		"user_id":     user.ID,
		"email":       user.Email,
		"role":        user.Role,
		"duration_ms": time.Since(startTime).Milliseconds(),
	})

	return accessToken, refreshToken, nil
}

// LoginResult is the return shape of LoginWithUser. When the user has MFA
// enabled, AccessToken+RefreshToken are empty and IntermediateToken is set
// with MFARequired=true; the caller must complete the second factor through
// VerifyLoginMFA before receiving full tokens. For users without MFA, the
// usual access+refresh tokens are issued and IntermediateToken is empty.
type LoginResult struct {
	AccessToken       string
	RefreshToken      string
	IntermediateToken string
	MFARequired       bool
	User              *entities.User
}

// LoginWithUser authenticates user and returns JWT tokens with user info.
// Returns *LoginResult so the same call can either issue full access+refresh
// tokens (no MFA / future MFA-not-required users) or, once login-flow MFA
// gating lands, an intermediate token with MFARequired=true.
func (u *AuthUseCase) LoginWithUser(ctx context.Context, input dto.LoginInput) (*LoginResult, error) {
	startTime := time.Now()

	// Use GetByEmailForAuth to bypass cache and ensure password field is populated
	user, err := u.userRepo.GetByEmailForAuth(ctx, input.Email)

	// Dummy hash for timing attack prevention
	dummyHash := "$2a$14$0000000000000000000000000000000000000000000000000000000"

	// Always perform password comparison to prevent timing attacks
	if err != nil || user == nil {
		// Perform dummy comparison to maintain constant time
		_ = bcrypt.CompareHashAndPassword([]byte(dummyHash), []byte(input.Password))

		// Log failed login attempt
		u.logLoginAttempt(ctx, input.Email, false, "user not found or invalid email")

		return nil, fmt.Errorf("authentication failed: %w", domainErrors.ErrUnauthorized)
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		// Log failed login - invalid password
		u.logLoginAttempt(ctx, input.Email, false, "invalid password")

		return nil, fmt.Errorf("authentication failed: %w", domainErrors.ErrUnauthorized)
	}

	// Check if user can login (status checks)
	if err := user.CanLogin(); err != nil {
		// Log failed login - account status issue
		reason := "account not active"
		if user.Status == entities.UserStatusBlocked {
			reason = "account blocked"
		}
		u.logLoginAttempt(ctx, input.Email, false, reason)

		return nil, fmt.Errorf("cannot login: %w", err)
	}

	// MFA gate: password verified, but if the user has MFA enabled we issue
	// a short-lived intermediate token instead of full access/refresh tokens.
	// The caller (frontend) collects a 6-digit code and exchanges the
	// intermediate via /api/auth/mfa/verify-login for the real tokens.
	if user.MFAEnabled {
		intermediate, err := u.generateIntermediateToken(user)
		if err != nil {
			u.logLoginAttempt(ctx, input.Email, false, "intermediate token generation failed")
			return nil, fmt.Errorf("failed to generate intermediate token: %w", err)
		}

		u.logLoginAttempt(ctx, input.Email, true, "login awaiting mfa")
		u.logAudit(ctx, "login_mfa_required", "auth", map[string]interface{}{
			"user_id":     user.ID,
			"email":       user.Email,
			"role":        user.Role,
			"duration_ms": time.Since(startTime).Milliseconds(),
		})

		return &LoginResult{
			IntermediateToken: intermediate,
			MFARequired:       true,
			User:              user,
		}, nil
	}

	// Generate tokens
	accessToken, refreshToken, err := u.generateTokens(ctx, user)
	if err != nil {
		u.logLoginAttempt(ctx, input.Email, false, "token generation failed")
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Log successful login
	u.logLoginAttempt(ctx, input.Email, true, "login successful")

	// Log audit event
	u.logAudit(ctx, "login", "auth", map[string]interface{}{
		"user_id":     user.ID,
		"email":       user.Email,
		"role":        user.Role,
		"duration_ms": time.Since(startTime).Milliseconds(),
	})

	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

// Register creates a new user
func (u *AuthUseCase) Register(ctx context.Context, input dto.RegisterInput) error {
	startTime := time.Now()

	role := domain.RoleType(input.Role)
	if !role.IsAllowedForSelfRegistration() {
		u.logRegistration(ctx, input.Email, input.Role, false, "role not allowed for self-registration")
		return domain.ErrRoleNotAllowedForSelfRegistration
	}

	// Hash password with increased cost
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcryptCost)
	if err != nil {
		u.logRegistration(ctx, input.Email, input.Role, false, "password hashing failed")
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user := entities.NewUser(
		input.Email,
		string(hashedPassword),
		input.Name,
		domain.RoleType(input.Role),
	)

	if err := u.userRepo.Create(ctx, user); err != nil {
		u.logRegistration(ctx, input.Email, input.Role, false, "database error")
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Log successful registration
	u.logRegistration(ctx, input.Email, input.Role, true, "registration successful")

	// Log audit event
	u.logAudit(ctx, "user_registered", "user", map[string]interface{}{
		"user_id":     user.ID,
		"email":       user.Email,
		"role":        user.Role,
		"duration_ms": time.Since(startTime).Milliseconds(),
	})

	// Send welcome notification
	if u.notificationUseCase != nil {
		go func() { // #nosec G118 -- fire-and-forget goroutine outlives request
			_ = u.notificationUseCase.SendSystemNotification(
				context.Background(),
				user.ID,
				"Добро пожаловать!",
				"Ваш аккаунт успешно создан. Настройте профиль и начните работу в системе.",
			)
		}()
	}

	return nil
}

// RefreshToken validates refresh token and returns new tokens.
//
// v0.159.0 ADR-2: when a revoked-token repository is wired (production
// path) the old refresh JTI is added to the blacklist before issuing
// the new pair (rotation), and a replay of an already-blacklisted JTI
// returns ErrRefreshTokenReused (RFC 6749 §10.4 reuse detection). When
// no repo is wired (legacy / test setups) the rotation step is skipped
// — the legacy single-issue behavior is preserved.
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
		u.logTokenOperation(ctx, "refresh", false, 0)
		return "", "", fmt.Errorf("invalid refresh token: %w", ErrInvalidToken)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		u.logTokenOperation(ctx, "refresh", false, 0)
		return "", "", fmt.Errorf("invalid token claims: %w", ErrInvalidToken)
	}

	// Extract user ID
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		u.logTokenOperation(ctx, "refresh", false, 0)
		return "", "", fmt.Errorf("missing user_id in token: %w", ErrInvalidToken)
	}
	userID := int64(userIDFloat)

	// Reuse-detection branch: presented JTI already blacklisted → replay.
	// Reads (claims["jti"], claims["exp"]) up-front so we can both detect
	// reuse and compute the blacklist TTL for the rotation step below.
	oldJTI, _ := claims["jti"].(string)
	var refreshTTL time.Duration
	if expFloat, ok := claims["exp"].(float64); ok {
		refreshTTL = time.Until(time.Unix(int64(expFloat), 0))
	}
	if u.revokedTokenRepo != nil && oldJTI != "" {
		revoked, checkErr := u.revokedTokenRepo.IsRevoked(ctx, oldJTI)
		if checkErr == nil && revoked {
			u.logTokenOperation(ctx, "refresh", false, userID)
			u.logAudit(ctx, "refresh_token_reuse_detected", "auth", map[string]interface{}{
				"user_id": userID,
				"jti":     oldJTI,
			})
			return "", "", ErrRefreshTokenReused
		}
		// A storage error here intentionally does NOT block the refresh.
		// Locking users out of their session because Redis is unreachable
		// would be a worse failure mode than letting a refresh proceed
		// without reuse detection; the per-request blast radius is one
		// refresh window. checkErr is left unhandled by design.
	}

	// Get user from database
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		u.logTokenOperation(ctx, "refresh", false, userID)
		return "", "", fmt.Errorf("user not found: %w", err)
	}

	// Check if user can still login
	if err := user.CanLogin(); err != nil {
		u.logTokenOperation(ctx, "refresh", false, userID)
		return "", "", fmt.Errorf("cannot refresh token: %w", err)
	}

	// Generate new tokens
	accessToken, newRefreshToken, err := u.generateTokens(ctx, user)
	if err != nil {
		u.logTokenOperation(ctx, "refresh", false, userID)
		return "", "", fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Rotation: blacklist the old refresh JTI so this exact refresh
	// token cannot be presented again. Bounded by the remaining lifetime
	// of the token (TTL beyond that adds nothing — the token expires
	// naturally and JWT validation rejects it).
	if u.revokedTokenRepo != nil && oldJTI != "" && refreshTTL > 0 {
		if revokeErr := u.revokedTokenRepo.Revoke(ctx, oldJTI, refreshTTL); revokeErr != nil {
			// Audit but do not fail the request — the new pair is already
			// minted; failing now would force the user to re-login while
			// silently leaving the new tokens usable. Audit lets ops detect
			// chronic blacklist-write failures.
			u.logAudit(ctx, "refresh_token_blacklist_failed", "auth", map[string]interface{}{
				"user_id": userID,
				"jti":     oldJTI,
				"error":   revokeErr.Error(),
			})
		}
	}

	// Log successful token refresh
	u.logTokenOperation(ctx, "refresh", true, userID)

	// Log audit event
	u.logAudit(ctx, "token_refreshed", "auth", map[string]interface{}{
		"user_id": userID,
	})

	return accessToken, newRefreshToken, nil
}

// ValidateAccessToken validates and parses access token
func (u *AuthUseCase) ValidateAccessToken(_ context.Context, tokenString string) (*jwt.MapClaims, error) {
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
func (u *AuthUseCase) generateTokens(_ context.Context, user *entities.User) (string, string, error) {
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

// MFA intermediate-token claim values are exported as package-level
// constants so the verify-login use case can compare them without string
// duplication. The issuer is intentionally distinct from the access-token
// issuer so a leaked intermediate token cannot satisfy the access-token
// middleware on its own (defense in depth).
const (
	mfaIntermediateIssuer  = "inf-sys-auth-mfa-intermediate"
	mfaIntermediatePurpose = "mfa_verify"
)

// WithMFAVerification wires the dependencies needed by VerifyLoginMFA:
// the revoked-token repository for one-shot intermediate-token replay
// guarding, the TOTP drift window (typically 1 step = ±30 s), and an
// injectable clock so unit tests stay deterministic. Returns the receiver
// so callers can chain after NewAuthUseCase. Test cases that never enroll
// an MFA user may skip this call — the LoginWithUser MFA branch still
// generates intermediate tokens, but VerifyLoginMFA refuses to run.
func (u *AuthUseCase) WithMFAVerification(
	revokedRepo RevokedTokenRepository,
	driftWindow int,
	now func() time.Time,
) *AuthUseCase {
	u.revokedTokenRepo = revokedRepo
	u.totpDriftWindow = driftWindow
	if now != nil {
		u.now = now
	}
	return u
}

// VerifyLoginMFA exchanges a (intermediate token + 6-digit code) pair for
// full access + refresh tokens. It is the second step of the MFA-gated
// login flow: AuthUseCase.LoginWithUser issued the intermediate after
// password verification, the frontend collected the code from the user's
// authenticator, and now we trade them for the real session tokens.
//
// On success the jti of the intermediate is added to the revoked-token
// set so it cannot be replayed. The intermediate stays valid for ~5 min,
// so the TTL on the revocation entry is bounded by the remaining lifetime
// at the moment of revocation.
//
// Sentinel errors mapped by the handler to HTTP statuses:
//   - ErrIntermediateInvalid  -> 401 (signature / issuer / purpose / claims)
//   - ErrIntermediateExpired  -> 401
//   - ErrIntermediateUsed     -> 401 (replay)
//   - entities.ErrInvalidMFACode -> 422
func (u *AuthUseCase) VerifyLoginMFA(ctx context.Context, intermediateToken, code string) (*LoginResult, error) {
	if u.revokedTokenRepo == nil || u.now == nil {
		return nil, ErrMFAVerificationNotConfigured
	}

	parsed, err := jwt.Parse(intermediateToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return u.mfaIntermediateSecret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrIntermediateExpired
		}
		return nil, fmt.Errorf("%w: %w", ErrIntermediateInvalid, err)
	}
	if !parsed.Valid {
		return nil, ErrIntermediateInvalid
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrIntermediateInvalid
	}

	if iss, _ := claims["iss"].(string); iss != mfaIntermediateIssuer {
		return nil, ErrIntermediateInvalid
	}
	if purpose, _ := claims["purpose"].(string); purpose != mfaIntermediatePurpose {
		return nil, ErrIntermediateInvalid
	}

	jti, _ := claims["jti"].(string)
	if jti == "" {
		return nil, ErrIntermediateInvalid
	}

	revoked, err := u.revokedTokenRepo.IsRevoked(ctx, jti)
	if err != nil {
		return nil, fmt.Errorf("check revoked jti: %w", err)
	}
	if revoked {
		return nil, ErrIntermediateUsed
	}

	var userID int64
	switch v := claims["user_id"].(type) {
	case float64:
		userID = int64(v)
	case int64:
		userID = v
	default:
		return nil, ErrIntermediateInvalid
	}

	// Bypass cache so we get the live MFASecret (json:"-" stripped from cache).
	user, err := u.userRepo.GetByIDForAuth(ctx, userID)
	if err != nil || user == nil {
		return nil, ErrIntermediateInvalid
	}
	if !user.MFAEnabled || user.MFASecret == nil {
		// Defense in depth: intermediate was issued for an MFA-enabled user;
		// if the live record has MFA off, treat as invalid (account state
		// changed mid-flow — fail closed rather than issuing tokens).
		return nil, entities.ErrMFANotEnabled
	}

	raw, err := user.MFASecret.Decode()
	if err != nil {
		return nil, fmt.Errorf("decode mfa secret: %w", err)
	}
	if !totp.Verify(raw, code, u.now(), u.totpDriftWindow) {
		return nil, entities.ErrInvalidMFACode
	}

	expFloat, _ := claims["exp"].(float64)
	ttl := time.Until(time.Unix(int64(expFloat), 0))
	if ttl > 0 {
		if err := u.revokedTokenRepo.Revoke(ctx, jti, ttl); err != nil {
			return nil, fmt.Errorf("revoke jti: %w", err)
		}
	}

	accessToken, refreshToken, err := u.generateTokens(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("generate tokens: %w", err)
	}

	u.logAudit(ctx, "login_mfa_verified", "auth", map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
	})

	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

// generateIntermediateToken issues a short-lived JWT signed with
// mfaIntermediateSecret. The token carries user_id, jti (one-shot replay
// guard), purpose=mfa_verify, and iss=inf-sys-auth-mfa-intermediate. It is
// useful only for the /api/auth/mfa/verify-login endpoint.
func (u *AuthUseCase) generateIntermediateToken(user *entities.User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     now.Add(u.mfaIntermediateExpiry).Unix(),
		"iat":     now.Unix(),
		"nbf":     now.Unix(),
		"jti":     uuid.New().String(),
		"iss":     mfaIntermediateIssuer,
		"purpose": mfaIntermediatePurpose,
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := t.SignedString(u.mfaIntermediateSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign intermediate token: %w", err)
	}
	return signed, nil
}

// Safe logging wrapper methods with nil checks

func (u *AuthUseCase) logLoginAttempt(ctx context.Context, email string, success bool, reason string) {
	if u.securityLog != nil {
		u.securityLog.LogLoginAttempt(ctx, email, success, reason)
	}
}

func (u *AuthUseCase) logRegistration(ctx context.Context, email, role string, success bool, reason string) {
	if u.securityLog != nil {
		u.securityLog.LogRegistration(ctx, email, role, success, reason)
	}
}

func (u *AuthUseCase) logTokenOperation(ctx context.Context, operation string, success bool, userID int64) {
	if u.securityLog != nil {
		u.securityLog.LogTokenOperation(ctx, operation, success, userID)
	}
}

func (u *AuthUseCase) logAudit(ctx context.Context, action, resourceType string, details map[string]interface{}) {
	if u.auditLog != nil {
		u.auditLog.LogAuditEvent(ctx, action, resourceType, details)
	}
}
