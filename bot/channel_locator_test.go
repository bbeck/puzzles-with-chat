package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestChannelLocator_ProcessPayload(t *testing.T) {
	tests := []struct {
		name     string
		payload  ChannelsPayload
		expected []Update
	}{
		{
			name:    "no puzzles",
			payload: ChannelsPayload{},
		},
		{
			name: "single puzzle",
			payload: ChannelsPayload{
				"crossword": {
					{Name: "channel", Status: "solving"},
				},
			},
			expected: []Update{
				{Application: "crossword", Channel: "channel", Status: "solving"},
			},
		},
		{
			name: "multiple puzzles",
			payload: ChannelsPayload{
				"acrostic": {
					{Name: "channel1", Status: "solving"},
				},
				"crossword": {
					{Name: "channel2", Status: "solving"},
					{Name: "channel3", Status: "solving"},
				},
				"spellingbee": {
					{Name: "channel4", Status: "solving"},
				},
			},
			expected: []Update{
				{Application: "acrostic", Channel: "channel1", Status: "solving"},
				{Application: "crossword", Channel: "channel2", Status: "solving"},
				{Application: "crossword", Channel: "channel3", Status: "solving"},
				{Application: "spellingbee", Channel: "channel4", Status: "solving"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var updateCalled bool
			update := func(updates []Update) {
				assert.False(t, updateCalled)
				assert.ElementsMatch(t, test.expected, updates)
				updateCalled = true
			}

			ProcessPayload(test.payload, update)
			assert.True(t, updateCalled)
		})
	}
}

func TestChannelLocator_Run(t *testing.T) {
	marshal := func(t *testing.T, v interface{}) []byte {
		bs, err := json.Marshal(v)
		require.NoError(t, err)
		return bs
	}

	tests := []struct {
		name                 string
		event                []byte
		expectedUpdateCalled bool
		expectedUpdates      []Update
		expectedFailCalled   bool
	}{
		{
			name: "ping event",
			event: marshal(t, Event{
				Kind: "ping",
			}),
		},
		{
			name: "channels event",
			event: marshal(t, Event{
				Kind: "channels",
				Payload: json.RawMessage(`{
					"acrostic": [
						{ "name": "channel1", "status": "solving" }
					],
					"crossword": [
						{ "name": "channel2", "status": "solving" },
						{ "name": "channel3", "status": "solving" }
					],
					"spellingbee": [
						{ "name": "channel4", "status": "solving" }
					]
				}`),
			}),
			expectedUpdateCalled: true,
			expectedUpdates: []Update{
				{Application: "acrostic", Channel: "channel1", Status: "solving"},
				{Application: "crossword", Channel: "channel2", Status: "solving"},
				{Application: "crossword", Channel: "channel3", Status: "solving"},
				{Application: "spellingbee", Channel: "channel4", Status: "solving"},
			},
		},
		{
			name: "empty channels event",
			event: marshal(t, Event{
				Kind: "channels",
			}),
			expectedUpdateCalled: true,
		},
		{
			name:               "error parsing event",
			event:              []byte(`not valid json`),
			expectedFailCalled: true,
		},
		{
			name: "unsupported event kind",
			event: marshal(t, Event{
				Kind: "unsupported",
			}),
			expectedFailCalled: true,
		},
		{
			name:               "error parsing payload",
			event:              []byte(`{"kind": "channels", "payload": true}`),
			expectedFailCalled: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			event := fmt.Sprintf("event:message\ndata:%s\n\n", test.event)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)

				_, err := w.Write([]byte(event))
				require.NoError(t, err)
			}))
			defer server.Close()

			parsed, err := url.Parse(server.URL)
			require.NoError(t, err)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var updateCalled bool
			update := func(updates []Update) {
				assert.False(t, updateCalled)
				assert.ElementsMatch(t, test.expectedUpdates, updates)
				updateCalled = true
				cancel()
			}

			var failCalled bool
			fail := func(e error) {
				assert.False(t, failCalled)
				failCalled = true
				cancel()
			}

			// Ensure that we cancel the context even if a callback isn't invoked.
			time.AfterFunc(10*time.Millisecond, cancel)

			locator := NewChannelLocator(parsed.Host)
			locator.Run(ctx, update, fail)

			assert.Equal(t, updateCalled, test.expectedUpdateCalled)
			assert.Equal(t, failCalled, test.expectedFailCalled)
		})
	}
}
