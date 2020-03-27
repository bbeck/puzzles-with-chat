package crossword

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
)

func TestChannelLocator_Run(t *testing.T) {
	data, err := json.Marshal(ChannelsMessage{
		Kind:     "channels",
		Channels: []string{"channel1", "channel2"},
	})
	require.NoError(t, err)

	response := fmt.Sprintf("event:message\ndata:%s\n\n", data)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)

		_, err = w.Write([]byte(response))
		require.NoError(t, err)
	}))
	defer server.Close()

	parsed, err := url.Parse(server.URL)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var channels []string
	onUpdate := func(cs []string) {
		cancel()
		channels = append(channels, cs...)
	}

	onError := func(err error) {
		cancel()
		assert.NoError(t, err)
	}

	locator := NewChannelLocator(parsed.Host)
	locator.Run(ctx, onUpdate, onError)
	assert.ElementsMatch(t, []string{"channel1", "channel2"}, channels)
}

func TestChannelLocator_Run_Error(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{
			name: "non-json event",
			data: "not json",
		},
		{
			name: "event has a kind different than channels",
			data: `{"kind": "not channels"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			responses := []string{
				fmt.Sprintf("event:message\ndata:%s\n\n", test.data),
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)

				for _, response := range responses {
					_, err := w.Write([]byte(response))
					require.NoError(t, err)
				}
			}))
			defer server.Close()

			parsed, err := url.Parse(server.URL)
			require.NoError(t, err)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			onUpdate := func(cs []string) {
				cancel()
				assert.Fail(t, "channels shouldn't have been updated")
			}

			onError := func(err error) {
				cancel()
			}

			locator := NewChannelLocator(parsed.Host)
			locator.Run(ctx, onUpdate, onError)
		})
	}
}
