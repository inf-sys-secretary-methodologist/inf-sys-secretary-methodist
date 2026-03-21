package websocket

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

func testLogger() *logging.Logger {
	return logging.NewLogger("error")
}

func TestNewHub(t *testing.T) {
	hub := NewHub(testLogger())
	assert.NotNil(t, hub)
	assert.NotNil(t, hub.clients)
	assert.NotNil(t, hub.conversations)
}

func TestHub_RegisterUnregister(t *testing.T) {
	hub := NewHub(testLogger())
	go hub.Run()

	client := &Client{
		hub:    hub,
		send:   make(chan []byte, 256),
		userID: 1,
		logger: testLogger(),
	}

	hub.Register(client)
	time.Sleep(50 * time.Millisecond)

	assert.True(t, hub.IsUserOnline(1))
	users := hub.GetOnlineUsers()
	assert.Contains(t, users, int64(1))

	hub.Unregister(client)
	time.Sleep(50 * time.Millisecond)

	assert.False(t, hub.IsUserOnline(1))
}

func TestHub_SubscribeUnsubscribe(t *testing.T) {
	hub := NewHub(testLogger())
	go hub.Run()

	client := &Client{
		hub:    hub,
		send:   make(chan []byte, 256),
		userID: 1,
		logger: testLogger(),
	}

	hub.Register(client)
	time.Sleep(50 * time.Millisecond)

	hub.Subscribe(client, 100)
	time.Sleep(50 * time.Millisecond)

	hub.mu.RLock()
	_, exists := hub.conversations[100]
	hub.mu.RUnlock()
	assert.True(t, exists)

	hub.Unsubscribe(client, 100)
	time.Sleep(50 * time.Millisecond)

	hub.Unregister(client)
	time.Sleep(50 * time.Millisecond)
}

func TestHub_BroadcastToConversation(t *testing.T) {
	hub := NewHub(testLogger())
	go hub.Run()

	client1 := &Client{hub: hub, send: make(chan []byte, 256), userID: 1, logger: testLogger()}
	client2 := &Client{hub: hub, send: make(chan []byte, 256), userID: 2, logger: testLogger()}

	hub.Register(client1)
	hub.Register(client2)
	time.Sleep(50 * time.Millisecond)

	hub.Subscribe(client1, 100)
	hub.Subscribe(client2, 100)
	time.Sleep(50 * time.Millisecond)

	event := &Event{
		Type:           EventTypeNewMessage,
		ConversationID: 100,
		UserID:         1,
	}
	hub.BroadcastToConversation(100, event, 1) // exclude user 1
	time.Sleep(50 * time.Millisecond)

	// client2 should receive the message
	select {
	case msg := <-client2.send:
		assert.NotEmpty(t, msg)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("client2 did not receive message")
	}

	// client1 should NOT receive the message (excluded)
	select {
	case <-client1.send:
		t.Fatal("client1 should not receive message")
	case <-time.After(100 * time.Millisecond):
		// expected
	}

	hub.Unregister(client1)
	hub.Unregister(client2)
	time.Sleep(50 * time.Millisecond)
}

func TestHub_SendToUser(t *testing.T) {
	hub := NewHub(testLogger())
	go hub.Run()

	client := &Client{hub: hub, send: make(chan []byte, 256), userID: 1, logger: testLogger()}
	hub.Register(client)
	time.Sleep(50 * time.Millisecond)

	event := &Event{Type: EventTypeUserOnline, UserID: 1}
	hub.SendToUser(1, event)

	select {
	case msg := <-client.send:
		assert.NotEmpty(t, msg)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("did not receive message")
	}

	// Send to non-existent user - should not panic
	hub.SendToUser(999, event)

	hub.Unregister(client)
	time.Sleep(50 * time.Millisecond)
}

func TestHub_IsUserOnline(t *testing.T) {
	hub := NewHub(testLogger())
	assert.False(t, hub.IsUserOnline(1))
}

func TestHub_GetOnlineUsers_Empty(t *testing.T) {
	hub := NewHub(testLogger())
	users := hub.GetOnlineUsers()
	assert.Empty(t, users)
}

func TestHub_BroadcastUserStatus(t *testing.T) {
	hub := NewHub(testLogger())
	go hub.Run()

	client := &Client{hub: hub, send: make(chan []byte, 256), userID: 1, logger: testLogger()}
	hub.Register(client)
	time.Sleep(50 * time.Millisecond)

	hub.Subscribe(client, 100)
	time.Sleep(50 * time.Millisecond)

	hub.BroadcastUserStatus(1, true)
	// Just verify it doesn't panic -- messages go to conversation members

	hub.BroadcastUserStatus(1, false)

	hub.Unregister(client)
	time.Sleep(50 * time.Millisecond)
}

func TestEventTypes(t *testing.T) {
	assert.Equal(t, EventType("new_message"), EventTypeNewMessage)
	assert.Equal(t, EventType("message_updated"), EventTypeMessageUpdated)
	assert.Equal(t, EventType("message_deleted"), EventTypeMessageDeleted)
	assert.Equal(t, EventType("typing"), EventTypeTyping)
	assert.Equal(t, EventType("stop_typing"), EventTypeStopTyping)
	assert.Equal(t, EventType("read"), EventTypeRead)
	assert.Equal(t, EventType("user_online"), EventTypeUserOnline)
	assert.Equal(t, EventType("user_offline"), EventTypeUserOffline)
	assert.Equal(t, EventType("conversation_updated"), EventTypeConvUpdated)
}

func TestClient_UserID(t *testing.T) {
	c := &Client{userID: 42}
	assert.Equal(t, int64(42), c.UserID())
}

func TestClient_Send(t *testing.T) {
	c := &Client{send: make(chan []byte, 256), userID: 1, logger: testLogger()}
	c.Send([]byte("hello"))
	msg := <-c.send
	assert.Equal(t, []byte("hello"), msg)
}

func TestClient_Send_BufferFull(t *testing.T) {
	c := &Client{send: make(chan []byte, 1), userID: 1, logger: testLogger()}
	c.send <- []byte("fill")
	// This should not block
	c.Send([]byte("overflow"))
}

func TestClient_handleMessage(t *testing.T) {
	hub := NewHub(testLogger())
	go hub.Run()

	client := &Client{hub: hub, send: make(chan []byte, 256), userID: 1, logger: testLogger()}
	hub.Register(client)
	time.Sleep(50 * time.Millisecond)

	// Test ping
	client.handleMessage(&ClientMessage{Type: "ping"})
	select {
	case msg := <-client.send:
		var result map[string]string
		_ = json.Unmarshal(msg, &result)
		assert.Equal(t, "pong", result["type"])
	case <-time.After(200 * time.Millisecond):
		t.Fatal("no pong response")
	}

	// Test subscribe
	client.handleMessage(&ClientMessage{Type: "subscribe", ConversationID: 100})
	time.Sleep(50 * time.Millisecond)

	// Test typing
	client.handleMessage(&ClientMessage{Type: "typing", ConversationID: 100})
	time.Sleep(50 * time.Millisecond)

	// Test stop_typing
	client.handleMessage(&ClientMessage{Type: "stop_typing", ConversationID: 100})
	time.Sleep(50 * time.Millisecond)

	// Test unsubscribe
	client.handleMessage(&ClientMessage{Type: "unsubscribe", ConversationID: 100})
	time.Sleep(50 * time.Millisecond)

	// Test unknown message type
	client.handleMessage(&ClientMessage{Type: "unknown"})

	// Test subscribe with no conversation ID
	client.handleMessage(&ClientMessage{Type: "subscribe", ConversationID: 0})
	client.handleMessage(&ClientMessage{Type: "unsubscribe", ConversationID: 0})
	client.handleMessage(&ClientMessage{Type: "typing", ConversationID: 0})
	client.handleMessage(&ClientMessage{Type: "stop_typing", ConversationID: 0})

	hub.Unregister(client)
	time.Sleep(50 * time.Millisecond)
}

func TestNewClient(t *testing.T) {
	hub := NewHub(testLogger())
	c := NewClient(hub, nil, 42, testLogger())
	assert.NotNil(t, c)
	assert.Equal(t, int64(42), c.userID)
	assert.NotNil(t, c.send)
}
