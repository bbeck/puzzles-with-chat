package crossword

import (
	"encoding/json"
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

func TestRoute_UpdateSetting(t *testing.T) {
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
	response := PUT("/setting/only_allow_correct_answers", `true`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	verify(func(s *Settings) { assert.True(t, s.OnlyAllowCorrectAnswers) })

	response = PUT("/setting/clues_to_show", `"none"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	verify(func(s *Settings) { assert.Equal(t, NoCluesVisible, s.CluesToShow) })

	response = PUT("/setting/clue_font_size", `"xlarge"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	verify(func(s *Settings) { assert.Equal(t, SizeXLarge, s.ClueFontSize) })
}

func TestRoute_UpdateSetting_ClearsIncorrectCells(t *testing.T) {
	// This acts as a small integration test toggling the OnlyAllowCorrectAnswers
	// setting and ensuring that it clears any incorrect answer cells.
	pool, conn, cleanup := NewRedisPool(t)
	defer cleanup()

	registry, events, cleanup := NewRegistry(t)
	defer cleanup()

	// Setup a cached entry for the date we're about to load to ensure that we
	// don't make a network call during the test.
	saved := XWordInfoPuzzleCache
	XWordInfoPuzzleCache = map[string]*Puzzle{
		"2018-12-31": LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
	}
	defer func() { XWordInfoPuzzleCache = saved }()

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

	response := PUT("/date", `{"date": "2018-12-31"}`, router)
	require.Equal(t, http.StatusOK, response.Code)
	drainEvents()

	// Toggle the status, it should transition to solving.
	response = PUT("/status", ``, router)
	require.Equal(t, http.StatusOK, response.Code)
	drainEvents()

	// Apply an incorrect across answer.
	response = PUT("/answer/1a", `"QNORA"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	drainEvents()

	// Set the OnlyAllowCorrectAnswers setting to true
	response = PUT("/setting/only_allow_correct_answers", `true`, router)
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

func TestRoute_UpdateSettings_Error(t *testing.T) {
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pool, _, cleanup := NewRedisPool(t)
			defer cleanup()

			router := gin.Default()
			RegisterRoutes(router, pool)

			response := PUT(fmt.Sprintf("/setting/%s", test.setting), test.json, router)
			assert.NotEqual(t, http.StatusOK, response.Code)
		})
	}
}

func TestRoute_UpdateCrosswordDate(t *testing.T) {
	// This acts as a small integration test updating the date of the crossword
	// we're working on and ensuring the proper values are written to the
	// database.
	pool, conn, cleanup := NewRedisPool(t)
	defer cleanup()

	registry, events, cleanup := NewRegistry(t)
	defer cleanup()

	// Setup a cached entry for the date we're about to load to ensure that we
	// don't make a network call during the test.
	saved := XWordInfoPuzzleCache
	XWordInfoPuzzleCache = map[string]*Puzzle{
		"2018-12-31": LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
	}
	defer func() { XWordInfoPuzzleCache = saved }()

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

	response := PUT("/date", `{"date": "2018-12-31"}`, router)
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

func TestRoute_UpdateCrosswordDate_Error(t *testing.T) {
	tests := []struct {
		name     string
		payload  string
		cache    map[string]*Puzzle
		expected int
	}{
		{
			name:     "malformed payload",
			payload:  `{"date": }`,
			expected: http.StatusBadRequest,
		},
		{
			name:     "bad date", // causes LoadFromNewYorkTimes to return an error
			payload:  `{"date": "bad"}`,
			cache:    map[string]*Puzzle{"bad": nil},
			expected: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pool, _, cleanup := NewRedisPool(t)
			defer cleanup()

			saved := XWordInfoPuzzleCache
			XWordInfoPuzzleCache = test.cache
			defer func() { XWordInfoPuzzleCache = saved }()

			router := gin.Default()
			RegisterRoutes(router, pool)

			response := PUT("/date", test.payload, router)
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

	// Setup a cached entry for the date we're about to load to ensure that we
	// don't make a network call during the test.
	saved := XWordInfoPuzzleCache
	XWordInfoPuzzleCache = map[string]*Puzzle{
		"2018-12-31": LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
	}
	defer func() { XWordInfoPuzzleCache = saved }()

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

	response := PUT("/date", `{"date": "2018-12-31"}`, router)
	require.Equal(t, http.StatusOK, response.Code)
	drainEvents()

	// Toggle the status, it should transition to solving.
	response = PUT("/status", ``, router)
	assert.Equal(t, http.StatusOK, response.Code)
	verify(func(state *State) {
		assert.Equal(t, StatusSolving, state.Status)
		assert.NotNil(t, state.LastStartTime)
	})

	// Toggle the status again, it should transition to paused. Make sure we
	// sleep for at least a nanosecond first so that the solve was unpaused for
	// a non-zero duration.
	time.Sleep(1 * time.Nanosecond)
	response = PUT("/status", ``, router)
	assert.Equal(t, http.StatusOK, response.Code)
	verify(func(state *State) {
		assert.Equal(t, StatusPaused, state.Status)
		assert.Nil(t, state.LastStartTime)
		assert.True(t, state.TotalSolveDuration.Seconds() > 0.)
	})

	// Toggle the status again, it should transition back to solving.
	response = PUT("/status", ``, router)
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
	response = PUT("/status", ``, router)
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

	// Setup a cached entry for the date we're about to load to ensure that we
	// don't make a network call during the test.
	saved := XWordInfoPuzzleCache
	XWordInfoPuzzleCache = map[string]*Puzzle{
		"2018-12-31": LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
	}
	defer func() { XWordInfoPuzzleCache = saved }()

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

	response := PUT("/date", `{"date": "2018-12-31"}`, router)
	require.Equal(t, http.StatusOK, response.Code)
	drainEvents()

	// Toggle the status, it should transition to solving.
	response = PUT("/status", ``, router)
	require.Equal(t, http.StatusOK, response.Code)
	drainEvents()

	// Apply a correct across answer.
	response = PUT("/answer/1a", `"QANDA"`, router)
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
	response = PUT("/answer/1d", `"QTIP"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	verify(func(state *State) {
		assert.True(t, state.DownCluesFilled[1])
		assert.Equal(t, "Q", state.Cells[0][0])
		assert.Equal(t, "T", state.Cells[1][0])
		assert.Equal(t, "I", state.Cells[2][0])
		assert.Equal(t, "P", state.Cells[3][0])
	})

	// Apply an incorrect across answer.
	response = PUT("/answer/6a", `"FLOOR"`, router)
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
	response = PUT("/answer/11d", `"HEYA"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	verify(func(state *State) {
		assert.True(t, state.DownCluesFilled[11])
		assert.Equal(t, "H", state.Cells[0][12])
		assert.Equal(t, "E", state.Cells[1][12])
		assert.Equal(t, "Y", state.Cells[2][12])
		assert.Equal(t, "A", state.Cells[3][12])
	})

	// Pause the solve.
	response = PUT("/status", ``, router)
	require.Equal(t, http.StatusOK, response.Code)
	drainEvents()

	// Try to apply an answer.
	response = PUT("/answer/6a", `"ATTIC"`, router)
	assert.Equal(t, http.StatusConflict, response.Code)
}

func TestRoute_UpdateCrosswordAnswer_OnlyAllowCorrectAnswers(t *testing.T) {
	// This acts as a small integration test toggling the status of a crossword
	// being solved.
	pool, conn, cleanup := NewRedisPool(t)
	defer cleanup()

	registry, events, cleanup := NewRegistry(t)
	defer cleanup()

	// Setup a cached entry for the date we're about to load to ensure that we
	// don't make a network call during the test.
	saved := XWordInfoPuzzleCache
	XWordInfoPuzzleCache = map[string]*Puzzle{
		"2018-12-31": LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
	}
	defer func() { XWordInfoPuzzleCache = saved }()

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

	response := PUT("/date", `{"date": "2018-12-31"}`, router)
	require.Equal(t, http.StatusOK, response.Code)
	drainEvents()

	// Toggle the status, it should transition to solving.
	response = PUT("/status", ``, router)
	require.Equal(t, http.StatusOK, response.Code)
	drainEvents()

	// Apply a correct across answer.
	response = PUT("/answer/1a", `"QANDA"`, router)
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
	response = PUT("/answer/1d", `"QTIP"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	verify(func(state *State) {
		assert.True(t, state.DownCluesFilled[1])
		assert.Equal(t, "Q", state.Cells[0][0])
		assert.Equal(t, "T", state.Cells[1][0])
		assert.Equal(t, "I", state.Cells[2][0])
		assert.Equal(t, "P", state.Cells[3][0])
	})

	// Apply an incorrect across answer.
	response = PUT("/answer/6a", `"FLOOR"`, router)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	state, err := GetState(conn, "channel")
	require.NoError(t, err)
	assert.False(t, state.AcrossCluesFilled[6])

	// Apply an incorrect down answer.
	response = PUT("/answer/11d", `"HEYA"`, router)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	state, err = GetState(conn, "channel")
	require.NoError(t, err)
	assert.False(t, state.DownCluesFilled[11])
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

			saved := XWordInfoPuzzleCache
			XWordInfoPuzzleCache = map[string]*Puzzle{
				"2018-12-31": LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
			}
			defer func() { XWordInfoPuzzleCache = saved }()

			router := gin.Default()
			RegisterRoutes(router, pool)

			response := PUT("/date", `{"date": "2018-12-31"}`, router)
			require.Equal(t, http.StatusOK, response.Code)

			// Toggle the status, it should transition to solving.
			response = PUT("/status", ``, router)
			require.Equal(t, http.StatusOK, response.Code)

			response = PUT("/answer/1a", test.json, router)
			assert.Equal(t, test.expected, response.Code)
		})
	}
}

func TestRoute_GetCrosswordEvents(t *testing.T) {
	// This acts as a small integration test ensuring that the event stream
	// receives the events put into a registry.
	pool, _, cleanup := NewRedisPool(t)
	defer cleanup()

	registry, _, cleanup := NewRegistry(t)
	defer cleanup()

	// Setup a cached entry for the date we're about to load to ensure that we
	// don't make a network call during the test.
	saved := XWordInfoPuzzleCache
	XWordInfoPuzzleCache = map[string]*Puzzle{
		"2018-12-31": LoadTestPuzzle(t, "xwordinfo-success-20181231.json"),
	}
	defer func() { XWordInfoPuzzleCache = saved }()

	router := gin.Default()
	RegisterRoutesWithRegistry(router, pool, registry)

	// Connect to the stream when there's no puzzle selected, we should receive
	// just the channel's settings.
	_, stop := SSE("/events", router)
	events := stop()
	assert.Equal(t, 1, len(events))
	assert.Equal(t, "settings", events[0].Kind)

	// Select a puzzle
	response := PUT("/date", `{"date": "2018-12-31"}`, router)
	require.Equal(t, http.StatusOK, response.Code)

	// Now reconnect to the stream and we should receive both the puzzle and the
	// channel's current state.
	flush, stop := SSE("/events", router)
	events = flush()
	assert.Equal(t, 2, len(events))
	assert.Equal(t, "settings", events[0].Kind)
	assert.Equal(t, "state", events[1].Kind)

	// Toggle the status to solving, this should cause the state to be sent again.
	response = PUT("/status", ``, router)
	require.Equal(t, http.StatusOK, response.Code)

	events = flush()
	assert.Equal(t, 1, len(events))
	assert.Equal(t, "state", events[0].Kind)

	// Now specify an answer, this should cause the state to be sent again.
	response = PUT("/answer/1a", `"QANDA"`, router)
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

func PUT(url, body string, router *gin.Engine) *TestResponseRecorder {
	url = path.Join("/channel/crossword", url)
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
func SSE(url string, router *gin.Engine) (flush func() []pubsub.Event, stop func() []pubsub.Event) {
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

	url = path.Join("/channel/crossword", url)
	request := httptest.NewRequest(http.MethodGet, url, nil)
	go router.ServeHTTP(recorder, request)

	return flush, stop
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

func CreateTestResponseRecorder() *TestResponseRecorder {
	return &TestResponseRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}
}
