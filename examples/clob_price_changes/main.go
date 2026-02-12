package main

import (
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/shopspring/decimal"

	polymarketrealtime "github.com/ivanzzeth/polymarket-go-real-time-data-client"
)

func main() {
	log.Println("=== CLOB Market Price Changes Demo ===")
	log.Println("This example demonstrates subscribing to real-time price changes from CLOB markets")
	log.Println()

	// Set proxy from environment if available
	var proxyURL *url.URL
	if proxyStr := os.Getenv("https_proxy"); proxyStr != "" {
		if parsed, err := url.Parse(proxyStr); err == nil {
			proxyURL = parsed
		}
	} else if proxyStr := os.Getenv("HTTPS_PROXY"); proxyStr != "" {
		if parsed, err := url.Parse(proxyStr); err == nil {
			proxyURL = parsed
		}
	} else if proxyStr := os.Getenv("http_proxy"); proxyStr != "" {
		if parsed, err := url.Parse(proxyStr); err == nil {
			proxyURL = parsed
		}
	} else if proxyStr := os.Getenv("HTTP_PROXY"); proxyStr != "" {
		if parsed, err := url.Parse(proxyStr); err == nil {
			proxyURL = parsed
		}
	}

	opts := []polymarketrealtime.ClientOption{
		polymarketrealtime.WithLogger(polymarketrealtime.NewLogger(polymarketrealtime.LogLevelInfo)),
		polymarketrealtime.WithAutoReconnect(true),
		polymarketrealtime.WithOnConnect(func() {
			log.Println("✅ Connected to Polymarket WebSocket")
		}),
		polymarketrealtime.WithOnDisconnect(func(err error) {
			log.Printf("❌ Disconnected: %v", err)
		}),
		polymarketrealtime.WithOnReconnect(func() {
			log.Println("🔄 Reconnected successfully")
		}),
	}
	if proxyURL != nil {
		opts = append(opts, polymarketrealtime.WithProxyURL(proxyURL))
	}

	// Create client
	client := polymarketrealtime.New(opts...)

	// Connect to the server
	log.Println("Connecting to Polymarket WebSocket...")
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Subscribe to price changes for specific token IDs
	// You can find token IDs on https://clob.polymarket.com/
	// Example: Presidential Election 2024 market
	tokenIDs := []string{
		"20244656410496119633637176580888572822795336808693757973566456523317429619143",
	}

	log.Println("\nSubscribing to price changes...")
	log.Printf("Token IDs: %v", tokenIDs)

	filter := polymarketrealtime.NewCLOBMarketFilter(tokenIDs...)
	if err := client.SubscribeToCLOBMarketPriceChanges(filter, func(priceChanges polymarketrealtime.PriceChanges) error {
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
	}); err != nil {
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
