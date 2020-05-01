package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestChannelManager_Update(t *testing.T) {
	tests := []struct {
		name            string
		initial         map[ID][]string // initial channels the manager knows about
		updates         map[ID][]string // updates to apply, one at a time
		expectedAdded   []string        // expected channels added globally
		expectedRemoved []string        // expected channels removed globally
	}{
		{
			name: "single integration: no added or removed channels",
			initial: map[ID][]string{
				"integration": {"a"},
			},
			updates: map[ID][]string{
				"integration": {"a"},
			},
		},
		{
			name: "single integration: one added channel",
			updates: map[ID][]string{
				"integration": {"a"},
			},
			expectedAdded: []string{"a"},
		},
		{
			name: "single integration: multiple added channels",
			updates: map[ID][]string{
				"integration": {"a", "b"},
			},
			expectedAdded: []string{"a", "b"},
		},
		{
			name: "single integration: one removed channel",
			initial: map[ID][]string{
				"integration": {"a"},
			},
			updates: map[ID][]string{
				"integration": {},
			},
			expectedRemoved: []string{"a"},
		},
		{
			name: "single integration: multiple removed channels",
			initial: map[ID][]string{
				"integration": {"a", "b"},
			},
			updates: map[ID][]string{
				"integration": {},
			},
			expectedRemoved: []string{"a", "b"},
		},
		{
			name: "single integration: one added and one removed channel",
			initial: map[ID][]string{
				"integration": {"b", "c"},
			},
			updates: map[ID][]string{
				"integration": {"a", "b"},
			},
			expectedAdded:   []string{"a"},
			expectedRemoved: []string{"c"},
		},
		{
			name: "single integration: multiple added and removed channel",
			initial: map[ID][]string{
				"integration": {"c", "d"},
			},
			updates: map[ID][]string{
				"integration": {"a", "b"},
			},
			expectedAdded:   []string{"a", "b"},
			expectedRemoved: []string{"c", "d"},
		},
		{
			name: "multiple integrations: no added channels",
			initial: map[ID][]string{
				"integration-1": {"a"},
			},
			updates: map[ID][]string{
				"integration-2": {"a"},
			},
		},
		{
			name: "multiple integrations: one added channel",
			initial: map[ID][]string{
				"integration-1": {"a"},
			},
			updates: map[ID][]string{
				"integration-2": {"b"},
			},
			expectedAdded: []string{"b"},
		},
		{
			name: "multiple integrations: multiple added channels",
			initial: map[ID][]string{
				"integration-1": {"a"},
			},
			updates: map[ID][]string{
				"integration-2": {"b"},
				"integration-3": {"c"},
			},
			expectedAdded: []string{"b", "c"},
		},
		{
			name: "multiple integrations: no removed channels",
			initial: map[ID][]string{
				"integration-1": {"a"},
				"integration-2": {"a"},
			},
			updates: map[ID][]string{
				"integration-2": {},
			},
		},
		{
			name: "multiple integrations: one removed channel",
			initial: map[ID][]string{
				"integration-1": {"a"},
				"integration-2": {"b"},
			},
			updates: map[ID][]string{
				"integration-2": {},
			},
			expectedRemoved: []string{"b"},
		},
		{
			name: "multiple integrations: multiple removed channels",
			initial: map[ID][]string{
				"integration-1": {"a"},
				"integration-2": {"b", "c"},
			},
			updates: map[ID][]string{
				"integration-1": {},
				"integration-2": {"b"},
			},
			expectedRemoved: []string{"a", "c"},
		},
		{
			name: "multiple integrations: one added and one removed channel",
			initial: map[ID][]string{
				"integration-1": {"a"},
				"integration-2": {"b"},
			},
			updates: map[ID][]string{
				"integration-1": {"c"},
				"integration-2": {"d"},
			},
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

			var updateID ID
			var updateChannels []string
			onUpdate := func(id ID, channels []string) {
				updateID = id
				updateChannels = channels
			}

			// Setup the channel manager with our initial set of channels and handlers.
			manager := newChannelManager(test.initial)
			manager.OnAddChannel = onAdd
			manager.OnRemoveChannel = onRemove
			manager.OnUpdateChannels = onUpdate

			// Apply each update one at a time making sure OnUpdate is called with
			// each.
			for id, channels := range test.updates {
				manager.Update(id, channels)

				assert.Equal(t, id, updateID)
				assert.Equal(t, channels, updateChannels)
			}

			// Verify globally we received the correct additions and removals.
			assert.ElementsMatch(t, test.expectedAdded, added)
			assert.ElementsMatch(t, test.expectedRemoved, removed)
		})
	}
}

func newChannelManager(initial map[ID][]string) *ChannelManager {
	// Setup the channel manager with our initial set of channels and handlers.
	var channels map[ID]map[string]struct{}
	if initial != nil {
		channels = make(map[ID]map[string]struct{})
		for id, cs := range initial {
			for _, channel := range cs {
				if channels[id] == nil {
					channels[id] = make(map[string]struct{})
				}

				channels[id][channel] = struct{}{}
			}
		}
	}

	return &ChannelManager{channels: channels}
}
