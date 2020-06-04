package acrostic

import (
	"fmt"
	"time"
)

// Puzzle represents an acrostic puzzle.  The puzzle is comprised of a
// grid which has dimensions (rows x cols) and demonstrates which cells of the
// acrostic are available for placing letters into and which are not.
// Additionally the puzzle contains a set of lettered clues and a mapping of
// answer letters their position within the grid.  Lastly a puzzle has
// various bits of interesting metadata such as the publication that the
// crossword is from, the date that it was published as well as the author(s).
type Puzzle struct {
	// A human readable description of the puzzle
	Description string `json:"description"`

	// The number of rows in the crossword grid.
	Rows int `json:"rows"`

	// The number of columns in the crossword grid.
	Cols int `json:"cols"`

	// The publisher of the crossword.
	Publisher string `json:"publisher"`

	// The date that the crossword was published.
	PublishedDate time.Time `json:"published"`

	// The cells of the acrostic as a 2D list, entries are the letter that belongs
	// in the cell.  If a cell cannot be inputted into then it will contain the
	// empty string.  The lists are first indexed by the row coordinate of the
	// cell and then by the column coordinate of the cell.
	Cells [][]string `json:"cells,omitempty"`

	// The block attribute for each of the cells in the acrostic as a 2D list.
	// Cells that cannot be inputted into will contain an entry of true, all other
	// cells will contain an entry of false.  Like cells the 2D list is first
	// indexed by the row coordinate of the cell and then by the column
	// coordinate.
	CellBlocks [][]bool `json:"cell_blocks"`

	// The numbers for each of the cells in the acrostic as a 2D list.  Cells that
	// cannot be inputted into will contain an entry of 0.  Like cells the 2D list
	// is first indexed by the row coordinate of the cell and then by the column
	// coordinate.
	CellNumbers [][]int `json:"cell_clue_numbers"`

	// The clue letter for each of the cells in the acrostic as a 2D list.
	// Cells that cannot be inputted into or that don't have a clue letter will
	// contain an empty string.  Like cells the 2D list is first indexed by the
	// row coordinate of the cell and then by the column coordinate.
	CellClueLetters [][]string `json:"cell_clue_letters"`

	// The clues indexed by the clue letter.
	Clues map[string]string `json:"clues"`

	// The clue numbers indexed by the clue letter.
	ClueNumbers map[string][]int `json:"clue_numbers"`
}

// WithoutSolution returns a copy of the puzzle that has the solution cells
// missing.  This makes it suitable to pass to a client that shouldn't know the
// answers to the puzzle.
func (p *Puzzle) WithoutSolution() *Puzzle {
	var puzzle Puzzle
	puzzle.Description = p.Description
	puzzle.Rows = p.Rows
	puzzle.Cols = p.Cols
	puzzle.Publisher = p.Publisher
	puzzle.PublishedDate = p.PublishedDate
	puzzle.Cells = nil
	puzzle.CellBlocks = p.CellBlocks
	puzzle.CellNumbers = p.CellNumbers
	puzzle.CellClueLetters = p.CellClueLetters
	puzzle.Clues = p.Clues

	return &puzzle
}

// GetCellCoordinates returns the x, y coordinates for a numbered cell.  If the
// cell doesn't exist then an error is returned.
func (p *Puzzle) GetCellCoordinates(num int) (int, int, error) {
	if num < 1 {
		return 0, 0, fmt.Errorf("cell number %d is out of bounds", num)
	}

	for y := num / p.Cols; y < p.Rows; y++ {
		for x := 0; x < p.Cols; x++ {
			if p.CellNumbers[y][x] == num {
				return x, y, nil
			}
		}
	}

	return 0, 0, fmt.Errorf("cell number %d is out of bounds", num)
}
