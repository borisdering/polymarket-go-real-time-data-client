package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	polymarketrealtime "github.com/ivanzzeth/polymarket-go-real-time-data-client"
)

func main() {
	log.Println("=== CLOB Market WebSocket Client Demo ===")
	log.Println("This example demonstrates subscribing to CLOB market data")
	log.Println()

	// Create typed subscription handler with client
	typedSub, client := polymarketrealtime.NewRealtimeTypedSubscriptionHandlerWithOptions(
		polymarketrealtime.WithLogger(polymarketrealtime.NewLogger()),
		polymarketrealtime.WithAutoReconnect(true),
		polymarketrealtime.WithOnConnect(func() {
			log.Println("✅ Connected to CLOB Market endpoint")
		}),
		polymarketrealtime.WithOnDisconnect(func(err error) {
			log.Printf("❌ Disconnected from CLOB Market endpoint: %v", err)
		}),
		polymarketrealtime.WithOnReconnect(func() {
			log.Println("🔄 Reconnected to CLOB Market endpoint")
		}),
	)

	// Connect to the server
	log.Println("Connecting to CLOB Market WebSocket...")
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Subscribe to market data for specific assets
	// Replace with your actual asset IDs
	assetIDs := []string{
		// Example asset IDs - replace with real ones
		// You can find asset IDs at https://clob.polymarket.com/
		"17123485321386672176776734800460321083207167824796999887262322064541574251437",
	}

	if len(assetIDs) > 0 {
		log.Println("Subscribing to market data...")

		// Create filter for CLOB market subscriptions
		filter := &polymarketrealtime.CLOBMarketFilter{
			TokenIDs: assetIDs,
		}

		// Subscribe to orderbook updates
		if err := typedSub.SubscribeToCLOBMarketAggOrderbook(filter, func(orderbook polymarketrealtime.AggOrderbook) error {
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
		}); err != nil {
			log.Fatalf("Failed to subscribe to orderbook: %v", err)
		}

		// Subscribe to price changes
		if err := typedSub.SubscribeToCLOBMarketPriceChanges(filter, func(priceChanges polymarketrealtime.PriceChanges) error {
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
		}); err != nil {
			log.Fatalf("Failed to subscribe to price changes: %v", err)
		}

		// Subscribe to last trade prices
		if err := typedSub.SubscribeToCLOBMarketLastTradePrice(filter, func(lastPrice polymarketrealtime.LastTradePrice) error {
			log.Printf("[Last Trade Price] Asset: %s, Market: %s, Side: %s, Price: %s, Size: %s",
				lastPrice.AssetID,
				lastPrice.Market,
				lastPrice.Side,
				lastPrice.Price.String(),
				lastPrice.Size.String(),
			)
			return nil
		}); err != nil {
			log.Fatalf("Failed to subscribe to last trade prices: %v", err)
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
