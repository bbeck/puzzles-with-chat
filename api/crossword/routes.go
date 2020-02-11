package crossword

import (
	"fmt"
	"github.com/bbeck/twitch-plays-crosswords/api/pubsub"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"io"
	"net/http"
	"time"
)

func RegisterRoutes(router gin.IRouter, pool *redis.Pool) {
	RegisterRoutesWithRegistry(router, pool, new(pubsub.Registry))
}

func RegisterRoutesWithRegistry(router gin.IRouter, pool *redis.Pool, registry *pubsub.Registry) {
	router.GET("/crossword", GetActiveCrosswords(pool))

	channel := router.Group("/crossword/:channel")
	{
		channel.PUT("/setting/:setting", UpdateCrosswordSetting(pool, registry))
		// TODO: This takes more than just the date as input, maybe it shouldn't be /date.
		channel.PUT("/date", UpdateCrosswordDate(pool, registry))
		channel.PUT("/status", ToggleCrosswordStatus(pool, registry))
		channel.PUT("/answer/:clue", UpdateCrosswordAnswer(pool, registry))
		channel.GET("/show/:clue", ShowCrosswordClue(registry))
		channel.GET("/events", GetCrosswordEvents(pool, registry))
	}
}

func GetActiveCrosswords(pool *redis.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn := pool.Get()
		defer func() { _ = conn.Close() }()

		names, err := GetChannelNamesWithState(conn)
		if err != nil {
			err = fmt.Errorf("unable to load channels with active crossword solves: %+v", err)
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, names)
	}
}

func UpdateCrosswordSetting(pool *redis.Pool, registry *pubsub.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		channel := c.Param("channel")
		setting := c.Param("setting")

		conn := pool.Get()
		defer func() { _ = conn.Close() }()

		// Load the existing settings for the channel so that we can apply the
		// updates to them.
		settings, err := GetSettings(conn, channel)
		if err != nil {
			err = fmt.Errorf("unable to read crossword settings for channel %s: %+v", channel, err)
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		// Apply the update to the settings in memory.
		var shouldClearIncorrectCells bool
		switch setting {
		case "only_allow_correct_answers":
			var value bool
			if err := c.BindJSON(&value); err != nil {
				err = fmt.Errorf("unable to parse crossword only correct answers setting json %v: %+v", value, err)
				_ = c.AbortWithError(http.StatusBadRequest, err)
				return
			}
			settings.OnlyAllowCorrectAnswers = value
			shouldClearIncorrectCells = value

		case "clues_to_show":
			var value ClueVisibility
			if err := c.BindJSON(&value); err != nil {
				err = fmt.Errorf("unable to parse crossword clue visibility setting json %s: %+v", value, err)
				_ = c.AbortWithError(http.StatusBadRequest, err)
				return
			}
			settings.CluesToShow = value

		case "clue_font_size":
			var value FontSize
			if err := c.BindJSON(&value); err != nil {
				err = fmt.Errorf("unable to parse crossword clue font size setting json %s: %+v", value, err)
				_ = c.AbortWithError(http.StatusBadRequest, err)
				return
			}
			settings.ClueFontSize = value

		case "show_notes":
			var value bool
			if err := c.BindJSON(&value); err != nil {
				err = fmt.Errorf("unable to parse crossword show notes setting json %v: %+v", value, err)
				_ = c.AbortWithError(http.StatusBadRequest, err)
				return
			}
			settings.ShowNotes = value

		default:
			err = fmt.Errorf("unrecognized crossword setting name %s", setting)
			_ = c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		// Save the settings back to the database.
		if err = SetSettings(conn, channel, settings); err != nil {
			err = fmt.Errorf("unable to save crossword settings for channel %s: %+v", channel, err)
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		// Load the state and clear any incorrect cells if we changed a setting that
		// requires this.  We do this after the setting is applied so that if there
		// was an error earlier we don't modify the solve's state.
		var updatedState *State
		if shouldClearIncorrectCells {
			state, err := GetState(conn, channel)
			if err != nil {
				err = fmt.Errorf("unable to load state for channel %s: %+v", channel, err)
				_ = c.AbortWithError(http.StatusNotFound, err)
				return
			}

			// There's no need to update cells if the puzzle hasn't been started yet
			// or is already complete.
			if state.Status != StatusCreated && state.Status != StatusComplete {
				if err := state.ClearIncorrectCells(); err != nil {
					err = fmt.Errorf("unable to clear incorrect cells for channel: %s: %+v", channel, err)
					_ = c.AbortWithError(http.StatusInternalServerError, err)
					return
				}

				if err := SetState(conn, channel, state); err != nil {
					err = fmt.Errorf("unable to save state for channel %s: %+v", channel, err)
					_ = c.AbortWithError(http.StatusInternalServerError, err)
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

func UpdateCrosswordDate(pool *redis.Pool, registry *pubsub.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		channel := c.Param("channel")

		// Since there are multiple ways to specify which crossword to solve we'll
		// parse the payload into a generic map instead of a specific object.
		var payload map[string]string
		if err := c.BindJSON(&payload); err != nil {
			err = fmt.Errorf("unable to read request body: %+v", err)
			_ = c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		var puzzle *Puzzle
		if date := payload["date"]; date != "" {
			p, err := LoadFromNewYorkTimes(date)
			if err != nil {
				err = fmt.Errorf("unable to load NYT puzzle for date %s: %+v", date, err)
				_ = c.AbortWithError(http.StatusInternalServerError, err)
				return
			}

			puzzle = p
		}

		if puzzle == nil {
			err := fmt.Errorf("unable to determine puzzle from payload: %+v", payload)
			_ = c.AbortWithError(http.StatusBadRequest, err)
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
			err = fmt.Errorf("unable to save state for channel %s: %+v", channel, err)
			_ = c.AbortWithError(http.StatusInternalServerError, err)
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

func ToggleCrosswordStatus(pool *redis.Pool, registry *pubsub.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		channel := c.Param("channel")

		conn := pool.Get()
		defer func() { _ = conn.Close() }()

		state, err := GetState(conn, channel)
		if err != nil {
			err = fmt.Errorf("unable to load state for channel %s: %+v", channel, err)
			_ = c.AbortWithError(http.StatusNotFound, err)
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
			err = fmt.Errorf("unable to toggle status for channel %s, puzzle is already solved", channel)
			_ = c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		if err := SetState(conn, channel, state); err != nil {
			err = fmt.Errorf("unable to save state for channel %s: %+v", channel, err)
			_ = c.AbortWithError(http.StatusInternalServerError, err)
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

func UpdateCrosswordAnswer(pool *redis.Pool, registry *pubsub.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		channel := c.Param("channel")
		clue := c.Param("clue")

		if c.Request.ContentLength > 1024 {
			c.AbortWithStatus(http.StatusRequestEntityTooLarge)
			return
		}

		var answer string
		if err := c.BindJSON(&answer); err != nil {
			err = fmt.Errorf("unable to read request body: %+v", err)
			_ = c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		if len(answer) == 0 {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		conn := pool.Get()
		defer func() { _ = conn.Close() }()

		state, err := GetState(conn, channel)
		if err != nil {
			err = fmt.Errorf("unable to load state for channel %s: %+v", channel, err)
			_ = c.AbortWithError(http.StatusNotFound, err)
			return
		}

		if state.Status != StatusSolving {
			c.AbortWithStatus(http.StatusConflict)
			return
		}

		settings, err := GetSettings(conn, channel)
		if err != nil {
			err = fmt.Errorf("unable to load settings for channel %s: %+v", channel, err)
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if err := state.ApplyAnswer(clue, answer, settings.OnlyAllowCorrectAnswers); err != nil {
			err = fmt.Errorf("unable to apply answer %s for clue %s for channel %s: %+v", answer, clue, channel, err)
			_ = c.AbortWithError(http.StatusBadRequest, err)
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
			err = fmt.Errorf("unable to save state for channel %s: %+v", channel, err)
			_ = c.AbortWithError(http.StatusInternalServerError, err)
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

// ShowCrosswordClue sends an event to all clients of a channel requesting that
// they update their view to make the specified clue visible.  If the specified
// clue isn't structured as a proper clue number and direction than an error
// will be returned.
func ShowCrosswordClue(registry *pubsub.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		channel := c.Param("channel")
		clue := c.Param("clue")

		_, _, err := ParseClue(clue)
		if err != nil {
			err = fmt.Errorf("malformed clue (%s): %+v", clue, err)
			_ = c.AbortWithError(http.StatusBadRequest, err)
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
func GetCrosswordEvents(pool *redis.Pool, registry *pubsub.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		channel := c.Param("channel")

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
			err = fmt.Errorf("unable to read settings for channel %s: %+v", channel, err)
			_ = c.AbortWithError(http.StatusInternalServerError, err)
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
			err = fmt.Errorf("unable to read state for channel %s: %+v", channel, err)
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		if state.Puzzle != nil {
			state.Puzzle = state.Puzzle.WithoutSolution()

			stream <- pubsub.Event{
				Kind:    "state",
				Payload: state,
			}
		}

		// TODO: Consider sending a periodic ping to keep the connection open?

		// Now that we've seeded the stream with the initialization events,
		// subscribe it to receive all future events for the channel.
		id, err := registry.Subscribe(pubsub.Channel(channel), stream)
		defer registry.Unsubscribe(pubsub.Channel(channel), id)
		if err != nil {
			err = fmt.Errorf("unable to subscribe client to channel %s: %+v", channel, err)
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Header("Cache-Control", "no-transform")
		c.Header("Connection", "keep-alive")
		c.Header("Content-Type", "text/event-stream")

		// Loop until the client disconnects sending them any events that are
		// queued for them.
		c.Stream(func(w io.Writer) bool {
			if msg, ok := <-stream; ok {
				c.SSEvent("message", msg)
				return true
			}

			return false
		})
	}
}
