package websocket

import (
	"encoding/json"
	"sync"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// EventType represents the type of WebSocket event.
type EventType string

const (
	EventTypeNewMessage     EventType = "new_message"
	EventTypeMessageUpdated EventType = "message_updated"
	EventTypeMessageDeleted EventType = "message_deleted"
	EventTypeTyping         EventType = "typing"
	EventTypeStopTyping     EventType = "stop_typing"
	EventTypeRead           EventType = "read"
	EventTypeUserOnline     EventType = "user_online"
	EventTypeUserOffline    EventType = "user_offline"
	EventTypeConvUpdated    EventType = "conversation_updated"
)

// Event represents a WebSocket event.
type Event struct {
	Type           EventType   `json:"type"`
	ConversationID int64       `json:"conversation_id,omitempty"`
	UserID         int64       `json:"user_id,omitempty"`
	Payload        interface{} `json:"payload,omitempty"`
}

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	// Registered clients by user ID
	clients map[int64]map[*Client]bool

	// Registered clients by conversation ID
	conversations map[int64]map[*Client]bool

	// Inbound messages from the clients
	broadcast chan *BroadcastMessage

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Subscribe client to conversation
	subscribe chan *Subscription

	// Unsubscribe client from conversation
	unsubscribe chan *Subscription

	// Mutex for thread safety
	mu sync.RWMutex

	// Logger
	logger *logging.Logger
}

// BroadcastMessage represents a message to broadcast.
type BroadcastMessage struct {
	ConversationID int64
	ExcludeUserID  int64
	Event          *Event
}

// Subscription represents a client subscription to a conversation.
type Subscription struct {
	Client         *Client
	ConversationID int64
}

// NewHub creates a new Hub.
func NewHub(logger *logging.Logger) *Hub {
	return &Hub{
		clients:       make(map[int64]map[*Client]bool),
		conversations: make(map[int64]map[*Client]bool),
		broadcast:     make(chan *BroadcastMessage, 256),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		subscribe:     make(chan *Subscription),
		unsubscribe:   make(chan *Subscription),
		logger:        logger,
	}
}

// Run starts the hub's main loop.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.userID] == nil {
				h.clients[client.userID] = make(map[*Client]bool)
			}
			h.clients[client.userID][client] = true
			h.mu.Unlock()

			h.logger.Info("client registered", map[string]interface{}{
				"user_id": client.userID,
			})

			// Broadcast user online status
			h.BroadcastUserStatus(client.userID, true)

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.userID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.send)
					if len(clients) == 0 {
						delete(h.clients, client.userID)
					}
				}
			}
			// Remove from all conversations
			for convID, clients := range h.conversations {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					if len(clients) == 0 {
						delete(h.conversations, convID)
					}
				}
			}
			h.mu.Unlock()

			h.logger.Info("client unregistered", map[string]interface{}{
				"user_id": client.userID,
			})

			// Broadcast user offline status if no more clients
			h.mu.RLock()
			hasClients := len(h.clients[client.userID]) > 0
			h.mu.RUnlock()
			if !hasClients {
				h.BroadcastUserStatus(client.userID, false)
			}

		case sub := <-h.subscribe:
			h.mu.Lock()
			if h.conversations[sub.ConversationID] == nil {
				h.conversations[sub.ConversationID] = make(map[*Client]bool)
			}
			h.conversations[sub.ConversationID][sub.Client] = true
			h.mu.Unlock()

			h.logger.Debug("client subscribed to conversation", map[string]interface{}{
				"user_id":         sub.Client.userID,
				"conversation_id": sub.ConversationID,
			})

		case sub := <-h.unsubscribe:
			h.mu.Lock()
			if clients, ok := h.conversations[sub.ConversationID]; ok {
				delete(clients, sub.Client)
				if len(clients) == 0 {
					delete(h.conversations, sub.ConversationID)
				}
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			clients := h.conversations[msg.ConversationID]
			h.mu.RUnlock()

			data, err := json.Marshal(msg.Event)
			if err != nil {
				h.logger.Error("failed to marshal event", map[string]interface{}{
					"error": err.Error(),
				})
				continue
			}

			for client := range clients {
				// Skip the sender if specified
				if msg.ExcludeUserID > 0 && client.userID == msg.ExcludeUserID {
					continue
				}

				select {
				case client.send <- data:
				default:
					h.mu.Lock()
					close(client.send)
					delete(h.clients[client.userID], client)
					delete(h.conversations[msg.ConversationID], client)
					h.mu.Unlock()
				}
			}
		}
	}
}

// Register registers a client with the hub.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister unregisters a client from the hub.
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// Subscribe subscribes a client to a conversation.
func (h *Hub) Subscribe(client *Client, conversationID int64) {
	h.subscribe <- &Subscription{Client: client, ConversationID: conversationID}
}

// Unsubscribe unsubscribes a client from a conversation.
func (h *Hub) Unsubscribe(client *Client, conversationID int64) {
	h.unsubscribe <- &Subscription{Client: client, ConversationID: conversationID}
}

// BroadcastToConversation sends an event to all clients in a conversation.
func (h *Hub) BroadcastToConversation(conversationID int64, event *Event, excludeUserID int64) {
	h.broadcast <- &BroadcastMessage{
		ConversationID: conversationID,
		ExcludeUserID:  excludeUserID,
		Event:          event,
	}
}

// SendToUser sends an event to a specific user.
func (h *Hub) SendToUser(userID int64, event *Event) {
	h.mu.RLock()
	clients := h.clients[userID]
	h.mu.RUnlock()

	data, err := json.Marshal(event)
	if err != nil {
		h.logger.Error("failed to marshal event", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	for client := range clients {
		select {
		case client.send <- data:
		default:
			h.mu.Lock()
			close(client.send)
			delete(h.clients[userID], client)
			h.mu.Unlock()
		}
	}
}

// BroadcastUserStatus broadcasts user online/offline status.
func (h *Hub) BroadcastUserStatus(userID int64, online bool) {
	eventType := EventTypeUserOffline
	if online {
		eventType = EventTypeUserOnline
	}

	event := &Event{
		Type:   eventType,
		UserID: userID,
	}

	// Broadcast to all conversations this user is part of
	h.mu.RLock()
	conversationIDs := make([]int64, 0)
	for convID, clients := range h.conversations {
		for client := range clients {
			if client.userID == userID {
				conversationIDs = append(conversationIDs, convID)
				break
			}
		}
	}
	h.mu.RUnlock()

	for _, convID := range conversationIDs {
		h.BroadcastToConversation(convID, event, userID)
	}
}

// IsUserOnline checks if a user has any active connections.
func (h *Hub) IsUserOnline(userID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients[userID]) > 0
}

// GetOnlineUsers returns a list of online user IDs.
func (h *Hub) GetOnlineUsers() []int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]int64, 0, len(h.clients))
	for userID := range h.clients {
		users = append(users, userID)
	}
	return users
}
