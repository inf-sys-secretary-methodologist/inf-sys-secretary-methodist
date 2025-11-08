package services

import (
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"

	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthService defines authentication domain service
type AuthService interface {
	ValidateCredentials(email, password string) (*entities.User, error)
	GenerateToken(user *entities.User) (string, error)
	ValidateToken(token string) (*entities.User, error)
}

// AuthorizationService defines authorization domain service
type AuthorizationService interface {
	CheckPermission(userCtx *entities.UserContext, resource entities.ResourceType, action entities.ActionType, resourceScope *Scope) bool
	CheckOwnership(userID string, resourceOwnerID string) bool
	CanApproveDocument(userCtx *entities.UserContext, documentType string, currentStep int) bool
}

type JWTService struct {
	secretKey       string
	refreshSecret   string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

type AuthorizationServiceImpl struct{}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type Scope struct {
	FacultyID *string `json:"faculty_id,omitempty"`
	GroupID   *string `json:"group_id,omitempty"`
}

func NewJWTService(secret, refreshSecret string, accessTTL, refreshTTL time.Duration) *JWTService {
	return &JWTService{
		secretKey:       secret,
		refreshSecret:   refreshSecret,
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}
}

func NewAuthorizationService() AuthorizationService {
	return &AuthorizationServiceImpl{}
}

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

func (s *AuthorizationServiceImpl) CheckPermission(userCtx *entities.UserContext, resource entities.ResourceType, action entities.ActionType, resourceScope *Scope) bool {
	if userCtx == nil {
		return false
	}

	// Базовая проверка разрешения
	if !userCtx.HasPermission(resource, action) {
		return false
	}

	accessLevel := userCtx.GetAccessLevel(resource, action)

	switch accessLevel {
	case entities.AccessFull:
		return true
	case entities.AccessLimited:
		return s.checkLimitedAccess(userCtx, resourceScope)
	case entities.AccessOwn:
		return s.checkOwnership(userCtx.UserID, resourceScope)
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

func (s *AuthorizationServiceImpl) checkOwnership(userID string, resourceScope *Scope) bool {
	// В реальной системе здесь была бы проверка владения ресурсом
	// Для примера возвращаем true
	return true
}

func (s *AuthorizationServiceImpl) CheckOwnership(userID string, resourceOwnerID string) bool {
	return userID == resourceOwnerID
}

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
		return userCtx.Role == entities.RoleMethodist || userCtx.Role == entities.RoleTeacher
	case 2: // Методическое согласование
		return userCtx.Role == entities.RoleMethodist
	case 3: // Административное согласование
		return userCtx.Role == entities.RoleSecretary
	case 4: // Утверждение
		return userCtx.Role == entities.RoleAdmin || userCtx.Role == entities.RoleMethodist
	default:
		return false
	}
}

func (s *AuthorizationServiceImpl) canApproveReport(userCtx *entities.UserContext, currentStep int) bool {
	// Упрощенная логика для отчетов
	return userCtx.HasPermission(entities.ResourceReports, entities.ActionApprove)
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
