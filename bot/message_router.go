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

	// A mapping of active integrations on a per channel basis.
	channels map[string]map[ID]struct{}
}

func NewMessageRouter(handlers map[ID]MessageHandler) *MessageRouter {
	return &MessageRouter{
		handlers: handlers,
		channels: make(map[string]map[ID]struct{}),
	}
}

// UpdateChannels updates the active integrations of the router.  This will
// cause messages to start/stop being delivered to message handlers.
func (r *MessageRouter) UpdateChannels(update map[ID][]string) {
	r.Lock()
	defer r.Unlock()

	// Index the update by channel instead of by integration.
	r.channels = make(map[string]map[ID]struct{})
	for id, channels := range update {
		for _, channel := range channels {
			m := r.channels[channel]
			if m == nil {
				m = make(map[ID]struct{})
				r.channels[channel] = m
			}

			m[id] = struct{}{}
		}
	}
}

// HandleChannelMessage takes a message that was sent to a channel and passes
// it onto the handlers for the integrations that are active for the channel.
func (r *MessageRouter) HandleChannelMessage(channel, userid, username, message string) {
	r.Lock()
	defer r.Unlock()

	for id := range r.channels[channel] {
		handler := r.handlers[id]
		if handler != nil {
			handler.HandleChannelMessage(channel, userid, username, message)
		}
	}
}
