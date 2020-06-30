package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMessageRouter(t *testing.T) {
	tests := []struct {
		name     string
		handlers map[ID]MessageHandler
	}{
		{
			name: "no integrations",
		},
		{
			name: "one integration",
			handlers: map[ID]MessageHandler{
				"integration-1": TestMessageHandler{id: "integration-1"},
			},
		},
		{
			name: "multiple integrations",
			handlers: map[ID]MessageHandler{
				"integration-1": TestMessageHandler{id: "integration-1"},
				"integration-2": TestMessageHandler{id: "integration-2"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := NewMessageRouter(test.handlers)

			assert.Equal(t, test.handlers, router.handlers)
		})
	}
}

func TestMessageRouter_AddIntegration(t *testing.T) {
	tests := []struct {
		name     string
		initial  map[string]map[ID]string // initial set of mappings
		adds     map[string]map[ID]string // adds to apply one at a time
		expected map[string]map[ID]string // expected statuses to see
	}{
		{
			name: "one integration",
			adds: map[string]map[ID]string{
				"channel": {"crossword": "solving"},
			},
			expected: map[string]map[ID]string{
				"channel": {"crossword": "solving"},
			},
		},
		{
			name: "existing integrations",
			initial: map[string]map[ID]string{
				"channel": {"acrostic": "complete"},
			},
			adds: map[string]map[ID]string{
				"channel": {"crossword": "solving"},
			},
			expected: map[string]map[ID]string{
				"channel": {
					"acrostic":  "complete",
					"crossword": "solving",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := &MessageRouter{statuses: test.initial}
			for channel, adds := range test.adds {
				for app, status := range adds {
					router.AddIntegration(app, channel, status)
				}
			}

			assert.Equal(t, test.expected, router.statuses)
		})
	}
}

func TestMessageRouter_RemoveIntegration(t *testing.T) {
	tests := []struct {
		name     string
		initial  map[string]map[ID]string // initial set of mappings
		removes  map[string][]ID          // removes to apply one at a time
		expected map[string]map[ID]string // expected statuses to see
	}{
		{
			name: "one integration",
			initial: map[string]map[ID]string{
				"channel": {"crossword": "complete"},
			},
			removes: map[string][]ID{
				"channel": {"crossword"},
			},
			expected: map[string]map[ID]string{},
		},
		{
			name: "remaining integrations",
			initial: map[string]map[ID]string{
				"channel": {
					"acrostic":  "complete",
					"crossword": "solving",
				},
			},
			removes: map[string][]ID{
				"channel": {"acrostic"},
			},
			expected: map[string]map[ID]string{
				"channel": {"crossword": "solving"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := &MessageRouter{statuses: test.initial}
			for channel, apps := range test.removes {
				for _, app := range apps {
					router.RemoveIntegration(app, channel)
				}
			}

			assert.Equal(t, test.expected, router.statuses)
		})
	}
}

func TestMessageRouter_UpdateIntegrationStatus(t *testing.T) {
	tests := []struct {
		name     string
		initial  map[string]map[ID]string // initial set of mappings
		updates  map[string]map[ID]string // updates to apply one at a time
		expected map[string]map[ID]string // expected statuses to see
	}{
		{
			name: "solving to complete",
			initial: map[string]map[ID]string{
				"channel": {"crossword": "solving"},
			},
			updates: map[string]map[ID]string{
				"channel": {"crossword": "complete"},
			},
			expected: map[string]map[ID]string{
				"channel": {"crossword": "complete"},
			},
		},
		{
			name: "multiple active applications",
			initial: map[string]map[ID]string{
				"channel": {
					"acrostic":  "selected",
					"crossword": "solving",
				},
			},
			updates: map[string]map[ID]string{
				"channel": {"crossword": "complete"},
			},
			expected: map[string]map[ID]string{
				"channel": {
					"acrostic":  "selected",
					"crossword": "complete",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := &MessageRouter{statuses: test.initial}
			for channel, apps := range test.updates {
				for app, status := range apps {
					router.UpdateIntegrationStatus(app, channel, status)
				}
			}

			assert.Equal(t, test.expected, router.statuses)
		})
	}
}

func TestMessageRouter_HandleChannelMessage(t *testing.T) {
	tests := []struct {
		name     string
		handlers []ID                     // which integrations should have handlers
		initial  map[string]map[ID]string // initial set of mappings
		channel  string                   // the channel to send a message from
		expected []ID                     // which integrations should receive the message
	}{
		{
			name:     "message from channel with no apps",
			handlers: []ID{"crossword"},
			channel:  "channel",
		},
		{
			name:     "message from channel with one app",
			handlers: []ID{"crossword"},
			initial: map[string]map[ID]string{
				"channel": {"crossword": "solving"},
			},
			channel:  "channel",
			expected: []ID{"crossword"},
		},
		{
			name:     "message from channel with multiple apps",
			handlers: []ID{"acrostic", "crossword"},
			initial: map[string]map[ID]string{
				"channel": {
					"acrostic":  "solving",
					"crossword": "solving",
				},
			},
			channel:  "channel",
			expected: []ID{"acrostic", "crossword"},
		},
		{
			name:     "message to different channel not received",
			handlers: []ID{"acrostic", "crossword"},
			initial: map[string]map[ID]string{
				"channel-1": {
					"acrostic":  "solving",
					"crossword": "solving",
				},
			},
			channel:  "channel-2",
			expected: []ID{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var called []ID // which integrations received messages

			handlers := make(map[ID]MessageHandler)
			for _, app := range test.handlers {
				app := app
				handlers[app] = TestMessageHandler{app, func() {
					called = append(called, app)
				}}
			}

			router := &MessageRouter{
				handlers: handlers,
				statuses: test.initial,
			}
			router.HandleChannelMessage(test.channel, "userid", "username", "message")
			assert.ElementsMatch(t, test.expected, called)
		})
	}
}

type TestMessageHandler struct {
	id ID
	fn func()
}

func (h TestMessageHandler) HandleChannelMessage(_, _, _ string) {
	h.fn()
}
