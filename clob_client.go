package polymarketrealtime

// ClobUserClient is a WebSocket client for the CLOB User endpoint
// It provides real-time updates for user-specific order and trade events
type ClobUserClient struct {
	*baseClient
}

// NewClobUserClient creates a new CLOB User client
// Example usage:
//
//	client := polymarketrealtime.NewClobUserClient(
//	    polymarketrealtime.WithLogger(polymarketrealtime.NewLogger()),
//	    polymarketrealtime.WithOnNewMessage(func(data []byte) {
//	        // Handle message
//	    }),
//	)
func NewClobUserClient(opts ...ClientOptions) *ClobUserClient {
	protocol := NewClobUserProtocol()
	base := newBaseClient(protocol, opts...)
	return &ClobUserClient{baseClient: base}
}

// Connect establishes a WebSocket connection to the CLOB User endpoint
func (c *ClobUserClient) Connect() error {
	return c.baseClient.connect()
}

// Disconnect closes the WebSocket connection
func (c *ClobUserClient) Disconnect() error {
	return c.baseClient.disconnect()
}

// Subscribe subscribes to CLOB User channels
// Markets should be specified in Subscription.Filters as market IDs
// Authentication must be provided via Subscription.ClobAuth
//
// Example:
//
//	auth := &polymarketrealtime.ClobAuth{
//	    Key: "your-api-key",
//	    Secret: "your-secret",
//	    Passphrase: "your-passphrase",
//	}
//	subscriptions := []polymarketrealtime.Subscription{
//	    {
//	        Topic: polymarketrealtime.TopicCLOBUser,
//	        Filters: "market-id-1",
//	        ClobAuth: auth,
//	    },
//	}
//	err := client.Subscribe(subscriptions)
func (c *ClobUserClient) Subscribe(subscriptions []Subscription) error {
	return c.baseClient.subscribe(subscriptions)
}

// ClobMarketClient is a WebSocket client for the CLOB Market endpoint
// It provides real-time market data including orderbook updates and price changes
type ClobMarketClient struct {
	*baseClient
}

// NewClobMarketClient creates a new CLOB Market client
// Example usage:
//
//	client := polymarketrealtime.NewClobMarketClient(
//	    polymarketrealtime.WithLogger(polymarketrealtime.NewLogger()),
//	    polymarketrealtime.WithOnNewMessage(func(data []byte) {
//	        // Handle message
//	    }),
//	)
func NewClobMarketClient(opts ...ClientOptions) *ClobMarketClient {
	protocol := NewClobMarketProtocol()
	base := newBaseClient(protocol, opts...)
	return &ClobMarketClient{baseClient: base}
}

// Connect establishes a WebSocket connection to the CLOB Market endpoint
func (c *ClobMarketClient) Connect() error {
	return c.baseClient.connect()
}

// Disconnect closes the WebSocket connection
func (c *ClobMarketClient) Disconnect() error {
	return c.baseClient.disconnect()
}

// Subscribe subscribes to CLOB Market channels
// Asset IDs should be specified in Subscription.Filters
//
// Example:
//
//	subscriptions := []polymarketrealtime.Subscription{
//	    {
//	        Topic: polymarketrealtime.TopicCLOBMarket,
//	        Filters: "asset-id-1",
//	    },
//	    {
//	        Topic: polymarketrealtime.TopicCLOBMarket,
//	        Filters: "asset-id-2",
//	    },
//	}
//	err := client.Subscribe(subscriptions)
func (c *ClobMarketClient) Subscribe(subscriptions []Subscription) error {
	return c.baseClient.subscribe(subscriptions)
}
