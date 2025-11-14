package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	polymarketrealtime "github.com/ivanzzeth/polymarket-go-real-time-data-client"
)

func main() {
	log.Println("=== WebSocket Reconnection Demo ===")
	log.Println("This example demonstrates automatic reconnection on connection failure")
	log.Println()

	// Track connection state
	var (
		connectCount    int
		disconnectCount int
		reconnectCount  int
	)

	// Create client with reconnection enabled
	client := polymarketrealtime.New(
		// Enable detailed logging to see reconnection attempts
		polymarketrealtime.WithLogger(polymarketrealtime.NewLogger(polymarketrealtime.LogLevelDebug)),

		// Configure reconnection behavior
		polymarketrealtime.WithAutoReconnect(true),
		polymarketrealtime.WithMaxReconnectAttempts(0), // Try up to 10 times
		polymarketrealtime.WithReconnectBackoff(
			1*time.Second,  // Initial backoff
			30*time.Second, // Max backoff
		),

		// Connection callback
		polymarketrealtime.WithOnConnect(func() {
			connectCount++
			if connectCount == 1 {
				log.Println("✅ Initial connection established")
			} else {
				log.Printf("✅ Reconnected successfully (total connects: %d)", connectCount)
			}
		}),

		// Disconnection callback
		polymarketrealtime.WithOnDisconnect(func(err error) {
			disconnectCount++
			log.Printf("❌ Connection lost (count: %d): %v", disconnectCount, err)
			log.Println("   Auto-reconnection will start...")
		}),

		// Reconnection success callback
		polymarketrealtime.WithOnReconnect(func() {
			reconnectCount++
			log.Printf("🔄 Reconnection successful (count: %d)", reconnectCount)
			log.Println("   Subscriptions have been restored")
		}),
	)

	// Connect to the server
	log.Println("Connecting to Polymarket WebSocket...")
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Subscribe to Bitcoin price
	log.Println("Subscribing to BTC price updates...")
	if err := client.SubscribeToCryptoPrices(polymarketrealtime.NewBTCPriceFilter(), func(price polymarketrealtime.CryptoPrice) error {
		log.Printf("[Price Update] %s = $%s", price.Symbol, price.Value.String())
		return nil
	}); err != nil {
		log.Printf("Failed to subscribe: %v", err)
	} else {
		log.Println("✅ Successfully subscribed to BTC prices")
	}

	log.Println()
	log.Println("=== Reconnection Test Instructions ===")
	log.Println("1. The client is now running and receiving price updates")
	log.Println("2. To test reconnection, you can:")
	log.Println("   - Temporarily disable your network connection")
	log.Println("   - Use a firewall to block the connection")
	log.Println("   - Wait for network issues")
	log.Println("3. Watch the logs to see automatic reconnection in action")
	log.Println("4. Press Ctrl+C to exit")
	log.Println()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("\n\n=== Connection Statistics ===")
	log.Printf("Total Connections: %d", connectCount)
	log.Printf("Total Disconnections: %d", disconnectCount)
	log.Printf("Total Reconnections: %d", reconnectCount)
	log.Println()

	log.Println("Shutting down...")
	if err := client.Disconnect(); err != nil {
		log.Printf("Error during disconnect: %v", err)
	}
	log.Println("✅ Disconnected successfully")
}
