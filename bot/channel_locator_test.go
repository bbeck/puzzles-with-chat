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
			name:     "no puzzles",
			event:    NewChannelsEvent("channels", nil, nil),
			expected: make(map[ID][]string),
		},
		{
			name:  "crossword puzzle only",
			event: NewChannelsEvent("channels", []string{"channel"}, nil),
			expected: map[ID][]string{
				"crossword": {"channel"},
			},
		},
		{
			name:  "spellingbee puzzle only",
			event: NewChannelsEvent("channels", nil, []string{"channel"}),
			expected: map[ID][]string{
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
		expectedChannels map[ID][]string
		expectedError    bool
	}{
		{
			name:  "ping event",
			event: marshal(t, NewChannelsEvent("ping", nil, nil)),
		},
		{
			name:  "channels event",
			event: marshal(t, NewChannelsEvent("channels", []string{"channel1"}, []string{"channel2"})),
			expectedChannels: map[ID][]string{
				"crossword":   {"channel1"},
				"spellingbee": {"channel2"},
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

			var channels map[ID][]string
			onUpdate := func(update map[ID][]string) {
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
	var payload struct {
		Crosswords []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"crossword"`
		SpellingBees []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"spellingbee"`
	}

	payload.Crosswords = make([]struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	}, 0)
	for _, name := range crosswords {
		payload.Crosswords = append(payload.Crosswords, struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		}{
			Name:   name,
			Status: "solving",
		})
	}

	payload.SpellingBees = make([]struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	}, 0)
	for _, name := range spellingbees {
		payload.SpellingBees = append(payload.SpellingBees, struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		}{
			Name:   name,
			Status: "solving",
		})
	}

	return ChannelsEvent{
		Kind:    kind,
		Payload: payload,
	}
}
