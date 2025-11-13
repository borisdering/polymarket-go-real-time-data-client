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

// Config contains all configuration options for the WebSocket client
type Config struct {
	Host                 string
	Logger               Logger
	PingInterval         time.Duration
	AutoReconnect        bool
	MaxReconnectAttempts int
	ReconnectBackoffInit time.Duration
	ReconnectBackoffMax  time.Duration
	OnConnectCallback    func()
	OnNewMessage         func([]byte)
	OnDisconnectCallback func(error)
	OnReconnectCallback  func()
}

type ClientOptions func(*Config)

// WithLogger sets the logger for the client
func WithLogger(l Logger) ClientOptions {
	return func(c *Config) {
		c.Logger = l
	}
}

// WithPingInterval sets the ping interval for the client
func WithPingInterval(interval time.Duration) ClientOptions {
	return func(c *Config) {
		c.PingInterval = interval
	}
}

// WithHost sets the host for the client
func WithHost(host string) ClientOptions {
	return func(c *Config) {
		c.Host = host
	}
}

// WithOnConnect sets the onConnect callback for the client
func WithOnConnect(f func()) ClientOptions {
	return func(c *Config) {
		c.OnConnectCallback = f
	}
}

func WithOnNewMessage(f func([]byte)) ClientOptions {
	return func(c *Config) {
		c.OnNewMessage = f
	}
}

// WithAutoReconnect enables or disables automatic reconnection on connection failures
func WithAutoReconnect(enabled bool) ClientOptions {
	return func(c *Config) {
		c.AutoReconnect = enabled
	}
}

// WithMaxReconnectAttempts sets the maximum number of reconnection attempts
// Set to 0 for infinite retries
func WithMaxReconnectAttempts(max int) ClientOptions {
	return func(c *Config) {
		c.MaxReconnectAttempts = max
	}
}

// WithReconnectBackoff sets the initial and maximum backoff duration for reconnection attempts
func WithReconnectBackoff(initial, max time.Duration) ClientOptions {
	return func(c *Config) {
		c.ReconnectBackoffInit = initial
		c.ReconnectBackoffMax = max
	}
}

// WithOnDisconnect sets a callback that is called when the connection is lost
func WithOnDisconnect(f func(error)) ClientOptions {
	return func(c *Config) {
		c.OnDisconnectCallback = f
	}
}

// WithOnReconnect sets a callback that is called when reconnection succeeds
func WithOnReconnect(f func()) ClientOptions {
	return func(c *Config) {
		c.OnReconnectCallback = f
	}
}

// MessageRouter defines the interface for routing messages to handlers
type MessageRouter interface {
	RouteMessage(data []byte) error
}

// WithRouter sets a message router that will automatically route all incoming messages
// This eliminates the need to manually call router.RouteMessage() in WithOnNewMessage
//
// Example usage:
//
//	typedSub := polymarketrealtime.NewRealtimeTypedSubscriptionHandler(nil)
//	client := polymarketrealtime.New(
//	    polymarketrealtime.WithRouter(typedSub.GetRouter()),
//	    polymarketrealtime.WithLogger(polymarketrealtime.NewLogger()),
//	)
//	typedSub.SetClient(client)
func WithRouter(router MessageRouter) ClientOptions {
	return func(c *Config) {
		if router != nil {
			// Wrap the existing OnNewMessage callback
			existingCallback := c.OnNewMessage
			c.OnNewMessage = func(data []byte) {
				// First call the router
				if err := router.RouteMessage(data); err != nil {
					// Log error but don't fail
					if c.Logger != nil {
						c.Logger.Error("Failed to route message: %v", err)
					}
				}

				// Then call the existing callback if any
				if existingCallback != nil {
					existingCallback(data)
				}
			}
		}
	}
}
