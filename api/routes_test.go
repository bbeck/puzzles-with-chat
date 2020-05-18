package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/alicebob/miniredis"
	"github.com/bbeck/puzzles-with-chat/api/crossword"
	"github.com/bbeck/puzzles-with-chat/api/model"
	"github.com/bbeck/puzzles-with-chat/api/pubsub"
	"github.com/bbeck/puzzles-with-chat/api/spellingbee"
	"github.com/go-chi/chi"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestRoute_GetChannels(t *testing.T) {
	// This acts as a small integration test ensuring that the active channels
	// event stream receives the events as new channels start and finish solves.
	router, pool, registry := NewTestRouter(t)
	conn := NewRedisConnection(t, pool)

	// Connect to the stream when there's no active solves happening, we should
	// receive an event that contains an empty list of channels.
	_, stop := SSE("/channels", router)
	events := stop()
	require.Equal(t, 1, len(events))
	assert.Equal(t, "channels", events[0].Kind)

	payload := ParsePayload(t, events[0].Payload)
	assert.Empty(t, payload["crossword"])
	assert.Empty(t, payload["spellingbee"])

	// Start a crossword.
	state1 := crossword.NewState(t, "xwordinfo-nyt-20181231.json")
	state1.Status = model.StatusSolving
	require.NoError(t, crossword.SetState(conn, "channel1", state1))

	// Now reconnect to the stream and we should receive one active channel.
	_, stop = SSE("/channels", router)
	events = stop()
	require.Equal(t, 1, len(events))
	assert.Equal(t, "channels", events[0].Kind)

	payload = ParsePayload(t, events[0].Payload)
	assert.ElementsMatch(t, []model.Channel{
		{
			Name:        "channel1",
			Status:      model.StatusSolving,
			Description: "New York Times puzzle from 2018-12-31",
		},
	}, payload["crossword"])
	assert.Empty(t, payload["spellingbee"])

	// Start a spelling bee on another channel.
	state2 := spellingbee.NewState(t, "nytbee-20180729.json")
	state2.Status = model.StatusSolving
	require.NoError(t, spellingbee.SetState(conn, "channel2", state2))

	// Now we expect there to be 2 channels in the stream.
	_, stop = SSE("/channels", router)
	events = stop()
	require.Equal(t, 1, len(events))
	assert.Equal(t, "channels", events[0].Kind)
	payload = ParsePayload(t, events[0].Payload)
	assert.ElementsMatch(t, []model.Channel{
		{
			Name:        "channel1",
			Status:      model.StatusSolving,
			Description: "New York Times puzzle from 2018-12-31",
		},
	}, payload["crossword"])
	assert.ElementsMatch(t, []model.Channel{
		{
			Name:        "channel2",
			Status:      model.StatusSolving,
			Description: "New York Times puzzle from 2018-07-29",
		},
	}, payload["spellingbee"])

	// Next remove the second channel from the database.
	_, err := conn.Do("DEL", spellingbee.StateKey("channel2"))
	require.NoError(t, err)

	// Now we expect there to be one channels in the stream.
	flush, stop := SSE("/channels", router)
	events = flush()
	require.Equal(t, 1, len(events))
	assert.Equal(t, "channels", events[0].Kind)

	payload = ParsePayload(t, events[0].Payload)
	assert.ElementsMatch(t, []model.Channel{
		{
			Name:        "channel1",
			Status:      model.StatusSolving,
			Description: "New York Times puzzle from 2018-12-31",
		},
	}, payload["crossword"])
	assert.Empty(t, payload["spellingbee"])

	// Now update the state of the first channel in the database and send an event
	// saying that it was updated.
	state1 = crossword.NewState(t, "xwordinfo-nyt-20181227-rebus.json")
	require.NoError(t, crossword.SetState(conn, "channel1", state1))
	registry.Publish(crossword.ChannelID("channel1"), crossword.StateEvent(state1))

	// Now we expect to have received another event in the stream.
	events = stop()
	require.Equal(t, 1, len(events))
	assert.Equal(t, "channels", events[0].Kind)

	payload = ParsePayload(t, events[0].Payload)
	assert.ElementsMatch(t, []model.Channel{
		{
			Name:        "channel1",
			Status:      model.StatusSelected,
			Description: "New York Times puzzle from 2018-12-27",
		},
	}, payload["crossword"])
	assert.Empty(t, payload["spellingbee"])
}

func TestRoute_GetChannels_Error(t *testing.T) {
	tests := []struct {
		name                    string
		loadActiveChannelsError error
	}{
		{
			name:                    "error loading channel names",
			loadActiveChannelsError: errors.New("forced error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router, pool, _ := NewTestRouter(t)
			conn := NewRedisConnection(t, pool)

			// Start a puzzle in the channel.
			state := crossword.NewState(t, "xwordinfo-nyt-20181231.json")
			state.Status = model.StatusSolving
			require.NoError(t, crossword.SetState(conn, "channel1", state))

			ForceErrorDuringActiveChannelsLoad(t, test.loadActiveChannelsError)

			// This won't start a background goroutine to send events because the
			// request will fail before reaching that part of the code.
			response := GET("/channels", router)
			assert.NotEqual(t, http.StatusOK, response.Code)
		})
	}
}

func TestChanged(t *testing.T) {
	tests := []struct {
		name     string
		before   map[string][]model.Channel
		after    map[string][]model.Channel
		expected bool
	}{
		{
			name:     "no categories",
			before:   map[string][]model.Channel{},
			after:    map[string][]model.Channel{},
			expected: false,
		},
		{
			name: "one category, no changes",
			before: map[string][]model.Channel{
				"crossword": {},
			},
			after: map[string][]model.Channel{
				"crossword": {},
			},
		},
		{
			name:   "one category added",
			before: map[string][]model.Channel{},
			after: map[string][]model.Channel{
				"crossword": {},
			},
			expected: true,
		},
		{
			name: "one channel added",
			before: map[string][]model.Channel{
				"crossword": {},
			},
			after: map[string][]model.Channel{
				"crossword": {{Name: "channel", Status: model.StatusSolving, Description: "description"}},
			},
			expected: true,
		},
		{
			name: "one channel removed",
			before: map[string][]model.Channel{
				"crossword": {{Name: "channel", Status: model.StatusSolving, Description: "description"}},
			},
			after: map[string][]model.Channel{
				"crossword": {},
			},
			expected: true,
		},
		{
			name: "channel status changed",
			before: map[string][]model.Channel{
				"crossword": {{Name: "channel", Status: model.StatusSolving, Description: "description"}},
			},
			after: map[string][]model.Channel{
				"crossword": {{Name: "channel", Status: model.StatusComplete, Description: "description"}},
			},
			expected: true,
		},
		{
			name: "channel description changed",
			before: map[string][]model.Channel{
				"crossword": {{Name: "channel", Status: model.StatusSolving, Description: "description1"}},
			},
			after: map[string][]model.Channel{
				"crossword": {{Name: "channel", Status: model.StatusSolving, Description: "description2"}},
			},
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			changed := Changed(test.before, test.after)
			assert.Equal(t, test.expected, changed)
		})
	}
}

// NewTestRouter will return a router configured with a redis pool and pubsub
// registry and wired together along with all of the routes for a spelling bee
// puzzle.
func NewTestRouter(t *testing.T) (chi.Router, *redis.Pool, *pubsub.Registry) {
	t.Helper()

	// Setup redis.
	server, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(server.Close)

	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", server.Addr())
		},
	}

	// Create the pubsub registry.
	registry := new(pubsub.Registry)

	// Setup the chi router and wire it up to the redis pool and pubsub registry.
	router := chi.NewRouter()
	RegisterRoutes(router, pool, registry)

	return router, pool, registry
}

// NewRedisConnection will return a connection to the provided connection pool.
// The returned connection will be configured to automatically close when the
// test completes.
func NewRedisConnection(t *testing.T, pool *redis.Pool) redis.Conn {
	t.Helper()

	conn := pool.Get()
	t.Cleanup(func() { _ = conn.Close() })

	return conn
}

func GET(url string, router chi.Router) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, url, nil)
	router.ServeHTTP(recorder, request)
	return recorder
}

// SSE performs a streaming request to the provided router.  Because the router
// won't immediately return, this request is done in a background goroutine.
// When the main thread wishes to read events that have been received thus far
// the flush method can be called and it will return any queued up events.  When
// the main thread wishes to close the connection to the router the stop method
// can be called and it will return any unread events.
func SSE(url string, router chi.Router) (flush func() []pubsub.Event, stop func() []pubsub.Event) {
	recorder := CreateTestResponseRecorder()

	ctx, cancel := context.WithCancel(context.Background())
	request := httptest.NewRequest(http.MethodGet, url, nil).WithContext(ctx)

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
		cancel()
		return flush()
	}

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

func ParsePayload(t *testing.T, v interface{}) map[string][]model.Channel {
	bs, err := json.Marshal(v)
	require.NoError(t, err)

	var payload map[string][]model.Channel
	require.NoError(t, json.Unmarshal(bs, &payload))

	return payload
}
