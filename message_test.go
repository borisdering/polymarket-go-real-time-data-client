package polymarketrealtime

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessage_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		message  Message
		expected string
	}{
		{
			name: "marshal message with timestamp",
			message: Message{
				ConnectionID: "conn123",
				Timestamp:    1762928533586,
				Time:         time.Unix(0, 1762928533586*int64(time.Millisecond)),
				Topic:        TopicActivity,
				Type:         MessageTypeOrdersMatched,
			},
			expected: `{"connection_id":"conn123","timestamp":1762928533586,"topic":"activity","type":"orders_matched"}`,
		},
		{
			name: "marshal message with zero timestamp",
			message: Message{
				ConnectionID: "conn456",
				Timestamp:    0,
				Time:         time.Time{},
				Topic:        TopicComments,
				Type:         MessageTypeCommentCreated,
			},
			expected: `{"connection_id":"conn456","timestamp":0,"topic":"comments","type":"comment_created"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.message)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(data))
		})
	}
}

func TestMessage_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected Message
	}{
		{
			name:     "unmarshal message with timestamp",
			jsonData: `{"connection_id":"conn123","timestamp":1762928533586,"topic":"activity","type":"orders_matched"}`,
			expected: Message{
				ConnectionID: "conn123",
				Timestamp:    1762928533586,
				Time:         time.Unix(0, 1762928533586*int64(time.Millisecond)),
				Topic:        TopicActivity,
				Type:         MessageTypeOrdersMatched,
			},
		},
		{
			name:     "unmarshal message with zero timestamp",
			jsonData: `{"connection_id":"conn456","timestamp":0,"topic":"comments","type":"comment_created"}`,
			expected: Message{
				ConnectionID: "conn456",
				Timestamp:    0,
				Time:         time.Time{},
				Topic:        TopicComments,
				Type:         MessageTypeCommentCreated,
			},
		},
		{
			name:     "unmarshal message with missing timestamp",
			jsonData: `{"connection_id":"conn789","topic":"rfq","type":"request_created"}`,
			expected: Message{
				ConnectionID: "conn789",
				Timestamp:    0,
				Time:         time.Time{},
				Topic:        TopicRfq,
				Type:         MessageTypeRequestCreated,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var message Message
			err := json.Unmarshal([]byte(tt.jsonData), &message)
			require.NoError(t, err)

			assert.Equal(t, tt.expected.ConnectionID, message.ConnectionID)
			assert.Equal(t, tt.expected.Timestamp, message.Timestamp)
			assert.Equal(t, tt.expected.Topic, message.Topic)
			assert.Equal(t, tt.expected.Type, message.Type)

			// Check time field - allow for small differences due to timezone handling
			if tt.expected.Timestamp != 0 {
				// For non-zero timestamps, check that the time is approximately correct
				expectedTime := tt.expected.Time
				actualTime := message.Time
				assert.WithinDuration(t, expectedTime, actualTime, time.Millisecond)
			} else {
				// For zero timestamps, check that time is zero
				assert.True(t, message.Time.IsZero())
			}
		})
	}
}

func TestMessage_UnmarshalJSON_InvalidData(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
	}{
		{
			name:     "invalid JSON",
			jsonData: `{"connection_id":"conn123","timestamp":"invalid","topic":"activity","type":"orders_matched"}`,
		},
		{
			name:     "malformed JSON",
			jsonData: `{"connection_id":"conn123","timestamp":1762928533586,"topic":"activity","type":"orders_matched"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var message Message
			err := json.Unmarshal([]byte(tt.jsonData), &message)
			assert.Error(t, err)
		})
	}
}

func TestMessage_RoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		message Message
	}{
		{
			name: "round trip with timestamp",
			message: Message{
				ConnectionID: "conn123",
				Timestamp:    1762928533586,
				Time:         time.Unix(0, 1762928533586*int64(time.Millisecond)),
				Topic:        TopicActivity,
				Type:         MessageTypeOrdersMatched,
			},
		},
		{
			name: "round trip with zero timestamp",
			message: Message{
				ConnectionID: "conn456",
				Timestamp:    0,
				Time:         time.Time{},
				Topic:        TopicComments,
				Type:         MessageTypeCommentCreated,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal the message
			data, err := json.Marshal(tt.message)
			require.NoError(t, err)

			// Unmarshal the data back
			var unmarshaledMessage Message
			err = json.Unmarshal(data, &unmarshaledMessage)
			require.NoError(t, err)

			// Check that the fields match
			assert.Equal(t, tt.message.ConnectionID, unmarshaledMessage.ConnectionID)
			assert.Equal(t, tt.message.Timestamp, unmarshaledMessage.Timestamp)
			assert.Equal(t, tt.message.Topic, unmarshaledMessage.Topic)
			assert.Equal(t, tt.message.Type, unmarshaledMessage.Type)

			// Check time field
			if tt.message.Timestamp != 0 {
				// For non-zero timestamps, check that the time is approximately correct
				expectedTime := tt.message.Time
				actualTime := unmarshaledMessage.Time
				assert.WithinDuration(t, expectedTime, actualTime, time.Millisecond)
			} else {
				// For zero timestamps, check that time is zero
				assert.True(t, unmarshaledMessage.Time.IsZero())
			}
		})
	}
}
