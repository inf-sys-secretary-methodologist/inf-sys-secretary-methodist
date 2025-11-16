package services

import (
	"context"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/services"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/composio"
)

// ComposioEmailService implements EmailService using Composio
type ComposioEmailService struct {
	client   *composio.Client
	entityID string // User ID for Composio authentication
}

// NewComposioEmailService creates a new email service using Composio
func NewComposioEmailService(apiKey, entityID string) services.EmailService {
	return &ComposioEmailService{
		client:   composio.NewClient(apiKey),
		entityID: entityID,
	}
}

// SendEmail sends an email to one or more recipients
func (s *ComposioEmailService) SendEmail(ctx context.Context, req *services.SendEmailRequest) error {
	if len(req.To) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}

	// Composio supports sending to one recipient at a time
	// For multiple recipients, we'll send to the first one and use CC for others
	recipientEmail := req.To[0]
	cc := req.CC
	if len(req.To) > 1 {
		cc = append(cc, req.To[1:]...)
	}

	emailReq := &composio.SendEmailRequest{
		RecipientEmail: recipientEmail,
		Subject:        req.Subject,
		Body:           req.Body,
		CC:             cc,
		BCC:            req.BCC,
		IsHTML:         req.IsHTML,
	}

	_, err := s.client.SendEmail(ctx, s.entityID, emailReq)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// SendWelcomeEmail sends a welcome email to a new user
func (s *ComposioEmailService) SendWelcomeEmail(ctx context.Context, recipientEmail, userName string) error {
	subject := "Добро пожаловать в систему Секретарь-Методист!"
	body := fmt.Sprintf(`
		<html>
			<body>
				<h2>Добро пожаловать, %s!</h2>
				<p>Ваш аккаунт успешно создан в системе Секретарь-Методист.</p>
				<p>Теперь вы можете войти в систему и начать работу.</p>
				<br>
				<p>С уважением,<br>Команда Секретарь-Методист</p>
			</body>
		</html>
	`, userName)

	req := &services.SendEmailRequest{
		To:      []string{recipientEmail},
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	}

	return s.SendEmail(ctx, req)
}

// SendPasswordResetEmail sends a password reset email
func (s *ComposioEmailService) SendPasswordResetEmail(ctx context.Context, recipientEmail, resetToken string) error {
	subject := "Сброс пароля - Секретарь-Методист"
	// В продакшне здесь должна быть ссылка на frontend с токеном
	resetURL := fmt.Sprintf("http://localhost:3000/reset-password?token=%s", resetToken)

	body := fmt.Sprintf(`
		<html>
			<body>
				<h2>Запрос на сброс пароля</h2>
				<p>Вы запросили сброс пароля для вашего аккаунта.</p>
				<p>Перейдите по ссылке ниже для создания нового пароля:</p>
				<p><a href="%s">Сбросить пароль</a></p>
				<p>Если вы не запрашивали сброс пароля, просто проигнорируйте это письмо.</p>
				<p>Ссылка действительна в течение 1 часа.</p>
				<br>
				<p>С уважением,<br>Команда Секретарь-Методист</p>
			</body>
		</html>
	`, resetURL)

	req := &services.SendEmailRequest{
		To:      []string{recipientEmail},
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	}

	return s.SendEmail(ctx, req)
}

// SendNotification sends a generic notification email
func (s *ComposioEmailService) SendNotification(ctx context.Context, recipientEmail, subject, body string) error {
	req := &services.SendEmailRequest{
		To:      []string{recipientEmail},
		Subject: subject,
		Body:    body,
		IsHTML:  false,
	}

	return s.SendEmail(ctx, req)
}
