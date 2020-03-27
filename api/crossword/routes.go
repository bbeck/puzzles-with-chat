package crossword

import (
	"encoding/json"
	"fmt"
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
	r.Get("/crossword/events", GetActiveCrosswordsEvents(pool))

	r.Route("/crossword/{channel}", func(r chi.Router) {
		r.Put("/", UpdateCrossword(pool, registry))
		r.Put("/setting/{setting}", UpdateCrosswordSetting(pool, registry))
		r.Put("/status", ToggleCrosswordStatus(pool, registry))
		r.Put("/answer/{clue}", UpdateCrosswordAnswer(pool, registry))
		r.Get("/show/{clue}", ShowCrosswordClue(registry))
		r.Get("/events", GetCrosswordEvents(pool, registry))
	})
}

func GetActiveCrosswordsEvents(pool *redis.Pool) http.HandlerFunc {
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
			log.Printf("unable to load channels with active crossword solves: %+v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if channels == nil {
			channels = []string{}
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
						log.Printf("unable to load channels with active crossword solves: %+v", err)

						// Don't exit the goroutine here since the client is still connected.
						// We'll just try again in the future.
						continue
					}

					if channels == nil {
						channels = []string{}
					}

					stream <- pubsub.Event{
						Kind:    "channels",
						Payload: channels,
					}
				}
			}
		}()

		EmitEvents(w, r, stream)
	}
}

// UpdateCrossword changes the crossword that's currently being solved for a
// channel.
func UpdateCrossword(pool *redis.Pool, registry *pubsub.Registry) http.HandlerFunc {
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
		if date := payload["new_york_times_date"]; date != "" {
			p, err := LoadFromNewYorkTimes(date)
			if err != nil {
				log.Printf("unable to load NYT puzzle for date %s: %+v", date, err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			puzzle = p
		}

		// Wall Street Journal date
		if date := payload["wall_street_journal_date"]; date != "" {
			p, err := LoadFromWallStreetJournal(date)
			if err != nil {
				log.Printf("unable to load WSJ puzzle for date %s: %+v", date, err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			puzzle = p
		}

		// .puz file upload
		if encoded := payload["puz_file_bytes"]; encoded != "" {
			p, err := LoadFromEncodedPuzFile(encoded)
			if err != nil {
				log.Printf("unable to load puzzle from bytes: %+v", err)
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
		cells := make([][]string, puzzle.Rows)
		for row := 0; row < puzzle.Rows; row++ {
			cells[row] = make([]string, puzzle.Cols)
		}

		state := &State{
			Status:            StatusCreated,
			Puzzle:            puzzle,
			Cells:             cells,
			AcrossCluesFilled: make(map[int]bool),
			DownCluesFilled:   make(map[int]bool),
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

		registry.Publish(pubsub.Channel(channel), pubsub.Event{
			Kind:    "state",
			Payload: state,
		})
	}
}

func UpdateCrosswordSetting(pool *redis.Pool, registry *pubsub.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		channel := chi.URLParam(r, "channel")
		setting := chi.URLParam(r, "setting")

		conn := pool.Get()
		defer func() { _ = conn.Close() }()

		// Load the existing settings for the channel so that we can apply the
		// updates to them.
		settings, err := GetSettings(conn, channel)
		if err != nil {
			log.Printf("unable to read crossword settings for channel %s: %+v", channel, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Apply the update to the settings in memory.
		var shouldClearIncorrectCells bool
		switch setting {
		case "only_allow_correct_answers":
			var value bool
			if err := render.DecodeJSON(r.Body, &value); err != nil {
				log.Printf("unable to parse crossword only correct answers setting json %v: %+v", value, err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			settings.OnlyAllowCorrectAnswers = value
			shouldClearIncorrectCells = value

		case "clues_to_show":
			var value ClueVisibility
			if err := render.DecodeJSON(r.Body, &value); err != nil {
				log.Printf("unable to parse crossword clue visibility setting json %s: %+v", value, err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			settings.CluesToShow = value

		case "clue_font_size":
			var value FontSize
			if err := render.DecodeJSON(r.Body, &value); err != nil {
				log.Printf("unable to parse crossword clue font size setting json %s: %+v", value, err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			settings.ClueFontSize = value

		case "show_notes":
			var value bool
			if err := render.DecodeJSON(r.Body, &value); err != nil {
				log.Printf("unable to parse crossword show notes setting json %v: %+v", value, err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			settings.ShowNotes = value

		default:
			log.Printf("unrecognized crossword setting name %s", setting)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Save the settings back to the database.
		if err = SetSettings(conn, channel, settings); err != nil {
			log.Printf("unable to save crossword settings for channel %s: %+v", channel, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Load the state and clear any incorrect cells if we changed a setting that
		// requires this.  We do this after the setting is applied so that if there
		// was an error earlier we don't modify the solve's state.
		var updatedState *State
		if shouldClearIncorrectCells {
			state, err := GetState(conn, channel)
			if err != nil {
				log.Printf("unable to load state for channel %s: %+v", channel, err)
				w.WriteHeader(http.StatusNotFound)
				return
			}

			// There's no need to update cells if the puzzle hasn't been started yet
			// or is already complete.
			if state.Status != StatusCreated && state.Status != StatusComplete {
				if err := state.ClearIncorrectCells(); err != nil {
					log.Printf("unable to clear incorrect cells for channel: %s: %+v", channel, err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				if err := SetState(conn, channel, state); err != nil {
					log.Printf("unable to save state for channel %s: %+v", channel, err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				updatedState = state
			}
		}

		// Now broadcast the new settings to all of the clients in the channel.
		registry.Publish(pubsub.Channel(channel), pubsub.Event{
			Kind:    "settings",
			Payload: settings,
		})

		if updatedState != nil {
			// Broadcast the updated state to all of the clients, making sure to not
			// include the answers.
			updatedState.Puzzle = updatedState.Puzzle.WithoutSolution()

			registry.Publish(pubsub.Channel(channel), pubsub.Event{
				Kind:    "state",
				Payload: updatedState,
			})
		}
	}
}

func ToggleCrosswordStatus(pool *redis.Pool, registry *pubsub.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		channel := chi.URLParam(r, "channel")

		conn := pool.Get()
		defer func() { _ = conn.Close() }()

		state, err := GetState(conn, channel)
		if err != nil {
			log.Printf("unable to load state for channel %s: %+v", channel, err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		now := time.Now()

		switch state.Status {
		case StatusCreated:
			state.Status = StatusSolving
			state.LastStartTime = &now

		case StatusPaused:
			state.Status = StatusSolving
			state.LastStartTime = &now

		case StatusSolving:
			state.Status = StatusPaused
			total := state.TotalSolveDuration.Nanoseconds() + now.Sub(*state.LastStartTime).Nanoseconds()
			state.LastStartTime = nil
			state.TotalSolveDuration = Duration{time.Duration(total)}

		case StatusComplete:
			// The puzzle is already solved, we can't toggle its state anymore
			log.Printf("unable to toggle status for channel %s, puzzle is already solved", channel)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := SetState(conn, channel, state); err != nil {
			log.Printf("unable to save state for channel %s: %+v", channel, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Broadcast to all of the clients that the puzzle status has been changed,
		// making sure to not include the answers.  It's okay to overwrite the
		// puzzle attribute because we just wrote this state instance to the
		// database and will be discarding it immediately publishing.
		state.Puzzle = state.Puzzle.WithoutSolution()

		registry.Publish(pubsub.Channel(channel), pubsub.Event{
			Kind:    "state",
			Payload: state,
		})
	}
}

func UpdateCrosswordAnswer(pool *redis.Pool, registry *pubsub.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		channel := chi.URLParam(r, "channel")
		clue := chi.URLParam(r, "clue")

		if r.ContentLength > 1024 {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}

		var answer string
		if err := render.DecodeJSON(r.Body, &answer); err != nil {
			log.Printf("unable to read request body: %+v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if len(answer) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		conn := pool.Get()
		defer func() { _ = conn.Close() }()

		state, err := GetState(conn, channel)
		if err != nil {
			log.Printf("unable to load state for channel %s: %+v", channel, err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if state.Status != StatusSolving {
			w.WriteHeader(http.StatusConflict)
			return
		}

		settings, err := GetSettings(conn, channel)
		if err != nil {
			log.Printf("unable to load settings for channel %s: %+v", channel, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := state.ApplyAnswer(clue, answer, settings.OnlyAllowCorrectAnswers); err != nil {
			log.Printf("unable to apply answer %s for clue %s for channel %s: %+v", answer, clue, channel, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// If we just solved the puzzle then we should stop the timer.
		if state.Status == StatusComplete {
			now := time.Now()
			total := state.TotalSolveDuration.Nanoseconds() + now.Sub(*state.LastStartTime).Nanoseconds()
			state.LastStartTime = nil
			state.TotalSolveDuration = Duration{time.Duration(total)}
		}

		// Save the updated state.
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

		registry.Publish(pubsub.Channel(channel), pubsub.Event{
			Kind:    "state",
			Payload: state,
		})

		// If we've just finished the solve then send a complete event as well.
		if state.Status == StatusComplete {
			registry.Publish(pubsub.Channel(channel), pubsub.Event{
				Kind:    "complete",
				Payload: nil,
			})
		}
	}
}

// ShowCrosswordClue sends an event to all clients of a channel requesting that
// they update their view to make the specified clue visible.  If the specified
// clue isn't structured as a proper clue number and direction than an error
// will be returned.
func ShowCrosswordClue(registry *pubsub.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		channel := chi.URLParam(r, "channel")
		clue := chi.URLParam(r, "clue")

		_, _, err := ParseClue(clue)
		if err != nil {
			log.Printf("malformed clue (%s): %+v", clue, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		registry.Publish(pubsub.Channel(channel), pubsub.Event{
			Kind:    "show_clue",
			Payload: clue,
		})
	}
}

// GetCrosswordEvents establishes an event stream with a client.  An event
// stream is server side event stream (SSE) with a client's browser that allows
// one way communication from the server to the client.  Clients that call into
// this handler will keep an open connection open to the server waiting to
// receive events as JSON objects.  The server can send events to all clients
// of a channel using the pubsub.Registry's Publish method.
func GetCrosswordEvents(pool *redis.Pool, registry *pubsub.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		channel := chi.URLParam(r, "channel")

		// Construct the stream that all events for this particular client will be
		// placed into.
		stream := make(chan pubsub.Event, 10)
		defer close(stream)

		// Setup a connection to redis so that we can read settings and the current
		// state of the solve.
		conn := pool.Get()
		defer func() { _ = conn.Close() }()

		// Always send the crossword settings if there are any.
		settings, err := GetSettings(conn, channel)
		if err != nil {
			log.Printf("unable to read settings for channel %s: %+v", channel, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		stream <- pubsub.Event{
			Kind:    "settings",
			Payload: settings,
		}

		// Send the current state of the solve if there is one, but make sure to
		// mask the solution to the puzzle.
		state, err := GetState(conn, channel)
		if err != nil {
			log.Printf("unable to read state for channel %s: %+v", channel, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if state.Puzzle != nil {
			state.Puzzle = state.Puzzle.WithoutSolution()

			stream <- pubsub.Event{
				Kind:    "state",
				Payload: state,
			}
		}

		// Now that we've seeded the stream with the initialization events,
		// subscribe it to receive all future events for the channel.
		id, err := registry.Subscribe(pubsub.Channel(channel), stream)
		defer registry.Unsubscribe(pubsub.Channel(channel), id)
		if err != nil {
			log.Printf("unable to subscribe client to channel %s: %+v", channel, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		EmitEvents(w, r, stream)
	}
}

func EmitEvents(w http.ResponseWriter, r *http.Request, events chan pubsub.Event) {
	w.Header().Set("Cache-Control", "no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	// Loop until the client disconnects sending them any events that are
	// queued for them.
	for {
		select {
		case <-r.Context().Done():
			// The client disconnected.
			return

		case msg, ok := <-events:
			if !ok {
				return
			}

			bs, err := json.Marshal(msg)
			if err != nil {
				log.Printf("unable to marshal event '%+v' to json: %+v\n", msg, err)
				return
			}

			if _, err := fmt.Fprintf(w, "event:message\ndata:%s\n\n", bs); err != nil {
				log.Printf("error while writing message to http.ResponseWriter: %+v", err)
				return
			}

			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}

		// TODO: Consider sending a periodic ping to keep the connection open?
	}
}
