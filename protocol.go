package polymarketrealtime

// Protocol defines the interface for endpoint-specific behavior
type Protocol interface {
	// GetDefaultHost returns the default WebSocket host for this protocol
	GetDefaultHost() string

	// FormatSubscription formats subscriptions into the protocol-specific message format
	FormatSubscription(subscriptions []Subscription) ([]byte, error)

	// ParseMessage parses a raw message into a structured Message format
	// Returns the parsed message and any error encountered
	ParseMessage(data []byte) (*Message, error)

	// GetProtocolName returns a human-readable name for this protocol
	GetProtocolName() string
}
