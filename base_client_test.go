//go:build integration
// +build integration

package polymarketrealtime

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestBaseClientSingleSubscription tests subscribing to a single topic
func TestBaseClientSingleSubscription(t *testing.T) {
	t.Log("=== Testing Single Subscription ===")

	var receivedSOL bool
	var mu sync.Mutex

	router := NewRealtimeMessageRouter()
	router.RegisterCryptoPriceHandler(func(price CryptoPrice) error {
		mu.Lock()
		defer mu.Unlock()

		if price.Symbol == CryptoSymbolSOLUSDT {
			receivedSOL = true
			t.Logf("✓ Received SOL price: %s", price.Value.String())
		}
		return nil
	})

	client := New(
		WithLogger(NewLogger()),
		WithAutoReconnect(false),
		WithOnConnect(func() {
			t.Log("✓ Connected")
		}),
		withOnNewMessage(func(data []byte) {
			if err := router.RouteMessage(data); err != nil {
				t.Logf("Error routing: %v", err)
			}
		}),
	)

	err := client.Connect()
	require.NoError(t, err)
	defer client.Disconnect()

	time.Sleep(500 * time.Millisecond)

	// Subscribe to SOL only
	t.Log("Subscribing to SOL...")
	err = client.Subscribe([]Subscription{
		{
			Topic:   TopicCryptoPrices,
			Type:    MessageTypeUpdate,
			Filters: `{"symbol":"solusdt"}`,
		},
	})
	require.NoError(t, err)

	// Wait for messages
	t.Log("Waiting for messages (15 seconds)...")
	time.Sleep(15 * time.Second)

	mu.Lock()
	gotSOL := receivedSOL
	mu.Unlock()

	require.True(t, gotSOL, "Should have received SOL price updates")
	t.Log("✓ Single subscription test passed")
}

// TestBaseClientSequentialSubscriptions tests subscribing to multiple topics sequentially
func TestBaseClientSequentialSubscriptions(t *testing.T) {
	t.Log("=== Testing Sequential Subscriptions ===")

	var receivedSOL, receivedETH bool
	var solCount, ethCount int
	var mu sync.Mutex

	router := NewRealtimeMessageRouter()
	router.RegisterCryptoPriceHandler(func(price CryptoPrice) error {
		mu.Lock()
		defer mu.Unlock()

		switch price.Symbol {
		case CryptoSymbolSOLUSDT:
			receivedSOL = true
			solCount++
			t.Logf("✓ Received SOL price #%d: %s", solCount, price.Value.String())
		case CryptoSymbolETHUSDT:
			receivedETH = true
			ethCount++
			t.Logf("✓ Received ETH price #%d: %s", ethCount, price.Value.String())
		}
		return nil
	})

	client := New(
		WithLogger(NewLogger()),
		WithAutoReconnect(false),
		WithOnConnect(func() {
			t.Log("✓ Connected")
		}),
		withOnNewMessage(func(data []byte) {
			if err := router.RouteMessage(data); err != nil {
				t.Logf("Error routing: %v", err)
			}
		}),
	)

	err := client.Connect()
	require.NoError(t, err)
	defer client.Disconnect()

	time.Sleep(500 * time.Millisecond)

	// Subscribe to SOL first
	t.Log("Step 1: Subscribing to SOL...")
	err = client.Subscribe([]Subscription{
		{
			Topic:   TopicCryptoPrices,
			Type:    MessageTypeUpdate,
			Filters: `{"symbol":"solusdt"}`,
		},
	})
	require.NoError(t, err)

	// Wait a bit
	time.Sleep(2 * time.Second)

	// Then subscribe to ETH
	t.Log("Step 2: Subscribing to ETH...")
	err = client.Subscribe([]Subscription{
		{
			Topic:   TopicCryptoPrices,
			Type:    MessageTypeUpdate,
			Filters: `{"symbol":"ethusdt"}`,
		},
	})
	require.NoError(t, err)

	// Wait for messages
	t.Log("Waiting for messages (20 seconds)...")
	time.Sleep(20 * time.Second)

	mu.Lock()
	gotSOL := receivedSOL
	gotETH := receivedETH
	finalSOLCount := solCount
	finalETHCount := ethCount
	mu.Unlock()

	t.Logf("Results: SOL=%v (count=%d), ETH=%v (count=%d)", gotSOL, finalSOLCount, gotETH, finalETHCount)

	require.True(t, gotSOL, "Should have received SOL price updates")
	require.True(t, gotETH, "Should have received ETH price updates")
	require.Greater(t, finalSOLCount, 0, "Should have received multiple SOL updates")
	require.Greater(t, finalETHCount, 0, "Should have received multiple ETH updates")
	t.Log("✓ Sequential subscriptions test passed")
}

// TestBaseClientBatchSubscription tests subscribing to multiple topics at once
func TestBaseClientBatchSubscription(t *testing.T) {
	t.Log("=== Testing Batch Subscription ===")

	var receivedSOL, receivedETH, receivedBTC bool
	var solCount, ethCount, btcCount int
	var mu sync.Mutex

	router := NewRealtimeMessageRouter()
	router.RegisterCryptoPriceHandler(func(price CryptoPrice) error {
		mu.Lock()
		defer mu.Unlock()

		switch price.Symbol {
		case CryptoSymbolSOLUSDT:
			receivedSOL = true
			solCount++
			t.Logf("✓ Received SOL price #%d: %s", solCount, price.Value.String())
		case CryptoSymbolETHUSDT:
			receivedETH = true
			ethCount++
			t.Logf("✓ Received ETH price #%d: %s", ethCount, price.Value.String())
		case CryptoSymbolBTCUSDT:
			receivedBTC = true
			btcCount++
			t.Logf("✓ Received BTC price #%d: %s", btcCount, price.Value.String())
		}
		return nil
	})

	client := New(
		WithLogger(NewLogger()),
		WithAutoReconnect(false),
		WithOnConnect(func() {
			t.Log("✓ Connected")
		}),
		withOnNewMessage(func(data []byte) {
			if err := router.RouteMessage(data); err != nil {
				t.Logf("Error routing: %v", err)
			}
		}),
	)

	err := client.Connect()
	require.NoError(t, err)
	defer client.Disconnect()

	time.Sleep(500 * time.Millisecond)

	// Subscribe to all three at once
	t.Log("Subscribing to SOL, ETH, and BTC all at once...")
	err = client.Subscribe([]Subscription{
		{
			Topic:   TopicCryptoPrices,
			Type:    MessageTypeUpdate,
			Filters: `{"symbol":"solusdt"}`,
		},
		{
			Topic:   TopicCryptoPrices,
			Type:    MessageTypeUpdate,
			Filters: `{"symbol":"ethusdt"}`,
		},
		{
			Topic:   TopicCryptoPrices,
			Type:    MessageTypeUpdate,
			Filters: `{"symbol":"btcusdt"}`,
		},
	})
	require.NoError(t, err)

	// Wait for messages
	t.Log("Waiting for messages (20 seconds)...")
	time.Sleep(20 * time.Second)

	mu.Lock()
	gotSOL := receivedSOL
	gotETH := receivedETH
	gotBTC := receivedBTC
	finalSOLCount := solCount
	finalETHCount := ethCount
	finalBTCCount := btcCount
	mu.Unlock()

	t.Logf("Results: SOL=%v (count=%d), ETH=%v (count=%d), BTC=%v (count=%d)",
		gotSOL, finalSOLCount, gotETH, finalETHCount, gotBTC, finalBTCCount)

	require.True(t, gotSOL, "Should have received SOL price updates")
	require.True(t, gotETH, "Should have received ETH price updates")
	require.True(t, gotBTC, "Should have received BTC price updates")
	require.Greater(t, finalSOLCount, 0, "Should have received multiple SOL updates")
	require.Greater(t, finalETHCount, 0, "Should have received multiple ETH updates")
	require.Greater(t, finalBTCCount, 0, "Should have received multiple BTC updates")
	t.Log("✓ Batch subscription test passed")
}

// TestBaseClientUnsubscribe tests unsubscribing from topics
func TestBaseClientUnsubscribe(t *testing.T) {
	t.Log("=== Testing Unsubscribe ===")

	var receivedSOL, receivedETH bool
	var solCount, ethCount int
	var ethCountAfterUnsub int
	var mu sync.Mutex

	router := NewRealtimeMessageRouter()
	router.RegisterCryptoPriceHandler(func(price CryptoPrice) error {
		mu.Lock()
		defer mu.Unlock()

		switch price.Symbol {
		case CryptoSymbolSOLUSDT:
			receivedSOL = true
			solCount++
			t.Logf("✓ Received SOL price #%d: %s", solCount, price.Value.String())
		case CryptoSymbolETHUSDT:
			receivedETH = true
			ethCount++
			t.Logf("✓ Received ETH price #%d: %s", ethCount, price.Value.String())
		}
		return nil
	})

	client := New(
		WithLogger(NewLogger()),
		WithAutoReconnect(false),
		WithOnConnect(func() {
			t.Log("✓ Connected")
		}),
		withOnNewMessage(func(data []byte) {
			if err := router.RouteMessage(data); err != nil {
				t.Logf("Error routing: %v", err)
			}
		}),
	)

	err := client.Connect()
	require.NoError(t, err)
	defer client.Disconnect()

	time.Sleep(500 * time.Millisecond)

	// Subscribe to both SOL and ETH
	t.Log("Step 1: Subscribing to SOL and ETH...")
	err = client.Subscribe([]Subscription{
		{
			Topic:   TopicCryptoPrices,
			Type:    MessageTypeUpdate,
			Filters: `{"symbol":"solusdt"}`,
		},
		{
			Topic:   TopicCryptoPrices,
			Type:    MessageTypeUpdate,
			Filters: `{"symbol":"ethusdt"}`,
		},
	})
	require.NoError(t, err)

	// Wait for messages
	t.Log("Waiting for messages (10 seconds)...")
	time.Sleep(10 * time.Second)

	mu.Lock()
	ethCountBeforeUnsub := ethCount
	mu.Unlock()

	t.Logf("Before unsubscribe: SOL count=%d, ETH count=%d", solCount, ethCountBeforeUnsub)

	// Unsubscribe from ETH
	t.Log("Step 2: Unsubscribing from ETH...")
	err = client.Unsubscribe([]Subscription{
		{
			Topic:   TopicCryptoPrices,
			Type:    MessageTypeUpdate,
			Filters: `{"symbol":"ethusdt"}`,
		},
	})
	require.NoError(t, err)

	// Wait more and verify ETH stops coming
	t.Log("Waiting for more messages (10 seconds)...")
	time.Sleep(10 * time.Second)

	mu.Lock()
	ethCountAfterUnsub = ethCount
	finalSOLCount := solCount
	mu.Unlock()

	t.Logf("After unsubscribe: SOL count=%d, ETH count=%d", finalSOLCount, ethCountAfterUnsub)

	require.True(t, receivedSOL, "Should have received SOL price updates")
	require.True(t, receivedETH, "Should have received ETH price updates initially")
	require.Greater(t, finalSOLCount, solCount, "SOL updates should continue after unsubscribe")
	require.Equal(t, ethCountBeforeUnsub, ethCountAfterUnsub, "ETH updates should stop after unsubscribe")
	t.Log("✓ Unsubscribe test passed")
}
