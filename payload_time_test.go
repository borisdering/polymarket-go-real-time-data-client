package polymarketrealtime

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestTradeTimeUnmarshal(t *testing.T) {
	jsonData := `{
		"asset": "123",
		"timestamp": 1762941635000,
		"price": "0.59",
		"size": "100",
		"side": "BUY",
		"slug": "test-market"
	}`

	var trade Trade
	err := json.Unmarshal([]byte(jsonData), &trade)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Check timestamp field
	if trade.Timestamp != 1762941635000 {
		t.Errorf("Expected Timestamp 1762941635000, got %d", trade.Timestamp)
	}

	// Check parsed time
	expectedTime := time.UnixMilli(1762941635000)
	if !trade.Time.Equal(expectedTime) {
		t.Errorf("Expected Time %v, got %v", expectedTime, trade.Time)
	}

	// Check time formatting
	timeStr := trade.Time.Format("2006-01-02 15:04:05")
	if timeStr == "" {
		t.Error("Time formatting failed")
	}
}

func TestCryptoPriceTimeUnmarshal(t *testing.T) {
	jsonData := `{
		"symbol": "btcusdt",
		"timestamp": 1762941635000,
		"value": "104572.89"
	}`

	var price CryptoPrice
	err := json.Unmarshal([]byte(jsonData), &price)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Check timestamp field
	if price.Timestamp != 1762941635000 {
		t.Errorf("Expected Timestamp 1762941635000, got %d", price.Timestamp)
	}

	// Check parsed time
	expectedTime := time.UnixMilli(1762941635000)
	if !price.Time.Equal(expectedTime) {
		t.Errorf("Expected Time %v, got %v", expectedTime, price.Time)
	}

	// Check value
	expectedValue := decimal.NewFromFloat(104572.89)
	if !price.Value.Equal(expectedValue) {
		t.Errorf("Expected Value %s, got %s", expectedValue.String(), price.Value.String())
	}
}

func TestEquityPriceTimeUnmarshal(t *testing.T) {
	jsonData := `{
		"symbol": "AAPL",
		"timestamp": 1762941635000,
		"value": "150.25"
	}`

	var price EquityPrice
	err := json.Unmarshal([]byte(jsonData), &price)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Check timestamp field
	if price.Timestamp != 1762941635000 {
		t.Errorf("Expected Timestamp 1762941635000, got %d", price.Timestamp)
	}

	// Check parsed time
	expectedTime := time.UnixMilli(1762941635000)
	if !price.Time.Equal(expectedTime) {
		t.Errorf("Expected Time %v, got %v", expectedTime, price.Time)
	}
}

func TestRFQRequestExpiryUnmarshal(t *testing.T) {
	// Note: RFQ expiry is in seconds, not milliseconds
	jsonData := `{
		"requestId": "test-123",
		"market": "0x123",
		"expiry": 1762941635,
		"price": "0.5",
		"sizeIn": "100",
		"sizeOut": "50"
	}`

	var request RFQRequest
	err := json.Unmarshal([]byte(jsonData), &request)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Check expiry field (seconds)
	if request.Expiry != 1762941635 {
		t.Errorf("Expected Expiry 1762941635, got %d", request.Expiry)
	}

	// Check parsed expiry time (seconds, not milliseconds)
	expectedTime := time.Unix(1762941635, 0)
	if !request.ExpiryTime.Equal(expectedTime) {
		t.Errorf("Expected ExpiryTime %v, got %v", expectedTime, request.ExpiryTime)
	}
}

func TestRFQQuoteExpiryUnmarshal(t *testing.T) {
	// Note: RFQ expiry is in seconds, not milliseconds
	jsonData := `{
		"quoteId": "quote-123",
		"requestId": "request-123",
		"expiry": 1762941635,
		"price": "0.5",
		"sizeIn": "100",
		"sizeOut": "50"
	}`

	var quote RFQQuote
	err := json.Unmarshal([]byte(jsonData), &quote)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Check expiry field (seconds)
	if quote.Expiry != 1762941635 {
		t.Errorf("Expected Expiry 1762941635, got %d", quote.Expiry)
	}

	// Check parsed expiry time (seconds, not milliseconds)
	expectedTime := time.Unix(1762941635, 0)
	if !quote.ExpiryTime.Equal(expectedTime) {
		t.Errorf("Expected ExpiryTime %v, got %v", expectedTime, quote.ExpiryTime)
	}
}

func TestTimeFieldsAreNotSerialized(t *testing.T) {
	// Create a CryptoPrice with time fields
	price := CryptoPrice{
		Symbol:    "btcusdt",
		Timestamp: 1762941635000,
		Time:      time.UnixMilli(1762941635000),
		Value:     decimal.NewFromFloat(104572.89),
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(price)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Parse JSON to verify Time field is not included
	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Check that "time" field is not in JSON
	if _, exists := result["time"]; exists {
		t.Error("Time field should not be serialized to JSON (json:\"-\" tag)")
	}

	// Check that timestamp field is present
	if _, exists := result["timestamp"]; !exists {
		t.Error("Timestamp field should be in JSON")
	}
}

func TestTimeComparison(t *testing.T) {
	jsonData := `{
		"symbol": "btcusdt",
		"timestamp": 1762941635000,
		"value": "104572.89"
	}`

	var price CryptoPrice
	err := json.Unmarshal([]byte(jsonData), &price)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Test time comparison
	now := time.Now()
	if price.Time.After(now) {
		t.Error("Price time should not be in the future")
	}

	// Test time arithmetic
	duration := now.Sub(price.Time)
	if duration < 0 {
		t.Error("Duration should be positive")
	}
}

func TestZeroTimestamp(t *testing.T) {
	jsonData := `{
		"symbol": "btcusdt",
		"timestamp": 0,
		"value": "104572.89"
	}`

	var price CryptoPrice
	err := json.Unmarshal([]byte(jsonData), &price)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Zero timestamp should result in zero time
	if !price.Time.IsZero() {
		t.Error("Expected zero time for zero timestamp")
	}
}
