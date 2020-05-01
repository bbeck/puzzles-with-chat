package main

import (
	"sync"
)

// A MessageRouter receives a message from a client and routes it to the
// appropriate MessageHandler instances.
type MessageRouter struct {
	sync.Mutex

	// The message handlers for each integration.
	handlers map[ID]MessageHandler

	// The active integrations for each channel.
	integrations map[string]map[ID]struct{}
}

func NewMessageRouter(integrations []Integration) *MessageRouter {
	handlers := make(map[ID]MessageHandler)
	for _, integration := range integrations {
		handlers[integration.ID] = integration.MessageHandler
	}

	return &MessageRouter{
		handlers:     handlers,
		integrations: make(map[string]map[ID]struct{}),
	}
}

// UpdateChannels notifies the message router of the current channel list as
// discovered by an integration.
func (r *MessageRouter) UpdateChannels(id ID, channels []string) {
	r.Lock()
	defer r.Unlock()

	// Remove all existing entries for this integration.
	for channel, m := range r.integrations {
		delete(m, id)

		// When we remove the last entry for a channel be sure to remove that
		// channel's entry so that this map doesn't grow forever.
		if len(m) == 0 {
			delete(r.integrations, channel)
		}
	}

	// And add it for each of the channels.
	for _, channel := range channels {
		m := r.integrations[channel]
		if m == nil {
			m = make(map[ID]struct{})
			r.integrations[channel] = m
		}

		m[id] = struct{}{}
	}
}

// HandleChannelMessage takes a message that was sent to a channel and passes
// it onto the handlers for the integrations that are active for the channel.
func (r *MessageRouter) HandleChannelMessage(channel, userid, username, message string) {
	r.Lock()
	defer r.Unlock()

	for id := range r.integrations[channel] {
		handler := r.handlers[id]
		if handler != nil {
			handler.HandleChannelMessage(channel, userid, username, message)
		}
	}
}
