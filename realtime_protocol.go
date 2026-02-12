package polymarketrealtime

import (
	"encoding/json"
	"fmt"
)

// realtimeProtocol implements the Protocol interface for the real-time data endpoint
type realtimeProtocol struct{}

// NewRealtimeProtocol creates a new real-time protocol instance
func NewRealtimeProtocol() Protocol {
	return &realtimeProtocol{}
}

// GetDefaultHost returns the default WebSocket host for real-time data
func (p *realtimeProtocol) GetDefaultHost() string {
	return "wss://ws-live-data.polymarket.com"
}

// FormatSubscription formats subscriptions into the real-time protocol message format
func (p *realtimeProtocol) FormatSubscription(subscriptions []Subscription) ([]byte, error) {
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

// FormatUnsubscribe formats unsubscriptions into the real-time protocol message format
func (p *realtimeProtocol) FormatUnsubscribe(subscriptions []Subscription) ([]byte, error) {
	type subscriptionMessage struct {
		Action        string         `json:"action"`
		Subscriptions []Subscription `json:"subscriptions"`
	}

	msg := subscriptionMessage{
		Action:        "unsubscribe",
		Subscriptions: subscriptions,
	}

	return json.Marshal(msg)
}

// ParseMessage parses a raw message into a structured Message format
func (p *realtimeProtocol) ParseMessage(data []byte) (*Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse message: %w", err)
	}
	return &msg, nil
}

// GetProtocolName returns a human-readable name for this protocol
func (p *realtimeProtocol) GetProtocolName() string {
	return "Real-Time Data"
}
