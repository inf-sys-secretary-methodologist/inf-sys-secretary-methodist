package websocket

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/config"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 4096
)

// checkOrigin validates the Origin header against allowed origins from config.
func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		// Allow connections without Origin header (e.g., from same origin)
		return true
	}

	cfg, err := config.Load()
	if err != nil {
		// If config fails to load, deny the connection for security
		return false
	}
	allowedOrigins := cfg.CORS.AllowedOrigins

	// Parse the origin URL
	originURL, err := url.Parse(origin)
	if err != nil {
		return false
	}

	for _, allowed := range allowedOrigins {
		// Handle wildcard
		if allowed == "*" {
			return true
		}

		allowedURL, err := url.Parse(allowed)
		if err != nil {
			continue
		}

		// Compare scheme and host (including port)
		if strings.EqualFold(originURL.Scheme, allowedURL.Scheme) &&
			strings.EqualFold(originURL.Host, allowedURL.Host) {
			return true
		}
	}

	return false
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkOrigin,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID int64
	logger *logging.Logger
}

// ClientMessage represents an incoming message from client.
type ClientMessage struct {
	Type           string          `json:"type"`
	ConversationID int64           `json:"conversation_id,omitempty"`
	Payload        json.RawMessage `json:"payload,omitempty"`
}

// NewClient creates a new websocket client.
func NewClient(hub *Hub, conn *websocket.Conn, userID int64, logger *logging.Logger) *Client {
	return &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: userID,
		logger: logger,
	}
}

// ReadPump pumps messages from the websocket connection to the hub.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Error("websocket error", map[string]interface{}{
					"error":   err.Error(),
					"user_id": c.userID,
				})
			}
			break
		}

		// Parse incoming message
		var clientMsg ClientMessage
		if err := json.Unmarshal(message, &clientMsg); err != nil {
			c.logger.Error("failed to parse client message", map[string]interface{}{
				"error":   err.Error(),
				"user_id": c.userID,
			})
			continue
		}

		// Handle different message types
		c.handleMessage(&clientMsg)
	}
}

// WritePump pumps messages from the hub to the websocket connection.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage handles incoming client messages.
func (c *Client) handleMessage(msg *ClientMessage) {
	switch msg.Type {
	case "subscribe":
		// Subscribe to a conversation
		if msg.ConversationID > 0 {
			c.hub.Subscribe(c, msg.ConversationID)
		}

	case "unsubscribe":
		// Unsubscribe from a conversation
		if msg.ConversationID > 0 {
			c.hub.Unsubscribe(c, msg.ConversationID)
		}

	case "typing":
		// Broadcast typing indicator
		if msg.ConversationID > 0 {
			c.hub.BroadcastToConversation(msg.ConversationID, &Event{
				Type:           EventTypeTyping,
				ConversationID: msg.ConversationID,
				UserID:         c.userID,
			}, c.userID)
		}

	case "stop_typing":
		// Broadcast stop typing
		if msg.ConversationID > 0 {
			c.hub.BroadcastToConversation(msg.ConversationID, &Event{
				Type:           EventTypeStopTyping,
				ConversationID: msg.ConversationID,
				UserID:         c.userID,
			}, c.userID)
		}

	case "ping":
		// Respond with pong
		pong, _ := json.Marshal(map[string]string{"type": "pong"})
		c.send <- pong

	default:
		c.logger.Debug("unknown message type", map[string]interface{}{
			"type":    msg.Type,
			"user_id": c.userID,
		})
	}
}

// UserID returns the user ID of the client.
func (c *Client) UserID() int64 {
	return c.userID
}

// Send sends a message to the client.
func (c *Client) Send(data []byte) {
	select {
	case c.send <- data:
	default:
		c.logger.Error("client send buffer full", map[string]interface{}{
			"user_id": c.userID,
		})
	}
}

// ServeWs handles websocket requests from the peer.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request, userID int64, logger *logging.Logger) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("failed to upgrade connection", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	client := NewClient(hub, conn, userID, logger)
	hub.Register(client)

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.WritePump()
	go client.ReadPump()
}
