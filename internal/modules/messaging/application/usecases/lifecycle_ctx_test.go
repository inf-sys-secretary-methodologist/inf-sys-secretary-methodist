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

// TestMessagingUseCase_SystemMessageTexts_RoutedThroughConfig pins
// v0.162.1 polish Item 4: system message content is sourced from the
// configured SystemMessageTexts value, not from string literals
// embedded в the usecase. Verifies all three lifecycle paths
// (group create / join / leave) read the configured field. Table-
// driven per CLAUDE.md ≥3-variant gate.
func TestMessagingUseCase_SystemMessageTexts_RoutedThroughConfig(t *testing.T) {
	cases := []struct {
		name     string
		texts    SystemMessageTexts
		exercise func(t *testing.T, uc *MessagingUseCase, msgRepo *MockMessageRepository, convRepo *MockConversationRepository)
		want     string
	}{
		{
			name:  "group_created",
			texts: SystemMessageTexts{GroupCreated: "Создана группа"},
			exercise: func(t *testing.T, uc *MessagingUseCase, msgRepo *MockMessageRepository, convRepo *MockConversationRepository) {
				t.Helper()
				convRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Conversation")).Return(nil)
				msgRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Message")).Return(nil)
				_, err := uc.CreateGroupConversation(context.Background(), 1, dto.CreateGroupConversationInput{
					Title:          "T",
					ParticipantIDs: []int64{2},
				})
				require.NoError(t, err)
			},
			want: "Создана группа",
		},
		{
			name:  "user_joined",
			texts: SystemMessageTexts{UserJoined: "Пользователь присоединился"},
			exercise: func(t *testing.T, uc *MessagingUseCase, msgRepo *MockMessageRepository, convRepo *MockConversationRepository) {
				t.Helper()
				conv := &entities.Conversation{
					ID:   1,
					Type: entities.ConversationTypeGroup,
					Participants: []entities.Participant{
						{UserID: 1, Role: entities.ParticipantRoleAdmin},
					},
				}
				convRepo.On("GetByID", mock.Anything, int64(1)).Return(conv, nil)
				convRepo.On("AddParticipant", mock.Anything, mock.AnythingOfType("*entities.Participant")).Return(nil)
				msgRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Message")).Return(nil)
				err := uc.AddParticipants(context.Background(), 1, 1, dto.AddParticipantsInput{
					UserIDs: []int64{42},
				})
				require.NoError(t, err)
			},
			want: "Пользователь присоединился",
		},
		{
			name:  "user_left",
			texts: SystemMessageTexts{UserLeft: "Пользователь покинул"},
			exercise: func(t *testing.T, uc *MessagingUseCase, msgRepo *MockMessageRepository, convRepo *MockConversationRepository) {
				t.Helper()
				conv := &entities.Conversation{
					ID:   1,
					Type: entities.ConversationTypeGroup,
					Participants: []entities.Participant{
						{UserID: 1, Role: entities.ParticipantRoleMember},
						{UserID: 2, Role: entities.ParticipantRoleAdmin},
					},
				}
				convRepo.On("GetByID", mock.Anything, int64(1)).Return(conv, nil)
				convRepo.On("RemoveParticipant", mock.Anything, int64(1), int64(1)).Return(nil)
				msgRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Message")).Return(nil)
				err := uc.LeaveConversation(context.Background(), 1, 1)
				require.NoError(t, err)
			},
			want: "Пользователь покинул",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockConvRepo := new(MockConversationRepository)
			mockMsgRepo := new(MockMessageRepository)
			logger := createTestLogger()
			hub := websocket.NewHub(logger)
			uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil).
				WithSystemMessageTexts(tc.texts)

			tc.exercise(t, uc, mockMsgRepo, mockConvRepo)

			// Find the system message create call and assert content
			// matches the configured text.
			found := false
			for _, call := range mockMsgRepo.Calls {
				if call.Method != "Create" {
					continue
				}
				msg, ok := call.Arguments[1].(*entities.Message)
				if !ok || msg.Type != entities.MessageTypeSystem {
					continue
				}
				assert.Equal(t, tc.want, msg.Content)
				found = true
			}
			assert.True(t, found, "expected a system message to be created")
		})
	}
}
