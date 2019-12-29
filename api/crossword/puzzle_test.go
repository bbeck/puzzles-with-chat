package crossword

import (
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
