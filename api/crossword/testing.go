package crossword

import (
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func load(t *testing.T, filename string) io.ReadCloser {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", filename))
	require.NoError(t, err)
	return f
}

func LoadTestPuzzle(t *testing.T, filename string) *Puzzle {
	t.Helper()
	in := load(t, filename)
	defer func() { _ = in.Close() }()

	puzzle, err := ParseXWordInfoResponse(in)
	require.NoError(t, err)

	return puzzle
}
