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

func RegisterRoutes(r chi.Router, pool *redis.Pool) {
	r.Get("/channels", GetChannels(pool))
}

// GetChannels establishes a SSE based stream with a client that contains the
// list of active channels across all puzzle types.  Events will be periodically
// sent to the stream containing the list of active channels, even if the list
// doesn't change.
func GetChannels(pool *redis.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Construct the stream that all events for this particular client will be
		// placed into.
		stream := make(chan pubsub.Event, 10)
		defer close(stream)

		// Setup a connection to redis so that we can read the current solves.
		conn := pool.Get()
		defer func() { _ = conn.Close() }()

		// Get our initial set of channels so that we can send a response to the
		// client immediately.
		channels, err := GetActiveChannels(conn)
		if err != nil {
			log.Printf("unable to load active crossword channels: %+v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Send the initial set of channels.
		stream <- pubsub.Event{
			Kind:    "channels",
			Payload: channels,
		}

		// Start a background goroutine for this client that sends updates to the
		// list of channels periodically.
		go func() {
			for {
				select {
				case <-r.Context().Done():
					// The client disconnected, the goroutine should exit.
					return

				case <-time.After(10 * time.Second):
					channels, err := GetActiveChannels(conn)
					if err != nil {
						log.Printf("unable to load active crossword channels: %+v", err)

						// Don't exit the goroutine here since the client is still connected.
						// We'll just try again in the future.
						continue
					}

					stream <- pubsub.Event{
						Kind:    "channels",
						Payload: channels,
					}
				}
			}
		}()

		pubsub.EmitEvents(r.Context(), w, stream)
	}
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
