// Package http contains HTTP request handlers for the auth module.
package http

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	authEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/interfaces/http/messages"
	emailServices "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/services"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/sanitization"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/validation"
)

// AuthHandler handles HTTP requests for authentication endpoints.
type AuthHandler struct {
	usecase      *usecases.AuthUseCase
	emailService emailServices.EmailService
	validator    *validation.Validator
	sanitizer    *sanitization.Sanitizer
}

// NewAuthHandler creates a new authentication handler.
func NewAuthHandler(usecase *usecases.AuthUseCase, emailService emailServices.EmailService) *AuthHandler {
	return &AuthHandler{
		usecase:      usecase,
		emailService: emailService,
		validator:    validation.NewValidator(),
		sanitizer:    sanitization.NewSanitizer(),
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var input dto.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Неверный формат запроса")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Sanitize inputs
	input.Name = h.sanitizer.SanitizeString(input.Name)
	input.Email = h.sanitizer.SanitizeEmail(input.Email)
	input.Role = h.sanitizer.SanitizeString(input.Role)

	// Additional validation with custom rules
	if err := h.validator.Validate(input); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	if err := h.usecase.Register(ctx, input); err != nil {
		if errors.Is(err, domain.ErrRoleNotAllowedForSelfRegistration) {
			c.JSON(http.StatusForbidden, response.Forbidden(messages.RoleNotAllowedForSelfRegistration))
			return
		}
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	// Send welcome email if email service is available
	if h.emailService != nil {
		// Use background context with timeout for async email sending
		emailCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // #nosec G118 -- fire-and-forget goroutine outlives request
		go func() {
			defer cancel()
			if err := h.emailService.SendWelcomeEmail(emailCtx, input.Email, input.Name); err != nil {
				log.Printf("[AuthHandler] Failed to send welcome email to %s: %v", input.Email, err)
			} else {
				log.Printf("[AuthHandler] Welcome email sent successfully to %s", input.Email)
			}
		}()
	}

	// Auto-login after successful registration
	loginInput := dto.LoginInput{
		Email:    input.Email,
		Password: input.Password,
	}

	result, err := h.usecase.LoginWithUser(ctx, loginInput)
	if err != nil {
		// Registration succeeded but auto-login failed
		resp := response.Success(gin.H{"message": "Пользователь успешно зарегистрирован. Пожалуйста, войдите."})
		c.JSON(http.StatusCreated, resp)
		return
	}

	resp := response.Success(gin.H{
		"token":        result.AccessToken,
		"refreshToken": result.RefreshToken,
		"user": gin.H{
			"id":          result.User.ID,
			"email":       result.User.Email,
			"name":        result.User.Name,
			"role":        result.User.Role,
			"mfa_enabled": result.User.MFAEnabled,
			"created_at":  result.User.CreatedAt,
			"updated_at":  result.User.UpdatedAt,
		},
	})
	c.JSON(http.StatusCreated, resp)
}

// Login handles user authentication
func (h *AuthHandler) Login(c *gin.Context) {
	var input dto.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Неверный формат запроса")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Sanitize inputs
	input.Email = h.sanitizer.SanitizeEmail(input.Email)

	// Additional validation with custom rules
	if err := h.validator.Validate(input); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	result, err := h.usecase.LoginWithUser(ctx, input)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	// MFA gate: when result.MFARequired the user must finish login by
	// posting the issued intermediate_token + 6-digit code to
	// /api/auth/mfa/verify-login. Withhold access/refresh until then.
	if result.MFARequired {
		resp := response.Success(gin.H{
			"mfa_required":       true,
			"intermediate_token": result.IntermediateToken,
			"user": gin.H{
				"id":          result.User.ID,
				"email":       result.User.Email,
				"name":        result.User.Name,
				"role":        result.User.Role,
				"mfa_enabled": result.User.MFAEnabled,
				"created_at":  result.User.CreatedAt,
				"updated_at":  result.User.UpdatedAt,
			},
		})
		c.JSON(http.StatusOK, resp)
		return
	}

	resp := response.Success(gin.H{
		"token":        result.AccessToken,
		"refreshToken": result.RefreshToken,
		"user": gin.H{
			"id":          result.User.ID,
			"email":       result.User.Email,
			"name":        result.User.Name,
			"role":        result.User.Role,
			"mfa_enabled": result.User.MFAEnabled,
			"created_at":  result.User.CreatedAt,
			"updated_at":  result.User.UpdatedAt,
		},
	})
	c.JSON(http.StatusOK, resp)
}

// VerifyMFALogin completes the MFA-gated login flow. Accepts the
// intermediate_token issued by Login when MFA is enabled plus the 6-digit
// TOTP code, and on success returns the full access+refresh pair. Status
// mapping (driven by sentinel errors from VerifyLoginMFA):
//   - 200: success
//   - 400: malformed body / non-numeric or non-6-digit code
//   - 401: invalid / expired / replayed intermediate
//   - 422: TOTP code mismatch
//   - 500: unexpected
func (h *AuthHandler) VerifyMFALogin(c *gin.Context) {
	var input struct {
		IntermediateToken string `json:"intermediate_token" binding:"required"`
		Code              string `json:"code" binding:"required,len=6,numeric"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Неверный формат запроса")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	result, err := h.usecase.VerifyLoginMFA(ctx, input.IntermediateToken, input.Code)
	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrIntermediateInvalid),
			errors.Is(err, usecases.ErrIntermediateExpired),
			errors.Is(err, usecases.ErrIntermediateUsed):
			resp := response.Unauthorized("Сессия MFA недействительна или истекла")
			c.JSON(http.StatusUnauthorized, resp)
			return
		case errors.Is(err, authEntities.ErrInvalidMFACode),
			errors.Is(err, authEntities.ErrMFANotEnabled):
			resp := response.ErrorResponse("INVALID_MFA_CODE", "Неверный код подтверждения")
			c.JSON(http.StatusUnprocessableEntity, resp)
			return
		default:
			httpErr := response.MapDomainError(err)
			c.JSON(httpErr.Status, httpErr.Response)
			return
		}
	}

	resp := response.Success(gin.H{
		"token":        result.AccessToken,
		"refreshToken": result.RefreshToken,
		"user": gin.H{
			"id":          result.User.ID,
			"email":       result.User.Email,
			"name":        result.User.Name,
			"role":        result.User.Role,
			"mfa_enabled": result.User.MFAEnabled,
			"created_at":  result.User.CreatedAt,
			"updated_at":  result.User.UpdatedAt,
		},
	})
	c.JSON(http.StatusOK, resp)
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var input dto.RefreshTokenInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Неверный формат запроса")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Sanitize inputs
	input.RefreshToken = h.sanitizer.SanitizeString(input.RefreshToken)

	// Additional validation with custom rules
	if err := h.validator.Validate(input); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	accessToken, refreshToken, err := h.usecase.RefreshToken(ctx, input.RefreshToken)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
	c.JSON(http.StatusOK, resp)
}
