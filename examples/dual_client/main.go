package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	polymarketdataclient "github.com/ivanzzeth/polymarket-go-real-time-data-client"
)

func main() {
	log.Println("=== Multi-Topic Demo: Real-Time Data + CLOB Market ===")
	log.Println("This example demonstrates subscribing to multiple topics simultaneously")
	log.Println()

	// ========== Create Unified Client ==========
	log.Println("Setting up unified client...")

	typedSub, client := polymarketdataclient.NewRealtimeTypedSubscriptionHandlerWithOptions(
		polymarketdataclient.WithLogger(polymarketdataclient.NewSilentLogger()),
		polymarketdataclient.WithAutoReconnect(true),
		polymarketdataclient.WithOnConnect(func() {
			log.Println("✅ Connected to WebSocket")
		}),
	)

	// Connect to the server
	log.Println("Connecting to WebSocket...")
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// ========== Subscribe to Data ==========
	log.Println("\nSubscribing to data streams...")

	// Subscribe to Bitcoin price updates
	if err := typedSub.SubscribeToCryptoPrices(func(price polymarketdataclient.CryptoPrice) error {
		log.Printf("[Crypto] Price: %s = $%s", price.Symbol, price.Value.String())
		return nil
	}, polymarketdataclient.NewBTCPriceFilter()); err != nil {
		log.Printf("Warning: Failed to subscribe to crypto prices: %v", err)
	} else {
		log.Println("✅ Subscribed to BTC price updates")
	}

	// Subscribe to CLOB market data
	// Replace with your actual asset IDs
	assetIDs := []string{
		// Add your asset IDs here
		// "0x1234567890abcdef...",
	}

	if len(assetIDs) > 0 {
		filter := &polymarketdataclient.CLOBMarketFilter{
			TokenIDs: assetIDs,
		}

		// Subscribe to price changes
		if err := typedSub.SubscribeToCLOBMarketPriceChanges(filter, func(priceChanges polymarketdataclient.PriceChanges) error {
			log.Printf("[CLOB] Price Changes: Market %s, %d changes", priceChanges.Market, len(priceChanges.PriceChange))
			return nil
		}); err != nil {
			log.Printf("Warning: Failed to subscribe to price changes: %v", err)
		}

		// Subscribe to orderbook
		if err := typedSub.SubscribeToCLOBMarketAggOrderbook(filter, func(orderbook polymarketdataclient.AggOrderbook) error {
			log.Printf("[CLOB] Orderbook: Asset %s, Bids: %d, Asks: %d",
				orderbook.AssetID, len(orderbook.Bids), len(orderbook.Asks))
			return nil
		}); err != nil {
			log.Printf("Warning: Failed to subscribe to orderbook: %v", err)
		} else {
			log.Println("✅ Subscribed to CLOB market data")
		}
	} else {
		log.Println("⚠️  No CLOB asset IDs specified. Add asset IDs to receive CLOB updates.")
	}

	// ========== Listen for Data ==========
	log.Println()
	log.Println("=== Multi-Topic Client Active ===")
	log.Println("Receiving data from multiple topics:")
	log.Println("  1. Crypto Prices: Bitcoin and other cryptocurrencies")
	log.Println("  2. CLOB Market: Orderbook updates, price changes")
	log.Println()
	log.Println("This demonstrates how you can combine multiple data sources")
	log.Println("in a single WebSocket connection!")
	log.Println()
	log.Println("Press Ctrl+C to exit")
	log.Println()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	// ========== Shutdown ==========
	log.Println("\n\nShutting down client...")

	if err := client.Disconnect(); err != nil {
		log.Printf("Error disconnecting client: %v", err)
	}

	log.Println("✅ Client disconnected successfully")
}
