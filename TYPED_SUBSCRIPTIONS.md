# Typed Subscriptions Guide

This guide explains how to use the typed subscription functions in the Polymarket Real-Time Data Client for Go.

## Overview

The typed subscription system provides two main components:

1. **TypedSubscriptionHandler** - Helper functions to subscribe to specific topics with explicit types
2. **TypedMessageRouter** - Router that automatically decodes messages and calls registered handlers with properly typed payloads

## Benefits

- **Type Safety**: All callbacks receive properly typed structs instead of raw JSON
- **Easy to Use**: Simple API for subscribing to different message types
- **Automatic Decoding**: Messages are automatically decoded from JSON to Go structs
- **Multiple Handlers**: You can register multiple handlers for the same message type

## Quick Start

### 1. Create a Message Router

```go
router := polymarketrealtime.NewTypedMessageRouter()
```

### 2. Register Handlers

Register typed handlers for the message types you're interested in:

```go
// Register a handler for CLOB trades
router.RegisterCLOBTradeHandler(func(trade polymarketrealtime.CLOBTrade) error {
    log.Printf("Trade: %s at price %s", trade.ID, trade.Price.String())
    return nil
})

// Register a handler for activity trades
router.RegisterActivityTradesHandler(func(trade polymarketrealtime.Trade) error {
    log.Printf("Market: %s, Side: %s, Price: %s", trade.Slug, trade.Side, trade.Price.String())
    return nil
})
```

### 3. Create WebSocket Client

Create a client that routes messages through your router:

```go
client := polymarketrealtime.New(
    polymarketrealtime.WithOnConnect(func() {
        log.Println("Connected!")
    }),
    polymarketrealtime.WithOnNewMessage(func(data []byte) {
        if err := router.RouteMessage(data); err != nil {
            log.Printf("Error: %v", err)
        }
    }),
)
```

### 4. Subscribe to Topics

Use the typed subscription handler to subscribe:

```go
typedSub := polymarketrealtime.NewTypedSubscriptionHandler(client)

// Subscribe to CLOB user trades (requires authentication)
clobAuth := polymarketrealtime.ClobAuth{
    Key:        "your-api-key",
    Secret:     "your-api-secret",
    Passphrase: "your-passphrase",
}
typedSub.SubscribeToCLOBUserTrades(clobAuth, nil)
```

## Available Subscriptions

### Activity Topic

#### Subscribe to Trades
```go
typedSub.SubscribeToActivityTrades(callback, `{"event_slug":"presidential-election-winner-2024"}`)
```

Handler registration:
```go
router.RegisterActivityTradesHandler(func(trade polymarketrealtime.Trade) error {
    // Handle trade
    return nil
})
```

#### Subscribe to Orders Matched
```go
typedSub.SubscribeToActivityOrdersMatched(callback, `{"market_slug":"your-market"}`)
```

Handler registration:
```go
router.RegisterActivityOrdersMatchedHandler(func(trade polymarketrealtime.Trade) error {
    // Handle matched order
    return nil
})
```

### Comments Topic

#### Subscribe to Comment Created
```go
typedSub.SubscribeToCommentCreated(callback, `{"parentEntityID":100,"parentEntityType":"Event"}`)
```

Handler registration:
```go
router.RegisterCommentCreatedHandler(func(comment polymarketrealtime.Comment) error {
    // Handle new comment
    return nil
})
```

#### Subscribe to Comment Removed
```go
typedSub.SubscribeToCommentRemoved(callback, `{"parentEntityID":100,"parentEntityType":"Event"}`)
```

#### Subscribe to Reaction Created/Removed
```go
typedSub.SubscribeToReactionCreated(callback, `{"parentEntityID":100,"parentEntityType":"Event"}`)
typedSub.SubscribeToReactionRemoved(callback, `{"parentEntityID":100,"parentEntityType":"Event"}`)
```

### RFQ Topic

#### Subscribe to RFQ Requests
```go
typedSub.SubscribeToRFQRequestCreated(callback)
typedSub.SubscribeToRFQRequestEdited(callback)
typedSub.SubscribeToRFQRequestCanceled(callback)
typedSub.SubscribeToRFQRequestExpired(callback)
```

Handler registration:
```go
router.RegisterRFQRequestHandler(func(request polymarketrealtime.RFQRequest) error {
    // Handle RFQ request
    return nil
})
```

#### Subscribe to RFQ Quotes
```go
typedSub.SubscribeToRFQQuoteCreated(callback)
typedSub.SubscribeToRFQQuoteEdited(callback)
typedSub.SubscribeToRFQQuoteCanceled(callback)
typedSub.SubscribeToRFQQuoteExpired(callback)
```

Handler registration:
```go
router.RegisterRFQQuoteHandler(func(quote polymarketrealtime.RFQQuote) error {
    // Handle RFQ quote
    return nil
})
```

### Crypto Prices Topic

#### Subscribe to Crypto Prices
```go
// Regular crypto prices
typedSub.SubscribeToCryptoPrices(callback, `{"symbol":"BTCUSDT"}`)

// Chainlink crypto prices
typedSub.SubscribeToCryptoPricesChainlink(callback, `{"symbol":"ETHUSDT"}`)
```

Handler registration:
```go
router.RegisterCryptoPriceHandler(func(price polymarketrealtime.CryptoPrice) error {
    log.Printf("Symbol: %s, Value: %s", price.Symbol, price.Value.String())
    return nil
})
```

Supported symbols:
- BTCUSDT
- ETHUSDT
- XRPUSDT
- SOLUSDT
- DOGEUSDT

### Equity Prices Topic

#### Subscribe to Equity Prices
```go
typedSub.SubscribeToEquityPrices(callback, `{"symbol":"AAPL"}`)
```

Handler registration:
```go
router.RegisterEquityPriceHandler(func(price polymarketrealtime.EquityPrice) error {
    log.Printf("Symbol: %s, Value: %s", price.Symbol, price.Value.String())
    return nil
})
```

Supported symbols:
- AAPL, TSLA, MSFT, GOOGL, AMZN, META, NVDA, NFLX, PLTR, OPEN, RKLB, ABNB

### CLOB User Topic (Requires Authentication)

#### Subscribe to User Orders
```go
clobAuth := polymarketrealtime.ClobAuth{
    Key:        "your-api-key",
    Secret:     "your-api-secret",
    Passphrase: "your-passphrase",
}
typedSub.SubscribeToCLOBUserOrders(clobAuth, callback)
```

Handler registration:
```go
router.RegisterCLOBOrderHandler(func(order polymarketrealtime.CLOBOrder) error {
    log.Printf("Order: %s, Status: %s, Price: %s", order.ID, order.Status, order.Price.String())
    return nil
})
```

#### Subscribe to User Trades
```go
typedSub.SubscribeToCLOBUserTrades(clobAuth, callback)
```

Handler registration:
```go
router.RegisterCLOBTradeHandler(func(trade polymarketrealtime.CLOBTrade) error {
    log.Printf("Trade: %s, Price: %s, Size: %s", trade.ID, trade.Price.String(), trade.Size.String())
    return nil
})
```

#### Subscribe to All User Events
```go
typedSub.SubscribeToCLOBUserAll(clobAuth)
```

### CLOB Market Topic

#### Subscribe to Price Changes
```go
// Filters are mandatory for price changes - must include token IDs
typedSub.SubscribeToCLOBMarketPriceChanges(`["100","200"]`, callback)
```

Handler registration:
```go
router.RegisterPriceChangesHandler(func(changes polymarketrealtime.PriceChanges) error {
    for _, change := range changes.PriceChange {
        log.Printf("Asset: %s, Price: %s, BestBid: %s, BestAsk: %s",
            change.AssetID, change.Price.String(), change.BestBid.String(), change.BestAsk.String())
    }
    return nil
})
```

#### Subscribe to Aggregated Orderbook
```go
typedSub.SubscribeToCLOBMarketAggOrderbook(callback, `["100","200"]`)
```

Handler registration:
```go
router.RegisterAggOrderbookHandler(func(orderbook polymarketrealtime.AggOrderbook) error {
    log.Printf("Market: %s, Bids: %d, Asks: %d", orderbook.Market, len(orderbook.Bids), len(orderbook.Asks))
    return nil
})
```

#### Subscribe to Last Trade Price
```go
typedSub.SubscribeToCLOBMarketLastTradePrice(callback, `["100","200"]`)
```

Handler registration:
```go
router.RegisterLastTradePriceHandler(func(lastPrice polymarketrealtime.LastTradePrice) error {
    log.Printf("Last price: %s for market %s", lastPrice.Price.String(), lastPrice.Market)
    return nil
})
```

#### Subscribe to Tick Size Changes
```go
typedSub.SubscribeToCLOBMarketTickSizeChange(callback, `["100","200"]`)
```

Handler registration:
```go
router.RegisterTickSizeChangeHandler(func(change polymarketrealtime.TickSizeChange) error {
    log.Printf("Tick size changed from %s to %s", change.OldTickSize.String(), change.NewTickSize.String())
    return nil
})
```

#### Subscribe to Market Events
```go
typedSub.SubscribeToCLOBMarketCreated(callback)
typedSub.SubscribeToCLOBMarketResolved(callback)
```

Handler registration:
```go
router.RegisterClobMarketHandler(func(market polymarketrealtime.ClobMarket) error {
    log.Printf("Market: %s, TickSize: %s, NegRisk: %v", market.Market, market.TickSize.String(), market.NegRisk)
    return nil
})
```

## Complete Example

```go
package main

import (
    "log"
    "time"

    polymarketrealtime "github.com/ivanzzeth/polymarket-go-real-time-data-client"
)

func main() {
    // Create router
    router := polymarketrealtime.NewTypedMessageRouter()

    // Register handlers
    router.RegisterCLOBTradeHandler(func(trade polymarketrealtime.CLOBTrade) error {
        log.Printf("Trade: %s at %s", trade.ID, trade.Price.String())
        return nil
    })

    router.RegisterActivityTradesHandler(func(trade polymarketrealtime.Trade) error {
        log.Printf("Activity: %s %s at %s", trade.Side, trade.Slug, trade.Price.String())
        return nil
    })

    // Create client
    client := polymarketrealtime.New(
        polymarketrealtime.WithOnNewMessage(func(data []byte) {
            router.RouteMessage(data)
        }),
    )

    // Connect
    if err := client.Connect(); err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect()

    // Subscribe
    typedSub := polymarketrealtime.NewTypedSubscriptionHandler(client)

    // Subscribe to activity trades
    typedSub.SubscribeToActivityTrades(nil)

    // Subscribe to CLOB user data (with auth)
    clobAuth := polymarketrealtime.ClobAuth{
        Key:        "your-key",
        Secret:     "your-secret",
        Passphrase: "your-passphrase",
    }
    typedSub.SubscribeToCLOBUserAll(clobAuth)

    // Keep running
    time.Sleep(60 * time.Second)
}
```

## Error Handling

Handlers should return an error if processing fails:

```go
router.RegisterCLOBTradeHandler(func(trade polymarketrealtime.CLOBTrade) error {
    if err := processTradeInDatabase(trade); err != nil {
        return fmt.Errorf("failed to process trade: %w", err)
    }
    return nil
})
```

## Multiple Handlers

You can register multiple handlers for the same message type:

```go
// Handler 1: Log to console
router.RegisterCLOBTradeHandler(func(trade polymarketrealtime.CLOBTrade) error {
    log.Printf("Trade received: %s", trade.ID)
    return nil
})

// Handler 2: Save to database
router.RegisterCLOBTradeHandler(func(trade polymarketrealtime.CLOBTrade) error {
    return database.SaveTrade(trade)
})

// Handler 3: Send notification
router.RegisterCLOBTradeHandler(func(trade polymarketrealtime.CLOBTrade) error {
    return notifier.SendTradeAlert(trade)
})
```

All handlers will be called in the order they were registered.

## Filter Examples

### Activity Filters
```go
// By event slug
`{"event_slug":"presidential-election-winner-2024"}`

// By market slug
`{"market_slug":"will-trump-win-2024"}`
```

### Comments Filters
```go
// By parent entity
`{"parentEntityID":100,"parentEntityType":"Event"}`
`{"parentEntityID":200,"parentEntityType":"Series"}`
```

### Price Filters
```go
// Crypto prices
`{"symbol":"BTCUSDT"}`
`{"symbol":"ETHUSDT"}`

// Equity prices
`{"symbol":"AAPL"}`
`{"symbol":"TSLA"}`
```

### CLOB Market Filters
```go
// Token IDs (required for price_change)
`["100","200","300"]`

// Single token
`["12345"]`
```

## See Also

- [examples/typed_subscriptions/main.go](examples/typed_subscriptions/main.go) - Comprehensive example
- [examples/simple_typed/main.go](examples/simple_typed/main.go) - Simple CLOB user example
- [examples/clob_auth/main.go](examples/clob_auth/main.go) - Original authentication example
