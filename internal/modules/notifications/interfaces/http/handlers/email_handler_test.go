package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/services"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// fakeEmailService is a minimal EmailService fake. Tests substitute
// the SendEmail / SendWelcomeEmail callbacks per case so the happy
// path + service-failure branches can both be exercised без a real
// SMTP transport.
type fakeEmailService struct {
	sendEmail        func(ctx context.Context, req *services.SendEmailRequest) error
	sendWelcomeEmail func(ctx context.Context, recipientEmail, userName string) error
	sendEmailReq     *services.SendEmailRequest
	welcomeEmail     string
	welcomeName      string
}

func (s *fakeEmailService) SendEmail(ctx context.Context, req *services.SendEmailRequest) error {
	s.sendEmailReq = req
	if s.sendEmail != nil {
		return s.sendEmail(ctx, req)
	}
	return nil
}

func (s *fakeEmailService) SendWelcomeEmail(ctx context.Context, recipientEmail, userName string) error {
	s.welcomeEmail = recipientEmail
	s.welcomeName = userName
	if s.sendWelcomeEmail != nil {
		return s.sendWelcomeEmail(ctx, recipientEmail, userName)
	}
	return nil
}

func (s *fakeEmailService) SendPasswordResetEmail(_ context.Context, _, _ string) error {
	return nil
}

func (s *fakeEmailService) SendNotification(_ context.Context, _, _, _ string) error {
	return nil
}

func TestEmailHandler_SendEmail_InvalidJSON(t *testing.T) {
	handler := NewEmailHandler(nil)
	r := gin.New()
	r.POST("/email/send", handler.SendEmail)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/email/send", strings.NewReader(`{bad`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmailHandler_SendEmail_MissingTo(t *testing.T) {
	handler := NewEmailHandler(nil)
	r := gin.New()
	r.POST("/email/send", handler.SendEmail)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/email/send",
		strings.NewReader(`{"subject":"test","body":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmailHandler_SendWelcomeEmail_InvalidJSON(t *testing.T) {
	handler := NewEmailHandler(nil)
	r := gin.New()
	r.POST("/email/welcome", handler.SendWelcomeEmail)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/email/welcome", strings.NewReader(`{bad`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmailHandler_SendWelcomeEmail_MissingFields(t *testing.T) {
	handler := NewEmailHandler(nil)
	r := gin.New()
	r.POST("/email/welcome", handler.SendWelcomeEmail)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/email/welcome",
		strings.NewReader(`{"email":"not-valid"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmailHandler_SendEmail_HappyPath(t *testing.T) {
	svc := &fakeEmailService{}
	handler := NewEmailHandler(svc)
	r := gin.New()
	r.POST("/email/send", handler.SendEmail)

	body := `{
		"to": ["recipient@example.com"],
		"cc": ["cc@example.com"],
		"bcc": ["bcc@example.com"],
		"subject": "Test subject",
		"body": "Test body",
		"is_html": false
	}`

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/email/send", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, svc.sendEmailReq)
	assert.Equal(t, []string{"recipient@example.com"}, svc.sendEmailReq.To)
	assert.Equal(t, "Test subject", svc.sendEmailReq.Subject)
	assert.Equal(t, "Test body", svc.sendEmailReq.Body)
	assert.False(t, svc.sendEmailReq.IsHTML)
	assert.Contains(t, w.Body.String(), "Email отправлен успешно")
}

func TestEmailHandler_SendEmail_HTMLBodyNotSanitized(t *testing.T) {
	// IsHTML=true skips body sanitisation so the rendered email keeps
	// the original markup. Pin via service request inspection.
	svc := &fakeEmailService{}
	handler := NewEmailHandler(svc)
	r := gin.New()
	r.POST("/email/send", handler.SendEmail)

	body := `{
		"to": ["recipient@example.com"],
		"subject": "HTML mail",
		"body": "<p>Hello <strong>world</strong></p>",
		"is_html": true
	}`

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/email/send", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, svc.sendEmailReq)
	assert.True(t, svc.sendEmailReq.IsHTML)
	assert.Contains(t, svc.sendEmailReq.Body, "<strong>")
}

func TestEmailHandler_SendEmail_ServiceError(t *testing.T) {
	svc := &fakeEmailService{
		sendEmail: func(_ context.Context, _ *services.SendEmailRequest) error {
			return errors.New("smtp transport down")
		},
	}
	handler := NewEmailHandler(svc)
	r := gin.New()
	r.POST("/email/send", handler.SendEmail)

	body := `{"to":["recipient@example.com"],"subject":"X","body":"Y"}`

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/email/send", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// MapDomainError на unknown error returns 500 internal error.
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestEmailHandler_SendWelcomeEmail_HappyPath(t *testing.T) {
	svc := &fakeEmailService{}
	handler := NewEmailHandler(svc)
	r := gin.New()
	r.POST("/email/welcome", handler.SendWelcomeEmail)

	body := `{"email":"newuser@example.com","name":"Иван Иванов"}`

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/email/welcome", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "newuser@example.com", svc.welcomeEmail)
	assert.Equal(t, "Иван Иванов", svc.welcomeName)
	assert.Contains(t, w.Body.String(), "Приветственное письмо")
}

func TestEmailHandler_SendWelcomeEmail_ServiceError(t *testing.T) {
	svc := &fakeEmailService{
		sendWelcomeEmail: func(_ context.Context, _, _ string) error {
			return errors.New("smtp transport down")
		},
	}
	handler := NewEmailHandler(svc)
	r := gin.New()
	r.POST("/email/welcome", handler.SendWelcomeEmail)

	body := `{"email":"newuser@example.com","name":"Test User"}`

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/email/welcome", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
