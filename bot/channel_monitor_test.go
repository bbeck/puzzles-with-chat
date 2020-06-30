package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestChannelMonitor_ComputeAddedChannels(t *testing.T) {
	tests := []struct {
		name     string
		before   []Update
		after    []Update
		expected []string
	}{
		{
			name: "no channels",
		},
		{
			name: "single added channel",
			after: []Update{
				{Application: "application", Channel: "a", Status: "solving"},
			},
			expected: []string{"a"},
		},
		{
			name: "multiple added channels",
			after: []Update{
				{Application: "application", Channel: "a", Status: "solving"},
				{Application: "application", Channel: "b", Status: "solving"},
			},
			expected: []string{"a", "b"},
		},
		{
			name: "single added channel, multiple applications",
			after: []Update{
				{Application: "application1", Channel: "a", Status: "solving"},
				{Application: "application2", Channel: "a", Status: "solving"},
			},
			expected: []string{"a"},
		},
		{
			name: "no added channels",
			before: []Update{
				{Application: "application", Channel: "a", Status: "solving"},
				{Application: "application", Channel: "b", Status: "solving"},
			},
			after: []Update{
				{Application: "application", Channel: "a", Status: "solving"},
				{Application: "application", Channel: "b", Status: "solving"},
			},
			expected: []string{},
		},
		{
			name: "no added channels, status change",
			before: []Update{
				{Application: "application", Channel: "a", Status: "solving"},
			},
			after: []Update{
				{Application: "application", Channel: "a", Status: "complete"},
			},
			expected: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := ComputeAddedChannels(test.before, test.after)
			assert.ElementsMatch(t, test.expected, actual)
		})
	}
}

func TestChannelMonitor_ComputeRemovedChannels(t *testing.T) {
	tests := []struct {
		name     string
		before   []Update
		after    []Update
		expected []string
	}{
		{
			name: "no channels",
		},
		{
			name: "single removed channel",
			before: []Update{
				{Application: "application", Channel: "a", Status: "solving"},
			},
			expected: []string{"a"},
		},
		{
			name: "multiple removed channels",
			before: []Update{
				{Application: "application", Channel: "a", Status: "solving"},
				{Application: "application", Channel: "b", Status: "solving"},
			},
			expected: []string{"a", "b"},
		},
		{
			name: "single removed channel, multiple applications",
			before: []Update{
				{Application: "application1", Channel: "a", Status: "solving"},
				{Application: "application2", Channel: "a", Status: "solving"},
			},
			expected: []string{"a"},
		},
		{
			name: "no removed channels",
			before: []Update{
				{Application: "application", Channel: "a", Status: "solving"},
				{Application: "application", Channel: "b", Status: "solving"},
			},
			after: []Update{
				{Application: "application", Channel: "a", Status: "solving"},
				{Application: "application", Channel: "b", Status: "solving"},
			},
			expected: []string{},
		},
		{
			name: "no removed channels, status change",
			before: []Update{
				{Application: "application", Channel: "a", Status: "solving"},
			},
			after: []Update{
				{Application: "application", Channel: "a", Status: "complete"},
			},
			expected: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := ComputeRemovedChannels(test.before, test.after)
			assert.ElementsMatch(t, test.expected, actual)
		})
	}
}

func TestChannelMonitor_ComputeChangedIntegrations(t *testing.T) {
	tests := []struct {
		name            string
		before          []Update
		after           []Update
		expectedAdded   []Update
		expectedRemoved []Update
		expectedChanged map[Update]Update
	}{
		{
			name: "no channels",
		},
		{
			name: "no changes",
			before: []Update{
				{Application: "application", Channel: "a", Status: "solving"},
				{Application: "application", Channel: "b", Status: "solving"},
			},
			after: []Update{
				{Application: "application", Channel: "a", Status: "solving"},
				{Application: "application", Channel: "b", Status: "solving"},
			},
		},
		{
			name: "single integration added",
			after: []Update{
				{Application: "application", Channel: "a", Status: "solving"},
			},
			expectedAdded: []Update{
				{Application: "application", Channel: "a", Status: "solving"},
			},
		},
		{
			name: "single integration removed",
			before: []Update{
				{Application: "application", Channel: "a", Status: "solving"},
			},
			expectedRemoved: []Update{
				{Application: "application", Channel: "a", Status: "solving"},
			},
		},
		{
			name: "single integration, status change",
			before: []Update{
				{Application: "application", Channel: "a", Status: "solving"},
			},
			after: []Update{
				{Application: "application", Channel: "a", Status: "complete"},
			},
			expectedChanged: map[Update]Update{
				{Application: "application", Channel: "a", Status: "solving"}: {Application: "application", Channel: "a", Status: "complete"},
			},
		},
		{
			name: "multiple integrations added",
			after: []Update{
				{Application: "application", Channel: "a", Status: "complete"},
				{Application: "application", Channel: "b", Status: "complete"},
			},
			expectedAdded: []Update{
				{Application: "application", Channel: "a", Status: "complete"},
				{Application: "application", Channel: "b", Status: "complete"},
			},
		},
		{
			name: "multiple integrations removed",
			before: []Update{
				{Application: "application", Channel: "a", Status: "complete"},
				{Application: "application", Channel: "b", Status: "complete"},
			},
			expectedRemoved: []Update{
				{Application: "application", Channel: "a", Status: "complete"},
				{Application: "application", Channel: "b", Status: "complete"},
			},
		},
		{
			name: "multiple integrations, status change",
			before: []Update{
				{Application: "application", Channel: "a", Status: "solving"},
				{Application: "application", Channel: "b", Status: "solving"},
			},
			after: []Update{
				{Application: "application", Channel: "a", Status: "complete"},
				{Application: "application", Channel: "b", Status: "complete"},
			},
			expectedChanged: map[Update]Update{
				{Application: "application", Channel: "a", Status: "solving"}: {Application: "application", Channel: "a", Status: "complete"},
				{Application: "application", Channel: "b", Status: "solving"}: {Application: "application", Channel: "b", Status: "complete"},
			},
		},
		{
			name: "many changes",
			before: []Update{
				{Application: "application", Channel: "a", Status: "solving"},
				{Application: "application", Channel: "b", Status: "solving"},
				{Application: "application", Channel: "c", Status: "solving"},
			},
			after: []Update{
				{Application: "application", Channel: "b", Status: "solving"},
				{Application: "application", Channel: "c", Status: "complete"},
				{Application: "application", Channel: "d", Status: "solving"},
			},
			expectedAdded: []Update{
				{Application: "application", Channel: "d", Status: "solving"},
			},
			expectedRemoved: []Update{
				{Application: "application", Channel: "a", Status: "solving"},
			},
			expectedChanged: map[Update]Update{
				{Application: "application", Channel: "c", Status: "solving"}: {Application: "application", Channel: "c", Status: "complete"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			added, removed, changed := ComputeChangedIntegrations(test.before, test.after)
			assert.ElementsMatch(t, test.expectedAdded, added)
			assert.ElementsMatch(t, test.expectedRemoved, removed)
			MapEquals(t, test.expectedChanged, changed)
		})
	}
}

func TestChannelMonitor_Update(t *testing.T) {
	tests := []struct {
		name   string
		before []Update
		after  []Update
		verify func(*testing.T, *CallbackRecorder)
	}{
		{
			name: "no channels",
			verify: func(t *testing.T, r *CallbackRecorder) {
				assert.Empty(t, r.ChannelAdds)
				assert.Empty(t, r.ChannelRemoves)
				assert.Empty(t, r.IntegrationAdds)
				assert.Empty(t, r.IntegrationRemoves)
				assert.Empty(t, r.IntegrationUpdates)
			},
		},
		{
			name: "no changes",
			before: []Update{
				{Application: "application", Channel: "channel", Status: "solving"},
			},
			after: []Update{
				{Application: "application", Channel: "channel", Status: "solving"},
			},
			verify: func(t *testing.T, r *CallbackRecorder) {
				assert.Empty(t, r.ChannelAdds)
				assert.Empty(t, r.ChannelRemoves)
				assert.Empty(t, r.IntegrationAdds)
				assert.Empty(t, r.IntegrationRemoves)
				assert.Empty(t, r.IntegrationUpdates)
			},
		},
		{
			name: "added channel",
			after: []Update{
				{Application: "application", Channel: "channel", Status: "solving"},
			},
			verify: func(t *testing.T, r *CallbackRecorder) {
				assert.Equal(t, []string{"channel"}, r.ChannelAdds)
				assert.Empty(t, r.ChannelRemoves)
				assert.Equal(t, []Update{
					{Application: "application", Channel: "channel", Status: "solving"},
				}, r.IntegrationAdds)
				assert.Empty(t, r.IntegrationRemoves)
				assert.Empty(t, r.IntegrationUpdates)
			},
		},
		{
			name: "removed channel",
			before: []Update{
				{Application: "application", Channel: "channel", Status: "solving"},
			},
			verify: func(t *testing.T, r *CallbackRecorder) {
				assert.Empty(t, r.ChannelAdds)
				assert.Equal(t, []string{"channel"}, r.ChannelRemoves)
				assert.Empty(t, r.IntegrationAdds)
				assert.Equal(t, []Update{
					{Application: "application", Channel: "channel"},
				}, r.IntegrationRemoves)
				assert.Empty(t, r.IntegrationUpdates)
			},
		},
		{
			name: "changed integration",
			before: []Update{
				{Application: "application", Channel: "channel", Status: "solving"},
			},
			after: []Update{
				{Application: "application", Channel: "channel", Status: "complete"},
			},
			verify: func(t *testing.T, r *CallbackRecorder) {
				assert.Empty(t, r.ChannelAdds)
				assert.Empty(t, r.ChannelRemoves)
				assert.Empty(t, r.IntegrationAdds)
				assert.Empty(t, r.IntegrationRemoves)
				assert.Equal(t, map[Update]string{
					{Application: "application", Channel: "channel", Status: "solving"}: "complete",
				}, r.IntegrationUpdates)
			},
		},
		{
			name: "multiple changes",
			before: []Update{
				{Application: "application", Channel: "channel1", Status: "solving"},
				{Application: "application", Channel: "channel2", Status: "solving"},
			},
			after: []Update{
				{Application: "application", Channel: "channel2", Status: "complete"},
				{Application: "application", Channel: "channel3", Status: "solving"},
			},
			verify: func(t *testing.T, r *CallbackRecorder) {
				assert.Equal(t, []string{"channel3"}, r.ChannelAdds)
				assert.Equal(t, []string{"channel1"}, r.ChannelRemoves)
				assert.Equal(t, []Update{
					{Application: "application", Channel: "channel3", Status: "solving"},
				}, r.IntegrationAdds)
				assert.Equal(t, []Update{
					{Application: "application", Channel: "channel1"},
				}, r.IntegrationRemoves)
				assert.Equal(t, map[Update]string{
					{Application: "application", Channel: "channel2", Status: "solving"}: "complete",
				}, r.IntegrationUpdates)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := &CallbackRecorder{}
			monitor := ChannelMonitor{
				current:              test.before,
				OnChannelAdded:       recorder.OnChannelAdded,
				OnChannelRemoved:     recorder.OnChannelRemoved,
				OnIntegrationAdded:   recorder.OnIntegrationAdded,
				OnIntegrationRemoved: recorder.OnIntegrationRemoved,
				OnIntegrationUpdated: recorder.OnIntegrationUpdated,
			}
			monitor.Update(test.after)
			test.verify(t, recorder)
		})
	}
}

type CallbackRecorder struct {
	ChannelAdds        []string
	ChannelRemoves     []string
	IntegrationAdds    []Update
	IntegrationRemoves []Update
	IntegrationUpdates map[Update]string
}

func (r *CallbackRecorder) OnChannelAdded(channel string) {
	r.ChannelAdds = append(r.ChannelAdds, channel)
}
func (r *CallbackRecorder) OnChannelRemoved(channel string) {
	r.ChannelRemoves = append(r.ChannelRemoves, channel)
}
func (r *CallbackRecorder) OnIntegrationAdded(app ID, channel string, status string) {
	r.IntegrationAdds = append(r.IntegrationAdds, Update{
		Application: app,
		Channel:     channel,
		Status:      status,
	})
}
func (r *CallbackRecorder) OnIntegrationRemoved(app ID, channel string) {
	r.IntegrationRemoves = append(r.IntegrationRemoves, Update{
		Application: app,
		Channel:     channel,
	})
}
func (r *CallbackRecorder) OnIntegrationUpdated(app ID, channel string, oldStatus, newStatus string) {
	if r.IntegrationUpdates == nil {
		r.IntegrationUpdates = make(map[Update]string)
	}

	r.IntegrationUpdates[Update{
		Application: app,
		Channel:     channel,
		Status:      oldStatus,
	}] = newStatus
}

func MapEquals(t *testing.T, expected, actual map[Update]Update) {
	if expected == nil {
		expected = make(map[Update]Update)
	}
	if actual == nil {
		actual = make(map[Update]Update)
	}

	assert.Equal(t, expected, actual)
}
