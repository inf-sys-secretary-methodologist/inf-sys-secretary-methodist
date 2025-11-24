// Package services contains domain services for the auth module.
package services

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
)

// AuthService defines authentication domain service
type AuthService interface {
	ValidateCredentials(email, password string) (*entities.User, error)
	GenerateToken(user *entities.User) (string, error)
	ValidateToken(token string) (*entities.User, error)
}

// AuthorizationService defines authorization domain service
type AuthorizationService interface {
	CheckPermission(userCtx *entities.UserContext, resource domain.ResourceType, action domain.ActionType, resourceScope *Scope) bool
	CheckOwnership(userID int64, resourceOwnerID int64) bool
	CanApproveDocument(userCtx *entities.UserContext, documentType string, currentStep int) bool
}

// JWTService provides JWT token generation and validation.
type JWTService struct {
	secretKey       string
	refreshSecret   string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

// AuthorizationServiceImpl implements the AuthorizationService interface.
type AuthorizationServiceImpl struct{}

// TokenPair represents a pair of access and refresh tokens.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// Scope represents resource scope for authorization checks.
type Scope struct {
	FacultyID *string `json:"faculty_id,omitempty"`
	GroupID   *string `json:"group_id,omitempty"`
}

// NewJWTService creates a new JWT service instance.
func NewJWTService(secret, refreshSecret string, accessTTL, refreshTTL time.Duration) *JWTService {
	return &JWTService{
		secretKey:       secret,
		refreshSecret:   refreshSecret,
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}
}

// NewAuthorizationService creates a new authorization service instance.
func NewAuthorizationService() AuthorizationService {
	return &AuthorizationServiceImpl{}
}

// GenerateTokens generates access and refresh tokens for a user.
func (s *JWTService) GenerateTokens(userID int64, role string) (*TokenPair, error) {
	accessClaims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(s.accessTokenTTL).Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString([]byte(s.secretKey))
	if err != nil {
		return nil, err
	}

	refreshClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(s.refreshTokenTTL).Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshToken.SignedString([]byte(s.refreshSecret))
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
	}, nil
}

// CheckPermission checks if user has permission to perform an action on a resource.
func (s *AuthorizationServiceImpl) CheckPermission(userCtx *entities.UserContext, resource domain.ResourceType, action domain.ActionType, resourceScope *Scope) bool {
	if userCtx == nil {
		return false
	}

	// Базовая проверка разрешения
	if !userCtx.HasPermission(resource, action) {
		return false
	}

	accessLevel := userCtx.GetAccessLevel(resource, action)

	switch accessLevel {
	case domain.AccessFull:
		return true
	case domain.AccessLimited:
		return s.checkLimitedAccess(userCtx, resourceScope)
	case domain.AccessOwn:
		return s.checkOwnershipInternal(userCtx.UserID, resourceScope)
	default:
		return false
	}
}

func (s *AuthorizationServiceImpl) checkLimitedAccess(userCtx *entities.UserContext, resourceScope *Scope) bool {
	if resourceScope == nil {
		return true
	}

	// Проверка доступа в рамках факультета
	if resourceScope.FacultyID != nil && userCtx.FacultyID != nil {
		return *resourceScope.FacultyID == *userCtx.FacultyID
	}

	// Проверка доступа в рамках группы
	if resourceScope.GroupID != nil && userCtx.GroupID != nil {
		return *resourceScope.GroupID == *userCtx.GroupID
	}

	return true
}

func (s *AuthorizationServiceImpl) checkOwnershipInternal(_ int64, _ *Scope) bool {
	// В реальной системе здесь была бы проверка владения ресурсом
	// Для примера возвращаем true
	return true
}

// CheckOwnership checks if user owns the resource.
func (s *AuthorizationServiceImpl) CheckOwnership(userID int64, resourceOwnerID int64) bool {
	return userID == resourceOwnerID
}

// CanApproveDocument checks if user can approve a document at current step.
func (s *AuthorizationServiceImpl) CanApproveDocument(userCtx *entities.UserContext, documentType string, currentStep int) bool {
	if userCtx == nil {
		return false
	}

	// Логика проверки возможности согласования документа на основе workflow
	switch documentType {
	case "curriculum":
		return s.canApproveCurriculum(userCtx, currentStep)
	case "report":
		return s.canApproveReport(userCtx, currentStep)
	default:
		return false
	}
}

func (s *AuthorizationServiceImpl) canApproveCurriculum(userCtx *entities.UserContext, currentStep int) bool {
	// Упрощенная логика согласования учебных планов
	switch currentStep {
	case 1: // Первичное согласование
		return userCtx.Role == domain.RoleMethodist || userCtx.Role == domain.RoleTeacher
	case 2: // Методическое согласование
		return userCtx.Role == domain.RoleMethodist
	case 3: // Административное согласование
		return userCtx.Role == domain.RoleAcademicSecretary
	case 4: // Утверждение
		return userCtx.Role == domain.RoleSystemAdmin || userCtx.Role == domain.RoleMethodist
	default:
		return false
	}
}

func (s *AuthorizationServiceImpl) canApproveReport(userCtx *entities.UserContext, _ int) bool {
	// Упрощенная логика для отчетов
	return userCtx.HasPermission(domain.ResourceReports, domain.ActionApprove)
}

// HashPassword hashes a password using bcrypt.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash checks if password matches the hash.
func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
