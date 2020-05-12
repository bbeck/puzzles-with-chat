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
	Payload interface{} `json:"payload,omitempty"`
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
	functions map[ClientID]func(Channel, Event) bool
	streams   map[ClientID]chan<- Event
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

	fn := func(c Channel, e Event) bool {
		return c == channel
	}

	return r.SubscribeMatching(fn, stream)
}

// SubscribeMatching adds a new client stream for all events published that are
// for a channel that matches a provided function.  The provided stream will be
// associated with the subscription and begin to receive all matching events
// once the subscribe call returns.  A unique client identifier will be returned
// so that in the future the client can be unsubscribed to stop events from
// being received.
//
// If desired, the provided stream can be used externally to target events for
// a single client.  It can also be passed into the subscribe method with events
// already in it so that a set of initialization events can be sent to the
// client before any published events.
//
// NOTE: The passed in stream should not be closed prior to the client being
// unsubscribed from the registry.
func (r *Registry) SubscribeMatching(fn func(Channel, Event) bool, stream chan<- Event) (ClientID, error) {
	if fn == nil {
		return "", errors.New("empty channel function")
	}

	if stream == nil {
		return "", errors.New("nil event stream")
	}

	if cap(stream) == 0 {
		return "", errors.New("event stream must have a non-zero capacity")
	}

	// Generate a new id for the client, do this before acquiring the lock to
	// minimize the amount of time that we hold the lock.
	id := ClientID(xid.New().String())

	r.Lock()
	defer r.Unlock()

	if r.functions == nil {
		r.functions = make(map[ClientID]func(Channel, Event) bool)
	}
	r.functions[id] = fn

	if r.streams == nil {
		r.streams = make(map[ClientID]chan<- Event)
	}
	r.streams[id] = stream

	return id, nil
}

// Unsubscribe removes a client from a particular channel.  Once this method
// returns no further events will be received on that client's stream.
func (r *Registry) Unsubscribe(id ClientID) {
	r.Lock()
	defer r.Unlock()

	delete(r.functions, id)
	delete(r.streams, id)
}

// Publish sends an event to all subscribed clients of a given channel.  If a
// client's stream is full the event will be skipped.
func (r *Registry) Publish(channel Channel, event Event) {
	r.Lock()
	defer r.Unlock()

	for id, fn := range r.functions {
		if fn(channel, event) {
			stream := r.streams[id]

			// Perform a non-blocking send to the stream so that we can detect the
			// situation where a client has a full stream and isn't draining events from
			// it.  When this happens we'll continue sending to the remaining clients.
			select {
			case stream <- event: // success
			default: // failure
			}
		}
	}
}
