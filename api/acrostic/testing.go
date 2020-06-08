package acrostic

import (
	"github.com/alicebob/miniredis"
	"github.com/bbeck/puzzles-with-chat/api/model"
	"github.com/bbeck/puzzles-with-chat/api/pubsub"
	"github.com/go-chi/chi"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// A cached puzzle to use instead of fetching a puzzle.  This is used by test
// cases to ensure that no network calls are made when loading puzzles.
var testPuzzle *Puzzle = nil

// A cached error to use instead of fetching a puzzle.  A cached puzzle takes
// precedence over a cached error.  This is used by test cases to force an
// error to be returned instead of a network call.
var testPuzzleLoadError error = nil

// A cached error to use instead of reading state from the database.
var testSettingsLoadError error = nil

// A cached error to use instead of writing settings to the database.
var testSettingsSaveError error = nil

// A cached error to use instead of writing state to the database.
var testStateLoadError error = nil

// A cached error to use instead of writing state to the database.
var testStateSaveError error = nil

// load will read a file from the testdata directory.
func load(t *testing.T, filename string) io.ReadCloser {
	t.Helper()

	// First see if the file is in a testdata directory within the current
	// directory.
	if f, err := os.Open(filepath.Join("testdata", filename)); err == nil {
		return f
	}

	// It wasn't present, we might be running from a parent directory.  Try each
	// puzzle type's test directory, one at at ime.
	types := []string{"acrostic", "crossword", "spellingbee"}
	for _, t := range types {
		if f, err := os.Open(filepath.Join(t, "testdata", filename)); err == nil {
			return f
		}
	}

	t.Errorf("unable to locate test input with filename: %s", filename)
	return nil
}

// LoadTestPuzzle loads a puzzle from the testdata directory.
func LoadTestPuzzle(t *testing.T, filename string) *Puzzle {
	t.Helper()

	in := load(t, filename)
	defer func() { _ = in.Close() }()

	var puzzle *Puzzle
	var err error
	switch {
	case strings.HasPrefix(filename, "xwordinfo-"):
		puzzle, err = ParseXWordInfoResponse(in)

	default:
		assert.Failf(t, "unrecognized filename prefix", "filename: %s", filename)
	}

	require.NoError(t, err)
	return puzzle
}

// ForcePuzzleToBeLoaded sets up a cached version of a puzzle using a file from
// the testdata directory.
func ForcePuzzleToBeLoaded(t *testing.T, filename string) {
	t.Helper()

	testPuzzle = LoadTestPuzzle(t, filename)
	t.Cleanup(func() { testPuzzle = nil })
}

// ForceErrorDuringLoad sets up an error to be returned when an attempt is made
// to load a puzzle.
func ForceErrorDuringPuzzleLoad(t *testing.T, err error) {
	t.Helper()

	testPuzzleLoadError = err
	t.Cleanup(func() { testPuzzleLoadError = nil })
}

// ForceErrorDuringSettingsLoad sets up an error to be returned when an attempt
// is made to load settings.
func ForceErrorDuringSettingsLoad(t *testing.T, err error) {
	t.Helper()

	testSettingsLoadError = err
	t.Cleanup(func() { testSettingsLoadError = nil })
}

// ForceErrorDuringSettingsSave sets up an error to be returned when an attempt
// is made to save settings.
func ForceErrorDuringSettingsSave(t *testing.T, err error) {
	t.Helper()

	testSettingsSaveError = err
	t.Cleanup(func() { testSettingsSaveError = nil })
}

// ForceErrorDuringStateLoad sets up an error to be returned when an attempt
// is made to load state.
func ForceErrorDuringStateLoad(t *testing.T, err error) {
	t.Helper()

	testStateLoadError = err
	t.Cleanup(func() { testStateLoadError = nil })
}

// ForceErrorDuringStateSave sets up an error to be returned when an attempt
// is made to save state.
func ForceErrorDuringStateSave(t *testing.T, err error) {
	t.Helper()

	testStateSaveError = err
	t.Cleanup(func() { testStateSaveError = nil })
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

// NewEventSubscription will return a channel of events that are subscribed to
// the specified channel.  The subscription will be configured to automatically
// unsubscribe when the test completes.
func NewEventSubscription(t *testing.T, registry *pubsub.Registry, channel string) <-chan pubsub.Event {
	t.Helper()

	events := make(chan pubsub.Event, 10)
	id, err := registry.Subscribe(ChannelID(channel), events)
	require.NoError(t, err)

	t.Cleanup(func() { registry.Unsubscribe(id) })
	return events
}

// NewState creates a new acrostic puzzle state that has been properly
// initialized with the puzzle corresponding to the provided filename.
func NewState(t *testing.T, filename string) State {
	puzzle := LoadTestPuzzle(t, filename)

	cells := make([][]string, puzzle.Rows)
	for col := 0; col < puzzle.Rows; col++ {
		cells[col] = make([]string, puzzle.Cols)
	}

	now := time.Now()
	return State{
		Status:        model.StatusSelected,
		Puzzle:        puzzle,
		Cells:         cells,
		CluesFilled:   make(map[string]bool),
		LastStartTime: &now,
	}
}
