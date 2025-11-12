package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	polymarketdataclient "github.com/ivanzzeth/polymarket-go-real-time-data-client"
)

func main() {
	log.Println("=== Dual Client Demo: Real-Time Data + CLOB Market ===")
	log.Println("This example demonstrates using both endpoints simultaneously")
	log.Println()

	// ========== Real-Time Data Client ==========
	log.Println("Setting up Real-Time Data client...")

	realtimeRouter := polymarketdataclient.NewTypedMessageRouter()

	realtimeRouter.RegisterCryptoPriceHandler(func(price polymarketdataclient.CryptoPrice) error {
		log.Printf("[Real-Time] Crypto Price: %s = $%s", price.Symbol, price.Value.String())
		return nil
	})

	realtimeClient := polymarketdataclient.New(
		polymarketdataclient.WithLogger(polymarketdataclient.NewSilentLogger()),
		polymarketdataclient.WithAutoReconnect(true),
		polymarketdataclient.WithOnConnect(func() {
			log.Println("✅ Connected to Real-Time Data endpoint")
		}),
		polymarketdataclient.WithOnNewMessage(func(data []byte) {
			realtimeRouter.RouteMessage(data)
		}),
	)

	// ========== CLOB Market Client ==========
	log.Println("Setting up CLOB Market client...")

	clobRouter := polymarketdataclient.NewClobMarketMessageRouter()

	clobRouter.RegisterPriceChangesHandler(func(priceChanges polymarketdataclient.PriceChanges) error {
		log.Printf("[CLOB Market] Price Changes: Market %s, %d changes", priceChanges.Market, len(priceChanges.PriceChange))
		return nil
	})

	clobRouter.RegisterOrderbookHandler(func(orderbook polymarketdataclient.AggOrderbook) error {
		log.Printf("[CLOB Market] Orderbook: Asset %s, Bids: %d, Asks: %d",
			orderbook.AssetID, len(orderbook.Bids), len(orderbook.Asks))
		return nil
	})

	clobClient := polymarketdataclient.NewClobMarketClient(
		polymarketdataclient.WithLogger(polymarketdataclient.NewSilentLogger()),
		polymarketdataclient.WithAutoReconnect(true),
		polymarketdataclient.WithOnConnect(func() {
			log.Println("✅ Connected to CLOB Market endpoint")
		}),
		polymarketdataclient.WithOnNewMessage(func(data []byte) {
			clobRouter.RouteMessage(data)
		}),
	)

	// ========== Connect Both Clients ==========
	log.Println("\nConnecting to both endpoints...")

	// Connect to Real-Time Data
	if err := realtimeClient.Connect(); err != nil {
		log.Fatalf("Failed to connect to Real-Time Data: %v", err)
	}

	// Connect to CLOB Market
	if err := clobClient.Connect(); err != nil {
		log.Fatalf("Failed to connect to CLOB Market: %v", err)
	}

	// ========== Subscribe to Data ==========
	log.Println("\nSubscing to data streams...")

	// Subscribe to Bitcoin price updates
	realtimeTypedSub := polymarketdataclient.NewTypedSubscriptionHandler(realtimeClient)
	if err := realtimeTypedSub.SubscribeToCryptoPrices(nil, `{"symbol":"btcusdt"}`); err != nil {
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
		clobTypedSub := polymarketdataclient.NewClobMarketTypedSubscriptionHandler(clobClient)
		if err := clobTypedSub.SubscribeToAllMarketData(assetIDs); err != nil {
			log.Printf("Warning: Failed to subscribe to CLOB market data: %v", err)
		} else {
			log.Println("✅ Subscribed to CLOB market data")
		}
	} else {
		log.Println("⚠️  No CLOB asset IDs specified. Add asset IDs to receive CLOB updates.")
	}

	// ========== Listen for Data ==========
	log.Println()
	log.Println("=== Dual Client Active ===")
	log.Println("Receiving data from both endpoints:")
	log.Println("  1. Real-Time Data: Crypto prices, activity, comments, etc.")
	log.Println("  2. CLOB Market: Orderbook updates, price changes, trades")
	log.Println()
	log.Println("This demonstrates how you can combine multiple data sources")
	log.Println("for comprehensive market intelligence!")
	log.Println()
	log.Println("Press Ctrl+C to exit")
	log.Println()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	// ========== Shutdown ==========
	log.Println("\n\nShutting down both clients...")

	if err := realtimeClient.Disconnect(); err != nil {
		log.Printf("Error disconnecting Real-Time client: %v", err)
	}

	if err := clobClient.Disconnect(); err != nil {
		log.Printf("Error disconnecting CLOB client: %v", err)
	}

	log.Println("✅ All clients disconnected successfully")
}
