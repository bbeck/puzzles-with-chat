package pubsub

import (
	"errors"
	"github.com/rs/xid"
	"sync"
)

// Event encapsulates an event that can be sent to all subscribed clients of a
// registry.
type Event struct {
	Kind    string      `json:"kind"`
	Payload interface{} `json:"payload"`
}

// Channel represents the segment of clients that a subscription is for or that
// an event should be delivered to.
type Channel string

// ClientID represents the identifier of a client that has subscribed to the
// registry.
type ClientID string

// Registry manages the event streams for subscribed clients across different
// channels.  Internally the registry keeps track of each subscribed client for
// a channel using an identifier.  This identifier can be used by the client to
// unsubscribe to stop receiving any future events.  The registry is safe to
// access from multiple goroutines.
type Registry struct {
	sync.Mutex
	streams map[Channel]map[ClientID]chan<- Event
}

// Subscribe adds a new client stream for a particular channel.  The provided
// stream will be associated with the channel and begin to receive all events
// for that channel.  A unique client identifier will be returned so that in the
// future the client can be unsubscribed to stop events from being received.
//
// If desired, the provided stream can be used externally to target events for
// a single client.  It can also be passed into the subscribe method with events
// already in it so that a set of initialization events can be sent to the
// client before any published events.
//
// NOTE: The passed in stream should not be closed prior to the client being
// unsubscribed from the registry.
func (r *Registry) Subscribe(channel Channel, stream chan<- Event) (ClientID, error) {
	if channel == "" {
		return "", errors.New("empty channel")
	}

	if stream == nil {
		return "", errors.New("nil event stream")
	}

	if cap(stream) == 0 {
		return "", errors.New("event stream must have a non-zero capacity")
	}

	r.Lock()
	defer r.Unlock()

	if r.streams == nil {
		r.streams = make(map[Channel]map[ClientID]chan<- Event)
	}

	channelStreams := r.streams[channel]
	if channelStreams == nil {
		channelStreams = make(map[ClientID]chan<- Event)
		r.streams[channel] = channelStreams
	}

	// Generate a new id for the client
	id := ClientID(xid.New().String())
	channelStreams[id] = stream

	return id, nil
}

func (r *Registry) Unsubscribe(channel Channel, id ClientID) {
	r.Lock()
	defer r.Unlock()

	// We haven't had any clients subscribe yet, nothing to do
	if r.streams == nil {
		return
	}

	delete(r.streams[channel], id)
}

func (r *Registry) Publish(channel Channel, event Event) {
	r.Lock()
	defer r.Unlock()

	// We haven't had any clients subscribe yet, nothing to do
	if r.streams == nil {
		return
	}

	for _, stream := range r.streams[channel] {
		// Perform a non-blocking send to the stream so that we can detect the
		// situation where a client has a full stream and isn't draining events from
		// it.  When this happens we'll continue sending to the remaining clients.
		select {
		case stream <- event: // success
		default: // failure
			// TODO: Return an error of some sort here?
		}
	}
}
