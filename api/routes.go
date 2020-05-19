package main

import (
	"github.com/bbeck/puzzles-with-chat/api/crossword"
	"github.com/bbeck/puzzles-with-chat/api/model"
	"github.com/bbeck/puzzles-with-chat/api/pubsub"
	"github.com/bbeck/puzzles-with-chat/api/spellingbee"
	"github.com/go-chi/chi"
	"github.com/gomodule/redigo/redis"
	"log"
	"net/http"
	"testing"
	"time"
)

func RegisterRoutes(r chi.Router, pool *redis.Pool, registry *pubsub.Registry) {
	r.Get("/channels", GetChannels(pool, registry))
}

// GetChannels establishes a SSE based stream with a client that contains the
// list of active channels across all puzzle types.  Events will be periodically
// sent to the stream containing the list of active channels, even if the list
// doesn't change.
func GetChannels(pool *redis.Pool, registry *pubsub.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Construct the stream that all events for this particular client will be
		// placed into.
		stream := make(chan pubsub.Event, 10)
		defer close(stream)

		// Setup a connection to redis so that we can read the current solves.
		conn := pool.Get()
		defer func() { _ = conn.Close() }()

		// Setup a subscription in the registry to be able to see all state update
		// events regardless of which channel or puzzle type it's for.  This will
		// allow us to know when to look for an updated set of channels to send to
		// anyone monitoring the list of active channels.
		events := make(chan pubsub.Event, 10)
		defer close(events)

		id, err := registry.SubscribeMatching(func(channel pubsub.Channel, event pubsub.Event) bool {
			return event.Kind == "state"
		}, events)
		defer registry.Unsubscribe(id)

		// Get our initial set of channels so that we can send a response to the
		// client immediately.
		channels, err := GetActiveChannels(conn)
		if err != nil {
			log.Printf("unable to load active crossword channels: %+v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Send the initial set of channels.
		stream <- ChannelsEvent(channels)

		// Start a background goroutine for this client that sends updates to the
		// list of channels periodically.
		go func(channels map[string][]model.Channel) {
			// Use a new connection to redis since this goroutine might live slightly
			// longer than the GetChannels method call.  This ensures that we don't
			// attempt to use a connection that's already been closed.
			conn := pool.Get()
			defer func() { _ = conn.Close() }()

			for {
				select {
				case <-r.Context().Done():
					// The client disconnected, the goroutine should exit.
					return

				case <-events:
					// A channel has updated it's state.  Check if anything has changed.
					current, err := GetActiveChannels(conn)
					if err != nil {
						log.Printf("unable to load active crossword channels: %+v", err)

						// Don't exit the goroutine here since the client is still connected.
						// We'll just try again in the future.
						continue
					}

					if Changed(channels, current) {
						channels = current
						stream <- ChannelsEvent(channels)
					}

				case <-time.After(1 * time.Minute):
					// If we haven't sent any channels to the user in awhile then do so
					// now.  This handles the case where changes happen to the underlying
					// database out of band (such as a TTL expiring an entry) from our
					// application.
					channels, err = GetActiveChannels(conn)
					if err != nil {
						log.Printf("unable to load active crossword channels: %+v", err)

						// Don't exit the goroutine here since the client is still connected.
						// We'll just try again in the future.
						continue
					}

					stream <- ChannelsEvent(channels)
				}
			}
		}(channels)

		pubsub.EmitEvents(r.Context(), w, stream)
	}
}

// Changed compares two sets of active channels and determines if anything has
// changed or not.
func Changed(before, after map[string][]model.Channel) bool {
	if len(before) != len(after) {
		return true
	}

	same := func(as, bs []model.Channel) bool {
		if len(as) != len(bs) {
			return false
		}

		seen := make(map[string]model.Channel)
		for _, a := range as {
			seen[a.Name] = a
		}

		for _, b := range bs {
			if seen[b.Name] != b {
				return false
			}
		}

		return true
	}

	for k := range after {
		if !same(before[k], after[k]) {
			return true
		}
	}

	return false
}

// GetActiveChannels loads from the database all channels that have states and
// returns them indexed by the puzzle type that they belong to.  If for some
// reason a channel can't be loaded from the database then an error is returned.
func GetActiveChannels(conn redis.Conn) (map[string][]model.Channel, error) {
	if testActiveChannelsLoadError != nil {
		return nil, testActiveChannelsLoadError
	}

	crosswords, err := crossword.GetAllChannels(conn)
	if err != nil {
		return nil, err
	}

	spellingbees, err := spellingbee.GetAllChannels(conn)
	if err != nil {
		return nil, err
	}

	return map[string][]model.Channel{
		"crossword":   crosswords,
		"spellingbee": spellingbees,
	}, nil
}

var testActiveChannelsLoadError error

// ForceErrorDuringActiveChannelsLoad sets up an error to be returned when an
// attempt is made to load active channels.
func ForceErrorDuringActiveChannelsLoad(t *testing.T, err error) {
	t.Helper()

	testActiveChannelsLoadError = err
	t.Cleanup(func() { testActiveChannelsLoadError = nil })
}

func ChannelsEvent(channels map[string][]model.Channel) pubsub.Event {
	return pubsub.Event{
		Kind:    "channels",
		Payload: channels,
	}
}
