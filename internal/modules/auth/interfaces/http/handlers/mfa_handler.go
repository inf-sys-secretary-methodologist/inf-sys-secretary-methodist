package http

import (
	"context"

	"github.com/gin-gonic/gin"
)

// MFAService is the narrow interface MFAHandler depends on; the concrete
// implementation is *usecases.MFAUseCase. Keeping it local to the handler
// keeps test mocks small and follows DIP.
type MFAService interface {
	BeginEnrollment(ctx context.Context, userID int64) (otpAuthURI string, secret string, err error)
	ConfirmEnrollment(ctx context.Context, userID int64, code string) error
	Disable(ctx context.Context, userID int64, code string) error
}

// MFAHandler exposes /api/auth/mfa/{begin,confirm,disable} endpoints.
type MFAHandler struct {
	svc MFAService
}

// NewMFAHandler builds a handler. Panics on nil dependency (failure-closed).
func NewMFAHandler(svc MFAService) *MFAHandler {
	if svc == nil {
		panic("mfa handler: svc is nil")
	}
	return &MFAHandler{svc: svc}
}

// Begin handles POST /api/auth/mfa/begin.
func (h *MFAHandler) Begin(_ *gin.Context) {
	// RED stub
}

// Confirm handles POST /api/auth/mfa/confirm with body {"code": "123456"}.
func (h *MFAHandler) Confirm(_ *gin.Context) {
	// RED stub
}

// Disable handles POST /api/auth/mfa/disable with body {"code": "123456"}.
func (h *MFAHandler) Disable(_ *gin.Context) {
	// RED stub
}
