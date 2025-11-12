package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	polymarketdataclient "github.com/ivanzzeth/polymarket-go-real-time-data-client"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Create a typed message router
	router := polymarketdataclient.NewTypedMessageRouter()

	// Register handlers for different message types

	// 1. Activity Trades Handler
	router.RegisterActivityTradesHandler(func(trade polymarketdataclient.Trade) error {
		log.Printf("[Activity Trade] Market: %s, Side: %s, Price: %s, Size: %s",
			trade.Slug, trade.Side, trade.Price.String(), trade.Size.String())
		return nil
	})

	// 2. Activity Orders Matched Handler
	router.RegisterActivityOrdersMatchedHandler(func(trade polymarketdataclient.Trade) error {
		log.Printf("[Orders Matched] Market: %s, Outcome: %s, Price: %s",
			trade.Slug, trade.Outcome, trade.Price.String())
		return nil
	})

	// 3. Comment Created Handler
	router.RegisterCommentCreatedHandler(func(comment polymarketdataclient.Comment) error {
		log.Printf("[Comment Created] ID: %s, Body: %s, User: %s",
			comment.ID, comment.Body, comment.UserAddress)
		return nil
	})

	// 4. Crypto Price Handler
	router.RegisterCryptoPriceHandler(func(price polymarketdataclient.CryptoPrice) error {
		log.Printf("[Crypto Price] Symbol: %s, Value: %s, Time: %s",
			price.Symbol, price.Value.String(), price.Time.Format("2006-01-02 15:04:05"))
		return nil
	})

	// 5. Equity Price Handler
	router.RegisterEquityPriceHandler(func(price polymarketdataclient.EquityPrice) error {
		log.Printf("[Equity Price] Symbol: %s, Value: %s, Time: %s",
			price.Symbol, price.Value.String(), price.Time.Format("2006-01-02 15:04:05"))
		return nil
	})

	// 6. CLOB Order Handler (requires authentication)
	router.RegisterCLOBOrderHandler(func(order polymarketdataclient.CLOBOrder) error {
		log.Printf("[CLOB Order] ID: %s, Market: %s, Side: %s, Price: %s, Size: %s, Status: %s",
			order.ID, order.Market, order.Side, order.Price.String(), order.OriginalSize.String(), order.Status)
		return nil
	})

	// 7. CLOB Trade Handler (requires authentication)
	router.RegisterCLOBTradeHandler(func(trade polymarketdataclient.CLOBTrade) error {
		log.Printf("[CLOB Trade] ID: %s, Market: %s, Side: %s, Price: %s, Size: %s, Status: %s",
			trade.ID, trade.Market, trade.Side, trade.Price.String(), trade.Size.String(), trade.Status)
		return nil
	})

	// 8. Price Changes Handler
	router.RegisterPriceChangesHandler(func(changes polymarketdataclient.PriceChanges) error {
		log.Printf("[Price Changes] Market: %s, Changes: %d", changes.Market, len(changes.PriceChange))
		for _, change := range changes.PriceChange {
			log.Printf("  - Asset: %s, Price: %s, Side: %s, BestBid: %s, BestAsk: %s",
				change.AssetID, change.Price.String(), change.Side, change.BestBid.String(), change.BestAsk.String())
		}
		return nil
	})

	// 9. Aggregated Orderbook Handler
	router.RegisterAggOrderbookHandler(func(orderbook polymarketdataclient.AggOrderbook) error {
		log.Printf("[Agg Orderbook] Market: %s, Asset: %s, Bids: %d, Asks: %d",
			orderbook.Market, orderbook.AssetID, len(orderbook.Bids), len(orderbook.Asks))
		return nil
	})

	// 10. Last Trade Price Handler
	router.RegisterLastTradePriceHandler(func(lastPrice polymarketdataclient.LastTradePrice) error {
		log.Printf("[Last Trade Price] Market: %s, Asset: %s, Price: %s, Side: %s",
			lastPrice.Market, lastPrice.AssetID, lastPrice.Price.String(), lastPrice.Side)
		return nil
	})

	// Create the WebSocket client with the router
	client := polymarketdataclient.New(
		// polymarketdataclient.WithLogger(polymarketdataclient.NewLogger()),
		polymarketdataclient.WithOnConnect(func() {
			log.Println("Connected to Polymarket WebSocket!")
		}),
		polymarketdataclient.WithOnNewMessage(func(data []byte) {
			// Route the message to the appropriate typed handler
			if err := router.RouteMessage(data); err != nil {
				log.Printf("Error routing message: %v", err)
			}
		}),
	)

	// Connect to the server
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Create typed subscription handler
	typedSub := polymarketdataclient.NewTypedSubscriptionHandler(client)

	// Example 1: Subscribe to activity trades for a specific market
	// Filter by event_slug or market_slug
	if err := typedSub.SubscribeToActivityTrades(nil, `{"event_slug":"presidential-election-winner-2024"}`); err != nil {
		log.Printf("Failed to subscribe to activity trades: %v", err)
	}

	// Example 2: Subscribe to activity orders matched
	if err := typedSub.SubscribeToActivityOrdersMatched(nil); err != nil {
		log.Printf("Failed to subscribe to orders matched: %v", err)
	}

	// Example 3: Subscribe to comments for a specific event
	// Filter by parentEntityID and parentEntityType
	if err := typedSub.SubscribeToCommentCreated(nil, `{"parentEntityID":100,"parentEntityType":"Event"}`); err != nil {
		log.Printf("Failed to subscribe to comments: %v", err)
	}

	// Example 4: Subscribe to crypto prices
	// Subscribe to Bitcoin price updates
	if err := typedSub.SubscribeToCryptoPrices(nil, `{"symbol":"BTCUSDT"}`); err != nil {
		log.Printf("Failed to subscribe to crypto prices: %v", err)
	}

	// Example 5: Subscribe to equity prices
	// Subscribe to Apple stock price updates
	if err := typedSub.SubscribeToEquityPrices(nil, `{"symbol":"AAPL"}`); err != nil {
		log.Printf("Failed to subscribe to equity prices: %v", err)
	}

	// Example 6: Subscribe to CLOB user data (requires authentication)
	apiKey := os.Getenv("API_KEY")
	apiSecret := os.Getenv("API_SECRET")
	apiPassphrase := os.Getenv("API_PASSPHRASE")

	if apiKey != "" && apiSecret != "" && apiPassphrase != "" {
		clobAuth := polymarketdataclient.ClobAuth{
			Key:        apiKey,
			Secret:     apiSecret,
			Passphrase: apiPassphrase,
		}

		// Subscribe to user orders
		if err := typedSub.SubscribeToCLOBUserOrders(clobAuth, nil); err != nil {
			log.Printf("Failed to subscribe to CLOB user orders: %v", err)
		}

		// Subscribe to user trades
		if err := typedSub.SubscribeToCLOBUserTrades(clobAuth, nil); err != nil {
			log.Printf("Failed to subscribe to CLOB user trades: %v", err)
		}
	} else {
		log.Println("Skipping CLOB user subscriptions (credentials not provided)")
	}

	// Example 7: Subscribe to CLOB market price changes
	// Filters are mandatory - provide token IDs
	if err := typedSub.SubscribeToCLOBMarketPriceChanges(`["100","200"]`, nil); err != nil {
		log.Printf("Failed to subscribe to price changes: %v", err)
	}

	// Example 8: Subscribe to aggregated orderbook
	if err := typedSub.SubscribeToCLOBMarketAggOrderbook(nil, `["100","200"]`); err != nil {
		log.Printf("Failed to subscribe to agg orderbook: %v", err)
	}

	// Example 9: Subscribe to last trade price
	if err := typedSub.SubscribeToCLOBMarketLastTradePrice(nil, `["100","200"]`); err != nil {
		log.Printf("Failed to subscribe to last trade price: %v", err)
	}

	// Example 10: Subscribe to market created/resolved events
	if err := typedSub.SubscribeToCLOBMarketCreated(nil); err != nil {
		log.Printf("Failed to subscribe to market created: %v", err)
	}

	if err := typedSub.SubscribeToCLOBMarketResolved(nil); err != nil {
		log.Printf("Failed to subscribe to market resolved: %v", err)
	}

	log.Println("Subscriptions complete. Listening for messages...")
	log.Println("Press Ctrl+C to exit")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Keep running until interrupted
	<-sigChan

	log.Println("\nShutting down...")

	// Give some time for graceful shutdown
	time.Sleep(1 * time.Second)

	// Disconnect from the server
	if err := client.Disconnect(); err != nil {
		log.Printf("Error disconnecting: %v", err)
	}

	log.Println("Disconnected successfully")
}
