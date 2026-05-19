package services

// v0.153.10 Phase 6 backfill — covers NewNotificationNotifier
// constructor + NotifyNewMessage nil-usecase fast-return branch.
// The non-nil branch is DIP-blocked by concrete *NotificationUseCase
// without a narrow port; defer to a refactor release.
// No production change.

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNotificationNotifier_StoresUsecasePointer(t *testing.T) {
	// nil is valid — constructor just stores the pointer (no panic).
	n := NewNotificationNotifier(nil)
	require.NotNil(t, n)
}

func TestNotifyNewMessage_NilUsecaseShortCircuits(t *testing.T) {
	// Defensive nil-check at start of NotifyNewMessage returns nil
	// without dereferencing — so the messaging module stays usable
	// when notifications wiring is absent (e.g. dev/test profiles).
	n := NewNotificationNotifier(nil)
	err := n.NotifyNewMessage(context.Background(), 7, "Иван", "hello", 42, 100)
	assert.NoError(t, err)
}
