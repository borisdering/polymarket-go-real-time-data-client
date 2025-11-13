package polymarketrealtime

import (
	"encoding/json"
	"fmt"
)

// ClobUserTypedSubscriptionHandler provides typed subscription methods for CLOB User client
type ClobUserTypedSubscriptionHandler struct {
	client *ClobUserClient
}

// NewClobUserTypedSubscriptionHandler creates a new typed subscription handler for CLOB User
func NewClobUserTypedSubscriptionHandler(client *ClobUserClient) *ClobUserTypedSubscriptionHandler {
	return &ClobUserTypedSubscriptionHandler{client: client}
}

// SubscribeToUserOrders subscribes to user order updates with authentication
// markets: List of market IDs to subscribe to
// auth: CLOB authentication credentials
func (h *ClobUserTypedSubscriptionHandler) SubscribeToUserOrders(markets []string, auth *ClobAuth) error {
	subscriptions := make([]Subscription, len(markets))
	for i, market := range markets {
		subscriptions[i] = Subscription{
			Topic:    TopicCLOBUser,
			Type:     MessageTypeCLOBOrder,
			Filters:  market,
			ClobAuth: auth,
		}
	}
	return h.client.Subscribe(subscriptions)
}

// SubscribeToUserTrades subscribes to user trade updates with authentication
func (h *ClobUserTypedSubscriptionHandler) SubscribeToUserTrades(markets []string, auth *ClobAuth) error {
	subscriptions := make([]Subscription, len(markets))
	for i, market := range markets {
		subscriptions[i] = Subscription{
			Topic:    TopicCLOBUser,
			Type:     MessageTypeCLOBTrade,
			Filters:  market,
			ClobAuth: auth,
		}
	}
	return h.client.Subscribe(subscriptions)
}

// SubscribeToAllUserEvents subscribes to all user events (orders and trades)
func (h *ClobUserTypedSubscriptionHandler) SubscribeToAllUserEvents(markets []string, auth *ClobAuth) error {
	subscriptions := make([]Subscription, 0, len(markets)*2)
	for _, market := range markets {
		subscriptions = append(subscriptions,
			Subscription{
				Topic:    TopicCLOBUser,
				Type:     MessageTypeCLOBOrder,
				Filters:  market,
				ClobAuth: auth,
			},
			Subscription{
				Topic:    TopicCLOBUser,
				Type:     MessageTypeCLOBTrade,
				Filters:  market,
				ClobAuth: auth,
			},
		)
	}
	return h.client.Subscribe(subscriptions)
}

// ClobMarketTypedSubscriptionHandler provides typed subscription methods for CLOB Market client
type ClobMarketTypedSubscriptionHandler struct {
	client *ClobMarketClient
}

// NewClobMarketTypedSubscriptionHandler creates a new typed subscription handler for CLOB Market
func NewClobMarketTypedSubscriptionHandler(client *ClobMarketClient) *ClobMarketTypedSubscriptionHandler {
	return &ClobMarketTypedSubscriptionHandler{client: client}
}

// SubscribeToOrderbook subscribes to orderbook updates for specified assets
func (h *ClobMarketTypedSubscriptionHandler) SubscribeToOrderbook(assetIDs []string) error {
	subscriptions := make([]Subscription, len(assetIDs))
	for i, assetID := range assetIDs {
		subscriptions[i] = Subscription{
			Topic:   TopicCLOBMarket,
			Type:    MessageTypeCLOBAggOrderbook,
			Filters: assetID,
		}
	}
	return h.client.Subscribe(subscriptions)
}

// SubscribeToPriceChanges subscribes to price change updates for specified assets
func (h *ClobMarketTypedSubscriptionHandler) SubscribeToPriceChanges(assetIDs []string) error {
	subscriptions := make([]Subscription, len(assetIDs))
	for i, assetID := range assetIDs {
		subscriptions[i] = Subscription{
			Topic:   TopicCLOBMarket,
			Type:    MessageTypeCLOBPriceChanges,
			Filters: assetID,
		}
	}
	return h.client.Subscribe(subscriptions)
}

// SubscribeToLastTradePrices subscribes to last trade price updates for specified assets
func (h *ClobMarketTypedSubscriptionHandler) SubscribeToLastTradePrices(assetIDs []string) error {
	subscriptions := make([]Subscription, len(assetIDs))
	for i, assetID := range assetIDs {
		subscriptions[i] = Subscription{
			Topic:   TopicCLOBMarket,
			Type:    MessageTypeCLOBLastTradePrice,
			Filters: assetID,
		}
	}
	return h.client.Subscribe(subscriptions)
}

// SubscribeToAllMarketData subscribes to all market data types for specified assets
func (h *ClobMarketTypedSubscriptionHandler) SubscribeToAllMarketData(assetIDs []string) error {
	subscriptions := make([]Subscription, 0, len(assetIDs)*3)
	for _, assetID := range assetIDs {
		subscriptions = append(subscriptions,
			Subscription{
				Topic:   TopicCLOBMarket,
				Type:    MessageTypeCLOBAggOrderbook,
				Filters: assetID,
			},
			Subscription{
				Topic:   TopicCLOBMarket,
				Type:    MessageTypeCLOBPriceChanges,
				Filters: assetID,
			},
			Subscription{
				Topic:   TopicCLOBMarket,
				Type:    MessageTypeCLOBLastTradePrice,
				Filters: assetID,
			},
		)
	}
	return h.client.Subscribe(subscriptions)
}

// ClobUserMessageRouter routes CLOB User messages to registered typed handlers
type ClobUserMessageRouter struct {
	orderHandlers []CLOBOrderCallback
	tradeHandlers []CLOBTradeCallback
}

// NewClobUserMessageRouter creates a new message router for CLOB User messages
func NewClobUserMessageRouter() *ClobUserMessageRouter {
	return &ClobUserMessageRouter{
		orderHandlers: make([]CLOBOrderCallback, 0),
		tradeHandlers: make([]CLOBTradeCallback, 0),
	}
}

// RegisterOrderHandler registers a callback for CLOB order events
func (r *ClobUserMessageRouter) RegisterOrderHandler(handler CLOBOrderCallback) {
	r.orderHandlers = append(r.orderHandlers, handler)
}

// RegisterTradeHandler registers a callback for CLOB trade events
func (r *ClobUserMessageRouter) RegisterTradeHandler(handler CLOBTradeCallback) {
	r.tradeHandlers = append(r.tradeHandlers, handler)
}

// RouteMessage routes a raw message to the appropriate handlers
func (r *ClobUserMessageRouter) RouteMessage(data []byte) error {
	// Parse the message to determine event type
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	eventType, ok := raw["event_type"].(string)
	if !ok {
		return nil // Skip messages without event_type
	}

	switch eventType {
	case "order":
		return r.routeOrderMessage(data)
	case "trade":
		return r.routeTradeMessage(data)
	}

	return nil
}

func (r *ClobUserMessageRouter) routeOrderMessage(data []byte) error {
	var order CLOBOrder
	if err := json.Unmarshal(data, &order); err != nil {
		return err
	}

	for _, handler := range r.orderHandlers {
		if err := handler(order); err != nil {
			return err
		}
	}

	return nil
}

func (r *ClobUserMessageRouter) routeTradeMessage(data []byte) error {
	var trade CLOBTrade
	if err := json.Unmarshal(data, &trade); err != nil {
		return err
	}

	for _, handler := range r.tradeHandlers {
		if err := handler(trade); err != nil {
			return err
		}
	}

	return nil
}

// ClobMarketMessageRouter routes CLOB Market messages to registered typed handlers
type ClobMarketMessageRouter struct {
	orderbookHandlers      []AggOrderbookCallback
	priceChangesHandlers   []PriceChangesCallback
	lastTradePriceHandlers []LastTradePriceCallback
}

// NewClobMarketMessageRouter creates a new message router for CLOB Market messages
func NewClobMarketMessageRouter() *ClobMarketMessageRouter {
	return &ClobMarketMessageRouter{
		orderbookHandlers:      make([]AggOrderbookCallback, 0),
		priceChangesHandlers:   make([]PriceChangesCallback, 0),
		lastTradePriceHandlers: make([]LastTradePriceCallback, 0),
	}
}

// RegisterOrderbookHandler registers a callback for orderbook updates
func (r *ClobMarketMessageRouter) RegisterOrderbookHandler(handler AggOrderbookCallback) {
	r.orderbookHandlers = append(r.orderbookHandlers, handler)
}

// RegisterPriceChangesHandler registers a callback for price change updates
func (r *ClobMarketMessageRouter) RegisterPriceChangesHandler(handler PriceChangesCallback) {
	r.priceChangesHandlers = append(r.priceChangesHandlers, handler)
}

// RegisterLastTradePriceHandler registers a callback for last trade price updates
func (r *ClobMarketMessageRouter) RegisterLastTradePriceHandler(handler LastTradePriceCallback) {
	r.lastTradePriceHandlers = append(r.lastTradePriceHandlers, handler)
}

// RouteMessage routes a raw message to the appropriate handlers
func (r *ClobMarketMessageRouter) RouteMessage(data []byte) error {
	// Parse the message to determine event type
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("err: %v, raw=%s", err, string(data))
	}

	eventType, ok := raw["event_type"].(string)
	if !ok {
		return nil // Skip messages without event_type
	}

	switch eventType {
	case "book":
		return r.routeOrderbookMessage(data)
	case "price_change":
		return r.routePriceChangesMessage(data)
	case "last_trade_price":
		return r.routeLastTradePriceMessage(data)
	}

	return nil
}

func (r *ClobMarketMessageRouter) routeOrderbookMessage(data []byte) error {
	var orderbook AggOrderbook
	if err := json.Unmarshal(data, &orderbook); err != nil {
		return err
	}

	for _, handler := range r.orderbookHandlers {
		if err := handler(orderbook); err != nil {
			return err
		}
	}

	return nil
}

func (r *ClobMarketMessageRouter) routePriceChangesMessage(data []byte) error {
	var priceChanges PriceChanges
	if err := json.Unmarshal(data, &priceChanges); err != nil {
		return err
	}

	for _, handler := range r.priceChangesHandlers {
		if err := handler(priceChanges); err != nil {
			return err
		}
	}

	return nil
}

func (r *ClobMarketMessageRouter) routeLastTradePriceMessage(data []byte) error {
	var lastTradePrice LastTradePrice
	if err := json.Unmarshal(data, &lastTradePrice); err != nil {
		return err
	}

	for _, handler := range r.lastTradePriceHandlers {
		if err := handler(lastTradePrice); err != nil {
			return err
		}
	}

	return nil
}
