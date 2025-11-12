package polymarketrealtime

import (
	"encoding/json"
	"fmt"
)

// ClobUserProtocol implements the Protocol interface for the CLOB User endpoint
type ClobUserProtocol struct{}

// NewClobUserProtocol creates a new CLOB User protocol instance
func NewClobUserProtocol() Protocol {
	return &ClobUserProtocol{}
}

// GetDefaultHost returns the default WebSocket host for CLOB User channel
func (p *ClobUserProtocol) GetDefaultHost() string {
	return "wss://ws-subscriptions-clob.polymarket.com/ws/user"
}

// FormatSubscription formats subscriptions into the CLOB User protocol message format
func (p *ClobUserProtocol) FormatSubscription(subscriptions []Subscription) ([]byte, error) {
	if len(subscriptions) == 0 {
		return nil, fmt.Errorf("no subscriptions provided")
	}

	// Extract markets and auth from subscriptions
	markets := make([]string, 0)
	var auth *ClobAuth

	for _, sub := range subscriptions {
		if sub.Filters != "" {
			markets = append(markets, sub.Filters)
		}
		if sub.ClobAuth != nil {
			auth = sub.ClobAuth
		}
	}

	if auth == nil {
		return nil, fmt.Errorf("CLOB User subscription requires authentication")
	}

	// CLOB User subscription format:
	// {
	//   "markets": ["market1", "market2"],
	//   "type": "USER",
	//   "auth": {
	//     "apiKey": "...",
	//     "secret": "...",
	//     "passphrase": "..."
	//   }
	// }
	type clobUserAuth struct {
		APIKey     string `json:"apiKey"`
		Secret     string `json:"secret"`
		Passphrase string `json:"passphrase"`
	}

	type clobUserSubscription struct {
		Markets []string     `json:"markets"`
		Type    string       `json:"type"`
		Auth    clobUserAuth `json:"auth"`
	}

	msg := clobUserSubscription{
		Markets: markets,
		Type:    "USER",
		Auth: clobUserAuth{
			APIKey:     auth.Key,
			Secret:     auth.Secret,
			Passphrase: auth.Passphrase,
		},
	}

	return json.Marshal(msg)
}

// ParseMessage parses a raw message from CLOB User endpoint
func (p *ClobUserProtocol) ParseMessage(data []byte) (*Message, error) {
	// CLOB User messages have an event_type field
	// We need to map it to our Message structure
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse CLOB message: %w", err)
	}

	// Extract event_type and convert to our Message format
	eventType, ok := raw["event_type"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid event_type field")
	}

	// Map CLOB event types to our message structure
	msg := &Message{
		Topic: TopicCLOBUser,
		Type:  MessageType(eventType),
	}

	// Re-marshal the data to store as payload
	payload, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	msg.Payload = payload

	return msg, nil
}

// GetProtocolName returns a human-readable name for this protocol
func (p *ClobUserProtocol) GetProtocolName() string {
	return "CLOB User"
}

// ClobMarketProtocol implements the Protocol interface for the CLOB Market endpoint
type ClobMarketProtocol struct{}

// NewClobMarketProtocol creates a new CLOB Market protocol instance
func NewClobMarketProtocol() Protocol {
	return &ClobMarketProtocol{}
}

// GetDefaultHost returns the default WebSocket host for CLOB Market channel
func (p *ClobMarketProtocol) GetDefaultHost() string {
	return "wss://ws-subscriptions-clob.polymarket.com/ws/market"
}

// FormatSubscription formats subscriptions into the CLOB Market protocol message format
func (p *ClobMarketProtocol) FormatSubscription(subscriptions []Subscription) ([]byte, error) {
	if len(subscriptions) == 0 {
		return nil, fmt.Errorf("no subscriptions provided")
	}

	// Extract asset IDs from subscriptions
	assetIDs := make([]string, 0)

	for _, sub := range subscriptions {
		if sub.Filters != "" {
			assetIDs = append(assetIDs, sub.Filters)
		}
	}

	// CLOB Market subscription format:
	// {
	//   "assets_ids": ["asset1", "asset2"],
	//   "type": "MARKET"
	// }
	type clobMarketSubscription struct {
		AssetsIDs []string `json:"assets_ids"`
		Type      string   `json:"type"`
	}

	msg := clobMarketSubscription{
		AssetsIDs: assetIDs,
		Type:      "MARKET",
	}

	return json.Marshal(msg)
}

// ParseMessage parses a raw message from CLOB Market endpoint
func (p *ClobMarketProtocol) ParseMessage(data []byte) (*Message, error) {
	// CLOB Market messages have an event_type field
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse CLOB message: %w", err)
	}

	// Extract event_type
	eventType, ok := raw["event_type"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid event_type field")
	}

	// Map to our Message structure
	msg := &Message{
		Topic: TopicCLOBMarket,
		Type:  MessageType(eventType),
	}

	// Re-marshal the data to store as payload
	payload, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	msg.Payload = payload

	return msg, nil
}

// GetProtocolName returns a human-readable name for this protocol
func (p *ClobMarketProtocol) GetProtocolName() string {
	return "CLOB Market"
}
