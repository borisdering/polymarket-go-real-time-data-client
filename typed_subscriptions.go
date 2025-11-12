package polymarketrealtime

import (
	"encoding/json"
	"fmt"
)

// RealtimeTypedSubscriptionHandler provides type-safe subscription handlers for different message types
type RealtimeTypedSubscriptionHandler struct {
	client Client
}

// NewRealtimeTypedSubscriptionHandler creates a new typed subscription handler
func NewRealtimeTypedSubscriptionHandler(client Client) *RealtimeTypedSubscriptionHandler {
	return &RealtimeTypedSubscriptionHandler{
		client: client,
	}
}

// Activity subscription handlers

// ActivityTradesCallback is the callback function for activity trades messages
type ActivityTradesCallback func(trade Trade) error

// SubscribeToActivityTrades subscribes to activity trades with a typed callback
func (h *RealtimeTypedSubscriptionHandler) SubscribeToActivityTrades(callback ActivityTradesCallback, filters ...string) error {
	filter := ""
	if len(filters) > 0 {
		filter = filters[0]
	}

	return h.client.Subscribe([]Subscription{
		{
			Topic:   TopicActivity,
			Type:    MessageTypeTrades,
			Filters: filter,
		},
	})
}

// ActivityOrdersMatchedCallback is the callback function for activity orders matched messages
type ActivityOrdersMatchedCallback func(trade Trade) error

// SubscribeToActivityOrdersMatched subscribes to activity orders matched with a typed callback
func (h *RealtimeTypedSubscriptionHandler) SubscribeToActivityOrdersMatched(callback ActivityOrdersMatchedCallback, filters ...string) error {
	filter := ""
	if len(filters) > 0 {
		filter = filters[0]
	}

	return h.client.Subscribe([]Subscription{
		{
			Topic:   TopicActivity,
			Type:    MessageTypeOrdersMatched,
			Filters: filter,
		},
	})
}

// Comments subscription handlers

// CommentCreatedCallback is the callback function for comment created messages
type CommentCreatedCallback func(comment Comment) error

// SubscribeToCommentCreated subscribes to comment created events with a typed callback
func (h *RealtimeTypedSubscriptionHandler) SubscribeToCommentCreated(callback CommentCreatedCallback, filters ...string) error {
	filter := ""
	if len(filters) > 0 {
		filter = filters[0]
	}

	return h.client.Subscribe([]Subscription{
		{
			Topic:   TopicComments,
			Type:    MessageTypeCommentCreated,
			Filters: filter,
		},
	})
}

// CommentRemovedCallback is the callback function for comment removed messages
type CommentRemovedCallback func(comment Comment) error

// SubscribeToCommentRemoved subscribes to comment removed events with a typed callback
func (h *RealtimeTypedSubscriptionHandler) SubscribeToCommentRemoved(callback CommentRemovedCallback, filters ...string) error {
	filter := ""
	if len(filters) > 0 {
		filter = filters[0]
	}

	return h.client.Subscribe([]Subscription{
		{
			Topic:   TopicComments,
			Type:    MessageTypeCommentRemoved,
			Filters: filter,
		},
	})
}

// ReactionCreatedCallback is the callback function for reaction created messages
type ReactionCreatedCallback func(reaction Reaction) error

// SubscribeToReactionCreated subscribes to reaction created events with a typed callback
func (h *RealtimeTypedSubscriptionHandler) SubscribeToReactionCreated(callback ReactionCreatedCallback, filters ...string) error {
	filter := ""
	if len(filters) > 0 {
		filter = filters[0]
	}

	return h.client.Subscribe([]Subscription{
		{
			Topic:   TopicComments,
			Type:    MessageTypeReactionCreated,
			Filters: filter,
		},
	})
}

// ReactionRemovedCallback is the callback function for reaction removed messages
type ReactionRemovedCallback func(reaction Reaction) error

// SubscribeToReactionRemoved subscribes to reaction removed events with a typed callback
func (h *RealtimeTypedSubscriptionHandler) SubscribeToReactionRemoved(callback ReactionRemovedCallback, filters ...string) error {
	filter := ""
	if len(filters) > 0 {
		filter = filters[0]
	}

	return h.client.Subscribe([]Subscription{
		{
			Topic:   TopicComments,
			Type:    MessageTypeReactionRemoved,
			Filters: filter,
		},
	})
}

// RFQ subscription handlers

// RFQRequestCallback is the callback function for RFQ request messages
type RFQRequestCallback func(request RFQRequest) error

// SubscribeToRFQRequestCreated subscribes to RFQ request created events
func (h *RealtimeTypedSubscriptionHandler) SubscribeToRFQRequestCreated(callback RFQRequestCallback) error {
	return h.client.Subscribe([]Subscription{
		{
			Topic: TopicRfq,
			Type:  MessageTypeRequestCreated,
		},
	})
}

// SubscribeToRFQRequestEdited subscribes to RFQ request edited events
func (h *RealtimeTypedSubscriptionHandler) SubscribeToRFQRequestEdited(callback RFQRequestCallback) error {
	return h.client.Subscribe([]Subscription{
		{
			Topic: TopicRfq,
			Type:  MessageTypeRequestEdited,
		},
	})
}

// SubscribeToRFQRequestCanceled subscribes to RFQ request canceled events
func (h *RealtimeTypedSubscriptionHandler) SubscribeToRFQRequestCanceled(callback RFQRequestCallback) error {
	return h.client.Subscribe([]Subscription{
		{
			Topic: TopicRfq,
			Type:  MessageTypeRequestCanceled,
		},
	})
}

// SubscribeToRFQRequestExpired subscribes to RFQ request expired events
func (h *RealtimeTypedSubscriptionHandler) SubscribeToRFQRequestExpired(callback RFQRequestCallback) error {
	return h.client.Subscribe([]Subscription{
		{
			Topic: TopicRfq,
			Type:  MessageTypeRequestExpired,
		},
	})
}

// RFQQuoteCallback is the callback function for RFQ quote messages
type RFQQuoteCallback func(quote RFQQuote) error

// SubscribeToRFQQuoteCreated subscribes to RFQ quote created events
func (h *RealtimeTypedSubscriptionHandler) SubscribeToRFQQuoteCreated(callback RFQQuoteCallback) error {
	return h.client.Subscribe([]Subscription{
		{
			Topic: TopicRfq,
			Type:  MessageTypeQuoteCreated,
		},
	})
}

// SubscribeToRFQQuoteEdited subscribes to RFQ quote edited events
func (h *RealtimeTypedSubscriptionHandler) SubscribeToRFQQuoteEdited(callback RFQQuoteCallback) error {
	return h.client.Subscribe([]Subscription{
		{
			Topic: TopicRfq,
			Type:  MessageTypeQuoteEdited,
		},
	})
}

// SubscribeToRFQQuoteCanceled subscribes to RFQ quote canceled events
func (h *RealtimeTypedSubscriptionHandler) SubscribeToRFQQuoteCanceled(callback RFQQuoteCallback) error {
	return h.client.Subscribe([]Subscription{
		{
			Topic: TopicRfq,
			Type:  MessageTypeQuoteCanceled,
		},
	})
}

// SubscribeToRFQQuoteExpired subscribes to RFQ quote expired events
func (h *RealtimeTypedSubscriptionHandler) SubscribeToRFQQuoteExpired(callback RFQQuoteCallback) error {
	return h.client.Subscribe([]Subscription{
		{
			Topic: TopicRfq,
			Type:  MessageTypeQuoteExpired,
		},
	})
}

// Crypto Prices subscription handlers

// CryptoPriceCallback is the callback function for crypto price update messages
type CryptoPriceCallback func(price CryptoPrice) error

// SubscribeToCryptoPrices subscribes to crypto price updates
// filters example: `{"symbol":"BTCUSDT"}`
func (h *RealtimeTypedSubscriptionHandler) SubscribeToCryptoPrices(callback CryptoPriceCallback, filters ...string) error {
	filter := ""
	if len(filters) > 0 {
		filter = filters[0]
	}

	return h.client.Subscribe([]Subscription{
		{
			Topic:   TopicCryptoPrices,
			Type:    MessageTypeUpdate,
			Filters: filter,
		},
	})
}

// SubscribeToCryptoPricesChainlink subscribes to crypto price updates from Chainlink
// filters example: `{"symbol":"BTCUSDT"}`
func (h *RealtimeTypedSubscriptionHandler) SubscribeToCryptoPricesChainlink(callback CryptoPriceCallback, filters ...string) error {
	filter := ""
	if len(filters) > 0 {
		filter = filters[0]
	}

	return h.client.Subscribe([]Subscription{
		{
			Topic:   TopicCryptoPricesChainlink,
			Type:    MessageTypeUpdate,
			Filters: filter,
		},
	})
}

// Equity Prices subscription handlers

// EquityPriceCallback is the callback function for equity price update messages
type EquityPriceCallback func(price EquityPrice) error

// SubscribeToEquityPrices subscribes to equity price updates
// filters example: `{"symbol":"AAPL"}`
func (h *RealtimeTypedSubscriptionHandler) SubscribeToEquityPrices(callback EquityPriceCallback, filters ...string) error {
	filter := ""
	if len(filters) > 0 {
		filter = filters[0]
	}

	return h.client.Subscribe([]Subscription{
		{
			Topic:   TopicEquityPrices,
			Type:    MessageTypeUpdate,
			Filters: filter,
		},
	})
}

// CLOB User subscription handlers

// CLOBOrderCallback is the callback function for CLOB order messages
type CLOBOrderCallback func(order CLOBOrder) error

// SubscribeToCLOBUserOrders subscribes to CLOB user orders
func (h *RealtimeTypedSubscriptionHandler) SubscribeToCLOBUserOrders(auth ClobAuth, callback CLOBOrderCallback) error {
	return h.client.Subscribe([]Subscription{
		{
			Topic:    TopicClobUser,
			Type:     MessageTypeOrder,
			ClobAuth: &auth,
		},
	})
}

// CLOBTradeCallback is the callback function for CLOB trade messages
type CLOBTradeCallback func(trade CLOBTrade) error

// SubscribeToCLOBUserTrades subscribes to CLOB user trades
func (h *RealtimeTypedSubscriptionHandler) SubscribeToCLOBUserTrades(auth ClobAuth, callback CLOBTradeCallback) error {
	return h.client.Subscribe([]Subscription{
		{
			Topic:    TopicClobUser,
			Type:     MessageTypeTrade,
			ClobAuth: &auth,
		},
	})
}

// SubscribeToCLOBUserAll subscribes to all CLOB user messages (orders and trades)
func (h *RealtimeTypedSubscriptionHandler) SubscribeToCLOBUserAll(auth ClobAuth) error {
	return h.client.Subscribe([]Subscription{
		{
			Topic:    TopicClobUser,
			Type:     MessageTypeAll,
			ClobAuth: &auth,
		},
	})
}

// CLOB Market subscription handlers

// PriceChangesCallback is the callback function for price changes messages
type PriceChangesCallback func(changes PriceChanges) error

// SubscribeToCLOBMarketPriceChanges subscribes to CLOB market price changes
// filters are mandatory and should contain token IDs, example: `["100","200"]`
func (h *RealtimeTypedSubscriptionHandler) SubscribeToCLOBMarketPriceChanges(filters string, callback PriceChangesCallback) error {
	if filters == "" {
		return fmt.Errorf("filters are mandatory for price_change subscription")
	}

	return h.client.Subscribe([]Subscription{
		{
			Topic:   TopicClobMarket,
			Type:    MessageTypePriceChange,
			Filters: filters,
		},
	})
}

// AggOrderbookCallback is the callback function for aggregated orderbook messages
type AggOrderbookCallback func(orderbook AggOrderbook) error

// SubscribeToCLOBMarketAggOrderbook subscribes to CLOB market aggregated orderbook
// filters example: `["100","200"]` (token IDs)
func (h *RealtimeTypedSubscriptionHandler) SubscribeToCLOBMarketAggOrderbook(callback AggOrderbookCallback, filters ...string) error {
	filter := ""
	if len(filters) > 0 {
		filter = filters[0]
	}

	return h.client.Subscribe([]Subscription{
		{
			Topic:   TopicClobMarket,
			Type:    MessageTypeAggOrderbook,
			Filters: filter,
		},
	})
}

// LastTradePriceCallback is the callback function for last trade price messages
type LastTradePriceCallback func(lastPrice LastTradePrice) error

// SubscribeToCLOBMarketLastTradePrice subscribes to CLOB market last trade price
// filters example: `["100","200"]` (token IDs)
func (h *RealtimeTypedSubscriptionHandler) SubscribeToCLOBMarketLastTradePrice(callback LastTradePriceCallback, filters ...string) error {
	filter := ""
	if len(filters) > 0 {
		filter = filters[0]
	}

	return h.client.Subscribe([]Subscription{
		{
			Topic:   TopicClobMarket,
			Type:    MessageTypeLastTradePrice,
			Filters: filter,
		},
	})
}

// TickSizeChangeCallback is the callback function for tick size change messages
type TickSizeChangeCallback func(change TickSizeChange) error

// SubscribeToCLOBMarketTickSizeChange subscribes to CLOB market tick size changes
// filters example: `["100","200"]` (token IDs)
func (h *RealtimeTypedSubscriptionHandler) SubscribeToCLOBMarketTickSizeChange(callback TickSizeChangeCallback, filters ...string) error {
	filter := ""
	if len(filters) > 0 {
		filter = filters[0]
	}

	return h.client.Subscribe([]Subscription{
		{
			Topic:   TopicClobMarket,
			Type:    MessageTypeTickSizeChange,
			Filters: filter,
		},
	})
}

// ClobMarketCallback is the callback function for CLOB market messages
type ClobMarketCallback func(market ClobMarket) error

// SubscribeToCLOBMarketCreated subscribes to CLOB market created events
func (h *RealtimeTypedSubscriptionHandler) SubscribeToCLOBMarketCreated(callback ClobMarketCallback) error {
	return h.client.Subscribe([]Subscription{
		{
			Topic: TopicClobMarket,
			Type:  MessageTypeMarketCreated,
		},
	})
}

// SubscribeToCLOBMarketResolved subscribes to CLOB market resolved events
func (h *RealtimeTypedSubscriptionHandler) SubscribeToCLOBMarketResolved(callback ClobMarketCallback) error {
	return h.client.Subscribe([]Subscription{
		{
			Topic: TopicClobMarket,
			Type:  MessageTypeMarketResolved,
		},
	})
}

// RealtimeMessageRouter routes incoming messages to registered typed callbacks
type RealtimeMessageRouter struct {
	// Activity handlers
	activityTradesHandlers        []ActivityTradesCallback
	activityOrdersMatchedHandlers []ActivityOrdersMatchedCallback

	// Comments handlers
	commentCreatedHandlers  []CommentCreatedCallback
	commentRemovedHandlers  []CommentRemovedCallback
	reactionCreatedHandlers []ReactionCreatedCallback
	reactionRemovedHandlers []ReactionRemovedCallback

	// RFQ handlers
	rfqRequestHandlers []RFQRequestCallback
	rfqQuoteHandlers   []RFQQuoteCallback

	// Price handlers
	cryptoPriceHandlers []CryptoPriceCallback
	equityPriceHandlers []EquityPriceCallback

	// CLOB User handlers
	clobOrderHandlers []CLOBOrderCallback
	clobTradeHandlers []CLOBTradeCallback

	// CLOB Market handlers
	priceChangesHandlers   []PriceChangesCallback
	aggOrderbookHandlers   []AggOrderbookCallback
	lastTradePriceHandlers []LastTradePriceCallback
	tickSizeChangeHandlers []TickSizeChangeCallback
	clobMarketHandlers     []ClobMarketCallback
}

// NewRealtimeMessageRouter creates a new typed message router
func NewRealtimeMessageRouter() *RealtimeMessageRouter {
	return &RealtimeMessageRouter{}
}

// RegisterActivityTradesHandler registers a handler for activity trades
func (r *RealtimeMessageRouter) RegisterActivityTradesHandler(handler ActivityTradesCallback) {
	r.activityTradesHandlers = append(r.activityTradesHandlers, handler)
}

// RegisterActivityOrdersMatchedHandler registers a handler for activity orders matched
func (r *RealtimeMessageRouter) RegisterActivityOrdersMatchedHandler(handler ActivityOrdersMatchedCallback) {
	r.activityOrdersMatchedHandlers = append(r.activityOrdersMatchedHandlers, handler)
}

// RegisterCommentCreatedHandler registers a handler for comment created events
func (r *RealtimeMessageRouter) RegisterCommentCreatedHandler(handler CommentCreatedCallback) {
	r.commentCreatedHandlers = append(r.commentCreatedHandlers, handler)
}

// RegisterCommentRemovedHandler registers a handler for comment removed events
func (r *RealtimeMessageRouter) RegisterCommentRemovedHandler(handler CommentRemovedCallback) {
	r.commentRemovedHandlers = append(r.commentRemovedHandlers, handler)
}

// RegisterReactionCreatedHandler registers a handler for reaction created events
func (r *RealtimeMessageRouter) RegisterReactionCreatedHandler(handler ReactionCreatedCallback) {
	r.reactionCreatedHandlers = append(r.reactionCreatedHandlers, handler)
}

// RegisterReactionRemovedHandler registers a handler for reaction removed events
func (r *RealtimeMessageRouter) RegisterReactionRemovedHandler(handler ReactionRemovedCallback) {
	r.reactionRemovedHandlers = append(r.reactionRemovedHandlers, handler)
}

// RegisterRFQRequestHandler registers a handler for RFQ request messages
func (r *RealtimeMessageRouter) RegisterRFQRequestHandler(handler RFQRequestCallback) {
	r.rfqRequestHandlers = append(r.rfqRequestHandlers, handler)
}

// RegisterRFQQuoteHandler registers a handler for RFQ quote messages
func (r *RealtimeMessageRouter) RegisterRFQQuoteHandler(handler RFQQuoteCallback) {
	r.rfqQuoteHandlers = append(r.rfqQuoteHandlers, handler)
}

// RegisterCryptoPriceHandler registers a handler for crypto price updates
func (r *RealtimeMessageRouter) RegisterCryptoPriceHandler(handler CryptoPriceCallback) {
	r.cryptoPriceHandlers = append(r.cryptoPriceHandlers, handler)
}

// RegisterEquityPriceHandler registers a handler for equity price updates
func (r *RealtimeMessageRouter) RegisterEquityPriceHandler(handler EquityPriceCallback) {
	r.equityPriceHandlers = append(r.equityPriceHandlers, handler)
}

// RegisterCLOBOrderHandler registers a handler for CLOB order messages
func (r *RealtimeMessageRouter) RegisterCLOBOrderHandler(handler CLOBOrderCallback) {
	r.clobOrderHandlers = append(r.clobOrderHandlers, handler)
}

// RegisterCLOBTradeHandler registers a handler for CLOB trade messages
func (r *RealtimeMessageRouter) RegisterCLOBTradeHandler(handler CLOBTradeCallback) {
	r.clobTradeHandlers = append(r.clobTradeHandlers, handler)
}

// RegisterPriceChangesHandler registers a handler for price changes
func (r *RealtimeMessageRouter) RegisterPriceChangesHandler(handler PriceChangesCallback) {
	r.priceChangesHandlers = append(r.priceChangesHandlers, handler)
}

// RegisterAggOrderbookHandler registers a handler for aggregated orderbook
func (r *RealtimeMessageRouter) RegisterAggOrderbookHandler(handler AggOrderbookCallback) {
	r.aggOrderbookHandlers = append(r.aggOrderbookHandlers, handler)
}

// RegisterLastTradePriceHandler registers a handler for last trade price
func (r *RealtimeMessageRouter) RegisterLastTradePriceHandler(handler LastTradePriceCallback) {
	r.lastTradePriceHandlers = append(r.lastTradePriceHandlers, handler)
}

// RegisterTickSizeChangeHandler registers a handler for tick size changes
func (r *RealtimeMessageRouter) RegisterTickSizeChangeHandler(handler TickSizeChangeCallback) {
	r.tickSizeChangeHandlers = append(r.tickSizeChangeHandlers, handler)
}

// RegisterClobMarketHandler registers a handler for CLOB market events
func (r *RealtimeMessageRouter) RegisterClobMarketHandler(handler ClobMarketCallback) {
	r.clobMarketHandlers = append(r.clobMarketHandlers, handler)
}

// RouteMessage routes a raw message to the appropriate typed handlers
func (r *RealtimeMessageRouter) RouteMessage(data []byte) error {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	switch msg.Topic {
	case TopicActivity:
		return r.routeActivityMessage(msg)
	case TopicComments:
		return r.routeCommentsMessage(msg)
	case TopicRfq:
		return r.routeRFQMessage(msg)
	case TopicCryptoPrices, TopicCryptoPricesChainlink:
		return r.routeCryptoPriceMessage(msg)
	case TopicEquityPrices:
		return r.routeEquityPriceMessage(msg)
	case TopicClobUser:
		return r.routeCLOBUserMessage(msg)
	case TopicClobMarket:
		return r.routeCLOBMarketMessage(msg)
	default:
		return fmt.Errorf("unknown topic: %s", msg.Topic)
	}
}

func (r *RealtimeMessageRouter) routeActivityMessage(msg Message) error {
	switch msg.Type {
	case MessageTypeTrades:
		var trade Trade
		if err := json.Unmarshal(msg.Payload, &trade); err != nil {
			return fmt.Errorf("failed to unmarshal trade: %w", err)
		}
		for _, handler := range r.activityTradesHandlers {
			if err := handler(trade); err != nil {
				return err
			}
		}
	case MessageTypeOrdersMatched:
		var trade Trade
		if err := json.Unmarshal(msg.Payload, &trade); err != nil {
			return fmt.Errorf("failed to unmarshal orders matched: %w", err)
		}
		for _, handler := range r.activityOrdersMatchedHandlers {
			if err := handler(trade); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *RealtimeMessageRouter) routeCommentsMessage(msg Message) error {
	switch msg.Type {
	case MessageTypeCommentCreated:
		var comment Comment
		if err := json.Unmarshal(msg.Payload, &comment); err != nil {
			return fmt.Errorf("failed to unmarshal comment: %w", err)
		}
		for _, handler := range r.commentCreatedHandlers {
			if err := handler(comment); err != nil {
				return err
			}
		}
	case MessageTypeCommentRemoved:
		var comment Comment
		if err := json.Unmarshal(msg.Payload, &comment); err != nil {
			return fmt.Errorf("failed to unmarshal comment: %w", err)
		}
		for _, handler := range r.commentRemovedHandlers {
			if err := handler(comment); err != nil {
				return err
			}
		}
	case MessageTypeReactionCreated:
		var reaction Reaction
		if err := json.Unmarshal(msg.Payload, &reaction); err != nil {
			return fmt.Errorf("failed to unmarshal reaction: %w", err)
		}
		for _, handler := range r.reactionCreatedHandlers {
			if err := handler(reaction); err != nil {
				return err
			}
		}
	case MessageTypeReactionRemoved:
		var reaction Reaction
		if err := json.Unmarshal(msg.Payload, &reaction); err != nil {
			return fmt.Errorf("failed to unmarshal reaction: %w", err)
		}
		for _, handler := range r.reactionRemovedHandlers {
			if err := handler(reaction); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *RealtimeMessageRouter) routeRFQMessage(msg Message) error {
	switch msg.Type {
	case MessageTypeRequestCreated, MessageTypeRequestEdited, MessageTypeRequestCanceled, MessageTypeRequestExpired:
		var request RFQRequest
		if err := json.Unmarshal(msg.Payload, &request); err != nil {
			return fmt.Errorf("failed to unmarshal RFQ request: %w", err)
		}
		for _, handler := range r.rfqRequestHandlers {
			if err := handler(request); err != nil {
				return err
			}
		}
	case MessageTypeQuoteCreated, MessageTypeQuoteEdited, MessageTypeQuoteCanceled, MessageTypeQuoteExpired:
		var quote RFQQuote
		if err := json.Unmarshal(msg.Payload, &quote); err != nil {
			return fmt.Errorf("failed to unmarshal RFQ quote: %w", err)
		}
		for _, handler := range r.rfqQuoteHandlers {
			if err := handler(quote); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *RealtimeMessageRouter) routeCryptoPriceMessage(msg Message) error {
	if msg.Type == MessageTypeUpdate {
		var price CryptoPrice
		if err := json.Unmarshal(msg.Payload, &price); err != nil {
			return fmt.Errorf("failed to unmarshal crypto price: %w", err)
		}
		for _, handler := range r.cryptoPriceHandlers {
			if err := handler(price); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *RealtimeMessageRouter) routeEquityPriceMessage(msg Message) error {
	if msg.Type == MessageTypeUpdate {
		var price EquityPrice
		if err := json.Unmarshal(msg.Payload, &price); err != nil {
			return fmt.Errorf("failed to unmarshal equity price: %w", err)
		}
		for _, handler := range r.equityPriceHandlers {
			if err := handler(price); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *RealtimeMessageRouter) routeCLOBUserMessage(msg Message) error {
	switch msg.Type {
	case MessageTypeOrder:
		var order CLOBOrder
		if err := json.Unmarshal(msg.Payload, &order); err != nil {
			return fmt.Errorf("failed to unmarshal CLOB order: %w", err)
		}
		for _, handler := range r.clobOrderHandlers {
			if err := handler(order); err != nil {
				return err
			}
		}
	case MessageTypeTrade:
		var trade CLOBTrade
		if err := json.Unmarshal(msg.Payload, &trade); err != nil {
			return fmt.Errorf("failed to unmarshal CLOB trade: %w", err)
		}
		for _, handler := range r.clobTradeHandlers {
			if err := handler(trade); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *RealtimeMessageRouter) routeCLOBMarketMessage(msg Message) error {
	switch msg.Type {
	case MessageTypePriceChange:
		var changes PriceChanges
		if err := json.Unmarshal(msg.Payload, &changes); err != nil {
			return fmt.Errorf("failed to unmarshal price changes: %w", err)
		}
		for _, handler := range r.priceChangesHandlers {
			if err := handler(changes); err != nil {
				return err
			}
		}
	case MessageTypeAggOrderbook:
		var orderbook AggOrderbook
		if err := json.Unmarshal(msg.Payload, &orderbook); err != nil {
			return fmt.Errorf("failed to unmarshal agg orderbook: %w", err)
		}
		for _, handler := range r.aggOrderbookHandlers {
			if err := handler(orderbook); err != nil {
				return err
			}
		}
	case MessageTypeLastTradePrice:
		var lastPrice LastTradePrice
		if err := json.Unmarshal(msg.Payload, &lastPrice); err != nil {
			return fmt.Errorf("failed to unmarshal last trade price: %w", err)
		}
		for _, handler := range r.lastTradePriceHandlers {
			if err := handler(lastPrice); err != nil {
				return err
			}
		}
	case MessageTypeTickSizeChange:
		var change TickSizeChange
		if err := json.Unmarshal(msg.Payload, &change); err != nil {
			return fmt.Errorf("failed to unmarshal tick size change: %w", err)
		}
		for _, handler := range r.tickSizeChangeHandlers {
			if err := handler(change); err != nil {
				return err
			}
		}
	case MessageTypeMarketCreated, MessageTypeMarketResolved:
		var market ClobMarket
		if err := json.Unmarshal(msg.Payload, &market); err != nil {
			return fmt.Errorf("failed to unmarshal CLOB market: %w", err)
		}
		for _, handler := range r.clobMarketHandlers {
			if err := handler(market); err != nil {
				return err
			}
		}
	}
	return nil
}
