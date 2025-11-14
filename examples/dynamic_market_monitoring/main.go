package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	polymarketrealtime "github.com/ivanzzeth/polymarket-go-real-time-data-client"
)

// MarketMonitor manages dynamic subscriptions for markets
type MarketMonitor struct {
	client              *polymarketrealtime.Client
	activeMarkets       map[string]bool // market ID -> active status
	marketSubscriptions map[string][]polymarketrealtime.Subscription
	mu                  sync.RWMutex
}

// NewMarketMonitor creates a new market monitor
func NewMarketMonitor(client *polymarketrealtime.Client) *MarketMonitor {
	return &MarketMonitor{
		client:              client,
		activeMarkets:       make(map[string]bool),
		marketSubscriptions: make(map[string][]polymarketrealtime.Subscription),
	}
}

// OnMarketCreated handles market creation events
func (m *MarketMonitor) OnMarketCreated(market polymarketrealtime.ClobMarket) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	marketID := market.Market
	log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("🆕 [Market Created] Market ID: %s", marketID)
	log.Printf("  Asset IDs: %v", market.AssetIDs)
	log.Printf("  Min Order Size: %s", market.MinOrderSize.String())
	log.Printf("  Tick Size: %s", market.TickSize.String())
	log.Printf("  Neg Risk: %v", market.NegRisk)
	log.Println()

	// Mark market as active
	m.activeMarkets[marketID] = true

	// Subscribe to all market data for this market
	log.Printf("📡 Subscribing to market data for market: %s", marketID)

	// Create filter for CLOB market subscriptions
	filter := &polymarketrealtime.CLOBMarketFilter{
		TokenIDs: market.AssetIDs,
	}

	// Subscribe to price changes
	if err := m.client.SubscribeToCLOBMarketPriceChanges(filter, nil); err != nil {
		log.Printf("❌ Failed to subscribe to price changes: %v", err)
		return err
	}
	log.Printf("  ✅ Subscribed to price changes")

	// Subscribe to orderbook updates
	if err := m.client.SubscribeToCLOBMarketAggOrderbook(filter, nil); err != nil {
		log.Printf("❌ Failed to subscribe to orderbook: %v", err)
		return err
	}
	log.Printf("  ✅ Subscribed to orderbook updates")

	// Subscribe to last trade prices
	if err := m.client.SubscribeToCLOBMarketLastTradePrice(filter, nil); err != nil {
		log.Printf("❌ Failed to subscribe to last trade prices: %v", err)
		return err
	}
	log.Printf("  ✅ Subscribed to last trade prices")

	// Store subscriptions for later unsubscription
	subscriptions := []polymarketrealtime.Subscription{}
	for _, assetID := range market.AssetIDs {
		subscriptions = append(subscriptions,
			polymarketrealtime.Subscription{
				Topic:   polymarketrealtime.TopicCLOBMarket,
				Type:    polymarketrealtime.MessageTypeCLOBPriceChanges,
				Filters: assetID,
			},
			polymarketrealtime.Subscription{
				Topic:   polymarketrealtime.TopicCLOBMarket,
				Type:    polymarketrealtime.MessageTypeCLOBAggOrderbook,
				Filters: assetID,
			},
			polymarketrealtime.Subscription{
				Topic:   polymarketrealtime.TopicCLOBMarket,
				Type:    polymarketrealtime.MessageTypeCLOBLastTradePrice,
				Filters: assetID,
			},
		)
	}
	m.marketSubscriptions[marketID] = subscriptions

	log.Printf("✅ Successfully subscribed to all data for market: %s", marketID)
	log.Println()

	return nil
}

// OnMarketResolved handles market resolution events
func (m *MarketMonitor) OnMarketResolved(market polymarketrealtime.ClobMarket) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	marketID := market.Market
	log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("🏁 [Market Resolved] Market ID: %s", marketID)
	log.Println()

	// Check if we were monitoring this market
	if !m.activeMarkets[marketID] {
		log.Printf("  ℹ️  Market was not being actively monitored")
		log.Println()
		return nil
	}

	// Mark market as inactive
	m.activeMarkets[marketID] = false

	// Unsubscribe from all market data
	subscriptions, ok := m.marketSubscriptions[marketID]
	if !ok || len(subscriptions) == 0 {
		log.Printf("  ℹ️  No subscriptions found for market: %s", marketID)
		log.Println()
		return nil
	}

	log.Printf("📡 Unsubscribing from market data for market: %s", marketID)

	if err := m.client.Unsubscribe(subscriptions); err != nil {
		log.Printf("❌ Failed to unsubscribe: %v", err)
		return err
	}

	// Clean up
	delete(m.marketSubscriptions, marketID)

	log.Printf("✅ Successfully unsubscribed from all data for market: %s", marketID)
	log.Println()

	return nil
}

// IsMarketActive checks if a market is currently active
func (m *MarketMonitor) IsMarketActive(marketID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activeMarkets[marketID]
}

// GetActiveMarketCount returns the number of active markets being monitored
func (m *MarketMonitor) GetActiveMarketCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, active := range m.activeMarkets {
		if active {
			count++
		}
	}
	return count
}

// ListActiveMarkets returns a list of active market IDs
func (m *MarketMonitor) ListActiveMarkets() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	markets := []string{}
	for marketID, active := range m.activeMarkets {
		if active {
			markets = append(markets, marketID)
		}
	}
	return markets
}

func main() {
	log.Println("=== Dynamic Market Monitoring Demo ===")
	log.Println("This example demonstrates monitoring market creation and resolution")
	log.Println("and dynamically managing subscriptions based on market lifecycle")
	log.Println()

	// Track message counts
	var priceChangeCount, orderbookCount, lastTradeCount int
	var mu sync.Mutex

	var monitor *MarketMonitor

	// Create message router for market data
	router := polymarketrealtime.NewRealtimeMessageRouter()
	router.RegisterPriceChangesHandler(func(priceChanges polymarketrealtime.PriceChanges) error {
		mu.Lock()
		priceChangeCount++
		count := priceChangeCount
		mu.Unlock()

		log.Printf("📊 [Price Changes #%d] Market: %s, Changes: %d", count, priceChanges.Market, len(priceChanges.PriceChange))
		for i, change := range priceChanges.PriceChange {
			log.Printf("    Change #%d: Asset=%s, Side=%s, Price=%s, Size=%s",
				i+1, change.AssetID, change.Side, change.Price.String(), change.Size.String())
		}
		return nil
	})

	router.RegisterAggOrderbookHandler(func(orderbook polymarketrealtime.AggOrderbook) error {
		mu.Lock()
		orderbookCount++
		count := orderbookCount
		mu.Unlock()

		log.Printf("📖 [Orderbook #%d] Market: %s, Asset: %s, Bids: %d, Asks: %d",
			count, orderbook.Market, orderbook.AssetID, len(orderbook.Bids), len(orderbook.Asks))
		return nil
	})

	router.RegisterLastTradePriceHandler(func(lastTrade polymarketrealtime.LastTradePrice) error {
		mu.Lock()
		lastTradeCount++
		count := lastTradeCount
		mu.Unlock()

		log.Printf("💰 [Last Trade #%d] Market: %s, Asset: %s, Price: %s, Size: %s",
			count, lastTrade.Market, lastTrade.AssetID, lastTrade.Price.String(), lastTrade.Size.String())
		return nil
	})

	// Create client
	client := polymarketrealtime.New(
		// polymarketrealtime.WithLogger(polymarketrealtime.NewLogger()),
		polymarketrealtime.WithAutoReconnect(true),
		polymarketrealtime.WithOnConnect(func() {
			log.Println("✅ Connected to CLOB Market WebSocket")
		}),
		polymarketrealtime.WithOnDisconnect(func(err error) {
			log.Printf("❌ Disconnected: %v", err)
		}),
		polymarketrealtime.WithOnReconnect(func() {
			log.Println("🔄 Reconnected successfully")
		}),
	)

	// Connect to the server
	log.Println("Connecting to CLOB Market WebSocket...")
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Create market monitor
	monitor = NewMarketMonitor(client)

	// Create separate client for market lifecycle events
	lifecycleClient := polymarketrealtime.New(
		// polymarketrealtime.WithHost("wss://ws-subscriptions-clob.polymarket.com/ws/market"),
		polymarketrealtime.WithLogger(polymarketrealtime.NewLogger(polymarketrealtime.LogLevelDebug)),
		polymarketrealtime.WithAutoReconnect(true),
	)

	// Connect the lifecycle client
	if err := lifecycleClient.Connect(); err != nil {
		log.Fatalf("Failed to connect lifecycle client: %v", err)
	}

	// Subscribe to market lifecycle events
	log.Println("📡 Subscribing to market lifecycle events...")

	if err := lifecycleClient.SubscribeToCLOBMarketCreated(monitor.OnMarketCreated); err != nil {
		log.Fatalf("Failed to subscribe to market created events: %v", err)
	}
	log.Println("✅ Subscribed to market_created events")

	if err := lifecycleClient.SubscribeToCLOBMarketResolved(monitor.OnMarketResolved); err != nil {
		log.Fatalf("Failed to subscribe to market resolved events: %v", err)
	}
	log.Println("✅ Subscribed to market_resolved events")

	log.Println()
	log.Println("=== Monitoring Market Events ===")
	log.Println("Waiting for market creation and resolution events...")
	log.Println("Press Ctrl+C to exit")
	log.Println()

	// Print statistics periodically
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			mu.Lock()
			pc := priceChangeCount
			oc := orderbookCount
			lt := lastTradeCount
			mu.Unlock()

			activeCount := monitor.GetActiveMarketCount()
			activeMarkets := monitor.ListActiveMarkets()

			log.Println()
			log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			log.Println("📊 Statistics Summary")
			log.Printf("  Active Markets: %d", activeCount)
			if len(activeMarkets) > 0 {
				log.Printf("  Market IDs: %v", activeMarkets)
			}
			log.Printf("  Price Changes Received: %d", pc)
			log.Printf("  Orderbook Updates Received: %d", oc)
			log.Printf("  Last Trade Updates Received: %d", lt)
			log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			log.Println()
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("Shutting down...")

	// Final statistics
	mu.Lock()
	pc := priceChangeCount
	oc := orderbookCount
	lt := lastTradeCount
	mu.Unlock()

	activeCount := monitor.GetActiveMarketCount()
	activeMarkets := monitor.ListActiveMarkets()

	log.Println()
	log.Println("📊 Final Statistics:")
	log.Printf("  Active Markets: %d", activeCount)
	if len(activeMarkets) > 0 {
		log.Printf("  Market IDs: %v", activeMarkets)
	}
	log.Printf("  Total Price Changes: %d", pc)
	log.Printf("  Total Orderbook Updates: %d", oc)
	log.Printf("  Total Last Trade Updates: %d", lt)
	log.Println()

	if err := client.Disconnect(); err != nil {
		log.Printf("Error during disconnect: %v", err)
	}
	if err := lifecycleClient.Disconnect(); err != nil {
		log.Printf("Error during lifecycle client disconnect: %v", err)
	}
	log.Println("✅ Disconnected successfully")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}
