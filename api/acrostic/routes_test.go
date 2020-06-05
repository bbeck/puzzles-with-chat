package acrostic

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"github.com/bbeck/puzzles-with-chat/api/model"
	"github.com/bbeck/puzzles-with-chat/api/pubsub"
	"github.com/go-chi/chi"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"sync"
	"testing"
	"time"
)

var Channel = ChannelClient{name: "channel"}

func TestRoute_UpdatePuzzle_NewYorkTimes(t *testing.T) {
	// This acts as a small integration test updating the date of the New York
	// Times acrostic we're working on and ensuring the proper values are written
	// to the database.
	router, pool, registry := NewTestRouter(t)
	events := NewEventSubscription(t, registry, Channel.name)

	// Force a specific puzzle to be loaded so we don't make a network call.
	ForcePuzzleToBeLoaded(t, "xwordinfo-nyt-20200524.json")

	response := Channel.PUT("/", `{"new_york_times_date": "2020-05-24"}`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.Equal(t, model.StatusSelected, state.Status)
		assert.NotNil(t, state.Puzzle)
		assert.Equal(t, 0, len(state.CluesFilled))
		assert.Nil(t, state.LastStartTime)
		assert.Equal(t, 0., state.TotalSolveDuration.Seconds())
	})
}

func TestRoute_UpdatePuzzle_JSONError(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected int
	}{
		{
			name:     "bad json",
			json:     `{"new_york_times_date": }`,
			expected: http.StatusBadRequest,
		},
		{
			name:     "invalid json",
			json:     `{}`,
			expected: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router, _, _ := NewTestRouter(t)
			ForcePuzzleToBeLoaded(t, "xwordinfo-nyt-20200524.json")

			response := Channel.PUT("/", test.json, router)
			assert.Equal(t, test.expected, response.Code)
		})
	}
}

func TestRoute_UpdatePuzzle_LoadSaveError(t *testing.T) {
	tests := []struct {
		name                 string
		json                 string
		forcePuzzleLoadError error
		forceStateSaveError  error
		expected             int
	}{
		{
			name:                 "nyt error loading puzzle",
			json:                 `{"new_york_times_date": "unused"}`,
			forcePuzzleLoadError: errors.New("forced error"),
			expected:             http.StatusInternalServerError,
		},
		{
			name:                "error saving state",
			json:                `{"new_york_times_date": "unused"}`,
			forceStateSaveError: errors.New("forced error"),
			expected:            http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router, _, _ := NewTestRouter(t)

			if test.forcePuzzleLoadError != nil {
				ForceErrorDuringPuzzleLoad(t, test.forcePuzzleLoadError)
			} else {
				ForcePuzzleToBeLoaded(t, "xwordinfo-nyt-20200524.json")
			}

			ForceErrorDuringStateSave(t, test.forceStateSaveError)

			response := Channel.PUT("/", test.json, router)
			assert.Equal(t, test.expected, response.Code)
		})
	}
}

func TestRoute_ToggleStatus(t *testing.T) {
	// This acts as a small integration test toggling the status of an acrostic
	// being solved.
	router, pool, registry := NewTestRouter(t)
	conn := NewRedisConnection(t, pool)
	events := NewEventSubscription(t, registry, Channel.name)

	// Set a state that has a puzzle selected but not yet started.
	state := NewState(t, "xwordinfo-nyt-20200524.json")
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
	state.Status = model.StatusComplete
	require.NoError(t, SetState(conn, Channel.name, state))

	// Try to toggle the status one more time.  Now that the puzzle is complete
	// it should return an HTTP error.
	response = Channel.PUT("/status", ``, router)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	state, err := GetState(conn, Channel.name)
	require.NoError(t, err)
	assert.Equal(t, model.StatusComplete, state.Status)
}

func TestRoute_ToggleStatus_Error(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  model.Status
		loadStateError error
		saveStateError error
	}{
		{
			name:          "status created",
			initialStatus: model.StatusCreated,
		},
		{
			name:           "error loading state",
			initialStatus:  model.StatusSelected,
			loadStateError: errors.New("forced error"),
		},
		{
			name:           "error saving state",
			initialStatus:  model.StatusSelected,
			saveStateError: errors.New("forced error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router, pool, _ := NewTestRouter(t)
			conn := NewRedisConnection(t, pool)

			state := NewState(t, "xwordinfo-nyt-20200524.json")
			state.Status = test.initialStatus
			require.NoError(t, SetState(conn, Channel.name, state))

			if test.loadStateError != nil {
				ForceErrorDuringStateLoad(t, test.loadStateError)
			}

			if test.saveStateError != nil {
				ForceErrorDuringStateSave(t, test.saveStateError)
			}

			response := Channel.PUT("/status", "", router)
			assert.NotEqual(t, http.StatusOK, response.Code)
		})
	}
}

// VerifyState performs common verifications for state objects in both event
// and database forms and then calls a custom verify function for test specific
// verifications.
func VerifyState(t *testing.T, pool *redis.Pool, events <-chan pubsub.Event, fn func(s State)) {
	t.Helper()

	// First check that we've received an event with the correct value
	select {
	case event := <-events:
		// Ignore any non-state events.
		if event.Kind != "state" {
			return
		}

		state := event.Payload.(State)
		assert.Nil(t, state.Puzzle.Cells) // Events should never have the solution
		fn(state)

	default:
		assert.Fail(t, "no state event available")
	}

	// Next check that the database has a valid settings object
	conn := NewRedisConnection(t, pool)
	state, err := GetState(conn, Channel.name)
	require.NoError(t, err)
	assert.NotNil(t, state.Puzzle.Cells) // Database should always have the solution
	fn(state)
}

// ChannelClient is a client that makes requests against the URL of a particular
// user's channel.
type ChannelClient struct {
	name string
}

func (c ChannelClient) GET(url string, router chi.Router) *httptest.ResponseRecorder {
	url = path.Join("/acrostic", c.name, url)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, url, nil)
	router.ServeHTTP(recorder, request)
	return recorder
}

func (c ChannelClient) PUT(url, body string, router chi.Router) *httptest.ResponseRecorder {
	url = path.Join("/acrostic", c.name, url)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, url, strings.NewReader(body))
	router.ServeHTTP(recorder, request)
	return recorder
}

// SSE performs a streaming request to the provided router.  Because the router
// won't immediately return, this request is done in a background goroutine.
// When the main thread wishes to read events that have been received thus far
// the flush method can be called and it will return any queued up events.  When
// the main thread wishes to close the connection to the router the stop method
// can be called and it will return any unread events.
func (c ChannelClient) SSE(url string, router chi.Router) (flush func() []pubsub.Event, stop func() []pubsub.Event) {
	url = path.Join("/acrostic", c.name, url)
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
