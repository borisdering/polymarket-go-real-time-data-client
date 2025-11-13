package polymarketrealtime

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

// writeRequest represents a request to write a message to the WebSocket
type writeRequest struct {
	messageType int
	data        []byte
}

// baseClient contains the common WebSocket infrastructure shared by all client types
type baseClient struct {
	// Protocol-specific behavior
	protocol Protocol

	// WebSocket connection
	conn   *websocket.Conn
	connMu sync.RWMutex

	// Configuration
	host                 string
	pingInterval         time.Duration
	autoReconnect        bool
	maxReconnectAttempts int
	reconnectBackoffInit time.Duration
	reconnectBackoffMax  time.Duration

	// Callbacks
	onConnectCallback    func()
	onNewMessage         func([]byte)
	onDisconnectCallback func(error)
	onReconnectCallback  func()

	// Logger
	logger Logger

	// Internal state
	internal struct {
		mu                sync.RWMutex
		connClosed        bool
		subscriptions     []Subscription
		reconnectAttempts int
		isReconnecting    bool
	}

	// Channels for control flow
	closeChan chan struct{}
	closeOnce sync.Once

	writeLoopOnce sync.Once
	writeChan     chan writeRequest // Channel for serializing WebSocket writes
}

// newBaseClient creates a new base client with the given protocol and options
func newBaseClient(protocol Protocol, opts ...ClientOptions) *baseClient {
	// Create config with defaults
	config := &Config{
		Host:                 protocol.GetDefaultHost(),
		Logger:               NewSilentLogger(),
		PingInterval:         defaultPingInterval,
		AutoReconnect:        defaultAutoReconnect,
		MaxReconnectAttempts: defaultMaxReconnectAttempts,
		ReconnectBackoffInit: defaultReconnectBackoffInit,
		ReconnectBackoffMax:  defaultReconnectBackoffMax,
	}

	// Apply user options to config
	for _, opt := range opts {
		opt(config)
	}

	// Create base client with config
	c := &baseClient{
		protocol:             protocol,
		closeChan:            make(chan struct{}),
		writeChan:            make(chan writeRequest, 100), // Buffered channel for writes
		host:                 config.Host,
		logger:               config.Logger,
		pingInterval:         config.PingInterval,
		autoReconnect:        config.AutoReconnect,
		maxReconnectAttempts: config.MaxReconnectAttempts,
		reconnectBackoffInit: config.ReconnectBackoffInit,
		reconnectBackoffMax:  config.ReconnectBackoffMax,
		onConnectCallback:    config.OnConnectCallback,
		onNewMessage:         config.OnNewMessage,
		onDisconnectCallback: config.OnDisconnectCallback,
		onReconnectCallback:  config.OnReconnectCallback,
	}

	return c
}

// connect establishes the WebSocket connection
func (c *baseClient) connect() error {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	// Check if already connected
	if c.conn != nil {
		c.internal.mu.RLock()
		isClosed := c.internal.connClosed
		c.internal.mu.RUnlock()

		if !isClosed {
			return errors.New("already connected")
		}
	}

	// Check if we're reconnecting - if not, this must be a fresh connection
	c.internal.mu.RLock()
	isReconnecting := c.internal.isReconnecting
	c.internal.mu.RUnlock()

	// If not reconnecting and connection was previously closed, reject
	if !isReconnecting {
		c.internal.mu.RLock()
		wasClosed := c.internal.connClosed
		c.internal.mu.RUnlock()

		if wasClosed && c.conn != nil {
			return errors.New("connection was closed, create a new client")
		}
	}

	c.logger.Debug("Connecting to %s (%s)", c.host, c.protocol.GetProtocolName())

	// Establish WebSocket connection
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.Dial(c.host, nil)
	if err != nil {
		return err
	}

	c.conn = conn

	// Reset connection state
	c.internal.mu.Lock()
	c.internal.connClosed = false
	c.internal.mu.Unlock()

	// Start goroutines
	go c.ping()
	go c.readMessages()
	go c.writeLoop()

	// Call connect callback
	if c.onConnectCallback != nil {
		c.onConnectCallback()
	}

	c.logger.Info("Connected to %s successfully", c.protocol.GetProtocolName())

	return nil
}

// disconnect closes the WebSocket connection
func (c *baseClient) disconnect() error {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	// Disable auto-reconnect when explicitly disconnecting
	c.internal.mu.Lock()
	c.autoReconnect = false
	c.internal.mu.Unlock()

	if c.conn == nil {
		return errors.New("not connected")
	}

	c.internal.mu.RLock()
	isClosed := c.internal.connClosed
	c.internal.mu.RUnlock()

	if isClosed {
		return errors.New("connection already closed")
	}

	c.logger.Debug("Disconnecting from %s", c.protocol.GetProtocolName())

	// Close the connection
	c.closeOnce.Do(func() {
		close(c.closeChan)
	})

	// Mark as closed
	c.internal.mu.Lock()
	c.internal.connClosed = true
	c.internal.mu.Unlock()

	// Send close message
	c.writeChan <- writeRequest{websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")}

	// Close connection
	err := c.conn.Close()
	if err != nil {
		return err
	}

	c.logger.Info("Disconnected from %s successfully", c.protocol.GetProtocolName())

	return nil
}

// subscribe sends subscription requests to the server
func (c *baseClient) subscribe(subscriptions []Subscription) error {
	c.connMu.RLock()
	defer c.connMu.RUnlock()

	if c.conn == nil {
		return errors.New("not connected")
	}

	c.internal.mu.RLock()
	isClosed := c.internal.connClosed
	c.internal.mu.RUnlock()

	if isClosed {
		return errors.New("connection closed")
	}

	// Add new subscriptions to stored list (with deduplication)
	c.internal.mu.Lock()

	// Create a map of existing subscriptions for quick lookup
	existing := make(map[string]bool)
	for _, sub := range c.internal.subscriptions {
		key := string(sub.Topic) + "|" + string(sub.Type) + "|" + sub.Filters
		existing[key] = true
	}

	// Only add subscriptions that don't already exist
	newSubs := []Subscription{}
	for _, sub := range subscriptions {
		key := string(sub.Topic) + "|" + string(sub.Type) + "|" + sub.Filters
		if !existing[key] {
			c.internal.subscriptions = append(c.internal.subscriptions, sub)
			newSubs = append(newSubs, sub)
			existing[key] = true
		}
	}

	totalCount := len(c.internal.subscriptions)
	c.internal.mu.Unlock()

	// Send each subscription individually
	// Note: Some topics (like crypto_prices, crypto_prices_chainlink, equity_prices) only support
	// ONE symbol per connection. For these topics, the last subscription will replace previous ones.
	// If you need multiple symbols for these topics, create separate client connections.
	//
	// Other topics (like clob_market, activity, comments) may support multiple subscriptions.
	for _, sub := range newSubs {
		// Format subscription message for this single subscription
		message, err := c.protocol.FormatSubscription([]Subscription{sub})
		if err != nil {
			return err
		}

		c.logger.Debug("Sending subscription message: %s", string(message))

		// Send subscription through write channel
		select {
		case c.writeChan <- writeRequest{messageType: websocket.TextMessage, data: message}:
			// Success
		case <-time.After(5 * time.Second):
			return errors.New("timeout sending subscription message")
		}
	}

	c.logger.Info("Subscribed to %d new channels (%d total)", len(newSubs), totalCount)
	return nil
}

// unsubscribe sends unsubscription requests to the server
func (c *baseClient) unsubscribe(subscriptions []Subscription) error {
	c.connMu.RLock()
	defer c.connMu.RUnlock()

	if c.conn == nil {
		return errors.New("not connected")
	}

	c.internal.mu.RLock()
	isClosed := c.internal.connClosed
	c.internal.mu.RUnlock()

	if isClosed {
		return errors.New("connection closed")
	}

	// First send unsubscribe message for the subscriptions to remove
	unsubMessage, err := c.protocol.FormatSubscription(subscriptions)
	if err != nil {
		return err
	}

	// Replace "subscribe" with "unsubscribe" in the message
	messageStr := string(unsubMessage)
	if strings.Contains(messageStr, `"action":"subscribe"`) {
		messageStr = strings.Replace(messageStr, `"action":"subscribe"`, `"action":"unsubscribe"`, 1)
	} else if strings.Contains(messageStr, `"type":"MARKET"`) {
		// CLOB protocol uses different format
		messageStr = strings.Replace(messageStr, `"type":"MARKET"`, `"type":"UNSUBSCRIBE"`, 1)
	}
	unsubMessage = []byte(messageStr)

	c.logger.Debug("Sending unsubscription message: %s", string(unsubMessage))

	// Send unsubscription through write channel
	select {
	case c.writeChan <- writeRequest{messageType: websocket.TextMessage, data: unsubMessage}:
		c.logger.Debug("Unsubscription message sent")
	case <-time.After(5 * time.Second):
		return errors.New("timeout sending unsubscription message")
	}

	// Remove subscriptions from stored list
	c.internal.mu.Lock()
	// Create a map for quick lookup
	toRemove := make(map[string]bool)
	for _, sub := range subscriptions {
		key := string(sub.Topic) + "|" + string(sub.Type) + "|" + sub.Filters
		toRemove[key] = true
	}

	// Filter out subscriptions to remove
	newSubs := make([]Subscription, 0)
	for _, sub := range c.internal.subscriptions {
		key := string(sub.Topic) + "|" + string(sub.Type) + "|" + sub.Filters
		if !toRemove[key] {
			newSubs = append(newSubs, sub)
		}
	}
	c.internal.subscriptions = newSubs
	c.internal.mu.Unlock()

	c.logger.Info("Unsubscribed from %d channels (%d remaining)", len(subscriptions), len(newSubs))
	return nil
}

// ping sends periodic ping messages to keep the connection alive
func (c *baseClient) ping() {
	ticker := time.NewTicker(c.pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.closeChan:
			c.logger.Debug("Ping routine stopped")
			return
		case <-ticker.C:
			c.connMu.RLock()
			conn := c.conn
			c.connMu.RUnlock()

			if conn == nil {
				continue
			}

			c.internal.mu.RLock()
			isClosed := c.internal.connClosed
			c.internal.mu.RUnlock()

			if isClosed {
				return
			}
			c.writeChan <- writeRequest{websocket.TextMessage, []byte("ping")}
			// c.logger.Debug("Ping sent")
		}
	}
}

// writeLoop serializes all WebSocket writes through a channel
func (c *baseClient) writeLoop() {
	c.writeLoopOnce.Do(func() {
		defer func() {
			c.logger.Debug("Write loop stopped")
		}()

		for {
			select {
			case <-c.closeChan:
				return
			case req := <-c.writeChan:
				c.connMu.RLock()
				conn := c.conn
				c.connMu.RUnlock()

				if conn == nil {
					continue
				}

				c.internal.mu.RLock()
				isClosed := c.internal.connClosed
				c.internal.mu.RUnlock()

				if isClosed {
					continue
				}

				c.logger.Debug("Write message %v: %v", req.messageType, string(req.data))

				err := conn.WriteMessage(req.messageType, req.data)
				if err != nil {
					c.logger.Error("Error writing message (type=%d): %v", req.messageType, err)
				}
			}
		}
	})
}

// readMessages reads messages from the WebSocket connection
func (c *baseClient) readMessages() {
	defer func() {
		c.logger.Debug("Read messages routine stopped")
	}()

	for {
		c.connMu.RLock()
		conn := c.conn
		c.connMu.RUnlock()

		if conn == nil {
			return
		}

		c.internal.mu.RLock()
		isClosed := c.internal.connClosed
		c.internal.mu.RUnlock()

		if isClosed {
			return
		}

		// Read message
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			c.logger.Error("Error reading message: %v", err)

			// Check if error is recoverable
			c.tryAutoReconnect(err)

			return
		}

		c.logger.Debug("Received raw message (len=%d): %s", len(message), string(message))

		// Handle different message types
		switch messageType {
		case websocket.TextMessage:
			// Skip empty messages
			if len(message) == 0 {
				c.logger.Debug("Received empty message, skipping")
				continue
			}

			// Skip ping/pong messages
			msgStr := string(message)
			if msgStr == "PING" || msgStr == "PONG" || msgStr == "ping" || msgStr == "pong" {
				c.logger.Debug("Received PING/PONG, skipping")
				continue
			}

			// // Skip non-JSON messages
			// if message[0] != '{' && message[0] != '[' {
			// 	c.logger.Debug("Received non-JSON message (len=%d, data=%q), skipping", len(message), msgStr)
			// 	continue
			// }

			c.logger.Debug("Received valid message (len=%d): %s", len(message), msgStr)

			// Call message callback
			if c.onNewMessage != nil {
				c.onNewMessage(message)
			}

		case websocket.BinaryMessage:
			c.logger.Debug("Received binary message (length: %d)", len(message))

			// Call message callback
			if c.onNewMessage != nil {
				c.onNewMessage(message)
			}

		case websocket.PongMessage:
			c.logger.Debug("Received pong")

		case websocket.CloseMessage:
			c.logger.Info("Received close message from server")
			c.tryAutoReconnect(nil)

			return

		default:
			c.logger.Debug("Received unknown message type: %d", messageType)
		}
	}
}

func (c *baseClient) tryAutoReconnect(err error) {
	if c.isRecoverableError(err) && c.autoReconnect {
		c.logger.Info("Connection lost, attempting to reconnect...")

		// Mark connection as closed
		c.internal.mu.Lock()
		c.internal.connClosed = true
		c.internal.mu.Unlock()

		// Call disconnect callback
		if c.onDisconnectCallback != nil {
			c.onDisconnectCallback(err)
		}

		// Trigger reconnection
		go c.reconnect()
	}
}

// isRecoverableError determines if an error is recoverable and should trigger reconnection
func (c *baseClient) isRecoverableError(err error) bool {
	if err == nil {
		return true
	}

	// Check for common network errors
	if errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrUnexpectedEOF) ||
		errors.Is(err, syscall.EPIPE) ||
		errors.Is(err, syscall.ECONNRESET) {
		return true
	}

	// Check for net.OpError
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		return true
	}

	// Check error message for common connection issues
	errMsg := err.Error()
	recoverableMessages := []string{
		"broken pipe",
		"connection reset",
		"connection refused",
		"i/o timeout",
		"use of closed network connection",
		"connection closed",
		"eof",
	}

	for _, msg := range recoverableMessages {
		if strings.Contains(strings.ToLower(errMsg), msg) {
			return true
		}
	}

	return false
}

// TODO: Fix reconnect
// reconnect attempts to reconnect to the WebSocket server with exponential backoff
func (c *baseClient) reconnect() {
	c.logger.Debug("Trying to reconnect...")
	c.internal.mu.Lock()

	// Check if already reconnecting
	if c.internal.isReconnecting {
		c.internal.mu.Unlock()
		return
	}

	// Mark as reconnecting
	c.internal.isReconnecting = true
	c.internal.reconnectAttempts = 0
	c.internal.mu.Unlock()

	// Exponential backoff parameters
	backoff := c.reconnectBackoffInit

	for {
		// Check if we should stop reconnecting
		c.internal.mu.RLock()
		if !c.autoReconnect {
			c.internal.mu.RUnlock()

			c.internal.mu.Lock()
			c.internal.isReconnecting = false
			c.internal.mu.Unlock()

			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Check max attempts
		if c.maxReconnectAttempts > 0 && c.internal.reconnectAttempts >= c.maxReconnectAttempts {
			c.internal.mu.RUnlock()
			c.logger.Error("Max reconnection attempts (%d) reached, giving up", c.maxReconnectAttempts)
			c.internal.mu.Lock()
			c.internal.isReconnecting = false
			c.internal.mu.Unlock()
			return
		}
		c.internal.mu.RUnlock()

		// Increment attempt counter
		c.internal.mu.Lock()
		c.internal.reconnectAttempts++
		attempt := c.internal.reconnectAttempts
		c.internal.mu.Unlock()

		c.logger.Info("Reconnection attempt %d (backoff: %v)", attempt, backoff)

		// Wait before attempting
		time.Sleep(backoff)

		// Attempt to reconnect
		err := c.connect()
		if err != nil {
			c.logger.Error("Reconnection attempt %d failed: %v", attempt, err)

			// Increase backoff exponentially
			backoff = backoff * 2
			if backoff > c.reconnectBackoffMax {
				backoff = c.reconnectBackoffMax
			}

			continue
		}

		// Reconnection successful
		c.logger.Info("Reconnection successful after %d attempts", attempt)

		// Restore subscriptions
		c.restoreSubscriptions()

		// Call reconnect callback
		if c.onReconnectCallback != nil {
			c.onReconnectCallback()
		}

		// Reset reconnection state
		c.internal.mu.Lock()
		c.internal.isReconnecting = false
		c.internal.reconnectAttempts = 0
		c.internal.connClosed = false
		c.internal.mu.Unlock()

		return
	}
}

// restoreSubscriptions restores all previous subscriptions after reconnection
func (c *baseClient) restoreSubscriptions() {
	c.internal.mu.RLock()
	subscriptions := c.internal.subscriptions
	c.internal.mu.RUnlock()

	if len(subscriptions) == 0 {
		c.logger.Debug("No subscriptions to restore")
		return
	}

	c.logger.Info("Restoring %d subscriptions", len(subscriptions))

	// Temporarily store subscriptions
	subs := make([]Subscription, len(subscriptions))
	copy(subs, subscriptions)

	// Clear subscriptions before re-subscribing
	c.internal.mu.Lock()
	c.internal.subscriptions = make([]Subscription, 0)
	c.internal.mu.Unlock()

	// Re-subscribe
	err := c.subscribe(subs)
	if err != nil {
		c.logger.Error("Failed to restore subscriptions: %v", err)
	} else {
		c.logger.Info("Successfully restored subscriptions")
	}
}

// Helper function to convert subscriptions to JSON (used by realtime protocol)
func subscriptionsToJSON(subscriptions []Subscription) ([]byte, error) {
	type subscriptionMessage struct {
		Action        string         `json:"action"`
		Subscriptions []Subscription `json:"subscriptions"`
	}

	msg := subscriptionMessage{
		Action:        "subscribe",
		Subscriptions: subscriptions,
	}

	return json.Marshal(msg)
}
