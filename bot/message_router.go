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

	// The status of each channel's integrations.
	statuses map[string]map[ID]string
}

func NewMessageRouter(handlers map[ID]MessageHandler) *MessageRouter {
	return &MessageRouter{handlers: handlers}
}

// AddIntegration updates the integration status for the provided channel.
func (r *MessageRouter) AddIntegration(app ID, channel string, status string) {
	r.Lock()
	defer r.Unlock()

	r.ensure(channel)
	r.statuses[channel][app] = status
}

// RemoveIntegration removes the specified integration for the provided channel.
func (r *MessageRouter) RemoveIntegration(app ID, channel string) {
	r.Lock()
	defer r.Unlock()

	r.ensure(channel)
	delete(r.statuses[channel], app)

	if len(r.statuses[channel]) == 0 {
		delete(r.statuses, channel)
	}
}

// UpdateChannel updates the active integrations for a channel.  This will
// cause messages to start/stop being delivered to message handlers.
func (r *MessageRouter) UpdateIntegrationStatus(app ID, channel string, status string) {
	r.Lock()
	defer r.Unlock()

	r.ensure(channel)
	r.statuses[channel][app] = status
}

// HandleChannelMessage takes a message that was sent to a channel and passes
// it onto the handlers for the integrations that are active for the channel.
func (r *MessageRouter) HandleChannelMessage(channel, _, _, message string) {
	r.Lock()
	defer r.Unlock()

	r.ensure(channel)
	for app, status := range r.statuses[channel] {
		handler := r.handlers[app]
		if handler != nil {
			handler.HandleChannelMessage(channel, status, message)
		}
	}
}

func (r *MessageRouter) ensure(channel string) {
	if r.statuses == nil {
		r.statuses = make(map[string]map[ID]string)
	}

	if r.statuses[channel] == nil {
		r.statuses[channel] = make(map[ID]string)
	}
}
