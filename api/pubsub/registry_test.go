package pubsub

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRegistry_Subscribe_Error(t *testing.T) {
	tests := []struct {
		name    string
		channel Channel
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
			registry := new(Registry)

			_, err := registry.Subscribe(test.channel, test.stream)
			assert.Error(t, err)
		})
	}
}

func TestRegistry_Unsubscribe_ClientStopsReceivingEvents(t *testing.T) {
	registry := new(Registry)

	stream := make(chan Event, 1)
	id, err := registry.Subscribe("channel", stream)
	require.NoError(t, err)

	registry.Publish("channel", Event{})
	assert.Equal(t, 1, len(receiveAll(stream)))

	registry.Unsubscribe("channel", id)

	registry.Publish("channel", Event{})
	assert.Equal(t, 0, len(receiveAll(stream)))
}

func TestRegistry_Unsubscribe_EmptyRegistry(t *testing.T) {
	registry := new(Registry)
	registry.Unsubscribe("channel", "id")
}

func TestRegistry_Unsubscribe_NonExistingClientID(t *testing.T) {
	registry := new(Registry)

	_, err := registry.Subscribe("channel", make(chan Event, 1))
	require.NoError(t, err)

	registry.Unsubscribe("channel", "id")
}

func TestRegistry_Unsubscribe_MultipleTimes(t *testing.T) {
	registry := new(Registry)

	id, err := registry.Subscribe("channel", make(chan Event, 1))
	require.NoError(t, err)

	registry.Unsubscribe("channel", id)
	registry.Unsubscribe("channel", id)
}

func TestRegistry_Publish(t *testing.T) {
	type client struct {
		channel  Channel
		expected []string
	}

	type event struct {
		channel Channel
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

			registry := new(Registry)

			// Create the streams for each client and register them with the registry.
			streams := make([]chan Event, len(test.clients))
			for i, client := range test.clients {
				streams[i] = make(chan Event, len(test.events))

				_, err := registry.Subscribe(client.channel, streams[i])
				require.NoError(t, err)
			}

			// Push all of the events to the registry.
			for _, event := range test.events {
				registry.Publish(event.channel, Event{Kind: event.kind})
			}

			// Verify that each client received its expected set of events.
			for i, client := range test.clients {
				stream := streams[i]
				received := receiveAll(stream)
				close(stream)

				assert.ElementsMatch(t, client.expected, received)
			}
		})
	}
}

func TestRegistry_Publish_SkipsPublishWhenStreamIsFull(t *testing.T) {
	registry := new(Registry)

	stream1 := make(chan Event, 1)
	_, err1 := registry.Subscribe("channel", stream1)
	require.NoError(t, err1)

	stream2 := make(chan Event, 2)
	_, err2 := registry.Subscribe("channel", stream2)
	require.NoError(t, err2)

	registry.Publish("channel", Event{})
	registry.Publish("channel", Event{})

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
