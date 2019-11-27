package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientRegistry_RegisterClient_Error(t *testing.T) {
	tests := []struct {
		name    string
		channel string
		stream  chan Event
	}{
		{
			name:    "empty channel",
			channel: "",
			stream:  make(chan Event, 10),
		},
		{
			name:    "nil stream",
			channel: "channel",
			stream:  nil,
		},
		{
			name:    "unbuffered stream",
			channel: "channel",
			stream:  make(chan Event),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			registry := new(ClientRegistry)

			_, err := registry.RegisterClient(test.channel, test.stream)
			assert.Error(t, err)
		})
	}
}

func TestClientRegistry_DeregisterClient_DeregisteredClientStopsReceivingEvents(t *testing.T) {
	registry := new(ClientRegistry)

	stream := make(chan Event, 1)
	id, err := registry.RegisterClient("channel", stream)
	require.NoError(t, err)

	registry.BroadcastEvent("channel", Event{})
	assert.Equal(t, 1, len(receiveAll(stream)))

	registry.DeregisterClient("channel", id)

	registry.BroadcastEvent("channel", Event{})
	assert.Equal(t, 0, len(receiveAll(stream)))
}

func TestClientRegistry_DeregisterClient_MultipleTimes(t *testing.T) {
	registry := new(ClientRegistry)

	id, err := registry.RegisterClient("channel", make(chan Event, 1))
	require.NoError(t, err)

	registry.DeregisterClient("channel", id)
	registry.DeregisterClient("channel", id)
}

func TestClientRegistry_DeregisterClient_NonExistingId(t *testing.T) {
	registry := new(ClientRegistry)

	_, err := registry.RegisterClient("channel", make(chan Event, 1))
	require.NoError(t, err)

	registry.DeregisterClient("channel", "id")
}

func TestClientRegistry_DeregisterClient_EmptyRegistry(t *testing.T) {
	registry := new(ClientRegistry)
	registry.DeregisterClient("channel", "id")
}

func TestClientRegistry_BroadcastEvent(t *testing.T) {
	type client struct {
		channel  string
		expected []string
	}

	type event struct {
		channel string
		kind    string
	}

	tests := []struct {
		name    string
		clients []client
		events  []event
	}{
		{
			name: "client receives event",
			clients: []client{
				{channel: "A", expected: []string{"e1"}},
			},
			events: []event{
				{channel: "A", kind: "e1"},
			},
		},
		{
			name: "multiple clients receive event",
			clients: []client{
				{channel: "A", expected: []string{"e1"}},
				{channel: "A", expected: []string{"e1"}},
			},
			events: []event{
				{channel: "A", kind: "e1"},
			},
		},
		{
			name: "client doesn't receive event for different channel",
			clients: []client{
				{channel: "A"},
			},
			events: []event{
				{channel: "B", kind: "e1"},
			},
		},
		{
			name: "no clients",
			events: []event{
				{channel: "A", kind: "e1"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.True(t, len(test.events) > 0)

			registry := new(ClientRegistry)

			// Create the streams for each client and register them with the registry.
			streams := make([]chan Event, len(test.clients))
			for i, client := range test.clients {
				streams[i] = make(chan Event, len(test.events))

				_, err := registry.RegisterClient(client.channel, streams[i])
				require.NoError(t, err)
			}

			// Push all of the events to the registry.
			for _, event := range test.events {
				registry.BroadcastEvent(event.channel, Event{Kind: event.kind})
			}

			// Verify that each client received its expected set of events.
			for i, client := range test.clients {
				stream := streams[i]
				received := receiveAll(stream)
				close(stream)

				assert.True(t, containsAll(received, client.expected))
			}
		})
	}
}

func TestClientRegistry_BroadcastEvent_SkipsFullStream(t *testing.T) {
	registry := new(ClientRegistry)

	stream1 := make(chan Event, 1)
	_, err1 := registry.RegisterClient("channel", stream1)
	require.NoError(t, err1)

	stream2 := make(chan Event, 2)
	_, err2 := registry.RegisterClient("channel", stream2)
	require.NoError(t, err2)

	registry.BroadcastEvent("channel", Event{})
	registry.BroadcastEvent("channel", Event{})

	assert.Equal(t, 1, len(receiveAll(stream1)))
	assert.Equal(t, 2, len(receiveAll(stream2)))
}

func receiveAll(c chan Event) []string {
	var kinds []string
	for {
		select {
		case event, ok := <-c:
			if !ok {
				return kinds
			}
			kinds = append(kinds, event.Kind)
		default:
			return kinds
		}
	}
}

func containsAll(haystack []string, needles []string) bool {
	found := make(map[string]bool)
	for _, hay := range haystack {
		found[hay] = true
	}

	for _, needle := range needles {
		if !found[needle] {
			return false
		}
	}

	return true
}
