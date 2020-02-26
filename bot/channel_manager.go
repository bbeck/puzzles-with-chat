package main

import (
	"sync"
	"time"
)

type ChannelManager struct {
	sync.Mutex

	// sent true when the manager has been asked to stop
	done chan bool

	// the current set of channels being monitored
	channels map[string]struct{}

	// helper to call when a new channel has been added to the monitored list
	OnAddChannel func(channel string)

	// helper to call when a channel has been removed from the monitored list
	OnRemoveChannel func(channel string)

	// helper to call when an error happens during polling
	OnError func(err error)

	// called to poll for the channels to be monitored
	Poll func() ([]string, error)
}

// NewChannelManager creates a new ChannelManager instance that polls the
// provided endpoint.
func NewChannelManager(poll func() ([]string, error)) *ChannelManager {
	return &ChannelManager{
		done:     make(chan bool, 1),
		channels: make(map[string]struct{}),
		Poll:     poll,
	}
}

// Run causes a ChannelManager to start polling in the background for the
// channels that should be monitored.
func (m *ChannelManager) Run(duration time.Duration) {
	// TODO: Instead of polling maybe have a SSE channel and subscribe to it?
	go func() {
		ticker := time.NewTicker(duration)

		for {
			select {
			case <-m.done:
				ticker.Stop()
				return

			case <-ticker.C:
				channels, err := m.Poll()
				if err != nil {
					if m.OnError != nil {
						m.OnError(err)
					}
					break
				}

				// Lock so that we have a consistent picture of the world.
				m.Lock()

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

				m.Unlock()
			}
		}
	}()
}

// Stop causes a ChannelManager to stop polling for the channels that should be
// monitored.  Stop should only be called a single time after Run has been
// called.
func (m *ChannelManager) Stop() {
	m.Lock()
	defer m.Unlock()

	m.done <- true
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
