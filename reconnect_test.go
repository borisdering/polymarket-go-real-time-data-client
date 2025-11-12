package polymarketrealtime

import (
	"errors"
	"io"
	"syscall"
	"testing"
	"time"
)

func TestIsRecoverableError(t *testing.T) {
	client := &client{}

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "EOF error",
			err:      io.EOF,
			expected: true,
		},
		{
			name:     "unexpected EOF",
			err:      io.ErrUnexpectedEOF,
			expected: true,
		},
		{
			name:     "broken pipe",
			err:      syscall.EPIPE,
			expected: true,
		},
		{
			name:     "connection reset",
			err:      syscall.ECONNRESET,
			expected: true,
		},
		{
			name:     "broken pipe error message",
			err:      errors.New("write tcp: broken pipe"),
			expected: true,
		},
		{
			name:     "connection reset error message",
			err:      errors.New("read tcp: connection reset by peer"),
			expected: true,
		},
		{
			name:     "EOF in message",
			err:      errors.New("unexpected EOF"),
			expected: true,
		},
		{
			name:     "non-recoverable error",
			err:      errors.New("authentication failed"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.isRecoverableError(tt.err)
			if result != tt.expected {
				t.Errorf("isRecoverableError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestReconnectionAttempts(t *testing.T) {
	// This test verifies that reconnection attempts can be tracked
	c := New()
	cl := c.(*client)

	// Verify initial state
	if cl.internal.reconnectAttempts != 0 {
		t.Errorf("Expected initial reconnectAttempts to be 0, got %d", cl.internal.reconnectAttempts)
	}

	// Verify isReconnecting flag
	if cl.internal.isReconnecting {
		t.Error("Expected isReconnecting to be false initially")
	}
}

func TestAutoReconnectConfiguration(t *testing.T) {
	// Test default configuration
	cl := New()
	c := cl.(*client)

	if !c.autoReconnect {
		t.Error("Expected autoReconnect to be enabled by default")
	}

	if c.maxReconnectAttempts != defaultMaxReconnectAttempts {
		t.Errorf("Expected maxReconnectAttempts to be %d, got %d", defaultMaxReconnectAttempts, c.maxReconnectAttempts)
	}

	if c.reconnectBackoffInit != defaultReconnectBackoffInit {
		t.Errorf("Expected reconnectBackoffInit to be %v, got %v", defaultReconnectBackoffInit, c.reconnectBackoffInit)
	}

	if c.reconnectBackoffMax != defaultReconnectBackoffMax {
		t.Errorf("Expected reconnectBackoffMax to be %v, got %v", defaultReconnectBackoffMax, c.reconnectBackoffMax)
	}
}

func TestCustomReconnectConfiguration(t *testing.T) {
	// Test custom configuration
	maxAttempts := 5
	backoffInit := 2 * time.Second
	backoffMax := 60 * time.Second

	cl := New(
		WithAutoReconnect(false),
		WithMaxReconnectAttempts(maxAttempts),
		WithReconnectBackoff(backoffInit, backoffMax),
	)
	c := cl.(*client)

	if c.autoReconnect {
		t.Error("Expected autoReconnect to be disabled")
	}

	if c.maxReconnectAttempts != maxAttempts {
		t.Errorf("Expected maxReconnectAttempts to be %d, got %d", maxAttempts, c.maxReconnectAttempts)
	}

	if c.reconnectBackoffInit != backoffInit {
		t.Errorf("Expected reconnectBackoffInit to be %v, got %v", backoffInit, c.reconnectBackoffInit)
	}

	if c.reconnectBackoffMax != backoffMax {
		t.Errorf("Expected reconnectBackoffMax to be %v, got %v", backoffMax, c.reconnectBackoffMax)
	}
}

func TestReconnectCallbacks(t *testing.T) {
	disconnectCalled := false
	reconnectCalled := false

	cl := New(
		WithOnDisconnect(func(err error) {
			disconnectCalled = true
		}),
		WithOnReconnect(func() {
			reconnectCalled = true
		}),
	)

	c := cl.(*client)

	// Test disconnect callback is set
	if c.onDisconnectCallback == nil {
		t.Error("Expected onDisconnectCallback to be set")
	}

	// Test reconnect callback is set
	if c.onReconnectCallback == nil {
		t.Error("Expected onReconnectCallback to be set")
	}

	// Call callbacks to verify they work
	if c.onDisconnectCallback != nil {
		c.onDisconnectCallback(errors.New("test error"))
	}

	if c.onReconnectCallback != nil {
		c.onReconnectCallback()
	}

	if !disconnectCalled {
		t.Error("Expected disconnect callback to be called")
	}

	if !reconnectCalled {
		t.Error("Expected reconnect callback to be called")
	}
}

func TestInfiniteRetries(t *testing.T) {
	// Test that maxReconnectAttempts = 0 means infinite retries
	cl := New(WithMaxReconnectAttempts(0))
	c := cl.(*client)

	if c.maxReconnectAttempts != 0 {
		t.Errorf("Expected maxReconnectAttempts to be 0 (infinite), got %d", c.maxReconnectAttempts)
	}
}
