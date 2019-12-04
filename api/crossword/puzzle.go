package crossword

import (
	"time"
)

// Puzzle represents a crossword puzzle.  The puzzle is comprised of a
// grid which has dimensions (rows x cols) and demonstrates which cells of the
// crossword are available for placing letters into and which are not.
// Additionally the puzzle contains a set of clues organized by number and
// whether or not they fill in cells going across or down.  Lastly a puzzle has
// various bits of interesting metadata such as the publication that the
// crossword is from, the date that it was published as well as the author(s).
type Puzzle struct {
	// The number of rows in the crossword grid.
	Rows int `json:"rows"`

	// The number of columns in the crossword grid.
	Cols int `json:"cols"`

	// The title of the crossword.
	Title string `json:"title"`

	// The publisher of the crossword.
	Publisher string `json:"publisher"`

	// The date that the crossword was published.
	PublishedDate time.Time `json:"published"`

	// The name of the author(s) of the crossword.
	Author string `json:"author"`

	// The cells of the crossword as a 2D list, entries are the letter (or letters
	// in the case of a rebus) that belong in the cell.  If a cell cannot be
	// inputted into then it will contain the empty string.  The lists are first
	// indexed by the row coordinate of the cell and then by the column coordinate
	// of the cell.
	Cells [][]string `json:"cells,omitempty"`

	// The clue numbers for each of the cells in the crossword as a 2D list.
	// Cells that cannot be inputted into or that don't have a clue number will
	// contain an entry of 0.  Like cells the 2D list is first indexed by the row
	// coordinate of the cell and then by the column coordinate.
	CellClueNumbers [][]int `json:"cell_clue_numbers"`

	// Whether or not a cell contains a circle for all of the cells in the
	// crossword as a 2D list.  Cells that should have a circle rendered in them
	// appear as true and those that shouldn't have a circle rendered in them will
	// appear as false.  Like cells the 2D list is first indexed by the row
	// coordinate of the cell and then by the column coordinate.
	CellCircles [][]bool `json:"cell_circles"`

	// The clues for the across answers indexed by the clue number.
	CluesAcross map[int]string `json:"clues_across"`

	// The clues for the down answers indexed by the clue number.
	CluesDown map[int]string `json:"clues_down"`

	// The notes for the clues of this crossword.  Often there is something
	// visually done when the crossword is published in a newspaper but that can't
	// be done online.  These notes describe the visual change so that the
	// crossword can be solved online.
	Notes string `json:"notes"`
}

// WithSolutionHidden temporarily removes the solution cells of the puzzle and
// passes it to the provided callback.  After the callback is returned, the
// solution cells are restored.
func (p *Puzzle) WithSolutionHidden(fn func(*Puzzle)) {
	cells := p.Cells
	p.Cells = nil
	fn(p)
	p.Cells = cells
}
