package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/shopspring/decimal"

	polymarketdataclient "github.com/ivanzzeth/polymarket-go-real-time-data-client"
)

func main() {
	log.Println("=== CLOB Market Price Changes Demo ===")
	log.Println("This example demonstrates subscribing to real-time price changes from CLOB markets")
	log.Println()

	// Create a message router for typed handling
	router := polymarketdataclient.NewRealtimeMessageRouter()

	// Register handler for price changes
	router.RegisterPriceChangesHandler(func(priceChanges polymarketdataclient.PriceChanges) error {
		log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		log.Printf("[Price Changes] Market: %s", priceChanges.Market)
		log.Printf("  Timestamp: %s", priceChanges.Timestamp)
		log.Printf("  Number of changes: %d", len(priceChanges.PriceChange))
		log.Println()

		// Display each price change in detail
		for i, change := range priceChanges.PriceChange {
			log.Printf("  Change #%d:", i+1)
			log.Printf("    Asset ID: %s", change.AssetID)
			log.Printf("    Side:     %s", change.Side)
			log.Printf("    Price:    %s", change.Price.String())
			log.Printf("    Size:     %s", change.Size.String())
			log.Printf("    Best Bid: %s", change.BestBid.String())
			log.Printf("    Best Ask: %s", change.BestAsk.String())
			log.Printf("    Hash:     %s", change.Hash)

			// Calculate spread
			if !change.BestAsk.IsZero() && !change.BestBid.IsZero() {
				spread := change.BestAsk.Sub(change.BestBid)
				spreadPercent := spread.Div(change.BestAsk).Mul(decimal.NewFromInt(100))
				log.Printf("    Spread:   %s (%.4f%%)", spread.String(), spreadPercent.InexactFloat64())
			}
			log.Println()
		}

		return nil
	})

	// Create client
	client := polymarketdataclient.New(
		polymarketdataclient.WithLogger(polymarketdataclient.NewLogger()),
		polymarketdataclient.WithAutoReconnect(true),
		polymarketdataclient.WithOnConnect(func() {
			log.Println("✅ Connected to Polymarket WebSocket")
		}),
		polymarketdataclient.WithOnDisconnect(func(err error) {
			log.Printf("❌ Disconnected: %v", err)
		}),
		polymarketdataclient.WithOnReconnect(func() {
			log.Println("🔄 Reconnected successfully")
		}),
	)

	// Connect to the server
	log.Println("Connecting to Polymarket WebSocket...")
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Subscribe to price changes for specific token IDs
	// You can find token IDs on https://clob.polymarket.com/
	// Example: Presidential Election 2024 market
	tokenIDs := []string{
		"87263549609442107846940966542305431913459334507901399837670195489634390247458",
	}

	log.Println("\nSubscribing to price changes...")
	log.Printf("Token IDs: %v", tokenIDs)

	filter := polymarketdataclient.NewCLOBMarketFilter(tokenIDs...)
	if err := client.SubscribeToCLOBMarketPriceChanges(filter, nil); err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	log.Println("✅ Successfully subscribed to price changes")
	log.Println()
	log.Println("=== Listening for Price Changes ===")
	log.Println("Waiting for updates... (this may take a moment)")
	log.Println("Press Ctrl+C to exit")
	log.Println()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("\n\nShutting down...")
	if err := client.Disconnect(); err != nil {
		log.Printf("Error during disconnect: %v", err)
	}
	log.Println("✅ Disconnected successfully")
}
