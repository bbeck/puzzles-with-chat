package crossword

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPuzzle_ConvertedJSON(t *testing.T) {
	tests := []struct {
		name   string
		input  io.ReadCloser
		verify func(t *testing.T, puzzle *Puzzle)
	}{
		{
			name:  "description",
			input: load(t, "puzzle-wp-20051206.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := "Crossword loaded from .puz file"
				assert.Equal(t, expected, puzzle.Description)
			},
		},
		{
			name:  "size",
			input: load(t, "puzzle-wp-20051206.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, 15, puzzle.Cols)
				assert.Equal(t, 15, puzzle.Rows)
			},
		},
		{
			name:  "title",
			input: load(t, "puzzle-wp-20051206.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, `December 6, 2005 - "Split Pea Soup"`, puzzle.Title)
			},
		},
		{
			name:  "publisher",
			input: load(t, "puzzle-wp-20051206.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, "", puzzle.Publisher) // No publisher in .puz files
			},
		},
		{
			name:  "published date",
			input: load(t, "puzzle-wp-20051206.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, time.Time{}, puzzle.PublishedDate) // No publish date in .puz files
			},
		},
		{
			name:  "author",
			input: load(t, "puzzle-wp-20051206.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, "Raymond Hamel", puzzle.Author)
			},
		},
		{
			name:  "cells",
			input: load(t, "puzzle-wp-20051206.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := [][]string{
					{"L", "A", "M", "B", "", "S", "P", "A", "T", "", "C", "A", "R", "V", "E"},
					{"O", "R", "A", "L", "", "A", "L", "E", "E", "", "O", "B", "I", "E", "S"},
					{"F", "I", "R", "E", "S", "T", "O", "R", "M", "", "Q", "U", "O", "T", "E"},
					{"T", "A", "K", "E", "A", "S", "W", "I", "P", "E", "A", "T", "", "", ""},
					{"", "", "", "P", "I", "C", "", "E", "T", "T", "U", "", "S", "P", "F"},
					{"P", "H", "D", "", "D", "O", "S", "", "", "A", "V", "A", "T", "A", "R"},
					{"R", "U", "E", "S", "", "R", "O", "O", "K", "", "I", "R", "A", "N", "I"},
					{"E", "S", "C", "A", "P", "E", "A", "T", "T", "E", "N", "T", "I", "O", "N"},
					{"S", "H", "A", "K", "E", "", "P", "O", "E", "M", "", "S", "N", "U", "G"},
					{"T", "U", "N", "E", "R", "S", "", "", "L", "B", "J", "", "S", "T", "E"},
					{"O", "P", "T", "", "F", "O", "A", "M", "", "E", "E", "L", "", "", ""},
					{"", "", "", "H", "O", "P", "E", "A", "N", "D", "F", "A", "I", "T", "H"},
					{"M", "A", "J", "O", "R", "", "G", "U", "I", "D", "E", "P", "O", "S", "T"},
					{"A", "X", "I", "O", "M", "", "I", "D", "L", "E", "", "S", "W", "A", "T"},
					{"T", "E", "M", "P", "S", "", "S", "E", "E", "D", "", "E", "A", "R", "P"},
				}
				assert.Equal(t, expected, puzzle.Cells)
			},
		},
		{
			name:  "blocks",
			input: load(t, "puzzle-wp-20051206.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := [][]bool{
					{false, false, false, false, true, false, false, false, false, true, false, false, false, false, false},
					{false, false, false, false, true, false, false, false, false, true, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, true, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, true, true, true},
					{true, true, true, false, false, false, true, false, false, false, false, true, false, false, false},
					{false, false, false, true, false, false, false, true, true, false, false, false, false, false, false},
					{false, false, false, false, true, false, false, false, false, true, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, true, false, false, false, false, true, false, false, false, false},
					{false, false, false, false, false, false, true, true, false, false, false, true, false, false, false},
					{false, false, false, true, false, false, false, false, true, false, false, false, true, true, true},
					{true, true, true, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, true, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, true, false, false, false, false, true, false, false, false, false},
					{false, false, false, false, false, true, false, false, false, false, true, false, false, false, false},
				}
				assert.Equal(t, expected, puzzle.CellBlocks)
			},
		},
		{
			name:  "cell clue numbers",
			input: load(t, "puzzle-wp-20051206.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := [][]int{
					{1, 2, 3, 4, 0, 5, 6, 7, 8, 0, 9, 10, 11, 12, 13},
					{14, 0, 0, 0, 0, 15, 0, 0, 0, 0, 16, 0, 0, 0, 0},
					{17, 0, 0, 0, 18, 0, 0, 0, 0, 0, 19, 0, 0, 0, 0},
					{20, 0, 0, 0, 0, 0, 0, 0, 0, 21, 0, 0, 0, 0, 0},
					{0, 0, 0, 22, 0, 0, 0, 23, 0, 0, 0, 0, 24, 25, 26},
					{27, 28, 29, 0, 30, 0, 31, 0, 0, 32, 0, 33, 0, 0, 0},
					{34, 0, 0, 35, 0, 36, 0, 37, 38, 0, 39, 0, 0, 0, 0},
					{40, 0, 0, 0, 41, 0, 0, 0, 0, 42, 0, 0, 0, 0, 0},
					{43, 0, 0, 0, 0, 0, 44, 0, 0, 0, 0, 45, 0, 0, 0},
					{46, 0, 0, 0, 0, 47, 0, 0, 48, 0, 49, 0, 50, 0, 0},
					{51, 0, 0, 0, 52, 0, 53, 54, 0, 55, 0, 56, 0, 0, 0},
					{0, 0, 0, 57, 0, 0, 0, 0, 58, 0, 0, 0, 59, 60, 61},
					{62, 63, 64, 0, 0, 0, 65, 0, 0, 0, 0, 0, 0, 0, 0},
					{66, 0, 0, 0, 0, 0, 67, 0, 0, 0, 0, 68, 0, 0, 0},
					{69, 0, 0, 0, 0, 0, 70, 0, 0, 0, 0, 71, 0, 0, 0},
				}
				assert.Equal(t, expected, puzzle.CellClueNumbers)
			},
		},
		{
			name:  "cell circles (none)",
			input: load(t, "puzzle-wp-20051206.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := [][]bool{
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
				}
				assert.Equal(t, expected, puzzle.CellCircles)
			},
		},
		{
			name:  "cell circles",
			input: load(t, "puzzle-nyt-20081006-nonsquare-with-circles.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := [][]bool{
					{true, true, true, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, true, true, true},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{true, true, true, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, true, true, true},
				}
				assert.Equal(t, expected, puzzle.CellCircles)
			},
		},
		{
			name:  "across clues",
			input: load(t, "puzzle-wp-20051206.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := map[int]string{
					1:  `Mary's pet`,
					5:  `Disagreement`,
					9:  `Cut, as a turkey`,
					14: `Kind of history`,
					15: `On the sheltered side`,
					16: `Theater awards`,
					17: `Outburst of controversy`,
					19: `Cite`,
					20: `Aim for`,
					22: `JPEG file, often`,
					23: `Rebuke to a backstabber`,
					24: `Tanning lotion tube letters`,
					27: `Graduate prog. award`,
					30: `Old Microsoft product`,
					32: `On-line game character`,
					34: `Regrets`,
					36: `Chess corner piece`,
					39: `Shiraz citizen`,
					40: `Go unnoticed`,
					43: `Get rid of`,
					44: `Haiku, for one`,
					45: `Warm and cozy`,
					46: `Signal receivers`,
					48: `"All the way with ___" (political slogan)`,
					50: `___ Anne de Beaupré`,
					51: `Choose`,
					52: `Insulation material`,
					55: `Shocking swimmer`,
					57: `Current sitcom set in Cleveland`,
					62: `Very important`,
					65: `Direction giver`,
					66: `"There's no such thing as a free lunch," e.g.`,
					67: `On strike`,
					68: `Police team`,
					69: `Office aides`,
					70: `Bit of bird chow`,
					71: `Holliday's marshal friend`,
				}
				assert.Equal(t, expected, puzzle.CluesAcross)
			},
		},
		{
			name:  "down clues",
			input: load(t, "puzzle-wp-20051206.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := map[int]string{
					1:  `Hit high in the air`,
					2:  `Caruso solo`,
					3:  `Simplified signature`,
					4:  `Censor`,
					5:  `College admission factor`,
					6:  `Tractor attachment`,
					7:  `Eagle's nest`,
					8:  `Entice`,
					9:  `French chicken dish`,
					10: `Touch`,
					11: `1983 Duran Duran hit`,
					12: `Doggy doc`,
					13: `Jargon suffix`,
					18: `Spoken`,
					21: `LAX listing`,
					24: `Soiled spots`,
					25: `Succeed`,
					26: `Green edge`,
					27: `Magic word`,
					28: `Keep quiet`,
					29: `Draw off sherry`,
					31: `Word before bubble or opera`,
					33: `Part of NEA`,
					35: `Drink with sushi`,
					37: `Ally of the Missouri`,
					38: `Record company name now licensed to an on-line pharmaceutical company`,
					41: `Sings and dances`,
					42: `Like some wartime journalists`,
					47: `Make soaking wet`,
					49: `Boss, usually following "El"`,
					53: `Protection`,
					54: `Bea Arthur sitcom`,
					56: `Fall into disuse`,
					57: `NBA target`,
					58: `West ___ virus`,
					59: `Big Ten school`,
					60: `Lenin foe`,
					61: `URL starter`,
					62: `Exercise accessory`,
					63: `Paul Bunyan's tool`,
					64: `Former NFL quarterback Kelly or Harbaugh`,
				}
				assert.Equal(t, expected, puzzle.CluesDown)
			},
		},
		{
			name:  "notes",
			input: load(t, "puzzle-nyt-20080912-notes.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, "TEEN PUZZLEMAKER WEEK", puzzle.Notes[:21])
			},
		},
		{
			name:  "rebus",
			input: load(t, "puzzle-nyt-20080914-rebus.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := [][]string{
					{"", "S", "T", "U", "A", "R", "T", "", "R", "O", "S", "I", "E", "", "", "", "O", "M", "E", "G", "A"},
					{"", "P", "A", "S", "T", "O", "R", "", "E", "R", "A", "S", "E", "", "C", "A", "R", "O", "L", "E", "R"},
					{"Y", "O", "K", "O", "O", "N", "O", "", "L", "I", "FEB", "A", "N", "", "A", "U", "G", "M", "E", "N", "T"},
					{"E", "T", "E", "", "B", "A", "JAN", "", "O", "B", "E", "Y", "", "S", "MAR", "T", "Y", "", "C", "O", "B"},
					{"A", "T", "T", "S", "", "", "A", "W", "A", "I", "T", "", "P", "O", "O", "R", "", "A", "T", "M", "O"},
					{"S", "E", "E", "I", "N", "G", "R", "E", "D", "", "", "G", "U", "T", "S", "Y", "", "T", "R", "I", "O"},
					{"T", "R", "A", "DEC", "O", "M", "M", "I", "S", "S", "I", "O", "N", "S", "", "", "R", "APR", "O", "C", "K"},
					{"", "", "", "A", "D", "A", "Y", "", "", "T", "O", "I", "T", "", "G", "A", "T", "O", "", "", ""},
					{"C", "A", "D", "R", "E", "", "", "S", "W", "A", "N", "N", "", "B", "E", "N", "E", "F", "I", "T", "S"},
					{"O", "T", "O", "", "A", "U", "S", "T", "I", "N", "", "G", "I", "A", "N", "T", "", "I", "T", "O", "O"},
					{"M", "A", "NOV", "E", "R", "B", "O", "A", "R", "D", "", "A", "N", "C", "I", "E", "N", "T", "MAY", "A", "N"},
					{"B", "R", "A", "Y", "", "O", "A", "T", "E", "S", "", "L", "I", "K", "E", "S", "O", "", "B", "I", "G"},
					{"S", "I", "N", "E", "W", "A", "V", "E", "", "F", "E", "L", "T", "S", "", "", "B", "A", "E", "R", "S"},
					{"", "", "", "D", "A", "T", "E", "", "N", "I", "K", "I", "", "", "M", "O", "L", "D", "", "", ""},
					{"D", "E", "C", "OCT", "S", "", "", "A", "D", "R", "E", "N", "A", "L", "I", "N", "E", "JUN", "K", "I", "E"},
					{"I", "M", "H", "O", "", "A", "B", "R", "A", "M", "", "", "B", "O", "N", "E", "S", "C", "A", "N", "S"},
					{"S", "U", "E", "R", "", "H", "A", "C", "K", "", "D", "W", "E", "L", "T", "", "", "T", "R", "O", "T"},
					{"C", "L", "E", "", "J", "O", "SEP", "H", "", "A", "R", "A", "T", "", "JUL", "E", "S", "", "A", "R", "E"},
					{"M", "A", "S", "C", "A", "R", "A", "", "F", "R", "AUG", "H", "T", "", "E", "P", "I", "S", "O", "D", "E"},
					{"A", "T", "E", "A", "W", "A", "Y", "", "A", "C", "H", "O", "O", "", "P", "E", "R", "U", "K", "E", ""},
					{"N", "E", "S", "T", "S", "", "", "", "M", "O", "T", "O", "R", "", "S", "E", "E", "D", "E", "R", ""},
				}
				assert.Equal(t, expected, puzzle.Cells)
			},
		},
		{
			name:  "non-square dimensions",
			input: load(t, "puzzle-nyt-20081006-nonsquare-with-circles.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, 9, puzzle.Rows)
				assert.Equal(t, 24, puzzle.Cols)
			},
		},
		{
			name:  "non-square cells",
			input: load(t, "puzzle-nyt-20081006-nonsquare-with-circles.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := [][]string{
					{"O", "N", "E", "G", "", "L", "E", "S", "", "D", "O", "L", "L", "A", "R", "", "H", "B", "O", "", "Z", "O", "N", "E"},
					{"P", "O", "T", "R", "O", "A", "S", "T", "", "E", "U", "G", "E", "N", "E", "", "E", "A", "R", "L", "O", "B", "E", "S"},
					{"T", "H", "E", "U", "N", "I", "T", "E", "D", "S", "T", "A", "T", "E", "S", "O", "F", "A", "M", "E", "R", "I", "C", "A"},
					{"", "", "", "N", "O", "R", "", "R", "I", "P", "", "", "", "", "O", "S", "T", "", "E", "B", "B", "", "", ""},
					{"", "S", "S", "T", "", "", "I", "N", "G", "O", "D", "W", "E", "T", "R", "U", "S", "T", "", "", "A", "M", "S", ""},
					{"E", "A", "U", "", "E", "S", "C", "", "", "T", "E", "A", "A", "C", "T", "", "", "Y", "O", "W", "", "A", "A", "H"},
					{"G", "R", "E", "A", "T", "S", "E", "A", "L", "", "W", "I", "S", "E", "", "B", "A", "L", "D", "E", "A", "G", "L", "E"},
					{"G", "A", "M", "M", "A", "R", "A", "Y", "S", "", "E", "V", "E", "L", "", "S", "T", "E", "E", "L", "D", "O", "O", "R"},
					{"O", "N", "E", "A", "L", "", "X", "E", "D", "", "Y", "E", "L", "L", "", "A", "E", "R", "", "D", "O", "O", "N", "E"},
				}
				assert.Equal(t, expected, puzzle.Cells)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer test.input.Close()

			puzzle := new(Puzzle)
			err := json.NewDecoder(test.input).Decode(puzzle)
			require.NoError(t, err)

			test.verify(t, puzzle)
		})
	}
}

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
			input:        load(t, "xwordinfo-nyt-20181231.json"),
			num:          1,
			direction:    "a",
			expectedMinX: 0,
			expectedMinY: 0,
			expectedMaxX: 4,
			expectedMaxY: 0,
		},
		{
			name:         "6a",
			input:        load(t, "xwordinfo-nyt-20181231.json"),
			num:          6,
			direction:    "a",
			expectedMinX: 6,
			expectedMinY: 0,
			expectedMaxX: 10,
			expectedMaxY: 0,
		},
		{
			name:         "11a",
			input:        load(t, "xwordinfo-nyt-20181231.json"),
			num:          11,
			direction:    "a",
			expectedMinX: 12,
			expectedMinY: 0,
			expectedMaxX: 14,
			expectedMaxY: 0,
		},
		{
			name:         "1d",
			input:        load(t, "xwordinfo-nyt-20181231.json"),
			num:          1,
			direction:    "d",
			expectedMinX: 0,
			expectedMinY: 0,
			expectedMaxX: 0,
			expectedMaxY: 3,
		},
		{
			name:         "30d",
			input:        load(t, "xwordinfo-nyt-20181231.json"),
			num:          30,
			direction:    "d",
			expectedMinX: 0,
			expectedMinY: 6,
			expectedMaxX: 0,
			expectedMaxY: 8,
		},
		{
			name:         "51d",
			input:        load(t, "xwordinfo-nyt-20181231.json"),
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
			input:     load(t, "xwordinfo-nyt-20181231.json"),
			num:       66,
			direction: "a",
		},
		{
			name:      "66d",
			input:     load(t, "xwordinfo-nyt-20181231.json"),
			num:       66,
			direction: "d",
		},
		{
			name:      "2a",
			input:     load(t, "xwordinfo-nyt-20181231.json"),
			num:       2,
			direction: "a",
		},
		{
			name:      "14d",
			input:     load(t, "xwordinfo-nyt-20181231.json"),
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
