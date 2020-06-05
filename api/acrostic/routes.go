package acrostic

import (
	"fmt"
	"github.com/bbeck/puzzles-with-chat/api/model"
	"github.com/bbeck/puzzles-with-chat/api/pubsub"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/gomodule/redigo/redis"
	"log"
	"net/http"
)

func RegisterRoutes(r chi.Router, pool *redis.Pool, registry *pubsub.Registry) {
	r.Route("/acrostic/{channel}", func(r chi.Router) {
		r.Put("/", UpdatePuzzle(pool, registry))
	})
}

// UpdatePuzzle changes the acrostic puzzle that's currently being solved for a
// channel.
func UpdatePuzzle(pool *redis.Pool, registry *pubsub.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		channel := chi.URLParam(r, "channel")

		// Since there could be multiple ways to specify which acrostic to solve
		// we'll parse the payload into a generic map instead of a specific object.
		var payload map[string]string
		if err := render.DecodeJSON(r.Body, &payload); err != nil {
			log.Printf("unable to read request body: %+v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var puzzle *Puzzle

		// New York Times date
		if date := payload["new_york_times_date"]; date != "" {
			p, err := LoadFromNewYorkTimes(date)
			if err != nil {
				log.Printf("unable to load NYT acrostic for date %s: %+v", date, err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			puzzle = p
		}

		if puzzle == nil {
			log.Printf("unable to determine acrostic from payload: %+v", payload)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		conn := pool.Get()
		defer func() { _ = conn.Close() }()

		// Save the puzzle to this channel's state
		cells := make([][]string, puzzle.Rows)
		for row := 0; row < puzzle.Rows; row++ {
			cells[row] = make([]string, puzzle.Cols)
		}

		state := State{
			Status:      model.StatusSelected,
			Puzzle:      puzzle,
			Cells:       cells,
			CluesFilled: make(map[string]bool),
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
		state.Puzzle = state.Puzzle.WithoutSolution()

		registry.Publish(ChannelID(channel), StateEvent(state))

		w.WriteHeader(http.StatusOK)
	}
}

func ChannelID(channel string) pubsub.Channel {
	channel = fmt.Sprintf("%s:acrostic", channel)
	return pubsub.Channel(channel)
}

func StateEvent(state State) pubsub.Event {
	return pubsub.Event{
		Kind:    "state",
		Payload: state,
	}
}
