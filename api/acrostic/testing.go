package acrostic

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A cached puzzle to use instead of fetching a puzzle.  This is used by test
// cases to ensure that no network calls are made when loading puzzles.
var testPuzzle *Puzzle = nil

// A cached error to use instead of fetching a puzzle.  A cached puzzle takes
// precedence over a cached error.  This is used by test cases to force an
// error to be returned instead of a network call.
var testPuzzleLoadError error = nil

// load will read a file from the testdata directory.
func load(t *testing.T, filename string) io.ReadCloser {
	t.Helper()

	// First see if the file is in a testdata directory within the current
	// directory.
	if f, err := os.Open(filepath.Join("testdata", filename)); err == nil {
		return f
	}

	// It wasn't present, we might be running from a parent directory.  Try
	// crossword/testdata/{filename}.
	f, err := os.Open(filepath.Join("crossword", "testdata", filename))
	require.NoError(t, err)

	return f
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
