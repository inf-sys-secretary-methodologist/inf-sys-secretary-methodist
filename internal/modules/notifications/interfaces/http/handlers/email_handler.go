// Package http contains HTTP request handlers for the notifications module.
package http

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/services"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/sanitization"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/validation"
)

type EmailHandler struct {
	emailService services.EmailService
	validator    *validation.Validator
	sanitizer    *sanitization.Sanitizer
}

func NewEmailHandler(emailService services.EmailService) *EmailHandler {
	return &EmailHandler{
		emailService: emailService,
		validator:    validation.NewValidator(),
		sanitizer:    sanitization.NewSanitizer(),
	}
}

// SendEmailInput represents the input for sending an email
type SendEmailInput struct {
	To      []string `json:"to" binding:"required,min=1"`
	CC      []string `json:"cc,omitempty"`
	BCC     []string `json:"bcc,omitempty"`
	Subject string   `json:"subject" binding:"required,min=1,max=200"`
	Body    string   `json:"body" binding:"required,min=1"`
	IsHTML  bool     `json:"is_html,omitempty"`
}

// SendEmail handles sending an email
func (h *EmailHandler) SendEmail(c *gin.Context) {
	var input SendEmailInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Неверный формат запроса")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Sanitize inputs
	for i := range input.To {
		input.To[i] = h.sanitizer.SanitizeEmail(input.To[i])
	}
	for i := range input.CC {
		input.CC[i] = h.sanitizer.SanitizeEmail(input.CC[i])
	}
	for i := range input.BCC {
		input.BCC[i] = h.sanitizer.SanitizeEmail(input.BCC[i])
	}
	input.Subject = h.sanitizer.SanitizeString(input.Subject)
	if !input.IsHTML {
		input.Body = h.sanitizer.SanitizeString(input.Body)
	}

	// Validate
	if err := h.validator.Validate(input); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Send email
	ctx := c.Request.Context()
	req := &services.SendEmailRequest{
		To:      input.To,
		CC:      input.CC,
		BCC:     input.BCC,
		Subject: input.Subject,
		Body:    input.Body,
		IsHTML:  input.IsHTML,
	}

	if err := h.emailService.SendEmail(ctx, req); err != nil {
		// Log the actual error for debugging
		log.Printf("[EmailHandler] Send email error: %v", err)
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(gin.H{
		"message": "Email отправлен успешно",
	})
	c.JSON(http.StatusOK, resp)
}

// SendWelcomeEmailInput represents the input for sending a welcome email
type SendWelcomeEmailInput struct {
	Email string `json:"email" binding:"required,email"`
	Name  string `json:"name" binding:"required,min=1,max=100"`
}

// SendWelcomeEmail handles sending a welcome email
func (h *EmailHandler) SendWelcomeEmail(c *gin.Context) {
	var input SendWelcomeEmailInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Неверный формат запроса")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Sanitize inputs
	input.Email = h.sanitizer.SanitizeEmail(input.Email)
	input.Name = h.sanitizer.SanitizeString(input.Name)

	// Validate
	if err := h.validator.Validate(input); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Send welcome email
	ctx := c.Request.Context()
	if err := h.emailService.SendWelcomeEmail(ctx, input.Email, input.Name); err != nil {
		// Log the actual error for debugging
		log.Printf("[EmailHandler] Send welcome email error: %v", err)
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(gin.H{
		"message": "Приветственное письмо отправлено успешно",
	})
	c.JSON(http.StatusOK, resp)
}
