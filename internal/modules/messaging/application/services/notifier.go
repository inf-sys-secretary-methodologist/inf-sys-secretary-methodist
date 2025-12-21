// Package services contains application-level services for the messaging module.
package services

import (
	"context"
	"fmt"

	notifDTO "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/application/dto"
	notifUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/application/usecases"
	notifEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
)

// NotificationNotifier implements MessageNotifier using the notifications module.
type NotificationNotifier struct {
	notificationUseCase *notifUsecases.NotificationUseCase
}

// NewNotificationNotifier creates a new notification notifier.
func NewNotificationNotifier(notificationUseCase *notifUsecases.NotificationUseCase) *NotificationNotifier {
	return &NotificationNotifier{
		notificationUseCase: notificationUseCase,
	}
}

// NotifyNewMessage sends a notification about a new message to the specified user.
func (n *NotificationNotifier) NotifyNewMessage(ctx context.Context, userID int64, senderName, content string, conversationID, messageID int64) error {
	if n.notificationUseCase == nil {
		return nil
	}

	input := &notifDTO.CreateNotificationInput{
		UserID:   userID,
		Type:     notifEntities.NotificationTypeInfo,
		Priority: notifEntities.PriorityNormal,
		Title:    fmt.Sprintf("Новое сообщение от %s", senderName),
		Message:  content,
		Link:     fmt.Sprintf("/messages/%d", conversationID),
		Metadata: map[string]any{
			"conversation_id": conversationID,
			"message_id":      messageID,
			"sender_name":     senderName,
		},
	}

	_, err := n.notificationUseCase.Create(ctx, input)
	return err
}
