package crossword

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alicebob/miniredis"
	"github.com/bbeck/twitch-plays-crosswords/api/pubsub"
	"github.com/gin-contrib/sse"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"
	"time"
)

var Global = CrosswordRoute{}
var Channel = ChannelRoute{"channel"}

func TestRoute_GetActiveCrosswords(t *testing.T) {
	// This acts as a small integration test creating crossword solves and making
	// sure they're returned by the /crossword handler.
	pool, _, cleanup := NewRedisPool(t)
	defer cleanup()

	registry, _, cleanup := NewRegistry(t)
	defer cleanup()

	// Force a specific puzzle to be loaded so we don't make a network call.
	cleanup = ForcePuzzleToBeLoaded(t, "xwordinfo-nyt-20181231.json")
	defer cleanup()

	router := gin.Default()
	RegisterRoutesWithRegistry(router, pool, registry)

	var names []string // The channel names of the active crossword solves

	// Make sure we have no active solves.
	response := Global.GET("/", router)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.NoError(t, response.JSON(&names))
	assert.Nil(t, names)

	// Start a crossword
	response = Channel.PUT("/", `{"new_york_times_date": "2018-12-31"}`, router)
	require.Equal(t, http.StatusOK, response.Code)

	// Make sure we have an active solve in our channel.
	response = Global.GET("/", router)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.NoError(t, response.JSON(&names))
	assert.Equal(t, []string{Channel.channel}, names)

	// Start a crossword in another channel.
	channel2 := ChannelRoute{"channel2"}
	response = channel2.PUT("/", `{"new_york_times_date": "2018-12-31"}`, router)
	require.Equal(t, http.StatusOK, response.Code)

	// We should now have 2 solves.
	response = Global.GET("/", router)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.NoError(t, response.JSON(&names))
	assert.Equal(t, []string{Channel.channel, channel2.channel}, names)
}

func TestRoute_UpdateCrosswordSetting(t *testing.T) {
	// This acts as a small integration test updating each setting in turn and
	// making sure the proper value is written to the database and that clients
	// receive events notifying them of the setting change.
	pool, conn, cleanup := NewRedisPool(t)
	defer cleanup()

	registry, events, cleanup := NewRegistry(t)
	defer cleanup()

	// Ensure that we have received the proper event and wrote the proper thing
	// to the database.
	verify := func(fn func(s *Settings)) {
		t.Helper()

		// First check that we've received an event with the correct value
		select {
		case event := <-events:
			require.Equal(t, "settings", event.Kind)
			fn(event.Payload.(*Settings))
		default:
			assert.Fail(t, "no settings event available")
		}

		// Next check that the database has a valid settings object
		settings, err := GetSettings(conn, "channel")
		require.NoError(t, err)
		fn(settings)
	}

	router := gin.Default()
	RegisterRoutesWithRegistry(router, pool, registry)

	// Update each setting, one at a time.
	response := Channel.PUT("/setting/only_allow_correct_answers", `true`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	verify(func(s *Settings) { assert.True(t, s.OnlyAllowCorrectAnswers) })

	response = Channel.PUT("/setting/clues_to_show", `"none"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	verify(func(s *Settings) { assert.Equal(t, NoCluesVisible, s.CluesToShow) })

	response = Channel.PUT("/setting/clue_font_size", `"xlarge"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	verify(func(s *Settings) { assert.Equal(t, SizeXLarge, s.ClueFontSize) })

	response = Channel.PUT("/setting/show_notes", `true`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	verify(func(s *Settings) { assert.True(t, s.ShowNotes) })
}

func TestRoute_UpdateCrosswordSetting_ClearsIncorrectCells(t *testing.T) {
	// This acts as a small integration test toggling the OnlyAllowCorrectAnswers
	// setting and ensuring that it clears any incorrect answer cells.
	pool, conn, cleanup := NewRedisPool(t)
	defer cleanup()

	registry, events, cleanup := NewRegistry(t)
	defer cleanup()

	// Force a specific puzzle to be loaded so we don't make a network call.
	cleanup = ForcePuzzleToBeLoaded(t, "xwordinfo-nyt-20181231.json")
	defer cleanup()

	// Ensure that we have received the proper event and wrote the proper thing
	// to the database.
	verify := func(fn func(s *State)) {
		t.Helper()

		// First check that we've received an event with the correct value
		select {
		case event := <-events:
			// Ignore the setting change event we receive
			if event.Kind == "settings" {
				return
			}

			require.Equal(t, "state", event.Kind)

			state := event.Payload.(*State)
			assert.Nil(t, state.Puzzle.Cells) // Events should never have the solution
			fn(state)

		default:
			assert.Fail(t, "no state event available")
		}

		// Next check that the database has a valid settings object
		state, err := GetState(conn, "channel")
		require.NoError(t, err)
		assert.NotNil(t, state.Puzzle.Cells) // Database should always have the solution
		fn(state)
	}

	drainEvents := func() {
		for {
			select {
			case <-events:
			default:
				return
			}
		}
	}

	router := gin.Default()
	RegisterRoutesWithRegistry(router, pool, registry)

	response := Channel.PUT("/", `{"new_york_times_date": "2018-12-31"}`, router)
	require.Equal(t, http.StatusOK, response.Code)
	drainEvents()

	// Toggle the status, it should transition to solving.
	response = Channel.PUT("/status", ``, router)
	require.Equal(t, http.StatusOK, response.Code)
	drainEvents()

	// Apply an incorrect across answer.
	response = Channel.PUT("/answer/1a", `"QNORA"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	drainEvents()

	// Set the OnlyAllowCorrectAnswers setting to true
	response = Channel.PUT("/setting/only_allow_correct_answers", `true`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	verify(func(state *State) {
		assert.False(t, state.AcrossCluesFilled[1])
		assert.Equal(t, "Q", state.Cells[0][0])
		assert.Equal(t, "", state.Cells[0][1])
		assert.Equal(t, "", state.Cells[0][2])
		assert.Equal(t, "", state.Cells[0][3])
		assert.Equal(t, "A", state.Cells[0][4])
	})
}

func TestRoute_UpdateCrosswordSetting_Error(t *testing.T) {
	tests := []struct {
		name    string
		setting string
		json    string
	}{
		{
			name:    "only_allow_correct_answers",
			setting: "only_allow_correct_answers",
			json:    `{`,
		},
		{
			name:    "clues_to_show",
			setting: "clues_to_show",
			json:    `{`,
		},
		{
			name:    "clue_font_size",
			setting: "clue_font_size",
			json:    `{`,
		},
		{
			name:    "invalid setting name",
			setting: "foo_bar_baz",
			json:    `false`,
		},
		{
			name:    "show_notes",
			setting: "show_notes",
			json:    `{`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pool, _, cleanup := NewRedisPool(t)
			defer cleanup()

			router := gin.Default()
			RegisterRoutes(router, pool)

			response := Channel.PUT(fmt.Sprintf("/setting/%s", test.setting), test.json, router)
			assert.NotEqual(t, http.StatusOK, response.Code)
		})
	}
}

func TestRoute_UpdateCrossword_NewYorkTimes(t *testing.T) {
	// This acts as a small integration test updating the date of the New York
	// Times crossword we're working on and ensuring the proper values are written
	// to the database.
	pool, conn, cleanup := NewRedisPool(t)
	defer cleanup()

	registry, events, cleanup := NewRegistry(t)
	defer cleanup()

	// Force a specific puzzle to be loaded so we don't make a network call.
	cleanup = ForcePuzzleToBeLoaded(t, "xwordinfo-nyt-20181231.json")
	defer cleanup()

	// Ensure that we have received the proper event and wrote the proper thing
	// to the database.
	verify := func(fn func(s *State)) {
		t.Helper()

		// First check that we've received an event with the correct value
		select {
		case event := <-events:
			require.Equal(t, "state", event.Kind)

			state := event.Payload.(*State)
			assert.Nil(t, state.Puzzle.Cells) // Events should never have the solution
			fn(state)

		default:
			assert.Fail(t, "no state event available")
		}

		// Next check that the database has a valid settings object
		state, err := GetState(conn, "channel")
		require.NoError(t, err)
		assert.NotNil(t, state.Puzzle.Cells) // Database should always have the solution
		fn(state)
	}

	router := gin.Default()
	RegisterRoutesWithRegistry(router, pool, registry)

	response := Channel.PUT("/", `{"new_york_times_date": "2018-12-31"}`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	verify(func(state *State) {
		assert.Equal(t, StatusCreated, state.Status)
		assert.NotNil(t, state.Puzzle)
		assert.Equal(t, 0, len(state.AcrossCluesFilled))
		assert.Equal(t, 0, len(state.DownCluesFilled))
		assert.Nil(t, state.LastStartTime)
		assert.Equal(t, 0., state.TotalSolveDuration.Seconds())
	})
}

func TestRoute_UpdateCrossword_WallStreetJournal(t *testing.T) {
	// This acts as a small integration test updating the date of the Wall Street
	// Journal crossword we're working on and ensuring the proper values are
	// written to the database.
	pool, conn, cleanup := NewRedisPool(t)
	defer cleanup()

	registry, events, cleanup := NewRegistry(t)
	defer cleanup()

	// Force a specific puzzle to be loaded so we don't make a network call.
	cleanup = ForcePuzzleToBeLoaded(t, "herbach-wsj-20190102.json")
	defer cleanup()

	// Ensure that we have received the proper event and wrote the proper thing
	// to the database.
	verify := func(fn func(s *State)) {
		t.Helper()

		// First check that we've received an event with the correct value
		select {
		case event := <-events:
			require.Equal(t, "state", event.Kind)

			state := event.Payload.(*State)
			assert.Nil(t, state.Puzzle.Cells) // Events should never have the solution
			fn(state)

		default:
			assert.Fail(t, "no state event available")
		}

		// Next check that the database has a valid settings object
		state, err := GetState(conn, "channel")
		require.NoError(t, err)
		assert.NotNil(t, state.Puzzle.Cells) // Database should always have the solution
		fn(state)
	}

	router := gin.Default()
	RegisterRoutesWithRegistry(router, pool, registry)

	response := Channel.PUT("/", `{"wall_street_journal_date": "2019-01-02"}`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	verify(func(state *State) {
		assert.Equal(t, StatusCreated, state.Status)
		assert.NotNil(t, state.Puzzle)
		assert.Equal(t, 0, len(state.AcrossCluesFilled))
		assert.Equal(t, 0, len(state.DownCluesFilled))
		assert.Nil(t, state.LastStartTime)
		assert.Equal(t, 0., state.TotalSolveDuration.Seconds())
	})
}

func TestRoute_UpdateCrossword_PuzFile(t *testing.T) {
	// This acts as a small integration test uploading a .puz fle of the crossword
	// we're working on and ensuring the proper values are written to the
	// database.
	pool, conn, cleanup := NewRedisPool(t)
	defer cleanup()

	registry, events, cleanup := NewRegistry(t)
	defer cleanup()

	// Force a specific puzzle to be loaded so we don't make a network call.
	cleanup = ForcePuzzleToBeLoaded(t, "converter-wp-20051206.json")
	defer cleanup()

	// Ensure that we have received the proper event and wrote the proper thing
	// to the database.
	verify := func(fn func(s *State)) {
		t.Helper()

		// First check that we've received an event with the correct value
		select {
		case event := <-events:
			require.Equal(t, "state", event.Kind)

			state := event.Payload.(*State)
			assert.Nil(t, state.Puzzle.Cells) // Events should never have the solution
			fn(state)

		default:
			assert.Fail(t, "no state event available")
		}

		// Next check that the database has a valid settings object
		state, err := GetState(conn, "channel")
		require.NoError(t, err)
		assert.NotNil(t, state.Puzzle.Cells) // Database should always have the solution
		fn(state)
	}

	router := gin.Default()
	RegisterRoutesWithRegistry(router, pool, registry)

	response := Channel.PUT("/", `{"puz_file_bytes": "unused"}`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	verify(func(state *State) {
		assert.Equal(t, StatusCreated, state.Status)
		assert.NotNil(t, state.Puzzle)
		assert.Equal(t, 0, len(state.AcrossCluesFilled))
		assert.Equal(t, 0, len(state.DownCluesFilled))
		assert.Nil(t, state.LastStartTime)
		assert.Equal(t, 0., state.TotalSolveDuration.Seconds())
	})
}

func TestRoute_UpdateCrossword_Error(t *testing.T) {
	tests := []struct {
		name        string
		payload     string
		forcedError error
		expected    int
	}{
		{
			name:     "bad json",
			payload:  `{"new_york_times_date": }`,
			expected: http.StatusBadRequest,
		},
		{
			name:     "invalid json",
			payload:  `{}`,
			expected: http.StatusBadRequest,
		},
		{
			name:        "nyt error",
			payload:     `{"new_york_times_date": "bad"}`,
			forcedError: errors.New("forced error"),
			expected:    http.StatusInternalServerError,
		},
		{
			name:        "wsj error",
			payload:     `{"wall_street_journal_date": "bad"}`,
			forcedError: errors.New("forced error"),
			expected:    http.StatusInternalServerError,
		},
		{
			name:        "puz error",
			payload:     `{"puz_file_bytes": "bad"}`,
			forcedError: errors.New("forced error"),
			expected:    http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pool, _, cleanup := NewRedisPool(t)
			defer cleanup()

			cleanup = ForceErrorDuringLoad(test.forcedError)
			defer cleanup()

			router := gin.Default()
			RegisterRoutes(router, pool)

			response := Channel.PUT("/", test.payload, router)
			assert.Equal(t, test.expected, response.Code)
		})
	}
}

func TestRoute_ToggleCrosswordStatus(t *testing.T) {
	// This acts as a small integration test toggling the status of a crossword
	// being solved.
	pool, conn, cleanup := NewRedisPool(t)
	defer cleanup()

	registry, events, cleanup := NewRegistry(t)
	defer cleanup()

	// Force a specific puzzle to be loaded so we don't make a network call.
	cleanup = ForcePuzzleToBeLoaded(t, "xwordinfo-nyt-20181231.json")
	defer cleanup()

	// Ensure that we have received the proper event and wrote the proper thing
	// to the database.
	verify := func(fn func(s *State)) {
		t.Helper()

		// First check that we've received an event with the correct value
		select {
		case event := <-events:
			require.Equal(t, "state", event.Kind)

			state := event.Payload.(*State)
			assert.Nil(t, state.Puzzle.Cells) // Events should never have the solution
			fn(state)

		default:
			assert.Fail(t, "no state event available")
		}

		// Next check that the database has a valid settings object
		state, err := GetState(conn, "channel")
		require.NoError(t, err)
		assert.NotNil(t, state.Puzzle.Cells) // Database should always have the solution
		fn(state)
	}

	drainEvents := func() {
		for {
			select {
			case <-events:
			default:
				return
			}
		}
	}

	router := gin.Default()
	RegisterRoutesWithRegistry(router, pool, registry)

	response := Channel.PUT("/", `{"new_york_times_date": "2018-12-31"}`, router)
	require.Equal(t, http.StatusOK, response.Code)
	drainEvents()

	// Toggle the status, it should transition to solving.
	response = Channel.PUT("/status", ``, router)
	assert.Equal(t, http.StatusOK, response.Code)
	verify(func(state *State) {
		assert.Equal(t, StatusSolving, state.Status)
		assert.NotNil(t, state.LastStartTime)
	})

	// Toggle the status again, it should transition to paused. Make sure we
	// sleep for at least a nanosecond first so that the solve was unpaused for
	// a non-zero duration.
	time.Sleep(1 * time.Nanosecond)
	response = Channel.PUT("/status", ``, router)
	assert.Equal(t, http.StatusOK, response.Code)
	verify(func(state *State) {
		assert.Equal(t, StatusPaused, state.Status)
		assert.Nil(t, state.LastStartTime)
		assert.True(t, state.TotalSolveDuration.Seconds() > 0.)
	})

	// Toggle the status again, it should transition back to solving.
	response = Channel.PUT("/status", ``, router)
	assert.Equal(t, http.StatusOK, response.Code)
	verify(func(state *State) {
		assert.Equal(t, StatusSolving, state.Status)
		assert.NotNil(t, state.LastStartTime)
		assert.True(t, state.TotalSolveDuration.Seconds() > 0.)
	})

	// Force the puzzle to be complete.
	state, err := GetState(conn, "channel")
	require.NoError(t, err)
	state.Status = StatusComplete
	require.NoError(t, SetState(conn, "channel", state))

	// Try to toggle the status one more time.  Now that the puzzle is complete
	// it should return a HTTP error.
	response = Channel.PUT("/status", ``, router)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	state, err = GetState(conn, "channel")
	require.NoError(t, err)
	assert.Equal(t, StatusComplete, state.Status)
}

func TestRoute_UpdateCrosswordAnswer_AllowIncorrectAnswers(t *testing.T) {
	// This acts as a small integration test toggling the status of a crossword
	// being solved.
	pool, conn, cleanup := NewRedisPool(t)
	defer cleanup()

	registry, events, cleanup := NewRegistry(t)
	defer cleanup()

	// Force a specific puzzle to be loaded so we don't make a network call.
	cleanup = ForcePuzzleToBeLoaded(t, "xwordinfo-nyt-20181231.json")
	defer cleanup()

	// Ensure that we have received the proper event and wrote the proper thing
	// to the database.
	verify := func(fn func(s *State)) {
		t.Helper()

		// First check that we've received an event with the correct value
		select {
		case event := <-events:
			require.Equal(t, "state", event.Kind)

			state := event.Payload.(*State)
			assert.Nil(t, state.Puzzle.Cells) // Events should never have the solution
			fn(state)

		default:
			assert.Fail(t, "no state event available")
		}

		// Next check that the database has a valid settings object
		state, err := GetState(conn, "channel")
		require.NoError(t, err)
		assert.NotNil(t, state.Puzzle.Cells) // Database should always have the solution
		fn(state)
	}

	drainEvents := func() {
		for {
			select {
			case <-events:
			default:
				return
			}
		}
	}

	router := gin.Default()
	RegisterRoutesWithRegistry(router, pool, registry)

	response := Channel.PUT("/", `{"new_york_times_date": "2018-12-31"}`, router)
	require.Equal(t, http.StatusOK, response.Code)
	drainEvents()

	// Toggle the status, it should transition to solving.
	response = Channel.PUT("/status", ``, router)
	require.Equal(t, http.StatusOK, response.Code)
	drainEvents()

	// Apply a correct across answer.
	response = Channel.PUT("/answer/1a", `"QANDA"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	verify(func(state *State) {
		assert.True(t, state.AcrossCluesFilled[1])
		assert.Equal(t, "Q", state.Cells[0][0])
		assert.Equal(t, "A", state.Cells[0][1])
		assert.Equal(t, "N", state.Cells[0][2])
		assert.Equal(t, "D", state.Cells[0][3])
		assert.Equal(t, "A", state.Cells[0][4])
	})

	// Apply a correct down answer.
	response = Channel.PUT("/answer/1d", `"QTIP"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	verify(func(state *State) {
		assert.True(t, state.DownCluesFilled[1])
		assert.Equal(t, "Q", state.Cells[0][0])
		assert.Equal(t, "T", state.Cells[1][0])
		assert.Equal(t, "I", state.Cells[2][0])
		assert.Equal(t, "P", state.Cells[3][0])
	})

	// Apply an incorrect across answer.
	response = Channel.PUT("/answer/6a", `"FLOOR"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	verify(func(state *State) {
		assert.True(t, state.AcrossCluesFilled[6])
		assert.Equal(t, "F", state.Cells[0][6])
		assert.Equal(t, "L", state.Cells[0][7])
		assert.Equal(t, "O", state.Cells[0][8])
		assert.Equal(t, "O", state.Cells[0][9])
		assert.Equal(t, "R", state.Cells[0][10])
	})

	// Apply an incorrect down answer.
	response = Channel.PUT("/answer/11d", `"HEYA"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	verify(func(state *State) {
		assert.True(t, state.DownCluesFilled[11])
		assert.Equal(t, "H", state.Cells[0][12])
		assert.Equal(t, "E", state.Cells[1][12])
		assert.Equal(t, "Y", state.Cells[2][12])
		assert.Equal(t, "A", state.Cells[3][12])
	})

	// Pause the solve.
	response = Channel.PUT("/status", ``, router)
	require.Equal(t, http.StatusOK, response.Code)
	drainEvents()

	// Try to apply an answer.
	response = Channel.PUT("/answer/6a", `"ATTIC"`, router)
	assert.Equal(t, http.StatusConflict, response.Code)
}

func TestRoute_UpdateCrosswordAnswer_OnlyAllowCorrectAnswers(t *testing.T) {
	// This acts as a small integration test toggling the status of a crossword
	// being solved.
	pool, conn, cleanup := NewRedisPool(t)
	defer cleanup()

	registry, events, cleanup := NewRegistry(t)
	defer cleanup()

	// Force a specific puzzle to be loaded so we don't make a network call.
	cleanup = ForcePuzzleToBeLoaded(t, "xwordinfo-nyt-20181231.json")
	defer cleanup()

	// Ensure that we have received the proper event and wrote the proper thing
	// to the database.
	verify := func(fn func(s *State)) {
		t.Helper()

		// First check that we've received an event with the correct value
		select {
		case event := <-events:
			require.Equal(t, "state", event.Kind)

			state := event.Payload.(*State)
			assert.Nil(t, state.Puzzle.Cells) // Events should never have the solution
			fn(state)

		default:
			assert.Fail(t, "no state event available")
		}

		// Next check that the database has a valid settings object
		state, err := GetState(conn, "channel")
		require.NoError(t, err)
		assert.NotNil(t, state.Puzzle.Cells) // Database should always have the solution
		fn(state)
	}

	drainEvents := func() {
		for {
			select {
			case <-events:
			default:
				return
			}
		}
	}

	router := gin.Default()
	RegisterRoutesWithRegistry(router, pool, registry)

	// Change the settings to require correct answers
	settings, err := GetSettings(conn, "channel")
	require.NoError(t, err)
	settings.OnlyAllowCorrectAnswers = true
	err = SetSettings(conn, "channel", settings)
	require.NoError(t, err)

	response := Channel.PUT("/", `{"new_york_times_date": "2018-12-31"}`, router)
	require.Equal(t, http.StatusOK, response.Code)
	drainEvents()

	// Toggle the status, it should transition to solving.
	response = Channel.PUT("/status", ``, router)
	require.Equal(t, http.StatusOK, response.Code)
	drainEvents()

	// Apply a correct across answer.
	response = Channel.PUT("/answer/1a", `"QANDA"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	verify(func(state *State) {
		assert.True(t, state.AcrossCluesFilled[1])
		assert.Equal(t, "Q", state.Cells[0][0])
		assert.Equal(t, "A", state.Cells[0][1])
		assert.Equal(t, "N", state.Cells[0][2])
		assert.Equal(t, "D", state.Cells[0][3])
		assert.Equal(t, "A", state.Cells[0][4])
	})

	// Apply a correct down answer.
	response = Channel.PUT("/answer/1d", `"QTIP"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	verify(func(state *State) {
		assert.True(t, state.DownCluesFilled[1])
		assert.Equal(t, "Q", state.Cells[0][0])
		assert.Equal(t, "T", state.Cells[1][0])
		assert.Equal(t, "I", state.Cells[2][0])
		assert.Equal(t, "P", state.Cells[3][0])
	})

	// Apply an incorrect across answer.
	response = Channel.PUT("/answer/6a", `"FLOOR"`, router)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	state, err := GetState(conn, "channel")
	require.NoError(t, err)
	assert.False(t, state.AcrossCluesFilled[6])

	// Apply an incorrect down answer.
	response = Channel.PUT("/answer/11d", `"HEYA"`, router)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	state, err = GetState(conn, "channel")
	require.NoError(t, err)
	assert.False(t, state.DownCluesFilled[11])
}

func TestRoute_UpdateCrosswordAnswer_SolvedPuzzleStopsTimer(t *testing.T) {
	// This acts as a small integration test ensuring that the timer stops
	// counting once the crossword has been solved.
	pool, conn, cleanup := NewRedisPool(t)
	defer cleanup()

	registry, events, cleanup := NewRegistry(t)
	defer cleanup()

	// Force a specific puzzle to be loaded so we don't make a network call.
	cleanup = ForcePuzzleToBeLoaded(t, "xwordinfo-nyt-20181231.json")
	defer cleanup()

	// Ensure that we have received the proper event and wrote the proper thing
	// to the database.
	verify := func(fn func(s *State)) {
		t.Helper()

		// First check that we've received an event with the correct value
		select {
		case event := <-events:
			require.Equal(t, "state", event.Kind)

			state := event.Payload.(*State)
			assert.Nil(t, state.Puzzle.Cells) // Events should never have the solution
			fn(state)

		default:
			assert.Fail(t, "no state event available")
		}

		// Next check that the database has a valid settings object
		state, err := GetState(conn, "channel")
		require.NoError(t, err)
		assert.NotNil(t, state.Puzzle.Cells) // Database should always have the solution
		fn(state)
	}

	drainEvents := func() {
		for {
			select {
			case <-events:
			default:
				return
			}
		}
	}

	router := gin.Default()
	RegisterRoutesWithRegistry(router, pool, registry)

	response := Channel.PUT("/", `{"new_york_times_date": "2018-12-31"}`, router)
	require.Equal(t, http.StatusOK, response.Code)
	drainEvents()

	// Toggle the status, it should transition to solving.
	response = Channel.PUT("/status", ``, router)
	require.Equal(t, http.StatusOK, response.Code)
	drainEvents()

	// Solve the entire puzzle except for the last answer.
	answers := map[string]string{
		"1a":  "Q AND A",
		"6a":  "ATTIC",
		"11a": "HON",
		"14a": "THIRD",
		"15a": "LAID ASIDE",
		"17a": "IM TOO OLD FOR THIS",
		"19a": "PERU",
		"20a": "LEAF",
		"21a": "PEONS",
		"22a": "DOG TAG",
		"24a": "LOL",
		"25a": "HAVE NO OOMPH",
		"30a": "MATTE",
		"33a": "IMPLORED",
		"35a": "ERR",
		"36a": "RANGE",
		"38a": "EMO",
		"39a": "WAIT HERE",
		"42a": "EGYPT",
		"44a": "BOO OFF STAGE",
		"47a": "ERS",
		"48a": "EUGENE",
		"51a": "SHARI",
		"54a": "SINN",
		"56a": "WING",
		"58a": "ITS A ZOO OUT THERE",
		"61a": "STEGOSAUR",
		"62a": "HIT ON",
		"63a": "IPA",
		"64a": "NURSE",
	}
	for clue, answer := range answers {
		response = Channel.PUT(fmt.Sprintf("/answer/%s", clue), fmt.Sprintf(`"%s"`, answer), router)
		assert.Equal(t, http.StatusOK, response.Code)
	}
	drainEvents()

	// Apply the last answer, but wait a bit first to ensure that a non-zero
	// amount of time has passed in the solve.
	time.Sleep(10 * time.Millisecond)

	response = Channel.PUT("/answer/65a", `"OZONE"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	verify(func(state *State) {
		require.Equal(t, StatusComplete, state.Status)
		assert.Nil(t, state.LastStartTime)
		assert.True(t, state.TotalSolveDuration.Seconds() > 0)
	})
}

func TestRoute_UpdateCrosswordAnswer_Error(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected int
	}{
		{
			name:     "empty answer",
			json:     `""`,
			expected: http.StatusBadRequest,
		},
		{
			name:     "too long answer",
			json:     `"` + RandomString(1023) + `"`,
			expected: http.StatusRequestEntityTooLarge,
		},
		{
			name:     "malformed json",
			json:     `"`,
			expected: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pool, _, cleanup := NewRedisPool(t)
			defer cleanup()

			// Force a specific puzzle to be loaded so we don't make a network call.
			cleanup = ForcePuzzleToBeLoaded(t, "xwordinfo-nyt-20181231.json")
			defer cleanup()

			router := gin.Default()
			RegisterRoutes(router, pool)

			response := Channel.PUT("/", `{"new_york_times_date": "2018-12-31"}`, router)
			require.Equal(t, http.StatusOK, response.Code)

			// Toggle the status, it should transition to solving.
			response = Channel.PUT("/status", ``, router)
			require.Equal(t, http.StatusOK, response.Code)

			response = Channel.PUT("/answer/1a", test.json, router)
			assert.Equal(t, test.expected, response.Code)
		})
	}
}

func TestRoute_ShowCrosswordClue(t *testing.T) {
	// This acts as a small integration test requesting clues to be shown and
	// making sure events are properly emitted.
	pool, _, cleanup := NewRedisPool(t)
	defer cleanup()

	registry, events, cleanup := NewRegistry(t)
	defer cleanup()

	// Force a specific puzzle to be loaded so we don't make a network call.
	cleanup = ForcePuzzleToBeLoaded(t, "xwordinfo-nyt-20181231.json")
	defer cleanup()

	// Ensure that we have received the proper event.
	verify := func(fn func(clue string)) {
		t.Helper()

		// First check that we've received an event with the correct value
		select {
		case event := <-events:
			require.Equal(t, "show_clue", event.Kind)
			fn(event.Payload.(string))

		default:
			assert.Fail(t, "no show_clue event available")
		}
	}

	drainEvents := func() {
		for {
			select {
			case <-events:
			default:
				return
			}
		}
	}

	router := gin.Default()
	RegisterRoutesWithRegistry(router, pool, registry)

	response := Channel.PUT("/", `{"new_york_times_date": "2018-12-31"}`, router)
	require.Equal(t, http.StatusOK, response.Code)
	drainEvents()

	// Request showing an across clue.
	response = Channel.GET("/show/1a", router)
	require.Equal(t, http.StatusOK, response.Code)
	verify(func(clue string) {
		assert.Equal(t, "1a", clue)
	})

	// Request showing a down clue.
	response = Channel.GET("/show/16d", router)
	require.Equal(t, http.StatusOK, response.Code)
	verify(func(clue string) {
		assert.Equal(t, "16d", clue)
	})

	// Request showing a malformed clue.
	response = Channel.GET("/show/1x", router)
	require.Equal(t, http.StatusBadRequest, response.Code)

	// Request showing a properly formed, but non-existent clue.  This doesn't
	// fail because it doesn't mutate the state of the puzzle in any way.
	response = Channel.GET("/show/999a", router)
	require.Equal(t, http.StatusOK, response.Code)
	verify(func(clue string) {
		assert.Equal(t, "999a", clue)
	})
}

func TestRoute_GetCrosswordEvents(t *testing.T) {
	// This acts as a small integration test ensuring that the event stream
	// receives the events put into a registry.
	pool, _, cleanup := NewRedisPool(t)
	defer cleanup()

	registry, _, cleanup := NewRegistry(t)
	defer cleanup()

	// Force a specific puzzle to be loaded so we don't make a network call.
	cleanup = ForcePuzzleToBeLoaded(t, "xwordinfo-nyt-20181231.json")
	defer cleanup()

	router := gin.Default()
	RegisterRoutesWithRegistry(router, pool, registry)

	// Connect to the stream when there's no puzzle selected, we should receive
	// just the channel's settings.
	_, stop := Channel.SSE("/events", router)
	events := stop()
	assert.Equal(t, 1, len(events))
	assert.Equal(t, "settings", events[0].Kind)

	// Select a puzzle
	response := Channel.PUT("/", `{"new_york_times_date": "2018-12-31"}`, router)
	require.Equal(t, http.StatusOK, response.Code)

	// Now reconnect to the stream and we should receive both the puzzle and the
	// channel's current state.
	flush, stop := Channel.SSE("/events", router)
	events = flush()
	assert.Equal(t, 2, len(events))
	assert.Equal(t, "settings", events[0].Kind)
	assert.Equal(t, "state", events[1].Kind)

	// Toggle the status to solving, this should cause the state to be sent again.
	response = Channel.PUT("/status", ``, router)
	require.Equal(t, http.StatusOK, response.Code)

	events = flush()
	assert.Equal(t, 1, len(events))
	assert.Equal(t, "state", events[0].Kind)

	// Now specify an answer, this should cause the state to be sent again.
	response = Channel.PUT("/answer/1a", `"QANDA"`, router)
	assert.Equal(t, http.StatusOK, response.Code)

	events = flush()
	assert.Equal(t, 1, len(events))
	assert.Equal(t, "state", events[0].Kind)

	// Disconnect, there shouldn't be any events anymore.
	events = stop()
	assert.Equal(t, 0, len(events))
}

func NewRedisPool(t *testing.T) (*redis.Pool, redis.Conn, func()) {
	s, err := miniredis.Run()
	require.NoError(t, err)

	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", s.Addr())
		},
	}

	conn := pool.Get()

	return pool, conn, func() {
		conn.Close()
		s.Close()
	}
}

func NewRegistry(t *testing.T) (*pubsub.Registry, <-chan pubsub.Event, func()) {
	registry := new(pubsub.Registry)
	stream := make(chan pubsub.Event, 10)

	id, err := registry.Subscribe("channel", stream)
	require.NoError(t, err)

	return registry, stream, func() {
		registry.Unsubscribe("channel", id)
		close(stream)
	}
}

// CrosswordRoute is a client that makes requests against the URL of the global
// crossword route, not associated with any particular channel.
type CrosswordRoute struct{}

func (r CrosswordRoute) GET(url string, router *gin.Engine) *TestResponseRecorder {
	url = path.Join("/crossword", url)
	request := httptest.NewRequest(http.MethodGet, url, nil)

	recorder := CreateTestResponseRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}
func (r CrosswordRoute) PUT(url, body string, router *gin.Engine) *TestResponseRecorder {
	url = path.Join("/crossword", url)
	request := httptest.NewRequest(http.MethodPut, url, strings.NewReader(body))

	recorder := CreateTestResponseRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

// SSE performs a streaming request to the provided router.  Because the router
// won't immediately return this request is done in a background goroutine.
// When the main thread wishes to read events that have been received thus far
// the flush method can be called and it will return any queued up events.  When
// the main thread wishes to close the connection to the router the stop method
// can be called and it will return any unread events.
func (r CrosswordRoute) SSE(url string, router *gin.Engine) (flush func() []pubsub.Event, stop func() []pubsub.Event) {
	url = path.Join("/crossword", url)
	recorder := CreateTestResponseRecorder()

	flush = func() []pubsub.Event {
		// Give the router a chance to write everything it needs to.
		time.Sleep(10 * time.Millisecond)

		messages, _ := sse.Decode(recorder.Body)

		var events []pubsub.Event
		for _, message := range messages {
			payload := message.Data.(string)

			var event pubsub.Event
			json.Unmarshal([]byte(payload), &event)
			events = append(events, event)
		}

		return events
	}

	stop = func() []pubsub.Event {
		// Give the router a chance to write everything it needs to.
		time.Sleep(10 * time.Millisecond)

		recorder.closeClient()
		return flush()
	}

	request := httptest.NewRequest(http.MethodGet, url, nil)
	go router.ServeHTTP(recorder, request)

	return flush, stop
}

// ChannelRoute is a client that makes requests against the URL of a particular
// user's channel.
type ChannelRoute struct {
	channel string
}

func (r ChannelRoute) GET(url string, router *gin.Engine) *TestResponseRecorder {
	url = path.Join(r.channel, url)
	return Global.GET(url, router)
}

func (r ChannelRoute) PUT(url, body string, router *gin.Engine) *TestResponseRecorder {
	url = path.Join(r.channel, url)
	return Global.PUT(url, body, router)
}

func (r ChannelRoute) SSE(url string, router *gin.Engine) (flush func() []pubsub.Event, stop func() []pubsub.Event) {
	url = path.Join(r.channel, url)
	return Global.SSE(url, router)
}

func RandomString(n int) string {
	var alphabet = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")

	bs := make([]rune, n)
	for i := range bs {
		bs[i] = alphabet[rand.Intn(len(alphabet))]
	}

	return string(bs)
}

// We wrap httptest.ResponseRecorder so that our version implements the
// http.CloseNotifier interface required by Gin.
type TestResponseRecorder struct {
	*httptest.ResponseRecorder
	closeChannel chan bool
}

func (r *TestResponseRecorder) CloseNotify() <-chan bool {
	return r.closeChannel
}

func (r *TestResponseRecorder) closeClient() {
	r.closeChannel <- true
}

func (r *TestResponseRecorder) JSON(target interface{}) error {
	return json.NewDecoder(r.Body).Decode(target)
}

func CreateTestResponseRecorder() *TestResponseRecorder {
	return &TestResponseRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}
}
