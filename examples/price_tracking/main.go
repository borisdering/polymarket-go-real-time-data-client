package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	polymarketdataclient "github.com/ivanzzeth/polymarket-go-real-time-data-client"
)

func main() {
	// Create a typed message router
	router := polymarketdataclient.NewTypedMessageRouter()

	// Track crypto prices
	router.RegisterCryptoPriceHandler(func(price polymarketdataclient.CryptoPrice) error {
		log.Printf("[Crypto] %s = $%s (time: %s)",
			price.Symbol,
			price.Value.String(),
			price.Time.Format("15:04:05.000"))
		return nil
	})

	// Track equity prices
	router.RegisterEquityPriceHandler(func(price polymarketdataclient.EquityPrice) error {
		log.Printf("[Equity] %s = $%s (time: %s)",
			price.Symbol,
			price.Value.String(),
			price.Time.Format("15:04:05.000"))
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
	defer client.Disconnect()

	// Create typed subscription handler
	typedSub := polymarketdataclient.NewTypedSubscriptionHandler(client)

	log.Println("Subscribing to price feeds...")

	// Subscribe to crypto prices
	cryptoSymbols := []string{"BTCUSDT", "ETHUSDT", "SOLUSDT"}
	for _, symbol := range cryptoSymbols {
		filter := `{"symbol":"` + symbol + `"}`
		if err := typedSub.SubscribeToCryptoPrices(nil, filter); err != nil {
			log.Printf("Failed to subscribe to %s: %v", symbol, err)
		} else {
			log.Printf("✓ Subscribed to %s", symbol)
		}
	}

	// Subscribe to equity prices
	equitySymbols := []string{"AAPL", "TSLA", "NVDA", "MSFT"}
	for _, symbol := range equitySymbols {
		filter := `{"symbol":"` + symbol + `"}`
		if err := typedSub.SubscribeToEquityPrices(nil, filter); err != nil {
			log.Printf("Failed to subscribe to %s: %v", symbol, err)
		} else {
			log.Printf("✓ Subscribed to %s", symbol)
		}
	}

	log.Println("=== Price Tracking Started ===")
	log.Println("Monitoring crypto and equity prices...")
	log.Println("Press Ctrl+C to exit")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("\n\nShutting down...")
	log.Println("✓ Disconnected successfully")
}
