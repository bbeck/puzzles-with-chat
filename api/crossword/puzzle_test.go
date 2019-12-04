package crossword

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPuzzle_WithSolutionHidden(t *testing.T) {
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
			p := &Puzzle{Cells: test.cells}
			p.WithSolutionHidden(func(puzzle *Puzzle) {
				assert.Nil(t, puzzle.Cells)
			})
			assert.Equal(t, test.cells, p.Cells)
		})
	}
}
