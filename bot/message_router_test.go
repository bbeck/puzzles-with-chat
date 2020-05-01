package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var PRESENT = struct{}{}

func TestNewMessageRouter(t *testing.T) {
	tests := []struct {
		name         string
		integrations []Integration
	}{
		{
			name: "no integrations",
		},
		{
			name: "one integration",
			integrations: []Integration{
				{
					ID:             "integration-1",
					MessageHandler: TestMessageHandler{id: "integration-1"},
				},
			},
		},
		{
			name: "multiple integrations",
			integrations: []Integration{
				{
					ID:             "integration-1",
					MessageHandler: TestMessageHandler{id: "integration-1"},
				},
				{
					ID:             "integration-2",
					MessageHandler: TestMessageHandler{id: "integration-2"},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := NewMessageRouter(test.integrations)

			assert.Equal(t, len(test.integrations), len(router.handlers))
			for _, integration := range test.integrations {
				require.NotNil(t, router.handlers[integration.ID])
				require.IsType(t, TestMessageHandler{}, router.handlers[integration.ID])
				assert.Equal(t, integration.ID, router.handlers[integration.ID].(TestMessageHandler).id)
			}
		})
	}
}

func TestMessageRouter_UpdateChannels(t *testing.T) {
	tests := []struct {
		name     string
		initial  map[string]map[ID]struct{} // the initial set of integrations
		updates  map[ID][]string            // the updates to apply, one at a time
		expected map[string]map[ID]struct{} // the expected integrations after the update
	}{
		{
			name:    "add single integration to channel",
			initial: make(map[string]map[ID]struct{}),
			updates: map[ID][]string{
				"integration": {"a"},
			},
			expected: map[string]map[ID]struct{}{
				"a": {"integration": PRESENT},
			},
		},
		{
			name:    "add single integrations to multiple channels",
			initial: make(map[string]map[ID]struct{}),
			updates: map[ID][]string{
				"integration-1": {"a"},
				"integration-2": {"b"},
			},
			expected: map[string]map[ID]struct{}{
				"a": {"integration-1": PRESENT},
				"b": {"integration-2": PRESENT},
			},
		},
		{
			name:    "add multiple integrations to one channel",
			initial: make(map[string]map[ID]struct{}),
			updates: map[ID][]string{
				"integration-1": {"a"},
				"integration-2": {"a"},
			},
			expected: map[string]map[ID]struct{}{
				"a": {"integration-1": PRESENT, "integration-2": PRESENT},
			},
		},
		{
			name: "change integration of one channel",
			initial: map[string]map[ID]struct{}{
				"a": {"integration-1": PRESENT},
			},
			updates: map[ID][]string{
				"integration-1": {},
				"integration-2": {"a"},
			},
			expected: map[string]map[ID]struct{}{
				"a": {"integration-2": PRESENT},
			},
		},
		{
			name: "remove integration of one channel",
			initial: map[string]map[ID]struct{}{
				"a": {"integration-1": PRESENT},
			},
			updates: map[ID][]string{
				"integration-1": {},
			},
			expected: make(map[string]map[ID]struct{}),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := &MessageRouter{
				integrations: test.initial,
			}

			for id, channels := range test.updates {
				router.UpdateChannels(id, channels)
			}

			assertIntegrationsEqual(t, test.expected, router.integrations)
		})
	}
}

func TestMessageRouter_HandleChannelMessage(t *testing.T) {
	tests := []struct {
		name         string
		handlers     []ID                       // which integrations should have handlers
		integrations map[string]map[ID]struct{} // mapping of channel to its integrations
		channel      string                     // the channel a message is received from
		expected     []ID                       // which integrations are expected to be called
	}{
		{
			name:     "message on channel with integration received",
			handlers: []ID{"integration-1"},
			integrations: map[string]map[ID]struct{}{
				"a": {"integration-1": struct{}{}},
			},
			channel:  "a",
			expected: []ID{"integration-1"},
		},
		{
			name:     "message sent to different channel not received",
			handlers: []ID{"integration-1"},
			integrations: map[string]map[ID]struct{}{
				"a": {"integration-1": struct{}{}},
			},
			channel:  "b",
			expected: []ID{},
		},
		{
			name:     "message sent multiple to integrations",
			handlers: []ID{"integration-1", "integration-2"},
			integrations: map[string]map[ID]struct{}{
				"a": {"integration-1": struct{}{}, "integration-2": struct{}{}},
			},
			channel:  "a",
			expected: []ID{"integration-1", "integration-2"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var called []ID // which integrations were called

			handlers := make(map[ID]MessageHandler)
			for _, id := range test.handlers {
				id := id
				handlers[id] = TestMessageHandler{id, func() {
					called = append(called, id)
				}}
			}

			router := &MessageRouter{
				handlers:     handlers,
				integrations: test.integrations,
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

func (h TestMessageHandler) HandleChannelMessage(channel, userid, username, message string) {
	if h.fn != nil {
		h.fn()
	}
}

func assertIntegrationsEqual(t *testing.T, expected, actual map[string]map[ID]struct{}) {
	var expectedChannels, actualChannels []string
	for channel := range expected {
		expectedChannels = append(expectedChannels, channel)
	}
	for channel := range actual {
		actualChannels = append(actualChannels, channel)
	}
	if !assert.ElementsMatch(t, expectedChannels, actualChannels) {
		return
	}

	for channel := range expected {
		var expectedIntegrations, actualIntegrations []ID
		for integration := range expected[channel] {
			expectedIntegrations = append(expectedIntegrations, integration)
		}
		for integration := range actual[channel] {
			actualIntegrations = append(actualIntegrations, integration)
		}

		assert.ElementsMatch(t, expectedIntegrations, actualIntegrations)
	}
}
