package main

import (
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	polymarketrealtime "github.com/ivanzzeth/polymarket-go-real-time-data-client"
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
		// polymarketrealtime.WithLogger(polymarketrealtime.NewLogger()),
		polymarketrealtime.WithOnConnect(func() {
			log.Println("✓ Connected to Polymarket WebSocket!")
		}),
	}
	if proxyURL != nil {
		opts = append(opts, polymarketrealtime.WithProxyURL(proxyURL))
	}

	// Create the WebSocket client
	client := polymarketrealtime.New(opts...)

	// Connect to the server
	log.Printf("Connecting to Polymarket to monitor %s...\n", symbolToMonitor)
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect()

	// Subscribe to a single crypto price with callback
	filter := polymarketrealtime.NewCryptoPriceFilter(symbolToMonitor)
	if err := client.SubscribeToCryptoPrices(filter, func(price polymarketrealtime.CryptoPrice) error {
		log.Printf("[Crypto] %s = $%s (time: %s)",
			price.Symbol,
			price.Value.String(),
			price.Time.Format("15:04:05.000"))
		return nil
	}); err != nil {
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
