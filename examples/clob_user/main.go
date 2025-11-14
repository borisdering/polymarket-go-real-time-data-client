package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	polymarketrealtime "github.com/ivanzzeth/polymarket-go-real-time-data-client"
)

func main() {
	log.Println("=== CLOB User WebSocket Client Demo ===")
	log.Println("This example demonstrates subscribing to user-specific CLOB events")
	log.Println()

	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// IMPORTANT: Set your CLOB API credentials
	// You can get these from https://clob.polymarket.com/
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("API_KEY not set in environment")
	}

	secret := os.Getenv("API_SECRET")
	if secret == "" {
		log.Fatal("API_SECRET not set in environment")
	}

	passphrase := os.Getenv("API_PASSPHRASE")
	if passphrase == "" {
		log.Fatal("API_PASSPHRASE not set in environment")
	}

	auth := polymarketrealtime.ClobAuth{
		Key:        apiKey,
		Secret:     secret,
		Passphrase: passphrase,
	}

	// Create client
	client := polymarketrealtime.New(
		polymarketrealtime.WithLogger(polymarketrealtime.NewLogger(polymarketrealtime.LogLevelDebug)),
		polymarketrealtime.WithAutoReconnect(true),
		polymarketrealtime.WithOnConnect(func() {
			log.Println("✅ Connected to CLOB User endpoint")
		}),
		polymarketrealtime.WithOnDisconnect(func(err error) {
			log.Printf("❌ Disconnected from CLOB User endpoint: %v", err)
		}),
		polymarketrealtime.WithOnReconnect(func() {
			log.Println("🔄 Reconnected to CLOB User endpoint")
		}),
	)

	// Connect to the server
	log.Println("Connecting to CLOB User WebSocket...")
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	log.Println("Subscribing to user orders and trades...")

	// Subscribe to user orders
	if err := client.SubscribeToCLOBUserOrders(&auth, func(order polymarketrealtime.CLOBOrder) error {
		log.Printf("[Order Update] Type: %s, Status: %s, Market: %s, Side: %s, Price: %s, Size: %s/%s",
			order.Type,
			order.Status,
			order.Market,
			order.Side,
			order.Price.String(),
			order.SizeMatched.String(),
			order.OriginalSize.String(),
		)
		return nil
	}); err != nil {
		log.Fatalf("Failed to subscribe to orders: %v", err)
	}

	// Subscribe to user trades
	if err := client.SubscribeToCLOBUserTrades(&auth, func(trade polymarketrealtime.CLOBTrade) error {
		log.Printf("[Trade Executed] Market: %s, Side: %s, Price: %s, Size: %s, Status: %s, TxHash: %s",
			trade.Market,
			trade.Side,
			trade.Price.String(),
			trade.Size.String(),
			trade.Status,
			trade.TransactionHash,
		)
		return nil
	}); err != nil {
		log.Fatalf("Failed to subscribe to trades: %v", err)
	}

	log.Println("✅ Successfully subscribed to user events")

	log.Println()
	log.Println("=== Listening for User Events ===")
	log.Println("- Order placements, cancellations, and fills")
	log.Println("- Trade executions and confirmations")
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
