package main

import (
	"sync"
)

type ChannelManager struct {
	sync.Mutex

	// the current set of channels being monitored
	channels map[string]struct{}

	// helper to call when a new channel has been added to the monitored list
	OnAddChannel func(channel string)

	// helper to call when a channel has been removed from the monitored list
	OnRemoveChannel func(channel string)
}

func (m *ChannelManager) Update(channels []string) {
	// Lock so that we have a consistent picture of the world.
	m.Lock()
	defer m.Unlock()

	if m.channels == nil {
		m.channels = make(map[string]struct{})
	}

	added, removed := m.Diff(channels)
	for _, channel := range added {
		if m.OnAddChannel != nil {
			m.OnAddChannel(channel)
		}
		m.channels[channel] = struct{}{}
	}
	for _, channel := range removed {
		if m.OnRemoveChannel != nil {
			m.OnRemoveChannel(channel)
		}
		delete(m.channels, channel)
	}
}

// Diff determines which channels are new and which are removed.
func (m *ChannelManager) Diff(channels []string) ([]string, []string) {
	var added []string
	seen := make(map[string]struct{})
	for _, channel := range channels {
		if _, ok := m.channels[channel]; !ok {
			added = append(added, channel)
		}

		seen[channel] = struct{}{}
	}

	var removed []string
	for channel := range m.channels {
		if _, ok := seen[channel]; !ok {
			removed = append(removed, channel)
		}
	}

	return added, removed
}
