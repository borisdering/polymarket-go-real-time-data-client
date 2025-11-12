package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	polymarketdataclient "github.com/ivanzzeth/polymarket-go-real-time-data-client"
)

func main() {
	log.Println("=== CLOB Market WebSocket Client Demo ===")
	log.Println("This example demonstrates subscribing to CLOB market data")
	log.Println()

	// Create a message router for typed handling
	router := polymarketdataclient.NewClobMarketMessageRouter()

	// Register handlers for different market data types
	router.RegisterOrderbookHandler(func(orderbook polymarketdataclient.AggOrderbook) error {
		log.Printf("[Orderbook Update] Asset: %s, Market: %s, Bids: %d levels, Asks: %d levels",
			orderbook.AssetID,
			orderbook.Market,
			len(orderbook.Bids),
			len(orderbook.Asks),
		)

		// Show top of book
		if len(orderbook.Bids) > 0 {
			log.Printf("  Best Bid: %s @ %s", orderbook.Bids[0].Size.String(), orderbook.Bids[0].Price.String())
		}
		if len(orderbook.Asks) > 0 {
			log.Printf("  Best Ask: %s @ %s", orderbook.Asks[0].Size.String(), orderbook.Asks[0].Price.String())
		}

		return nil
	})

	router.RegisterPriceChangesHandler(func(priceChanges polymarketdataclient.PriceChanges) error {
		log.Printf("[Price Changes] Market: %s, Changes: %d", priceChanges.Market, len(priceChanges.PriceChange))

		for _, change := range priceChanges.PriceChange {
			log.Printf("  Asset: %s, Side: %s, Price: %s, Size: %s, BestBid: %s, BestAsk: %s",
				change.AssetID,
				change.Side,
				change.Price.String(),
				change.Size.String(),
				change.BestBid.String(),
				change.BestAsk.String(),
			)
		}

		return nil
	})

	router.RegisterLastTradePriceHandler(func(lastPrice polymarketdataclient.LastTradePrice) error {
		log.Printf("[Last Trade Price] Asset: %s, Market: %s, Side: %s, Price: %s, Size: %s",
			lastPrice.AssetID,
			lastPrice.Market,
			lastPrice.Side,
			lastPrice.Price.String(),
			lastPrice.Size.String(),
		)
		return nil
	})

	// Create CLOB Market client
	client := polymarketdataclient.NewClobMarketClient(
		polymarketdataclient.WithLogger(polymarketdataclient.NewLogger()),
		polymarketdataclient.WithAutoReconnect(true),
		polymarketdataclient.WithOnConnect(func() {
			log.Println("✅ Connected to CLOB Market endpoint")
		}),
		polymarketdataclient.WithOnDisconnect(func(err error) {
			log.Printf("❌ Disconnected from CLOB Market endpoint: %v", err)
		}),
		polymarketdataclient.WithOnReconnect(func() {
			log.Println("🔄 Reconnected to CLOB Market endpoint")
		}),
		polymarketdataclient.WithOnNewMessage(func(data []byte) {
			// Route message to appropriate handler
			if err := router.RouteMessage(data); err != nil {
				log.Printf("Error routing message: %v", err)
			}
		}),
	)

	// Connect to the server
	log.Println("Connecting to CLOB Market WebSocket...")
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Create typed subscription handler
	typedSub := polymarketdataclient.NewClobMarketTypedSubscriptionHandler(client)

	// Subscribe to market data for specific assets
	// Replace with your actual asset IDs
	assetIDs := []string{
		// Example asset IDs - replace with real ones
		// You can find asset IDs at https://clob.polymarket.com/
		"17123485321386672176776734800460321083207167824796999887262322064541574251437",
	}

	if len(assetIDs) > 0 {
		log.Println("Subscribing to market data...")
		if err := typedSub.SubscribeToAllMarketData(assetIDs); err != nil {
			log.Fatalf("Failed to subscribe: %v", err)
		}
		log.Println("✅ Successfully subscribed to market data")
	} else {
		log.Println("⚠️  No asset IDs specified. Add asset IDs to the 'assetIDs' slice to receive updates.")
		log.Println("   You can find asset IDs at https://clob.polymarket.com/")
	}

	log.Println()
	log.Println("=== Listening for Market Data ===")
	log.Println("- Orderbook updates")
	log.Println("- Price changes")
	log.Println("- Last trade prices")
	log.Println("Press Ctrl+C to exit")
	log.Println()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("\nShutting down...")
	if err := client.Disconnect(); err != nil {
		log.Printf("Error during disconnect: %v", err)
	}
	log.Println("✅ Disconnected successfully")
}
