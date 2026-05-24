package usecases

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/infrastructure/websocket"
)

// spyMessageNotifier captures the ctx + arg snapshot of NotifyNewMessage
// invocations atomically so the v0.162.1 polish Item 2 test can assert
// the fire-and-forget goroutine uses the registered lifecycle ctx,
// not context.Background(). Distinct от MockMessageNotifier (testify
// mock с mock.Anything for ctx) so existing test setups stay intact.
type spyMessageNotifier struct {
	mu    sync.Mutex
	calls []spyNotifyCall
	count atomic.Int32
}

type spyNotifyCall struct {
	ctx     context.Context
	userID  int64
	content string
}

func (s *spyMessageNotifier) NotifyNewMessage(ctx context.Context, userID int64, _, content string, _, _ int64) error {
	s.mu.Lock()
	s.calls = append(s.calls, spyNotifyCall{ctx: ctx, userID: userID, content: content})
	s.mu.Unlock()
	s.count.Add(1)
	return nil
}

func (s *spyMessageNotifier) snapshot() []spyNotifyCall {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]spyNotifyCall, len(s.calls))
	copy(out, s.calls)
	return out
}

func (s *spyMessageNotifier) waitFor(t *testing.T, n int32) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if s.count.Load() >= n {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for notifier к receive %d calls; got %d", n, s.count.Load())
}

// TestMessagingUseCase_SendMessage_FanOutUsesLifecycleContext pins
// v0.162.1 polish guarantee #2: the SendMessage fan-out goroutine uses
// the ctx registered via WithLifecycleContext, not context.Background().
// Graceful shutdown can therefore cancel in-flight notification sends
// instead of leaking goroutines past server stop. Mirror к
// announcement_usecase_test.go::TestAnnouncementUseCase_Publish_FanOutUsesLifecycleContext.
func TestMessagingUseCase_SendMessage_FanOutUsesLifecycleContext(t *testing.T) {
	type ctxKey struct{}
	sentinel := "v0.162.1-lifecycle-ctx"
	lifecycleCtx := context.WithValue(context.Background(), ctxKey{}, sentinel)

	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	notifier := &spyMessageNotifier{}
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, notifier, nil).
		WithLifecycleContext(lifecycleCtx)

	participant := &entities.Participant{
		ConversationID: 1,
		UserID:         10,
		UserName:       "Sender",
		Role:           entities.ParticipantRoleMember,
	}
	// SendMessage signature: SendMessage(ctx, userID, conversationID, input)
	// — pass userID=10 (sender), conversationID=1.
	mockConvRepo.On("GetParticipant", mock.Anything, int64(1), int64(10)).Return(participant, nil)
	mockMsgRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Message")).Return(nil)
	mockConvRepo.On("GetParticipants", mock.Anything, int64(1)).Return([]entities.Participant{
		{UserID: 10, UserName: "Sender"},    // sender — skipped
		{UserID: 20, UserName: "Recipient"}, // notified
	}, nil)

	// Request ctx — deliberately different value-namespace from lifecycleCtx
	// so the assertion below only passes if the goroutine reads the
	// registered lifecycle ctx, not the request ctx by accident.
	requestCtx := context.Background()
	_, err := uc.SendMessage(requestCtx, 10, 1, dto.SendMessageInput{Content: "Hi"})
	require.NoError(t, err)

	notifier.waitFor(t, 1)
	calls := notifier.snapshot()
	require.Len(t, calls, 1)
	require.NotNil(t, calls[0].ctx)
	got, ok := calls[0].ctx.Value(ctxKey{}).(string)
	require.True(t, ok, "notifier ctx must carry lifecycle sentinel value")
	assert.Equal(t, sentinel, got)
}
