package spellingbee

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alicebob/miniredis"
	"github.com/bbeck/twitch-plays-crosswords/api/model"
	"github.com/bbeck/twitch-plays-crosswords/api/pubsub"
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

var Global = GlobalRoute{}
var Channel = ChannelRoute{channel: "channel"}

func TestRoute_GetChannels(t *testing.T) {
	// This acts as a small integration test ensuring that the active channels
	// event stream receives the events as new channels start and finish solves.
	pool, conn := NewRedisPool(t)

	router := chi.NewRouter()
	RegisterRoutes(router, pool)

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
	SetState(conn, Channel.channel, state)

	// Now reconnect to the stream and we should receive one active channel.
	_, stop = Global.SSE("/channels", router)
	events = stop()
	assert.Equal(t, 1, len(events))
	assert.Equal(t, "channels", events[0].Kind)
	assert.ElementsMatch(t, []string{Channel.channel}, events[0].Payload)

	// Start a puzzle on another channel.
	state = State{
		Status: model.StatusSolving,
		Puzzle: LoadTestPuzzle(t, "nytbee-20200408.html"),
		Words:  []string{},
	}
	SetState(conn, "channel2", state)

	// Now we expect there to be 2 channels in the stream.
	_, stop = Global.SSE("/channels", router)
	events = stop()
	assert.Equal(t, 1, len(events))
	assert.Equal(t, "channels", events[0].Kind)
	assert.ElementsMatch(t, []string{Channel.channel, "channel2"}, events[0].Payload)

	// Lastly remove the second channel from the database.
	_, err := conn.Do("DEL", StateKey("channel2"))
	require.NoError(t, err)

	// Now we expect there to be one channels in the stream.
	_, stop = Global.SSE("/channels", router)
	events = stop()
	assert.Equal(t, 1, len(events))
	assert.Equal(t, "channels", events[0].Kind)
	assert.ElementsMatch(t, []string{Channel.channel}, events[0].Payload)
}

func TestRoute_UpdatePuzzle_NYTBee(t *testing.T) {
	// This acts as a small integration test updating the date of the NYTBee
	// puzzle we're working on and ensuring the proper values are written
	// to the database.
	pool, conn := NewRedisPool(t)
	registry, events := NewRegistry(t)

	// Force a specific puzzle to be loaded so we don't make a network call.
	ForcePuzzleToBeLoaded(t, "nytbee-20200408.html")

	// Ensure that we have received the proper event and wrote the proper thing
	// to the database.
	verify := func(fn func(State)) {
		// First check that we've received an event with the correct value
		select {
		case event := <-events:
			require.Equal(t, "state", event.Kind)

			state := event.Payload.(State)

			// Event should never have the answers
			assert.Nil(t, state.Puzzle.OfficialAnswers)
			assert.Nil(t, state.Puzzle.UnofficialAnswers)
			fn(state)

		default:
			assert.Fail(t, "no state event available")
		}

		// Next check that the database has a valid settings object
		state, err := GetState(conn, "channel")
		require.NoError(t, err)

		// Database should always have the answers
		assert.NotNil(t, state.Puzzle.OfficialAnswers)
		assert.NotNil(t, state.Puzzle.UnofficialAnswers)
		fn(state)
	}

	router := chi.NewRouter()
	RegisterRoutesWithRegistry(router, pool, registry)

	response := Channel.PUT("/", `{"nytbee": "2020-04-08"}`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	verify(func(state State) {
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
			pool, _ := NewRedisPool(t)

			ForceErrorDuringLoad(t, test.forcedError)

			router := chi.NewRouter()
			RegisterRoutes(router, pool)

			response := Channel.PUT("/", test.payload, router)
			assert.Equal(t, test.expected, response.Code)
		})
	}
}

func TestRoute_UpdateSetting(t *testing.T) {
	// This acts as a small integration test updating each setting in turn and
	// making sure the proper value is written to the database and that clients
	// receive events notifying them of the setting change.
	pool, conn := NewRedisPool(t)
	registry, events := NewRegistry(t)

	// Ensure that we have received the proper event and wrote the proper thing
	// to the database.
	verify := func(t *testing.T, fn func(s Settings)) {
		t.Helper()

		// First check that we've received an event with the correct value
		select {
		case event := <-events:
			require.Equal(t, "settings", event.Kind)
			fn(event.Payload.(Settings))
		default:
			assert.Fail(t, "no settings event available")
		}

		// Next check that the database has a valid settings object
		settings, err := GetSettings(conn, "channel")
		require.NoError(t, err)
		fn(settings)
	}

	router := chi.NewRouter()
	RegisterRoutesWithRegistry(router, pool, registry)

	// Update each setting, one at a time.
	response := Channel.PUT("/setting/allow_unofficial_answers", `true`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	verify(t, func(s Settings) { assert.True(t, s.AllowUnofficialAnswers) })

	response = Channel.PUT("/setting/font_size", `"xlarge"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	verify(t, func(s Settings) { assert.Equal(t, model.FontSizeXLarge, s.FontSize) })
}

func TestRoute_UpdateSetting_ClearsUnofficialAnswers(t *testing.T) {
	// This acts as a small integration test toggling the AllowUnofficialAnswers
	// setting and ensuring that it removes any unofficial answers.
	pool, conn := NewRedisPool(t)
	registry, events := NewRegistry(t)

	// Ensure that we have received the proper event and wrote the proper thing
	// to the database.
	verify := func(t *testing.T, fn func(s State)) {
		t.Helper()

		// First check that we've received an event with the correct value
		select {
		case event := <-events:
			// Ignore the setting change event we receive
			if event.Kind == "settings" {
				return
			}
			require.Equal(t, "state", event.Kind)
			state := event.Payload.(State)

			// Events should never have the answers
			assert.Nil(t, state.Puzzle.OfficialAnswers)
			assert.Nil(t, state.Puzzle.UnofficialAnswers)
			fn(state)

		default:
			assert.Fail(t, "no state event available")
		}

		// Next check that the database has a valid settings object
		state, err := GetState(conn, "channel")
		require.NoError(t, err)

		// Database should always have the answers
		assert.NotNil(t, state.Puzzle.OfficialAnswers)
		assert.NotNil(t, state.Puzzle.UnofficialAnswers)
		fn(state)
	}

	router := chi.NewRouter()
	RegisterRoutesWithRegistry(router, pool, registry)

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
	SetState(conn, Channel.channel, state)

	// Set the AllowUnofficialAnswers setting to false
	response := Channel.PUT("/setting/allow_unofficial_answers", `false`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	verify(t, func(state State) {
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
			pool, _ := NewRedisPool(t)

			router := chi.NewRouter()
			RegisterRoutes(router, pool)

			response := Channel.PUT(fmt.Sprintf("/setting/%s", test.setting), test.json, router)
			assert.NotEqual(t, http.StatusOK, response.Code)
		})
	}
}

func NewRedisPool(t *testing.T) (*redis.Pool, redis.Conn) {
	t.Helper()

	s, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(s.Close)

	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", s.Addr())
		},
	}

	conn := pool.Get()
	t.Cleanup(func() { conn.Close() })

	return pool, conn
}

func NewRegistry(t *testing.T) (*pubsub.Registry, <-chan pubsub.Event) {
	t.Helper()

	registry := new(pubsub.Registry)
	stream := make(chan pubsub.Event, 10)
	t.Cleanup(func() { close(stream) })

	id, err := registry.Subscribe("channel", stream)
	require.NoError(t, err)
	t.Cleanup(func() { registry.Unsubscribe("channel", id) })

	return registry, stream
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
	channel string
}

func (r ChannelRoute) GET(url string, router chi.Router) *httptest.ResponseRecorder {
	url = path.Join(r.channel, url)
	return Global.GET(url, router)
}

func (r ChannelRoute) PUT(url, body string, router chi.Router) *httptest.ResponseRecorder {
	url = path.Join(r.channel, url)
	return Global.PUT(url, body, router)
}

func (r ChannelRoute) SSE(url string, router chi.Router) (flush func() []pubsub.Event, stop func() []pubsub.Event) {
	url = path.Join(r.channel, url)
	return Global.SSE(url, router)
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
