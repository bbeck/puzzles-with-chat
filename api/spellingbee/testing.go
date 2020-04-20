package spellingbee

import (
	"github.com/alicebob/miniredis"
	"github.com/bbeck/twitch-plays-crosswords/api/pubsub"
	"github.com/go-chi/chi"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// A cached puzzle to use instead of fetching a puzzle.  This is used by test
// cases to ensure that no network calls are made when loading puzzles.
var testCachedPuzzle *Puzzle = nil

// A cached error to use instead of fetching a puzzle.  A cached puzzle takes
// precedence over a cached error.  This is used by test cases to force an
// error to be returned instead of a network call.
var testCachedError error = nil

// load will read a file from the testdata directory.
func load(t *testing.T, filename string) io.ReadCloser {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", filename))
	require.NoError(t, err)
	return f
}

// LoadTestPuzzle loads a puzzle from the testdata directory.
func LoadTestPuzzle(t *testing.T, filename string) *Puzzle {
	t.Helper()

	in := load(t, filename)
	defer func() { _ = in.Close() }()

	puzzle, err := ParseNYTBeeResponse(in)
	require.NoError(t, err)
	return puzzle
}

// ForcePuzzleToBeLoaded sets up a cached version of a puzzle using a file from
// the testdata directory.
func ForcePuzzleToBeLoaded(t *testing.T, filename string) {
	t.Helper()

	testCachedPuzzle = LoadTestPuzzle(t, filename)
	t.Cleanup(func() { testCachedPuzzle = nil })
}

// ForceErrorDuringLoad sets up an error to be returned when an attempt is made
// to load a puzzle.
func ForceErrorDuringLoad(t *testing.T, err error) {
	t.Helper()

	testCachedError = err
	t.Cleanup(func() { testCachedError = nil })
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
	RegisterRoutesWithRegistry(router, pool, registry)

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

// NewEventSubscription will return a channel of events that are subscribed to
// the specified channel.  The subscription will be configured to automatically
// unsubscribe when the test completes.
func NewEventSubscription(t *testing.T, registry *pubsub.Registry, channel string) <-chan pubsub.Event {
	t.Helper()

	events := make(chan pubsub.Event, 10)
	id, err := registry.Subscribe(pubsub.Channel(channel), events)
	require.NoError(t, err)

	t.Cleanup(func() { registry.Unsubscribe(pubsub.Channel(channel), id) })
	return events
}
