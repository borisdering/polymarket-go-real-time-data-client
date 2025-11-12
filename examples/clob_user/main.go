package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	polymarketdataclient "github.com/ivanzzeth/polymarket-go-real-time-data-client"
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

	auth := &polymarketdataclient.ClobAuth{
		Key:        apiKey,
		Secret:     secret,
		Passphrase: passphrase,
	}

	// Create a message router for typed handling
	router := polymarketdataclient.NewClobUserMessageRouter()

	// Register handlers for different event types
	router.RegisterOrderHandler(func(order polymarketdataclient.CLOBOrder) error {
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
	})

	router.RegisterTradeHandler(func(trade polymarketdataclient.CLOBTrade) error {
		log.Printf("[Trade Executed] Market: %s, Side: %s, Price: %s, Size: %s, Status: %s, TxHash: %s",
			trade.Market,
			trade.Side,
			trade.Price.String(),
			trade.Size.String(),
			trade.Status,
			trade.TransactionHash,
		)
		return nil
	})

	// Create CLOB User client
	client := polymarketdataclient.NewClobUserClient(
		polymarketdataclient.WithLogger(polymarketdataclient.NewLogger()),
		polymarketdataclient.WithAutoReconnect(true),
		polymarketdataclient.WithOnConnect(func() {
			log.Println("✅ Connected to CLOB User endpoint")
		}),
		polymarketdataclient.WithOnDisconnect(func(err error) {
			log.Printf("❌ Disconnected from CLOB User endpoint: %v", err)
		}),
		polymarketdataclient.WithOnReconnect(func() {
			log.Println("🔄 Reconnected to CLOB User endpoint")
		}),
		polymarketdataclient.WithOnNewMessage(func(data []byte) {
			// Route message to appropriate handler
			if err := router.RouteMessage(data); err != nil {
				log.Printf("Error routing message: %v", err)
			}
		}),
	)

	// Connect to the server
	log.Println("Connecting to CLOB User WebSocket...")
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Create typed subscription handler
	typedSub := polymarketdataclient.NewClobUserTypedSubscriptionHandler(client)

	// Subscribe to user events for specific markets
	// Replace with your actual market IDs
	markets := []string{
		// Example market IDs - replace with real ones
		// "0x1234567890abcdef...",
		"0xf95edd63d6ece5e59b7e8d26bec3881fe96df161b40aa85991f4182c8fac559d",
	}

	if len(markets) > 0 {
		log.Println("Subscribing to user orders and trades...")
		if err := typedSub.SubscribeToAllUserEvents(markets, auth); err != nil {
			log.Fatalf("Failed to subscribe: %v", err)
		}
		log.Println("✅ Successfully subscribed to user events")
	} else {
		log.Println("⚠️  No markets specified. Add market IDs to the 'markets' slice to receive updates.")
		log.Println("   You can find market IDs at https://clob.polymarket.com/")
	}

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
