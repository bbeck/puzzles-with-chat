package crossword

import (
	"github.com/stretchr/testify/require"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPuzzle_WithoutSolution(t *testing.T) {
	tests := []struct {
		name  string
		cells [][]string
	}{
		{
			name: "nil cells",
		},
		{
			name:  "empty cells",
			cells: [][]string{},
		},
		{
			name: "non-empty cells",
			cells: [][]string{
				{"A", "B", "C"},
				{"D", "E", "F"},
				{"I", "H", "G"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			puzzle := &Puzzle{Cells: test.cells}
			assert.Nil(t, puzzle.WithoutSolution().Cells)
		})
	}
}

func TestPuzzle_GetAnswerCoordinates(t *testing.T) {
	tests := []struct {
		name                       string
		input                      io.ReadCloser
		num                        int
		direction                  string
		expectedMinX, expectedMinY int
		expectedMaxX, expectedMaxY int
	}{
		{
			name:         "1a",
			input:        load(t, "xwordinfo-success-20181231.json"),
			num:          1,
			direction:    "a",
			expectedMinX: 0,
			expectedMinY: 0,
			expectedMaxX: 4,
			expectedMaxY: 0,
		},
		{
			name:         "6a",
			input:        load(t, "xwordinfo-success-20181231.json"),
			num:          6,
			direction:    "a",
			expectedMinX: 6,
			expectedMinY: 0,
			expectedMaxX: 10,
			expectedMaxY: 0,
		},
		{
			name:         "11a",
			input:        load(t, "xwordinfo-success-20181231.json"),
			num:          11,
			direction:    "a",
			expectedMinX: 12,
			expectedMinY: 0,
			expectedMaxX: 14,
			expectedMaxY: 0,
		},
		{
			name:         "1d",
			input:        load(t, "xwordinfo-success-20181231.json"),
			num:          1,
			direction:    "d",
			expectedMinX: 0,
			expectedMinY: 0,
			expectedMaxX: 0,
			expectedMaxY: 3,
		},
		{
			name:         "30d",
			input:        load(t, "xwordinfo-success-20181231.json"),
			num:          30,
			direction:    "d",
			expectedMinX: 0,
			expectedMinY: 6,
			expectedMaxX: 0,
			expectedMaxY: 8,
		},
		{
			name:         "51d",
			input:        load(t, "xwordinfo-success-20181231.json"),
			num:          51,
			direction:    "d",
			expectedMinX: 0,
			expectedMinY: 11,
			expectedMaxX: 0,
			expectedMaxY: 14,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer test.input.Close()

			puzzle, err := ParseXWordInfoResponse(test.input)
			require.NoError(t, err)

			minX, minY, maxX, maxY, err := puzzle.GetAnswerCoordinates(test.num, test.direction)
			assert.NoError(t, err)
			assert.Equal(t, test.expectedMinX, minX)
			assert.Equal(t, test.expectedMinY, minY)
			assert.Equal(t, test.expectedMaxX, maxX)
			assert.Equal(t, test.expectedMaxY, maxY)
		})
	}
}

func TestPuzzle_GetAnswerCoordinates_Error(t *testing.T) {
	tests := []struct {
		name      string
		input     io.ReadCloser
		num       int
		direction string
	}{
		{
			name:      "66a",
			input:     load(t, "xwordinfo-success-20181231.json"),
			num:       66,
			direction: "a",
		},
		{
			name:      "66d",
			input:     load(t, "xwordinfo-success-20181231.json"),
			num:       66,
			direction: "d",
		},
		{
			name:      "2a",
			input:     load(t, "xwordinfo-success-20181231.json"),
			num:       2,
			direction: "a",
		},
		{
			name:      "14d",
			input:     load(t, "xwordinfo-success-20181231.json"),
			num:       14,
			direction: "d",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer test.input.Close()

			puzzle, err := ParseXWordInfoResponse(test.input)
			require.NoError(t, err)

			_, _, _, _, err = puzzle.GetAnswerCoordinates(test.num, test.direction)
			assert.Error(t, err)
		})
	}
}
