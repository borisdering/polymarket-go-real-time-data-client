//go:build integration
// +build integration

package polymarketrealtime

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegrationRealConnection tests with a real connection to Polymarket WebSocket
// Run with: go test -tags=integration -v -run TestIntegrationRealConnection
func TestIntegrationRealConnection(t *testing.T) {
	var connectedCalled bool
	var disconnectedCalled bool
	var messagesReceived [][]byte
	var mu sync.Mutex

	// Create client with real connection
	client := New(
		WithLogger(NewLogger()),
		WithAutoReconnect(false), // Disable auto-reconnect for controlled testing
		WithOnConnect(func() {
			mu.Lock()
			connectedCalled = true
			mu.Unlock()
			t.Log("✅ Connected to Polymarket WebSocket")
		}),
		WithOnDisconnect(func(err error) {
			mu.Lock()
			disconnectedCalled = true
			mu.Unlock()
			t.Logf("❌ Disconnected: %v", err)
		}),
		withOnNewMessage(func(data []byte) {
			mu.Lock()
			messagesReceived = append(messagesReceived, data)
			mu.Unlock()
			t.Logf("📨 Received message: %s", string(data))
		}),
	)

	// Test connection
	t.Run("Connect", func(t *testing.T) {
		err := client.Connect()
		require.NoError(t, err)
		time.Sleep(500 * time.Millisecond)

		mu.Lock()
		assert.True(t, connectedCalled, "OnConnect callback should be called")
		mu.Unlock()
	})

	// Test subscription to crypto prices
	t.Run("SubscribeToCryptoPrices", func(t *testing.T) {
		filter := NewCryptoPriceFilter(CryptoSymbolBTCUSDT)
		subscriptions := []Subscription{
			{
				Topic:   TopicCryptoPrices,
				Type:    MessageTypeUpdate,
				Filters: filter.ToJSON(),
			},
		}

		err := client.Subscribe(subscriptions)
		require.NoError(t, err)
		t.Log("✅ Subscribed to BTC price updates")

		// Wait for messages
		time.Sleep(5 * time.Second)

		mu.Lock()
		messageCount := len(messagesReceived)
		mu.Unlock()

		assert.Greater(t, messageCount, 0, "Should receive at least one message")
		t.Logf("📊 Received %d messages", messageCount)
	})

	// Test unsubscribe
	t.Run("Unsubscribe", func(t *testing.T) {
		filter := NewCryptoPriceFilter(CryptoSymbolBTCUSDT)
		subscriptions := []Subscription{
			{
				Topic:   TopicCryptoPrices,
				Type:    MessageTypeUpdate,
				Filters: filter.ToJSON(),
			},
		}

		err := client.Unsubscribe(subscriptions)
		require.NoError(t, err)
		t.Log("✅ Unsubscribed from BTC price updates")

		// Clear received messages
		mu.Lock()
		previousCount := len(messagesReceived)
		messagesReceived = [][]byte{}
		mu.Unlock()

		// Wait and verify no new messages
		time.Sleep(3 * time.Second)

		mu.Lock()
		newCount := len(messagesReceived)
		mu.Unlock()

		t.Logf("📊 Messages after unsubscribe: %d (previous: %d)", newCount, previousCount)
		// Note: We might still receive some messages due to buffering
	})

	// Test disconnect
	t.Run("Disconnect", func(t *testing.T) {
		err := client.Disconnect()
		require.NoError(t, err)
		time.Sleep(500 * time.Millisecond)

		mu.Lock()
		assert.True(t, disconnectedCalled, "OnDisconnect callback should be called")
		mu.Unlock()
	})
}

// TestIntegrationMultipleSubscriptionsSameTopic tests subscribing to the same topic multiple times
// Run with: go test -tags=integration -v -run TestIntegrationMultipleSubscriptionsSameTopic
func TestIntegrationMultipleSubscriptionsSameTopic(t *testing.T) {
	var messagesReceived [][]byte
	var mu sync.Mutex

	client := New(
		WithLogger(NewLogger()),
		WithAutoReconnect(false),
		withOnNewMessage(func(data []byte) {
			mu.Lock()
			messagesReceived = append(messagesReceived, data)
			mu.Unlock()
			t.Logf("📨 Received message: %s", string(data))
		}),
	)

	// Connect
	err := client.Connect()
	require.NoError(t, err)
	defer client.Disconnect()
	time.Sleep(500 * time.Millisecond)

	t.Run("SubscribeMultipleTimes", func(t *testing.T) {
		// Subscribe to BTC three times
		for i := 0; i < 3; i++ {
			filter := NewCryptoPriceFilter(CryptoSymbolBTCUSDT)
			subscriptions := []Subscription{
				{
					Topic:   TopicCryptoPrices,
					Type:    MessageTypeUpdate,
					Filters: filter.ToJSON(),
				},
			}

			err := client.Subscribe(subscriptions)
			require.NoError(t, err)
			t.Logf("✅ Subscription #%d to BTC sent", i+1)
			time.Sleep(500 * time.Millisecond)
		}

		// Wait for messages
		time.Sleep(5 * time.Second)

		mu.Lock()
		messageCount := len(messagesReceived)
		mu.Unlock()

		t.Logf("📊 Total messages received: %d", messageCount)
		// Note: We might receive duplicate messages if server processes all subscriptions
	})
}

// TestIntegrationMultipleTopics tests subscribing to multiple different topics
// Run with: go test -tags=integration -v -run TestIntegrationMultipleTopics
func TestIntegrationMultipleTopics(t *testing.T) {
	var btcMessages, ethMessages, solMessages int
	var mu sync.Mutex

	router := NewRealtimeMessageRouter()

	// Register handlers to count messages per topic
	router.RegisterCryptoPriceHandler(func(price CryptoPrice) error {
		mu.Lock()
		defer mu.Unlock()

		switch price.Symbol {
		case CryptoSymbolBTCUSDT:
			btcMessages++
			t.Logf("📊 BTC Price: %s", price.Value.String())
		case CryptoSymbolETHUSDT:
			ethMessages++
			t.Logf("📊 ETH Price: %s", price.Value.String())
		case CryptoSymbolSOLUSDT:
			solMessages++
			t.Logf("📊 SOL Price: %s", price.Value.String())
		}
		return nil
	})

	client := New(
		WithLogger(NewLogger()),
		WithAutoReconnect(false),
		withOnNewMessage(func(data []byte) {
			if err := router.RouteMessage(data); err != nil {
				t.Logf("⚠️ Error routing message: %v", err)
			}
		}),
	)

	// Connect
	err := client.Connect()
	require.NoError(t, err)
	defer client.Disconnect()
	time.Sleep(500 * time.Millisecond)

	// Create typed subscription handler
	typedSub := NewRealtimeTypedSubscriptionHandler(client)

	t.Run("SubscribeMultipleCryptos", func(t *testing.T) {
		// Subscribe to BTC
		err := typedSub.SubscribeToCryptoPrices(nil, NewBTCPriceFilter())
		require.NoError(t, err)
		t.Log("✅ Subscribed to BTC")
		time.Sleep(500 * time.Millisecond)

		// Subscribe to ETH
		err = typedSub.SubscribeToCryptoPrices(nil, NewETHPriceFilter())
		require.NoError(t, err)
		t.Log("✅ Subscribed to ETH")
		time.Sleep(500 * time.Millisecond)

		// Subscribe to SOL
		err = typedSub.SubscribeToCryptoPrices(nil, NewSOLPriceFilter())
		require.NoError(t, err)
		t.Log("✅ Subscribed to SOL")
		time.Sleep(500 * time.Millisecond)

		// Wait for messages
		time.Sleep(10 * time.Second)

		mu.Lock()
		t.Logf("📊 BTC messages: %d", btcMessages)
		t.Logf("📊 ETH messages: %d", ethMessages)
		t.Logf("📊 SOL messages: %d", solMessages)
		totalMessages := btcMessages + ethMessages + solMessages
		mu.Unlock()

		assert.Greater(t, totalMessages, 0, "Should receive at least one message from any crypto")
	})
}

// TestIntegrationCLOBMarketSubscription tests CLOB market data subscriptions
// Run with: go test -tags=integration -v -run TestIntegrationCLOBMarketSubscription
func TestIntegrationCLOBMarketSubscription(t *testing.T) {
	var priceChangesReceived int
	var mu sync.Mutex

	router := NewClobMarketMessageRouter()

	router.RegisterPriceChangesHandler(func(priceChanges PriceChanges) error {
		mu.Lock()
		defer mu.Unlock()
		priceChangesReceived++
		t.Logf("📊 Price Changes for market: %s, changes: %d", priceChanges.Market, len(priceChanges.PriceChange))
		for i, change := range priceChanges.PriceChange {
			t.Logf("  Change #%d: AssetID=%s, Side=%s, Price=%s, Size=%s",
				i+1, change.AssetID, change.Side, change.Price.String(), change.Size.String())
		}
		return nil
	})

	client := NewClobMarketClient(
		WithLogger(NewLogger()),
		WithAutoReconnect(false),
		WithOnConnect(func() {
			t.Log("✅ Connected to CLOB WebSocket")
		}),
		withOnNewMessage(func(data []byte) {
			if err := router.RouteMessage(data); err != nil {
				t.Logf("⚠️ Error routing message: %v", err)
			}
		}),
	)

	// Connect
	err := client.Connect()
	require.NoError(t, err)
	defer client.Disconnect()
	time.Sleep(500 * time.Millisecond)

	// Create typed subscription handler
	typedSub := NewClobMarketTypedSubscriptionHandler(client)

	t.Run("SubscribeToPriceChanges", func(t *testing.T) {
		// Use a real token ID - this is from a popular market
		// You may need to update this with a currently active token ID
		tokenIDs := []string{
			"87263549609442107846940966542305431913459334507901399837670195489634390247458",
		}

		err := typedSub.SubscribeToPriceChanges(tokenIDs)
		require.NoError(t, err)
		t.Log("✅ Subscribed to CLOB market price changes")

		// Wait for messages
		time.Sleep(30 * time.Second)

		mu.Lock()
		count := priceChangesReceived
		mu.Unlock()

		t.Logf("📊 Total price changes received: %d", count)
		// Note: Active markets should have price changes, but it depends on market activity
	})
}

// TestIntegrationResubscribeAfterUnsubscribe tests unsubscribing and then resubscribing
// Run with: go test -tags=integration -v -run TestIntegrationResubscribeAfterUnsubscribe
func TestIntegrationResubscribeAfterUnsubscribe(t *testing.T) {
	var messagesPhase1, messagesPhase2, messagesPhase3 int
	var mu sync.Mutex
	phase := 1

	client := New(
		WithLogger(NewLogger()),
		WithAutoReconnect(false),
		withOnNewMessage(func(data []byte) {
			mu.Lock()
			switch phase {
			case 1:
				messagesPhase1++
			case 2:
				messagesPhase2++
			case 3:
				messagesPhase3++
			}
			mu.Unlock()
		}),
	)

	// Connect
	err := client.Connect()
	require.NoError(t, err)
	defer client.Disconnect()
	time.Sleep(500 * time.Millisecond)

	filter := NewCryptoPriceFilter(CryptoSymbolBTCUSDT)
	subscription := []Subscription{
		{
			Topic:   TopicCryptoPrices,
			Type:    MessageTypeUpdate,
			Filters: filter.ToJSON(),
		},
	}

	// Phase 1: Subscribe and receive messages
	t.Run("Phase1_Subscribe", func(t *testing.T) {
		err := client.Subscribe(subscription)
		require.NoError(t, err)
		t.Log("✅ Phase 1: Subscribed")

		time.Sleep(5 * time.Second)

		mu.Lock()
		count := messagesPhase1
		mu.Unlock()

		t.Logf("📊 Phase 1 messages: %d", count)
		assert.Greater(t, count, 0, "Should receive messages in phase 1")
	})

	// Phase 2: Unsubscribe and verify no messages
	t.Run("Phase2_Unsubscribe", func(t *testing.T) {
		mu.Lock()
		phase = 2
		mu.Unlock()

		err := client.Unsubscribe(subscription)
		require.NoError(t, err)
		t.Log("✅ Phase 2: Unsubscribed")

		time.Sleep(5 * time.Second)

		mu.Lock()
		count := messagesPhase2
		mu.Unlock()

		t.Logf("📊 Phase 2 messages (should be minimal): %d", count)
		// Note: Might receive some buffered messages
	})

	// Phase 3: Resubscribe and receive messages again
	t.Run("Phase3_Resubscribe", func(t *testing.T) {
		mu.Lock()
		phase = 3
		mu.Unlock()

		err := client.Subscribe(subscription)
		require.NoError(t, err)
		t.Log("✅ Phase 3: Resubscribed")

		time.Sleep(5 * time.Second)

		mu.Lock()
		count := messagesPhase3
		mu.Unlock()

		t.Logf("📊 Phase 3 messages: %d", count)
		assert.Greater(t, count, 0, "Should receive messages again in phase 3")
	})

	mu.Lock()
	t.Logf("📊 Summary - Phase1: %d, Phase2: %d, Phase3: %d",
		messagesPhase1, messagesPhase2, messagesPhase3)
	mu.Unlock()
}

// TestIntegrationConcurrentSubscriptions tests concurrent subscription operations on real connection
// Run with: go test -tags=integration -v -run TestIntegrationConcurrentSubscriptions
func TestIntegrationConcurrentSubscriptions(t *testing.T) {
	var totalMessages int
	var mu sync.Mutex

	client := New(
		WithLogger(NewLogger()),
		WithAutoReconnect(false),
		withOnNewMessage(func(data []byte) {
			mu.Lock()
			totalMessages++
			mu.Unlock()
		}),
	)

	// Connect
	err := client.Connect()
	require.NoError(t, err)
	defer client.Disconnect()
	time.Sleep(500 * time.Millisecond)

	t.Run("ConcurrentSubscribe", func(t *testing.T) {
		var wg sync.WaitGroup
		cryptos := []string{
			CryptoSymbolBTCUSDT,
			CryptoSymbolETHUSDT,
			CryptoSymbolSOLUSDT,
		}

		// Subscribe concurrently
		for _, crypto := range cryptos {
			wg.Add(1)
			go func(symbol string) {
				defer wg.Done()

				filter := NewCryptoPriceFilter(symbol)
				subscription := []Subscription{
					{
						Topic:   TopicCryptoPrices,
						Type:    MessageTypeUpdate,
						Filters: filter.ToJSON(),
					},
				}

				err := client.Subscribe(subscription)
				if err != nil {
					t.Errorf("Failed to subscribe to %s: %v", symbol, err)
				} else {
					t.Logf("✅ Subscribed to %s", symbol)
				}
			}(crypto)
		}

		wg.Wait()
		t.Log("✅ All concurrent subscriptions completed")

		// Wait for messages
		time.Sleep(10 * time.Second)

		mu.Lock()
		count := totalMessages
		mu.Unlock()

		t.Logf("📊 Total messages from all subscriptions: %d", count)
		assert.Greater(t, count, 0, "Should receive messages from concurrent subscriptions")
	})
}
