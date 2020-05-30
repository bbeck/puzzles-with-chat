package crossword

import (
	"fmt"
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
	// A human readable description of the puzzle
	Description string `json:"description"`

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

	// The block attribute for each of the cells in the crossword as a 2D list.
	// Cells that cannot be inputted into will contain an entry of true, all other
	// cells will contain an entry of false.  Like cells the 2D list is first
	// indexed by the row coordinate of the cell and then by the column
	// coordinate.
	CellBlocks [][]bool `json:"cell_blocks"`

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

	// Whether or not a cell should be shaded for all of the cells in the
	// crossword as a 2D list.  Cells that should be shaded appear as true and
	// those that shouldn't be shaded will appear as false.  Like cells the 2D
	// list is first indexed by the row coordinate of the cell and then by the
	// column coordinate.
	CellShades [][]bool `json:"cell_shades"`

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

// WithoutSolution returns a copy of the puzzle that has the solution cells
// missing.  This makes it suitable to pass to a client that shouldn't know the
// answers to the puzzle.
func (p *Puzzle) WithoutSolution() *Puzzle {
	var puzzle Puzzle
	puzzle.Description = p.Description
	puzzle.Rows = p.Rows
	puzzle.Cols = p.Cols
	puzzle.Title = p.Title
	puzzle.Publisher = p.Publisher
	puzzle.PublishedDate = p.PublishedDate
	puzzle.Author = p.Author
	puzzle.Cells = nil
	puzzle.CellBlocks = p.CellBlocks
	puzzle.CellClueNumbers = p.CellClueNumbers
	puzzle.CellCircles = p.CellCircles
	puzzle.CellShades = p.CellShades
	puzzle.CluesAcross = p.CluesAcross
	puzzle.CluesDown = p.CluesDown
	puzzle.Notes = p.Notes

	return &puzzle
}

// GetAnswerCoordinates returns the min/max x/y coordinates for a clue.  If the
// clue doesn't exist then an error is returned.
func (p *Puzzle) GetAnswerCoordinates(num int, direction string) (int, int, int, int, error) {
	// First, make sure this is a valid clue.
	if direction == "a" {
		_, ok := p.CluesAcross[num]
		if !ok {
			return 0, 0, 0, 0, fmt.Errorf("invalid clue %d%s", num, direction)
		}
	} else {
		_, ok := p.CluesDown[num]
		if !ok {
			return 0, 0, 0, 0, fmt.Errorf("invalid clue %d%s", num, direction)
		}
	}

	// Find the x, y coordinate where the answer begins.
	var minX, minY int
	for y := 0; y < p.Rows; y++ {
		for x := 0; x < p.Cols; x++ {
			if p.CellClueNumbers[y][x] == num {
				minX = x
				minY = y
			}
		}
	}

	// Determine the direction to step.
	var dx, dy int
	if direction == "a" {
		dx = 1
	} else {
		dy = 1
	}

	// Now that we know the starting cell, let's traverse in the correct direction
	// until we run into a black cell or the edge of the puzzle.
	var maxX, maxY int
	for x, y := minX, minY; x < p.Cols && y < p.Rows; x, y = x+dx, y+dy {
		if p.CellBlocks[y][x] {
			break
		}

		maxX = x
		maxY = y
	}

	return minX, minY, maxX, maxY, nil
}
