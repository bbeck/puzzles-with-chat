package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"github.com/alicebob/miniredis"
	"github.com/bbeck/twitch-plays-crosswords/api/crossword"
	"github.com/bbeck/twitch-plays-crosswords/api/model"
	"github.com/bbeck/twitch-plays-crosswords/api/pubsub"
	"github.com/bbeck/twitch-plays-crosswords/api/spellingbee"
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
	router, pool, _ := NewTestRouter(t)
	conn := NewRedisConnection(t, pool)

	// Connect to the stream when there's no active solves happening, we should
	// receive an event that contains an empty list of channels.
	_, stop := SSE("/channels", router)
	events := stop()
	assert.Equal(t, 1, len(events))
	assert.Equal(t, "channels", events[0].Kind)

	payload := ParsePayload(t, events[0].Payload)
	assert.Empty(t, payload["crossword"])
	assert.Empty(t, payload["spellingbee"])

	// Start a crossword.
	state1 := crossword.State{Status: model.StatusSolving}
	require.NoError(t, crossword.SetState(conn, "channel1", state1))

	// Now reconnect to the stream and we should receive one active channel.
	_, stop = SSE("/channels", router)
	events = stop()
	assert.Equal(t, 1, len(events))
	assert.Equal(t, "channels", events[0].Kind)

	payload = ParsePayload(t, events[0].Payload)
	assert.ElementsMatch(t, []model.Channel{
		{Name: "channel1", Status: model.StatusSolving},
	}, payload["crossword"])
	assert.Empty(t, payload["spellingbee"])

	// Start a spelling bee on another channel.
	state2 := spellingbee.State{Status: model.StatusSolving}
	require.NoError(t, spellingbee.SetState(conn, "channel2", state2))

	// Now we expect there to be 2 channels in the stream.
	_, stop = SSE("/channels", router)
	events = stop()
	assert.Equal(t, 1, len(events))
	assert.Equal(t, "channels", events[0].Kind)
	payload = ParsePayload(t, events[0].Payload)
	assert.ElementsMatch(t, []model.Channel{
		{Name: "channel1", Status: model.StatusSolving},
	}, payload["crossword"])
	assert.ElementsMatch(t, []model.Channel{
		{Name: "channel2", Status: model.StatusSolving},
	}, payload["spellingbee"])

	// Lastly remove the second channel from the database.
	_, err := conn.Do("DEL", spellingbee.StateKey("channel2"))
	require.NoError(t, err)

	// Now we expect there to be one channels in the stream.
	_, stop = SSE("/channels", router)
	events = stop()
	assert.Equal(t, 1, len(events))
	assert.Equal(t, "channels", events[0].Kind)

	payload = ParsePayload(t, events[0].Payload)
	assert.ElementsMatch(t, []model.Channel{
		{Name: "channel1", Status: model.StatusSolving},
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
			state := crossword.State{Status: model.StatusSolving}
			require.NoError(t, crossword.SetState(conn, "channel1", state))

			ForceErrorDuringActiveChannelsLoad(t, test.loadActiveChannelsError)

			// This won't start a background goroutine to send events because the
			// request will fail before reaching that part of the code.
			response := GET("/channels", router)
			assert.NotEqual(t, http.StatusOK, response.Code)
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
	RegisterRoutes(router, pool)

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

func ParsePayload(t *testing.T, v interface{}) map[string][]model.Channel {
	bs, err := json.Marshal(v)
	require.NoError(t, err)

	var payload map[string][]model.Channel
	require.NoError(t, json.Unmarshal(bs, &payload))

	return payload
}
