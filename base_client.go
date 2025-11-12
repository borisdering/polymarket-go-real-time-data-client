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
}

// newBaseClient creates a new base client with the given protocol and options
func newBaseClient(protocol Protocol, opts ...ClientOptions) *baseClient {
	c := &baseClient{
		protocol:  protocol,
		closeChan: make(chan struct{}),
	}

	// Set defaults
	c.host = protocol.GetDefaultHost()
	c.pingInterval = defaultPingInterval
	c.logger = NewSilentLogger()
	c.autoReconnect = defaultAutoReconnect
	c.maxReconnectAttempts = defaultMaxReconnectAttempts
	c.reconnectBackoffInit = defaultReconnectBackoffInit
	c.reconnectBackoffMax = defaultReconnectBackoffMax

	// Apply user options using adapter
	adapter := &clientOptionsAdapter{baseClient: c}
	for _, opt := range opts {
		adapter.applyOption(opt)
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
	err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		c.logger.Debug("Error sending close message: %v", err)
	}

	// Close connection
	err = c.conn.Close()
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

	// Store subscriptions for potential reconnection
	c.internal.mu.Lock()
	c.internal.subscriptions = append(c.internal.subscriptions, subscriptions...)
	c.internal.mu.Unlock()

	// Format subscription message using protocol
	message, err := c.protocol.FormatSubscription(subscriptions)
	if err != nil {
		return err
	}

	c.logger.Debug("Sending subscription message: %s", string(message))

	// Send subscription
	err = c.conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		return err
	}

	c.logger.Info("Subscribed to %d channels", len(subscriptions))

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

			err := conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				c.logger.Error("Error sending ping: %v", err)
				continue
			}

			c.logger.Debug("Ping sent")
		}
	}
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

			return
		}

		// Handle different message types
		switch messageType {
		case websocket.TextMessage:
			c.logger.Debug("Received message: %s", string(message))

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
			return

		default:
			c.logger.Debug("Received unknown message type: %d", messageType)
		}
	}
}

// isRecoverableError determines if an error is recoverable and should trigger reconnection
func (c *baseClient) isRecoverableError(err error) bool {
	if err == nil {
		return false
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

// reconnect attempts to reconnect to the WebSocket server with exponential backoff
func (c *baseClient) reconnect() {
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
		if !c.autoReconnect || c.internal.connClosed {
			c.internal.mu.RUnlock()
			c.internal.mu.Lock()
			c.internal.isReconnecting = false
			c.internal.mu.Unlock()
			return
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

// applyOptions applies client options to base client (needed for option compatibility)
type clientOptionsAdapter struct {
	*baseClient
}

func (c *clientOptionsAdapter) applyOption(opt ClientOptions) {
	// Create a temporary client wrapper that satisfies the client interface
	tempClient := &client{
		host:                  c.host,
		pingInterval:          c.pingInterval,
		logger:                c.logger,
		autoReconnect:         c.autoReconnect,
		maxReconnectAttempts:  c.maxReconnectAttempts,
		reconnectBackoffInit:  c.reconnectBackoffInit,
		reconnectBackoffMax:   c.reconnectBackoffMax,
		onConnectCallback:     c.onConnectCallback,
		onNewMessage:          c.onNewMessage,
		onDisconnectCallback:  c.onDisconnectCallback,
		onReconnectCallback:   c.onReconnectCallback,
	}

	// Apply option
	opt(tempClient)

	// Copy back
	c.host = tempClient.host
	c.pingInterval = tempClient.pingInterval
	c.logger = tempClient.logger
	c.autoReconnect = tempClient.autoReconnect
	c.maxReconnectAttempts = tempClient.maxReconnectAttempts
	c.reconnectBackoffInit = tempClient.reconnectBackoffInit
	c.reconnectBackoffMax = tempClient.reconnectBackoffMax
	c.onConnectCallback = tempClient.onConnectCallback
	c.onNewMessage = tempClient.onNewMessage
	c.onDisconnectCallback = tempClient.onDisconnectCallback
	c.onReconnectCallback = tempClient.onReconnectCallback
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
