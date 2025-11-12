package main

import (
	"log"
	"os"
	"time"

	polymarketdataclient "github.com/ivanzzeth/polymarket-go-real-time-data-client"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Create a typed message router
	router := polymarketdataclient.NewRealtimeMessageRouter()

	// Register handler for CLOB user trades
	router.RegisterCLOBTradeHandler(func(trade polymarketdataclient.CLOBTrade) error {
		log.Printf("=== CLOB Trade ===")
		log.Printf("ID: %s", trade.ID)
		log.Printf("Market: %s", trade.Market)
		log.Printf("Asset ID: %s", trade.AssetID)
		log.Printf("Side: %s", trade.Side)
		log.Printf("Price: %s", trade.Price.String())
		log.Printf("Size: %s", trade.Size.String())
		log.Printf("Status: %s", trade.Status)
		log.Printf("Match Time: %s", trade.MatchTime)
		log.Printf("Transaction Hash: %s", trade.TransactionHash)
		log.Printf("Maker Orders Count: %d", len(trade.MakerOrders))
		log.Println()
		return nil
	})

	// Register handler for CLOB user orders
	router.RegisterCLOBOrderHandler(func(order polymarketdataclient.CLOBOrder) error {
		log.Printf("=== CLOB Order ===")
		log.Printf("ID: %s", order.ID)
		log.Printf("Market: %s", order.Market)
		log.Printf("Type: %s", order.Type)
		log.Printf("Side: %s", order.Side)
		log.Printf("Price: %s", order.Price.String())
		log.Printf("Original Size: %s", order.OriginalSize.String())
		log.Printf("Size Matched: %s", order.SizeMatched.String())
		log.Printf("Status: %s", order.Status)
		log.Printf("Outcome: %s", order.Outcome)
		log.Println()
		return nil
	})

	// Create the WebSocket client
	client := polymarketdataclient.New(
		polymarketdataclient.WithLogger(polymarketdataclient.NewSilentLogger()),
		polymarketdataclient.WithOnConnect(func() {
			log.Println("✓ Connected to Polymarket WebSocket!")
		}),
		polymarketdataclient.WithOnNewMessage(func(data []byte) {
			// Route the message to the appropriate typed handler
			if err := router.RouteMessage(data); err != nil {
				log.Printf("Error routing message: %v", err)
			}
		}),
	)

	// Connect to the server
	log.Println("Connecting to Polymarket...")
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Get credentials from environment
	apiKey := os.Getenv("API_KEY")
	apiSecret := os.Getenv("API_SECRET")
	apiPassphrase := os.Getenv("API_PASSPHRASE")

	if apiKey == "" || apiSecret == "" || apiPassphrase == "" {
		log.Fatal("API_KEY, API_SECRET, and API_PASSPHRASE must be set in environment or .env file")
	}

	// Create typed subscription handler
	typedSub := polymarketdataclient.NewRealtimeTypedSubscriptionHandler(client)

	// Create CLOB authentication
	clobAuth := polymarketdataclient.ClobAuth{
		Key:        apiKey,
		Secret:     apiSecret,
		Passphrase: apiPassphrase,
	}

	// Subscribe to all CLOB user events (orders and trades)
	log.Println("Subscribing to CLOB user events...")
	if err := typedSub.SubscribeToCLOBUserAll(clobAuth); err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	log.Println("✓ Successfully subscribed!")
	log.Println("Listening for orders and trades...")
	log.Println("Press Ctrl+C to exit")
	log.Println()

	// Keep the connection alive
	time.Sleep(60 * time.Second)

	// Clean up
	log.Println("\nDisconnecting...")
	if err := client.Disconnect(); err != nil {
		log.Printf("Error disconnecting: %v", err)
	}
	log.Println("✓ Disconnected successfully")
}
