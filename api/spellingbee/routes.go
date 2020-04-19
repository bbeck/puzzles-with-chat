package spellingbee

import (
	"github.com/bbeck/twitch-plays-crosswords/api/model"
	"github.com/bbeck/twitch-plays-crosswords/api/pubsub"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/gomodule/redigo/redis"
	"log"
	"net/http"
	"time"
)

func RegisterRoutes(router chi.Router, pool *redis.Pool) {
	RegisterRoutesWithRegistry(router, pool, new(pubsub.Registry))
}

func RegisterRoutesWithRegistry(r chi.Router, pool *redis.Pool, registry *pubsub.Registry) {
	r.Get("/spellingbee/channels", GetChannels(pool))

	r.Route("/spellingbee/{channel}", func(r chi.Router) {
		r.Put("/", UpdatePuzzle(pool, registry))
	})
}

// GetChannels establishes a SSE based stream with a client that contains the
// list of active spelling bee channels.  Events will be periodically sent to
// the stream containing the list of active channels, even if the list doesn't
// change.
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
		channels, err := GetChannelNamesWithState(conn)
		if err != nil {
			log.Printf("unable to load spelling bee channels: %+v", err)
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
					channels, err := GetChannelNamesWithState(conn)
					if err != nil {
						log.Printf("unable to load spelling bee channels: %+v", err)

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

// UpdatePuzzle changes the spelling bee puzzle that's currently being solved
// for a channel.
func UpdatePuzzle(pool *redis.Pool, registry *pubsub.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		channel := chi.URLParam(r, "channel")

		// Since there are multiple ways to specify which crossword to solve we'll
		// parse the payload into a generic map instead of a specific object.
		var payload map[string]string
		if err := render.DecodeJSON(r.Body, &payload); err != nil {
			log.Printf("unable to read request body: %+v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var puzzle *Puzzle

		// New York Times date
		if date := payload["nytbee"]; date != "" {
			p, err := LoadFromNYTBee(date)
			if err != nil {
				log.Printf("unable to load NYTBee puzzle for date %s: %+v", date, err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			puzzle = p
		}

		if puzzle == nil {
			log.Printf("unable to determine puzzle from payload: %+v", payload)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		conn := pool.Get()
		defer func() { _ = conn.Close() }()

		// Save the puzzle to this channel's state
		state := State{
			Status: model.StatusSelected,
			Puzzle: puzzle,
			Words:  make([]string, 0),
		}
		if err := SetState(conn, channel, state); err != nil {
			log.Printf("unable to save state for channel %s: %+v", channel, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Broadcast to all of the clients that the puzzle has been selected, making
		// sure to not include the answers.  It's okay to overwrite the puzzle
		// attribute because we just wrote this state instance to the database
		// and will be discarding it immediately publishing.
		state.Puzzle = state.Puzzle.WithoutAnswers()

		registry.Publish(pubsub.Channel(channel), pubsub.Event{
			Kind:    "state",
			Payload: state,
		})
	}
}
