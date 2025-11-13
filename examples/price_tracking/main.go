package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	polymarketdataclient "github.com/ivanzzeth/polymarket-go-real-time-data-client"
)

func main() {
	log.Println("=== Price Tracking Demo ===")
	log.Println()
	log.Println("IMPORTANT: Due to Polymarket API limitations, crypto_prices and equity_prices")
	log.Println("topics only support ONE symbol per WebSocket connection.")
	log.Println()
	log.Println("This example demonstrates the CORRECT way to monitor a SINGLE symbol.")
	log.Println("To monitor multiple symbols, you need separate client connections for each.")
	log.Println()

	// Symbol to monitor (change this to monitor different symbols)
	symbolToMonitor := "solusdt" // Options: btcusdt, ethusdt, solusdt, etc.

	// Create the WebSocket client
	client := polymarketdataclient.New(
		// polymarketdataclient.WithLogger(polymarketdataclient.NewLogger()),
		polymarketdataclient.WithOnConnect(func() {
			log.Println("✓ Connected to Polymarket WebSocket!")
		}),
	)

	// Connect to the server
	log.Printf("Connecting to Polymarket to monitor %s...\n", symbolToMonitor)
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect()

	// Subscribe to a single crypto price with callback
	filter := polymarketdataclient.NewCryptoPriceFilter(symbolToMonitor)
	if err := client.SubscribeToCryptoPrices(func(price polymarketdataclient.CryptoPrice) error {
		log.Printf("[Crypto] %s = $%s (time: %s)",
			price.Symbol,
			price.Value.String(),
			price.Time.Format("15:04:05.000"))
		return nil
	}, filter); err != nil {
		log.Fatalf("Failed to subscribe to %s: %v", symbolToMonitor, err)
	}
	log.Printf("✓ Subscribed to %s\n", symbolToMonitor)

	log.Println()
	log.Println("=== Price Tracking Started ===")
	log.Printf("Monitoring %s price updates...\n", symbolToMonitor)
	log.Println("Press Ctrl+C to exit")
	log.Println()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("\n\nShutting down...")
	log.Println("✓ Disconnected successfully")
}
