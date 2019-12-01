package main

import (
	"errors"
	"sync"

	"github.com/rs/xid"
)

// Event encapsulates an event that can be sent from the web application to a
// connected client.
type Event struct {
	Kind    string      `json:"kind"`
	Payload interface{} `json:"payload"`
}

// ClientRegistry manages the event streams for clients that are connected to
// the web application.  Internally the registry keeps track of which channels
// each client is connected to and provides a mechanism to send events to each
// connected client of a channel.
type ClientRegistry struct {
	sync.Mutex
	streams map[string]map[string]chan<- Event
}

// RegisterClient adds a new client stream for a particular channel.  The
// registered stream will be associated with the channel and begin to receive
// all events for that channel.  A unique identifier for this client will be
// returned so that in the future when the client disconnects from the server
// the DeregisterClient method can be called to remove their stream from the
// registry.
//
// NOTE: The passed in stream should not be closed prior to the client being
// deregistered from the registry.
func (r *ClientRegistry) RegisterClient(channel string, stream chan<- Event) (string, error) {
	if channel == "" {
		return "", errors.New("empty channel name")
	}

	if stream == nil {
		return "", errors.New("nil event stream")
	}

	if cap(stream) == 0 {
		return "", errors.New("event stream must be a buffered channel")
	}

	r.Lock()
	defer r.Unlock()

	if r.streams == nil {
		r.streams = make(map[string]map[string]chan<- Event)
	}

	channelStreams := r.streams[channel]
	if channelStreams == nil {
		channelStreams = make(map[string]chan<- Event)
		r.streams[channel] = channelStreams
	}

	// Register the stream and return the id to the caller.
	id := xid.New().String()
	channelStreams[id] = stream

	return id, nil
}

// DeregisterClient removes a client from a particular channel.  This ensures
// that events will not be delivered to that client in the future.
func (r *ClientRegistry) DeregisterClient(channel string, id string) {
	r.Lock()
	defer r.Unlock()

	if r.streams == nil {
		return
	}

	delete(r.streams[channel], id)
}

// BroadcastEvent will send an event to all clients of a particular channel.
func (r *ClientRegistry) BroadcastEvent(channel string, event Event) {
	r.Lock()
	defer r.Unlock()

	if r.streams == nil {
		return
	}

	for _, stream := range r.streams[channel] {
		// Perform a non-blocking send to the stream.  The send might fail if the
		// stream's buffer is full.  In that case we'll continue sending to the
		// remaining clients.
		select {
		case stream <- event:
		default:
			// TODO: Return an error of some sort here.
		}
	}
}
