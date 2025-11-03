package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/sanitization"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/validation"
)

type AuthHandler struct {
	usecase   *usecases.AuthUseCase
	validator *validation.Validator
	sanitizer *sanitization.Sanitizer
}

func NewAuthHandler(usecase *usecases.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		usecase:   usecase,
		validator: validation.NewValidator(),
		sanitizer: sanitization.NewSanitizer(),
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var input dto.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Invalid request format")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Sanitize inputs
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
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(gin.H{"message": "User registered successfully"})
	c.JSON(http.StatusCreated, resp)
}

// Login handles user authentication
func (h *AuthHandler) Login(c *gin.Context) {
	var input dto.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Invalid request format")
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
	accessToken, refreshToken, err := h.usecase.Login(ctx, input)
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

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var input dto.RefreshTokenInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Invalid request format")
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
