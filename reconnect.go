package polymarketrealtime

import (
	"errors"
	"io"
	"net"
	"strings"
	"syscall"
	"time"
)

// isRecoverableError determines if an error is recoverable and should trigger reconnection
func (c *client) isRecoverableError(err error) bool {
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
func (c *client) reconnect() {
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
		err := c.Connect()
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
func (c *client) restoreSubscriptions() {
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
	err := c.Subscribe(subs)
	if err != nil {
		c.logger.Error("Failed to restore subscriptions: %v", err)
	} else {
		c.logger.Info("Successfully restored subscriptions")
	}
}
