package polymarketrealtime

// Subscription represents a subscription to a topic and message type
type Subscription struct {
	Topic   Topic       `json:"topic"`
	Type    MessageType `json:"type"`
	Filters string      `json:"filters,omitempty"`

	ClobAuth  *ClobAuth  `json:"clob_auth,omitempty"`
	GammaAuth *GammaAuth `json:"gamma_auth,omitempty"`
}

// ClobAuth contains authentication information for CLOB subscriptions
type ClobAuth struct {
	Key        string `json:"key"`
	Secret     string `json:"secret"`
	Passphrase string `json:"passphrase"`
}

// GammaAuth contains authentication information for Gamma subscriptions
type GammaAuth struct {
	Address string `json:"address"`
}

// WsClient interface for backward compatibility
type WsClient interface {
	// Connect establishes a WebSocket connection to the server
	Connect() error

	// Disconnect closes the WebSocket connection
	Disconnect() error

	// Subscribe sends a subscription message to the server
	Subscribe(subscriptions []Subscription) error

	// Unsubscribe sends an unsubscription message to the server
	Unsubscribe(subscriptions []Subscription) error
}

// Client is a thin wrapper around baseClient for backward compatibility
// It uses the RealtimeProtocol which supports multiple topics
type Client struct {
	*baseClient

	*RealtimeTypedSubscriptionHandler
}

// New creates a new client using the baseClient infrastructure
// This provides backward compatibility while using the improved baseClient implementation
func New(opts ...ClientOption) *Client {
	protocol := NewRealtimeProtocol()

	router := NewRealtimeMessageRouter()
	opts = append(opts, withRouter(router))
	base := newBaseClient(protocol, opts...)
	cli := &Client{baseClient: base}

	handler := NewRealtimeTypedSubscriptionHandler(cli, router)

	cli.RealtimeTypedSubscriptionHandler = handler
	return cli
}

// Connect establishes a WebSocket connection to the server
func (c *Client) Connect() error {
	return c.baseClient.connect()
}

// Disconnect closes the WebSocket connection
func (c *Client) Disconnect() error {
	return c.baseClient.disconnect()
}

// Subscribe sends subscription requests to the server
func (c *Client) Subscribe(subscriptions []Subscription) error {
	return c.baseClient.subscribe(subscriptions)
}

// Unsubscribe sends unsubscription requests to the server
func (c *Client) Unsubscribe(subscriptions []Subscription) error {
	return c.baseClient.unsubscribe(subscriptions)
}
