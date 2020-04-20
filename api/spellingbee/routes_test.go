package spellingbee

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bbeck/twitch-plays-crosswords/api/model"
	"github.com/bbeck/twitch-plays-crosswords/api/pubsub"
	"github.com/go-chi/chi"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"sync"
	"testing"
	"time"
)

var Global = GlobalRoute{}
var Channel = ChannelRoute{name: "channel"}

func TestRoute_GetChannels(t *testing.T) {
	// This acts as a small integration test ensuring that the active channels
	// event stream receives the events as new channels start and finish solves.
	router, pool, _ := NewTestRouter(t)
	conn := NewRedisConnection(t, pool)

	// Connect to the stream when there's no active solves happening, we should
	// receive an event that contains an empty list of channels.
	_, stop := Global.SSE("/channels", router)
	events := stop()
	assert.Equal(t, 1, len(events))
	assert.Equal(t, "channels", events[0].Kind)
	assert.Empty(t, events[0].Payload)

	// Start a puzzle in the first channel.
	state := State{
		Status: model.StatusSolving,
		Puzzle: LoadTestPuzzle(t, "nytbee-20200408.html"),
		Words:  []string{},
	}
	require.NoError(t, SetState(conn, Channel.name, state))

	// Now reconnect to the stream and we should receive one active channel.
	_, stop = Global.SSE("/channels", router)
	events = stop()
	assert.Equal(t, 1, len(events))
	assert.Equal(t, "channels", events[0].Kind)
	assert.ElementsMatch(t, []string{Channel.name}, events[0].Payload)

	// Start a puzzle on another channel.
	state = State{
		Status: model.StatusSolving,
		Puzzle: LoadTestPuzzle(t, "nytbee-20200408.html"),
		Words:  []string{},
	}
	require.NoError(t, SetState(conn, "channel2", state))

	// Now we expect there to be 2 channels in the stream.
	_, stop = Global.SSE("/channels", router)
	events = stop()
	assert.Equal(t, 1, len(events))
	assert.Equal(t, "channels", events[0].Kind)
	assert.ElementsMatch(t, []string{Channel.name, "channel2"}, events[0].Payload)

	// Lastly remove the second channel from the database.
	_, err := conn.Do("DEL", StateKey("channel2"))
	require.NoError(t, err)

	// Now we expect there to be one channels in the stream.
	_, stop = Global.SSE("/channels", router)
	events = stop()
	assert.Equal(t, 1, len(events))
	assert.Equal(t, "channels", events[0].Kind)
	assert.ElementsMatch(t, []string{Channel.name}, events[0].Payload)
}

func TestRoute_UpdatePuzzle_NYTBee(t *testing.T) {
	// This acts as a small integration test updating the date of the NYTBee
	// puzzle we're working on and ensuring the proper values are written
	// to the database.
	router, pool, registry := NewTestRouter(t)
	events := NewEventSubscription(t, registry, Channel.name)

	// Force a specific puzzle to be loaded so we don't make a network call.
	ForcePuzzleToBeLoaded(t, "nytbee-20200408.html")

	response := Channel.PUT("/", `{"nytbee": "2020-04-08"}`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.Equal(t, model.StatusSelected, state.Status)
		assert.NotNil(t, state.Puzzle)
		assert.Nil(t, state.LastStartTime)
		assert.Equal(t, 0., state.TotalSolveDuration.Seconds())
	})
}

func TestRoute_UpdatePuzzle_Error(t *testing.T) {
	tests := []struct {
		name        string
		payload     string
		forcedError error
		expected    int
	}{
		{
			name:     "bad json",
			payload:  `{"nytbee": }`,
			expected: http.StatusBadRequest,
		},
		{
			name:     "invalid json",
			payload:  `{}`,
			expected: http.StatusBadRequest,
		},
		{
			name:        "nytbee error",
			payload:     `{"nytbee": "bad"}`,
			forcedError: errors.New("forced error"),
			expected:    http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router, _, _ := NewTestRouter(t)
			ForceErrorDuringLoad(t, test.forcedError)

			response := Channel.PUT("/", test.payload, router)
			assert.Equal(t, test.expected, response.Code)
		})
	}
}

func TestRoute_UpdateSetting(t *testing.T) {
	// This acts as a small integration test updating each setting in turn and
	// making sure the proper value is written to the database and that clients
	// receive events notifying them of the setting change.
	router, pool, registry := NewTestRouter(t)
	events := NewEventSubscription(t, registry, Channel.name)

	// Update each setting, one at a time.
	response := Channel.PUT("/setting/allow_unofficial_answers", `true`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifySettings(t, pool, events, func(s Settings) {
		assert.True(t, s.AllowUnofficialAnswers)
	})

	response = Channel.PUT("/setting/font_size", `"xlarge"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifySettings(t, pool, events, func(s Settings) {
		assert.Equal(t, model.FontSizeXLarge, s.FontSize)
	})
}

func TestRoute_UpdateSetting_ClearsUnofficialAnswers(t *testing.T) {
	// This acts as a small integration test toggling the AllowUnofficialAnswers
	// setting and ensuring that it removes any unofficial answers.
	router, pool, registry := NewTestRouter(t)
	conn := NewRedisConnection(t, pool)
	events := NewEventSubscription(t, registry, Channel.name)

	// Setup the state with some unofficial answers in it.
	state := State{
		Status: model.StatusSolving,
		Puzzle: LoadTestPuzzle(t, "nytbee-20200408.html"),
		Words: []string{
			"COCONUT",
			"CONCOCT",
			"CONCOCTOR",
			"CONTO",
		},
	}
	require.NoError(t, SetState(conn, Channel.name, state))

	// Set the AllowUnofficialAnswers setting to false
	response := Channel.PUT("/setting/allow_unofficial_answers", `false`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		expected := []string{"COCONUT", "CONCOCT"}
		assert.ElementsMatch(t, expected, state.Words)
	})
}

func TestRoute_UpdateSetting_Error(t *testing.T) {
	tests := []struct {
		name    string
		setting string
		json    string
	}{
		{
			name:    "allow_unofficial_answers",
			setting: "allow_unofficial_answers",
			json:    `{`,
		},
		{
			name:    "font_size",
			setting: "font_size",
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
			router, _, _ := NewTestRouter(t)

			response := Channel.PUT(fmt.Sprintf("/setting/%s", test.setting), test.json, router)
			assert.NotEqual(t, http.StatusOK, response.Code)
		})
	}
}

func TestRoute_ToggleStatus(t *testing.T) {
	// This acts as a small integration test toggling the status of a spelling bee
	// puzzle being solved.
	router, pool, registry := NewTestRouter(t)
	conn := NewRedisConnection(t, pool)
	events := NewEventSubscription(t, registry, Channel.name)

	// Start a puzzle on another channel in the selected state.
	state := State{
		Status: model.StatusSelected,
		Puzzle: LoadTestPuzzle(t, "nytbee-20200408.html"),
	}
	require.NoError(t, SetState(conn, Channel.name, state))

	// Toggle the status, it should transition to solving.
	response := Channel.PUT("/status", ``, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.Equal(t, model.StatusSolving, state.Status)
		assert.NotNil(t, state.LastStartTime)
	})

	// Toggle the status again, it should transition to paused. Make sure we
	// sleep for at least a nanosecond first so that the solve was unpaused for
	// a non-zero duration.
	time.Sleep(1 * time.Nanosecond)
	response = Channel.PUT("/status", ``, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.Equal(t, model.StatusPaused, state.Status)
		assert.Nil(t, state.LastStartTime)
		assert.True(t, state.TotalSolveDuration.Seconds() > 0.)
	})

	// Toggle the status again, it should transition back to solving.
	response = Channel.PUT("/status", ``, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.Equal(t, model.StatusSolving, state.Status)
		assert.NotNil(t, state.LastStartTime)
		assert.True(t, state.TotalSolveDuration.Seconds() > 0.)
	})

	// Force the puzzle to be complete.
	state, err := GetState(conn, Channel.name)
	require.NoError(t, err)
	state.Status = model.StatusComplete
	require.NoError(t, SetState(conn, Channel.name, state))

	// Try to toggle the status one more time.  Now that the puzzle is complete
	// it should return a HTTP error.
	response = Channel.PUT("/status", ``, router)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	state, err = GetState(conn, "channel")
	require.NoError(t, err)
	assert.Equal(t, model.StatusComplete, state.Status)
}

func TestRoute_AddAnswer_NoUnofficialAnswers(t *testing.T) {
	// This acts as a small integration test of adding answers to a spelling bee
	// puzzle being solved.
	router, pool, registry := NewTestRouter(t)
	conn := NewRedisConnection(t, pool)
	events := NewEventSubscription(t, registry, Channel.name)

	state := State{
		Status: model.StatusSelected,
		Puzzle: LoadTestPuzzle(t, "nytbee-20200408.html"),
	}
	require.NoError(t, SetState(conn, Channel.name, state))

	// Apply a correct answer before the puzzle has been started, should get an
	// error.
	response := Channel.POST("/answer", `"COCONUT"`, router)
	assert.Equal(t, http.StatusConflict, response.Code)

	// Transition to solving.
	state.Status = model.StatusSolving
	require.NoError(t, SetState(conn, Channel.name, state))

	// Now applying the answer should succeed.
	response = Channel.POST("/answer", `"COCONUT"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.Contains(t, state.Words, "COCONUT")
	})

	// Applying an incorrect answer should fail.
	response = Channel.POST("/answer", `"CCCC"`, router)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	// Applying an unofficial answer should fail.
	response = Channel.POST("/answer", `"CONCOCTOR"`, router)
	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestRoute_AddAnswer_AllowUnofficialAnswers(t *testing.T) {
	// This acts as a small integration test of adding answers to a spelling bee
	// puzzle being solved.
	router, pool, registry := NewTestRouter(t)
	conn := NewRedisConnection(t, pool)
	events := NewEventSubscription(t, registry, Channel.name)

	settings := Settings{
		AllowUnofficialAnswers: true,
	}
	require.NoError(t, SetSettings(conn, Channel.name, settings))

	state := State{
		Status: model.StatusSolving,
		Puzzle: LoadTestPuzzle(t, "nytbee-20200408.html"),
	}
	require.NoError(t, SetState(conn, Channel.name, state))

	// Applying an answer from the official list should succeed.
	response := Channel.POST("/answer", `"COCONUT"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.Contains(t, state.Words, "COCONUT")
	})

	// Applying an answer from the unofficial list should also succeed.
	response = Channel.POST("/answer", `"CONCOCTOR"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.Contains(t, state.Words, "CONCOCTOR")
	})

	// Applying an incorrect answer should fail.
	response = Channel.POST("/answer", `"CCCC"`, router)
	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestRoute_AddAnswer_SolvedPuzzleStopsTimer(t *testing.T) {
	// This acts as a small integration test ensuring that the timer stops
	// counting once the puzzle has been solved.
	router, pool, registry := NewTestRouter(t)
	conn := NewRedisConnection(t, pool)
	events := NewEventSubscription(t, registry, Channel.name)

	// Set the state to have all of the words except for one.
	now := time.Now()
	state := State{
		Status: model.StatusSolving,
		Puzzle: LoadTestPuzzle(t, "nytbee-20200408.html"),
		Words: []string{
			"CONCOCT",
			"CONTORT",
			"CONTOUR",
			"COOT",
			"COTTON",
			"COTTONY",
			"COUNT",
			"COUNTRY",
			"COUNTY",
			"COURT",
			"CROUTON",
			"CURT",
			"CUTOUT",
			"NUTTY",
			"ONTO",
			"OUTCRY",
			"OUTRO",
			"OUTRUN",
			"ROOT",
			"ROTO",
			"ROTOR",
			"ROUT",
			"RUNOUT",
			"RUNT",
			"RUNTY",
			"RUTTY",
			"TONY",
			"TOON",
			"TOOT",
			"TORN",
			"TORO",
			"TORT",
			"TOUR",
			"TOUT",
			"TROT",
			"TROUT",
			"TROY",
			"TRYOUT",
			"TURN",
			"TURNOUT",
			"TUTOR",
			"TUTU",
			"TYCOON",
			"TYRO",
			"UNCUT",
			"UNTO",
			"YURT",
		},
		LastStartTime: &now,
	}
	require.NoError(t, SetState(conn, Channel.name, state))

	// Apply the last answer, but wait a bit first to ensure that a non-zero
	// amount of time has passed in the solve.
	time.Sleep(2 * time.Millisecond)

	response := Channel.POST("/answer", `"COCONUT"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		require.Equal(t, model.StatusComplete, state.Status)
		assert.Nil(t, state.LastStartTime)
		assert.True(t, state.TotalSolveDuration.Seconds() > 0)
	})

}

func TestRoute_AddAnswer_Error(t *testing.T) {
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
			name:     "long answer",
			json:     `"` + RandomString(1023) + `"`,
			expected: http.StatusRequestEntityTooLarge,
		},
		{
			name:     "non-string answer",
			json:     `true`,
			expected: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router, pool, _ := NewTestRouter(t)
			conn := NewRedisConnection(t, pool)

			state := State{
				Status: model.StatusSolving,
				Puzzle: LoadTestPuzzle(t, "nytbee-20200408.html"),
			}
			require.NoError(t, SetState(conn, Channel.name, state))

			response := Channel.POST("/answer", test.json, router)
			assert.Equal(t, test.expected, response.Code)
		})
	}
}

// VerifySettings performs test specific verifications on the settings objects
// in both event and database forms.
func VerifySettings(t *testing.T, pool *redis.Pool, events <-chan pubsub.Event, fn func(s Settings)) {
	t.Helper()

	// First check that we've received an event with the correct value
	select {
	case event := <-events:
		// Ignore any non-settings events.
		if event.Kind != "settings" {
			return
		}
		fn(event.Payload.(Settings))

	default:
		assert.Fail(t, "no settings event available")
	}

	// Next check that the database has a valid settings object
	conn := NewRedisConnection(t, pool)
	settings, err := GetSettings(conn, "channel")
	require.NoError(t, err)
	fn(settings)
}

// VerifyState performs common verifications for state objects in both event
// and database forms and then calls a custom verify function for test specific
// verifications.
func VerifyState(t *testing.T, pool *redis.Pool, events <-chan pubsub.Event, fn func(State)) {
	t.Helper()

	// First check that we've received an event with the correct value
	select {
	case event := <-events:
		// Ignore any non-state events.
		if event.Kind != "state" {
			return
		}
		state := event.Payload.(State)

		// Event should never have the answers
		assert.Nil(t, state.Puzzle.OfficialAnswers)
		assert.Nil(t, state.Puzzle.UnofficialAnswers)
		fn(state)

	default:
		assert.Fail(t, "no state event available")
	}

	// Next check that the database has a valid settings object
	state, err := GetState(NewRedisConnection(t, pool), "channel")
	require.NoError(t, err)

	// Database should always have the answers
	assert.NotNil(t, state.Puzzle.OfficialAnswers)
	assert.NotNil(t, state.Puzzle.UnofficialAnswers)
	fn(state)
}

// GlobalRoute is a client that makes requests against the URL of the global
// spelling bee route, not associated with any particular channel.
type GlobalRoute struct{}

func (r GlobalRoute) GET(url string, router chi.Router) *httptest.ResponseRecorder {
	url = path.Join("/spellingbee", url)
	request := httptest.NewRequest(http.MethodGet, url, nil)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func (r GlobalRoute) PUT(url, body string, router chi.Router) *httptest.ResponseRecorder {
	url = path.Join("/spellingbee", url)
	request := httptest.NewRequest(http.MethodPut, url, strings.NewReader(body))

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func (r GlobalRoute) POST(url, body string, router chi.Router) *httptest.ResponseRecorder {
	url = path.Join("/spellingbee", url)
	request := httptest.NewRequest(http.MethodPost, url, strings.NewReader(body))

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

// SSE performs a streaming request to the provided router.  Because the router
// won't immediately return, this request is done in a background goroutine.
// When the main thread wishes to read events that have been received thus far
// the flush method can be called and it will return any queued up events.  When
// the main thread wishes to close the connection to the router the stop method
// can be called and it will return any unread events.
func (r GlobalRoute) SSE(url string, router chi.Router) (flush func() []pubsub.Event, stop func() []pubsub.Event) {
	url = path.Join("/spellingbee", url)
	recorder := CreateTestResponseRecorder()

	flush = func() []pubsub.Event {
		// Give the router a chance to write everything it needs to.
		time.Sleep(10 * time.Millisecond)

		reader, err := recorder.Body()
		if err != nil {
			return nil
		}

		var events []pubsub.Event
		for {
			bs, err := reader.ReadBytes('\n')
			if err != nil {
				break
			}

			if !bytes.HasPrefix(bs, []byte("data:")) {
				continue
			}

			var event pubsub.Event
			json.Unmarshal(bs[5:], &event)
			events = append(events, event)
		}

		return events
	}

	stop = func() []pubsub.Event {
		// Give the router a chance to write everything it needs to.
		time.Sleep(10 * time.Millisecond)

		recorder.Close()
		return flush()
	}

	request := httptest.NewRequest(http.MethodGet, url, nil)
	go router.ServeHTTP(recorder, request)

	return flush, stop
}

// ChannelRoute is a client that makes requests against the URL of a particular
// user's channel.
type ChannelRoute struct {
	name string
}

func (r ChannelRoute) GET(url string, router chi.Router) *httptest.ResponseRecorder {
	url = path.Join(r.name, url)
	return Global.GET(url, router)
}

func (r ChannelRoute) PUT(url, body string, router chi.Router) *httptest.ResponseRecorder {
	url = path.Join(r.name, url)
	return Global.PUT(url, body, router)
}

func (r ChannelRoute) POST(url, body string, router chi.Router) *httptest.ResponseRecorder {
	url = path.Join(r.name, url)
	return Global.POST(url, body, router)
}

func (r ChannelRoute) SSE(url string, router chi.Router) (flush func() []pubsub.Event, stop func() []pubsub.Event) {
	url = path.Join(r.name, url)
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

// Create a http.ResponseWriter that synchronizes whenever reads or writes
// happen so that there are no races in a multiple goroutine environment.
// Additionally implement the http.CloseNotifier interface so that requests can
// be stopped by tests.
type TestResponseRecorder struct {
	sync.Mutex
	headers http.Header
	body    *bytes.Buffer
	close   chan bool
}

func CreateTestResponseRecorder() *TestResponseRecorder {
	return &TestResponseRecorder{
		headers: make(http.Header),
		body:    new(bytes.Buffer),
		close:   make(chan bool, 1),
	}
}

func (r *TestResponseRecorder) Header() http.Header {
	r.Lock()
	defer r.Unlock()

	return r.headers
}

func (r *TestResponseRecorder) Write(bs []byte) (int, error) {
	r.Lock()
	defer r.Unlock()

	return r.body.Write(bs)
}

func (r *TestResponseRecorder) Body() (*bufio.Reader, error) {
	r.Lock()
	defer r.Unlock()

	bs, err := ioutil.ReadAll(r.body)
	if err != nil {
		return nil, err
	}
	r.body.Reset()
	return bufio.NewReader(bytes.NewReader(bs)), nil
}

func (r *TestResponseRecorder) CloseNotify() <-chan bool {
	r.Lock()
	defer r.Unlock()

	return r.close
}

func (r *TestResponseRecorder) Close() {
	r.Lock()
	defer r.Unlock()

	r.close <- true
}

func (r *TestResponseRecorder) WriteHeader(int) {
	// Not used
}
