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
	"strconv"
	"time"
)

func RegisterRoutes(r chi.Router, pool *redis.Pool, registry *pubsub.Registry) {
	r.Route("/acrostic/{channel}", func(r chi.Router) {
		r.Put("/", UpdatePuzzle(pool, registry))
		r.Put("/setting/{setting}", UpdateSetting(pool, registry))
		r.Put("/status", ToggleStatus(pool, registry))
		r.Put("/answer/{clue}", UpdateAnswer(pool, registry))
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

// UpdateSetting changes a specified acrostic setting to a new value.
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
			log.Printf("unable to read acrostic settings for channel %s: %+v", channel, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Apply the update to the settings in memory.
		var shouldClearIncorrectCells bool
		switch setting {
		case "only_allow_correct_answers":
			var value bool
			if err := render.DecodeJSON(r.Body, &value); err != nil {
				log.Printf("unable to parse acrostic only correct answers setting json %v: %+v", value, err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			settings.OnlyAllowCorrectAnswers = value
			shouldClearIncorrectCells = value

		case "clue_font_size":
			var value model.FontSize
			if err := render.DecodeJSON(r.Body, &value); err != nil {
				log.Printf("unable to parse acrostic clue font size setting json %s: %+v", value, err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			settings.ClueFontSize = value

		default:
			log.Printf("unrecognized acrostic setting name %s", setting)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Save the settings back to the database.
		if err = SetSettings(conn, channel, settings); err != nil {
			log.Printf("unable to save acrostic settings for channel %s: %+v", channel, err)
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

			// There's no need to update cells if the puzzle hasn't been selected or
			// started or is already complete.
			status := state.Status
			if status != model.StatusCreated && status != model.StatusSelected && status != model.StatusComplete {
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

				updatedState = &state
			}
		}

		// Now broadcast the new settings to all of the clients in the channel.
		registry.Publish(ChannelID(channel), SettingsEvent(settings))

		if updatedState != nil {
			// Broadcast the updated state to all of the clients, making sure to not
			// include the answers.
			updatedState.Puzzle = updatedState.Puzzle.WithoutSolution()

			registry.Publish(ChannelID(channel), StateEvent(*updatedState))
		}

		w.WriteHeader(http.StatusOK)
	}
}

// ToggleStatus changes the status of the current acrostic solve to a new
// status.  This effectively toggles between the solving and paused statuses as
// long as the solve is in a state that can be paused or resumed.
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
		state.Puzzle = state.Puzzle.WithoutSolution()

		registry.Publish(ChannelID(channel), StateEvent(state))

		w.WriteHeader(http.StatusOK)
	}
}

// UpdateAnswer applies an answer to either a given clue or given set of cells
// in the current acrostic solve.
func UpdateAnswer(pool *redis.Pool, registry *pubsub.Registry) http.HandlerFunc {
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

		// Determine if the user specified a clue letter or cell numbers.
		if start, err := strconv.Atoi(clue); err == nil {
			if err := state.ApplyCellAnswer(start, answer, settings.OnlyAllowCorrectAnswers); err != nil {
				log.Printf("unable to apply answer %s for cell %d for channel %s: %+v", answer, start, channel, err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		} else {
			if err := state.ApplyClueAnswer(clue, answer, settings.OnlyAllowCorrectAnswers); err != nil {
				log.Printf("unable to apply answer %s for clue %s for channel %s: %+v", answer, clue, channel, err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
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
		state.Puzzle = state.Puzzle.WithoutSolution()

		registry.Publish(ChannelID(channel), StateEvent(state))

		// If we've just finished the solve then send a complete event as well.
		if state.Status == model.StatusComplete {
			registry.Publish(ChannelID(channel), CompleteEvent())
		}

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

func CompleteEvent() pubsub.Event {
	return pubsub.Event{
		Kind: "complete",
	}
}

func SettingsEvent(settings Settings) pubsub.Event {
	return pubsub.Event{
		Kind:    "settings",
		Payload: settings,
	}
}
