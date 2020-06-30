package main

import (
	"sync"
)

// ChannelMonitor gets notified about current puzzles being solved and
// determines which channels have been added, removed or updated and notifies
// callbacks of the changes.
type ChannelMonitor struct {
	sync.Mutex

	// the set of channels being monitored on a per-integration basis with the
	// current status of each channel
	current []Update

	// callback to call when a channel has been added to the monitored list
	OnChannelAdded func(channel string)

	// callback to call when a channel has been removed from the monitored list
	OnChannelRemoved func(channel string)

	// callback to call when a channel's integration has been added
	OnIntegrationAdded func(app ID, channel string, status string)

	// callback to call when a channel's integration has been removed
	OnIntegrationRemoved func(app ID, channel string)

	// callback to call when a channel's integration has been updated
	OnIntegrationUpdated func(app ID, channel string, oldStatus, newStatus string)
}

// Update records the updated set of channels from the channel locator and
// calls into the appropriate callbacks based on changes from the last seen
// set of channels.
func (m *ChannelMonitor) Update(updates []Update) {
	// Lock so that we have a consistent picture of the world.
	m.Lock()
	defer m.Unlock()

	// Look for channels that weren't present previously
	added := ComputeAddedChannels(m.current, updates)
	for _, channel := range added {
		m.OnChannelAdded(channel)
	}

	// Look for channels that are no longer present
	removed := ComputeRemovedChannels(m.current, updates)
	for _, channel := range removed {
		m.OnChannelRemoved(channel)
	}

	// Determine which integrations have been added, removed, or changed
	intAdds, intRemoves, intUpdates := ComputeChangedIntegrations(m.current, updates)
	for _, add := range intAdds {
		m.OnIntegrationAdded(add.Application, add.Channel, add.Status)
	}
	for _, remove := range intRemoves {
		m.OnIntegrationRemoved(remove.Application, remove.Channel)
	}
	for before, after := range intUpdates {
		m.OnIntegrationUpdated(before.Application, before.Channel, before.Status, after.Status)
	}

	// Save the current set of updates to compare against next time.
	m.current = updates
}

// ComputeAddedChannels determines which channels have been added where there
// was previously no integration present.
func ComputeAddedChannels(before, after []Update) []string {
	added := make(map[string]struct{})
	for _, update := range after {
		var found bool
		for _, current := range before {
			if update.Channel == current.Channel {
				found = true
				break
			}
		}

		if !found {
			added[update.Channel] = struct{}{}
		}
	}

	channels := make([]string, 0)
	for channel := range added {
		channels = append(channels, channel)
	}

	return channels
}

// ComputeRemovedChannels determines which channels have been removed where
// there was previously an integration present.
func ComputeRemovedChannels(before, after []Update) []string {
	removed := make(map[string]struct{})
	for _, current := range before {
		var found bool
		for _, update := range after {
			if update.Channel == current.Channel {
				found = true
				break
			}
		}

		if !found {
			removed[current.Channel] = struct{}{}
		}
	}

	channels := make([]string, 0)
	for channel := range removed {
		channels = append(channels, channel)
	}

	return channels
}

// ComputeChangedIntegrations determines which application integrations for
// channels have been changed.  A change can be either an integration being
// added or removed, or the status of a solve changing.
func ComputeChangedIntegrations(before, after []Update) ([]Update, []Update, map[Update]Update) {
	// Find added integrations.  These are updates in the after set that don't
	// have a corresponding update in the before set.
	var added []Update
	for _, a := range after {
		var found bool
		for _, b := range before {
			if a.Application == b.Application && a.Channel == b.Channel {
				found = true
				break
			}
		}

		if !found {
			added = append(added, a)
		}
	}

	// Find removed integrations.  These are updates in the before set that don't
	// have a corresponding update in the after set.
	var removed []Update
	for _, b := range before {
		var found bool
		for _, a := range after {
			if a.Application == b.Application && a.Channel == b.Channel {
				found = true
				break
			}
		}

		if !found {
			removed = append(removed, b)
		}
	}

	// Find changed integrations.  These are updates in the before and after set
	// that have a changed status.
	changed := make(map[Update]Update)
	for _, b := range before {
		for _, a := range after {
			if a.Application == b.Application && a.Channel == b.Channel && a.Status != b.Status {
				changed[b] = a
			}
		}
	}

	return added, removed, changed
}
