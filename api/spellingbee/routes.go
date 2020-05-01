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
		r.Put("/setting/{setting}", UpdateSetting(pool, registry))
		r.Put("/status", ToggleStatus(pool, registry))
		r.Post("/answer", AddAnswer(pool, registry))
		r.Get("/events", GetEvents(pool, registry))
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
		channels, err := GetActiveChannelNames(conn)
		if err != nil {
			log.Printf("unable to load active spelling bee channels: %+v", err)
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
					channels, err := GetActiveChannelNames(conn)
					if err != nil {
						log.Printf("unable to load active spelling bee channels: %+v", err)

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

		// Since there are multiple ways to specify which puzzle to solve we'll
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

// UpdateSetting changes a specified spelling bee setting to a new value.
func UpdateSetting(pool *redis.Pool, registry *pubsub.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		channel := chi.URLParam(r, "channel")
		setting := chi.URLParam(r, "setting")

		conn := pool.Get()
		defer func() { _ = conn.Close() }()

		// Load the existing settings for the channel so that we can apply the
		// updates to them.
		settings, err := GetSettings(conn, channel)
		if err != nil {
			log.Printf("unable to load spelling bee settings for channel %s: %+v", channel, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Apply the update to the settings in memory.
		var shouldClearUnofficialWords bool
		switch setting {
		case "allow_unofficial_answers":
			var value bool
			if err := render.DecodeJSON(r.Body, &value); err != nil {
				log.Printf("unable to parse spelling bee allow unofficial answers setting json %v: %+v", value, err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			settings.AllowUnofficialAnswers = value
			shouldClearUnofficialWords = !value

		case "font_size":
			var value model.FontSize
			if err := render.DecodeJSON(r.Body, &value); err != nil {
				log.Printf("unable to parse spelling bee font size setting json %s: %+v", value, err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			settings.FontSize = value

		default:
			log.Printf("unrecognized spelling bee setting name %s", setting)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Save the settings back to the database.
		if err = SetSettings(conn, channel, settings); err != nil {
			log.Printf("unable to save spelling bee settings for channel %s: %+v", channel, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Load the state and clear any unofficial words if we changed a setting
		// requires this.  We do this after the setting is applied so that if there
		// was an error earlier we don't modify the solve's state.
		var updatedState *State
		if shouldClearUnofficialWords {
			state, err := GetState(conn, channel)
			if err != nil {
				log.Printf("unable to load state for channel %s: %+v", channel, err)
				w.WriteHeader(http.StatusNotFound)
				return
			}

			// There's no need to update cells if the puzzle hasn't been selected or
			// started or is already complete.
			status := state.Status
			if status != model.StatusCreated && status != model.StatusSelected && status != model.StatusComplete {
				state.ClearUnofficialAnswers()

				if err := SetState(conn, channel, state); err != nil {
					log.Printf("unable to save state for channel %s: %+v", channel, err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				updatedState = &state
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
			updatedState.Puzzle = updatedState.Puzzle.WithoutAnswers()

			registry.Publish(pubsub.Channel(channel), pubsub.Event{
				Kind:    "state",
				Payload: *updatedState,
			})
		}
	}
}

// ToggleStatus changes the status of the current puzzle solve to a new status.
// This effectively toggles between the solving and paused statuses as long as
// the solve is in a state that can be paused or resumed.
func ToggleStatus(pool *redis.Pool, registry *pubsub.Registry) http.HandlerFunc {
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
		case model.StatusCreated:
			log.Printf("unable to toggle status for channel %s, no puzzle selected", channel)
			w.WriteHeader(http.StatusBadRequest)
			return

		case model.StatusSelected:
			state.Status = model.StatusSolving
			state.LastStartTime = &now

		case model.StatusPaused:
			state.Status = model.StatusSolving
			state.LastStartTime = &now

		case model.StatusSolving:
			state.Status = model.StatusPaused
			total := state.TotalSolveDuration.Nanoseconds() + now.Sub(*state.LastStartTime).Nanoseconds()
			state.LastStartTime = nil
			state.TotalSolveDuration = model.Duration{Duration: time.Duration(total)}

		case model.StatusComplete:
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
		state.Puzzle = state.Puzzle.WithoutAnswers()

		registry.Publish(pubsub.Channel(channel), pubsub.Event{
			Kind:    "state",
			Payload: state,
		})
	}
}

// AddAnswer applies an answer to the puzzle solve.
func AddAnswer(pool *redis.Pool, registry *pubsub.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		channel := chi.URLParam(r, "channel")

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

		if state.Status != model.StatusSolving {
			w.WriteHeader(http.StatusConflict)
			return
		}

		settings, err := GetSettings(conn, channel)
		if err != nil {
			log.Printf("unable to load settings for channel %s: %+v", channel, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := state.ApplyAnswer(answer, settings.AllowUnofficialAnswers); err != nil {
			log.Printf("unable to apply answer %s for channel %s: %+v", answer, channel, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// If we just solved the puzzle then we should stop the timer.
		if state.Status == model.StatusComplete {
			now := time.Now()
			total := state.TotalSolveDuration.Nanoseconds() + now.Sub(*state.LastStartTime).Nanoseconds()
			state.LastStartTime = nil
			state.TotalSolveDuration = model.Duration{Duration: time.Duration(total)}
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
		state.Puzzle = state.Puzzle.WithoutAnswers()

		registry.Publish(pubsub.Channel(channel), pubsub.Event{
			Kind:    "state",
			Payload: state,
		})

		// If we've just finished the solve then send a complete event as well.
		if state.Status == model.StatusComplete {
			registry.Publish(pubsub.Channel(channel), pubsub.Event{
				Kind:    "complete",
				Payload: nil,
			})
		}
	}
}

// GetEvents establishes an event stream with a client.  An event stream is
// server side event stream (SSE) with a client's browser that allows one way
// communication from the server to the client.  Clients that call into this
// handler will keep an open connection open to the server waiting to receive
// events as JSON objects.  The server can send events to all clients of a
// channel using the pubsub.Registry's Publish method.
func GetEvents(pool *redis.Pool, registry *pubsub.Registry) http.HandlerFunc {
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

		// Always send the settings if there are any.
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
			state.Puzzle = state.Puzzle.WithoutAnswers()

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

		pubsub.EmitEvents(r.Context(), w, stream)
	}
}
