package polymarketrealtime

import "time"

const (
	defaultHost                 = "wss://ws-live-data.polymarket.com"
	defaultPingInterval         = 5 * time.Second
	defaultAutoReconnect        = true
	defaultMaxReconnectAttempts = 10 // 0 means infinite retries
	defaultReconnectBackoffInit = 1 * time.Second
	defaultReconnectBackoffMax  = 30 * time.Second
)

type ClientOptions func(*client)

// defaultOpts sets the default options for the client
func defaultOpts() ClientOptions {
	return func(c *client) {
		c.host = defaultHost
		c.pingInterval = defaultPingInterval
		c.logger = NewSilentLogger()
		c.autoReconnect = defaultAutoReconnect
		c.maxReconnectAttempts = defaultMaxReconnectAttempts
		c.reconnectBackoffInit = defaultReconnectBackoffInit
		c.reconnectBackoffMax = defaultReconnectBackoffMax
	}
}

// WithLogger sets the logger for the client
func WithLogger(l Logger) ClientOptions {
	return func(c *client) {
		c.logger = l
	}
}

// WithPingInterval sets the ping interval for the client
func WithPingInterval(interval time.Duration) ClientOptions {
	return func(c *client) {
		c.pingInterval = interval
	}
}

// WithHost sets the host for the client
func WithHost(host string) ClientOptions {
	return func(c *client) {
		c.host = host
	}
}

// WithOnConnect sets the onConnect callback for the client
func WithOnConnect(f func()) ClientOptions {
	return func(c *client) {
		c.onConnectCallback = f
	}
}

func WithOnNewMessage(f func([]byte)) ClientOptions {
	return func(c *client) {
		c.onNewMessage = f
	}
}

// WithAutoReconnect enables or disables automatic reconnection on connection failures
func WithAutoReconnect(enabled bool) ClientOptions {
	return func(c *client) {
		c.autoReconnect = enabled
	}
}

// WithMaxReconnectAttempts sets the maximum number of reconnection attempts
// Set to 0 for infinite retries
func WithMaxReconnectAttempts(max int) ClientOptions {
	return func(c *client) {
		c.maxReconnectAttempts = max
	}
}

// WithReconnectBackoff sets the initial and maximum backoff duration for reconnection attempts
func WithReconnectBackoff(initial, max time.Duration) ClientOptions {
	return func(c *client) {
		c.reconnectBackoffInit = initial
		c.reconnectBackoffMax = max
	}
}

// WithOnDisconnect sets a callback that is called when the connection is lost
func WithOnDisconnect(f func(error)) ClientOptions {
	return func(c *client) {
		c.onDisconnectCallback = f
	}
}

// WithOnReconnect sets a callback that is called when reconnection succeeds
func WithOnReconnect(f func()) ClientOptions {
	return func(c *client) {
		c.onReconnectCallback = f
	}
}
