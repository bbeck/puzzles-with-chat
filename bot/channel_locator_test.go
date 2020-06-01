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

func TestProcessEvent(t *testing.T) {
	tests := []struct {
		name     string
		event    ChannelsEvent
		expected map[ID][]string
	}{
		{
			name:  "no puzzles",
			event: NewChannelsEvent("channels", nil, nil),
			expected: map[ID][]string{
				"crossword":   nil,
				"spellingbee": nil,
			},
		},
		{
			name:  "crossword puzzle only",
			event: NewChannelsEvent("channels", []string{"channel"}, nil),
			expected: map[ID][]string{
				"crossword":   {"channel"},
				"spellingbee": nil,
			},
		},
		{
			name:  "spellingbee puzzle only",
			event: NewChannelsEvent("channels", nil, []string{"channel"}),
			expected: map[ID][]string{
				"crossword":   nil,
				"spellingbee": {"channel"},
			},
		},
		{
			name:  "multiple puzzles (different channels)",
			event: NewChannelsEvent("channels", []string{"channel1"}, []string{"channel2"}),
			expected: map[ID][]string{
				"crossword":   {"channel1"},
				"spellingbee": {"channel2"},
			},
		},
		{
			name:  "multiple puzzles (same channels)",
			event: NewChannelsEvent("channels", []string{"channel"}, []string{"channel"}),
			expected: map[ID][]string{
				"crossword":   {"channel"},
				"spellingbee": {"channel"},
			},
		},
		{
			name:     "ping event",
			event:    NewChannelsEvent("ping", nil, nil),
			expected: nil, // update shouldn't be called
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bs, err := json.Marshal(test.event)
			require.NoError(t, err)

			channels, err := ProcessEvent(bs)
			require.NoError(t, err)
			assert.Equal(t, test.expected, channels)
		})
	}
}

func TestProcessEvent_Error(t *testing.T) {
	tests := []struct {
		name  string
		event []byte
	}{
		{
			name:  "invalid json",
			event: []byte("not json"),
		},
		{
			name:  "event has a kind different than channels",
			event: []byte(`{"kind":"not channels"}`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ProcessEvent(test.event)
			assert.Error(t, err)
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
		name             string
		event            []byte
		expectedUpdate   bool
		expectedChannels map[ID][]string
		expectedError    bool
	}{
		{
			name:  "ping event",
			event: marshal(t, NewChannelsEvent("ping", nil, nil)),
		},
		{
			name:           "channels event",
			event:          marshal(t, NewChannelsEvent("channels", []string{"channel1"}, []string{"channel2"})),
			expectedUpdate: true,
			expectedChannels: map[ID][]string{
				"crossword":   {"channel1"},
				"spellingbee": {"channel2"},
			},
		},
		{
			name:           "empty channels event",
			event:          marshal(t, NewChannelsEvent("channels", nil, nil)),
			expectedUpdate: true,
			expectedChannels: map[ID][]string{
				"crossword":   nil,
				"spellingbee": nil,
			},
		},
		{
			name:          "error parsing json",
			event:         []byte("not valid json"),
			expectedError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := fmt.Sprintf("event:message\ndata:%s\n\n", test.event)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)

				_, err := w.Write([]byte(response))
				require.NoError(t, err)
			}))
			defer server.Close()

			parsed, err := url.Parse(server.URL)
			require.NoError(t, err)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var onUpdateCalled bool
			var channels map[ID][]string
			onUpdate := func(update map[ID][]string) {
				onUpdateCalled = true
				channels = update
				cancel()
			}

			onError := func(e error) {
				err = e
				cancel()
			}

			// Ensure that we cancel the context even if a callback isn't invoked.
			time.AfterFunc(10*time.Millisecond, cancel)

			locator := NewChannelLocator(parsed.Host)
			locator.Run(ctx, onUpdate, onError)

			assert.Equal(t, test.expectedUpdate, onUpdateCalled)
			assert.Equal(t, test.expectedChannels, channels)
			if !test.expectedError {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func NewChannelsEvent(kind string, crosswords []string, spellingbees []string) ChannelsEvent {
	type Channel = struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	}

	var event ChannelsEvent
	event.Kind = kind

	event.Payload.Crosswords = make([]Channel, len(crosswords))
	for i, name := range crosswords {
		event.Payload.Crosswords[i] = Channel{
			Name:   name,
			Status: "solving",
		}
	}

	event.Payload.SpellingBees = make([]Channel, len(spellingbees))
	for i, name := range spellingbees {
		event.Payload.SpellingBees[i] = Channel{
			Name:   name,
			Status: "solving",
		}
	}
	for i, payload := range event.Payload.Crosswords {
		payload.Name = crosswords[i]
		payload.Status = "solving"
	}

	return event
}
