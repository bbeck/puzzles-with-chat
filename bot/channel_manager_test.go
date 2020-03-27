package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestChannelManager_Update(t *testing.T) {
	tests := []struct {
		name            string
		initial         []string
		channels        []string
		expectedAdded   []string
		expectedRemoved []string
	}{
		{
			name:          "no initial channels",
			channels:      []string{"a"},
			expectedAdded: []string{"a"},
		},
		{
			name:          "one added channel",
			channels:      []string{"a"},
			expectedAdded: []string{"a"},
		},
		{
			name:          "multiple added channels",
			channels:      []string{"a", "b"},
			expectedAdded: []string{"a", "b"},
		},
		{
			name:            "one removed channel",
			initial:         []string{"a"},
			channels:        []string{},
			expectedRemoved: []string{"a"},
		},
		{
			name:            "multiple removed channels",
			initial:         []string{"a", "b"},
			channels:        []string{},
			expectedRemoved: []string{"a", "b"},
		},
		{
			name:            "one added and removed channel",
			initial:         []string{"a", "b"},
			channels:        []string{"b", "c"},
			expectedAdded:   []string{"c"},
			expectedRemoved: []string{"a"},
		},
		{
			name:            "multiple added and removed channel",
			initial:         []string{"a", "b"},
			channels:        []string{"c", "d"},
			expectedAdded:   []string{"c", "d"},
			expectedRemoved: []string{"a", "b"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var added []string
			onAdd := func(channel string) {
				added = append(added, channel)
			}

			var removed []string
			onRemove := func(channel string) {
				removed = append(removed, channel)
			}

			// Setup the channel manager with our initial set of channels and handlers.
			manager := newChannelManager(test.initial)
			manager.OnAddChannel = onAdd
			manager.OnRemoveChannel = onRemove
			manager.Update(test.channels)

			assert.ElementsMatch(t, test.expectedAdded, added)
			assert.ElementsMatch(t, test.expectedRemoved, removed)
		})
	}
}

func newChannelManager(initial []string) *ChannelManager {
	// Setup the channel manager with our initial set of channels and handlers.
	var channels map[string]struct{}
	if initial != nil {
		channels = make(map[string]struct{})
		for _, channel := range initial {
			channels[channel] = struct{}{}
		}
	}

	return &ChannelManager{channels: channels}
}
