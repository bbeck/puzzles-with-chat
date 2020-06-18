package crossword

import (
	"bytes"
	"io"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseXWordInfoClue(t *testing.T) {
	tests := []struct {
		input  string
		number int
		clue   string
	}{
		{
			input:  "1. clue",
			number: 1,
			clue:   "clue",
		},
		{
			input:  "2. multiple words in clue",
			number: 2,
			clue:   "multiple words in clue",
		},
		{
			input:  "3.  leading whitespace",
			number: 3,
			clue:   "leading whitespace",
		},
		{
			input:  "4. trailing whitespace ",
			number: 4,
			clue:   "trailing whitespace",
		},
		{
			input:  "5.   leading and trailing whitespace ",
			number: 5,
			clue:   "leading and trailing whitespace",
		},
		{
			input:  "6. internal  whitespace  preserved",
			number: 6,
			clue:   "internal  whitespace  preserved",
		},
		{
			input:  "7.  \t\n multiple types of whitespace \t\n",
			number: 7,
			clue:   "multiple types of whitespace",
		},
		{
			input:  "8. 4.0 is a great one",
			number: 8,
			clue:   "4.0 is a great one",
		},
		{
			input:  "9. &quot;Look out!&quot;",
			number: 9,
			clue:   `"Look out!"`,
		},
		{
			input:  "10. ___ raving mad",
			number: 10,
			clue:   "___ raving mad",
		},
		{
			input:  "11.", // no clue text, this sometimes happens with theme clues
			number: 11,
		},
		{
			input:  "12. \t\n ", // only whitespace
			number: 12,
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			num, clue, err := ParseXWordInfoClue(test.input)
			require.NoError(t, err)
			assert.Equal(t, test.number, num)
			assert.Equal(t, test.clue, clue)
		})
	}
}

func TestParseXWordInfoClue_Error(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty clue",
			input: "",
		},
		{
			name:  "no number",
			input: "clue",
		},
		{
			name:  "clue letter instead of number",
			input: "A. clue",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, err := ParseXWordInfoClue(test.input)
			assert.Error(t, err)
		})
	}
}

func TestParseXWordInfoResponse(t *testing.T) {
	tests := []struct {
		name   string
		input  io.ReadCloser
		verify func(t *testing.T, puzzle *Puzzle)
	}{
		{
			name:  "description",
			input: load(t, "xwordinfo-nyt-20181231.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := "New York Times puzzle from 2018-12-31"
				assert.Equal(t, expected, puzzle.Description)
			},
		},
		{
			name:  "size",
			input: load(t, "xwordinfo-nyt-20181231.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, 15, puzzle.Cols)
				assert.Equal(t, 15, puzzle.Rows)
			},
		},
		{
			name:  "title",
			input: load(t, "xwordinfo-nyt-20181231.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, "NY Times, Mon, Dec 31, 2018", puzzle.Title)
			},
		},
		{
			name:  "publisher",
			input: load(t, "xwordinfo-nyt-20181231.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, "The New York Times", puzzle.Publisher)
			},
		},
		{
			name:  "published date",
			input: load(t, "xwordinfo-nyt-20181231.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, 2018, puzzle.PublishedDate.Year())
				assert.Equal(t, time.December, puzzle.PublishedDate.Month())
				assert.Equal(t, 31, puzzle.PublishedDate.Day())
			},
		},
		{
			name:  "author",
			input: load(t, "xwordinfo-nyt-20181231.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, "Brian Thomas", puzzle.Author)
			},
		},
		{
			name:  "cells",
			input: load(t, "xwordinfo-nyt-20181231.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := [][]string{
					{"Q", "A", "N", "D", "A", "", "A", "T", "T", "I", "C", "", "H", "O", "N"},
					{"T", "H", "I", "R", "D", "", "L", "A", "I", "D", "A", "S", "I", "D", "E"},
					{"I", "M", "T", "O", "O", "O", "L", "D", "F", "O", "R", "T", "H", "I", "S"},
					{"P", "E", "R", "U", "", "L", "E", "A", "F", "", "P", "E", "O", "N", "S"},
					{"", "D", "O", "G", "T", "A", "G", "", "", "L", "O", "L", "", "", ""},
					{"", "", "", "H", "A", "V", "E", "N", "O", "O", "O", "M", "P", "H", ""},
					{"M", "A", "T", "T", "E", "", "", "I", "M", "P", "L", "O", "R", "E", "D"},
					{"E", "R", "R", "", "", "R", "A", "N", "G", "E", "", "", "E", "M", "O"},
					{"W", "A", "I", "T", "H", "E", "R", "E", "", "", "E", "G", "Y", "P", "T"},
					{"", "B", "O", "O", "O", "F", "F", "S", "T", "A", "G", "E", "", "", ""},
					{"", "", "", "E", "R", "S", "", "", "E", "U", "G", "E", "N", "E", ""},
					{"S", "H", "A", "R", "I", "", "S", "I", "N", "N", "", "W", "I", "N", "G"},
					{"I", "T", "S", "A", "Z", "O", "O", "O", "U", "T", "T", "H", "E", "R", "E"},
					{"S", "T", "E", "G", "O", "S", "A", "U", "R", "", "H", "I", "T", "O", "N"},
					{"I", "P", "A", "", "N", "U", "R", "S", "E", "", "O", "Z", "O", "N", "E"},
				}
				assert.Equal(t, expected, puzzle.Cells)
			},
		},
		{
			name:  "blocks",
			input: load(t, "xwordinfo-nyt-20181231.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := [][]bool{
					{false, false, false, false, false, true, false, false, false, false, false, true, false, false, false},
					{false, false, false, false, false, true, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, true, false, false, false, false, true, false, false, false, false, false},
					{true, false, false, false, false, false, false, true, true, false, false, false, true, true, true},
					{true, true, true, false, false, false, false, false, false, false, false, false, false, false, true},
					{false, false, false, false, false, true, true, false, false, false, false, false, false, false, false},
					{false, false, false, true, true, false, false, false, false, false, true, true, false, false, false},
					{false, false, false, false, false, false, false, false, true, true, false, false, false, false, false},
					{true, false, false, false, false, false, false, false, false, false, false, false, true, true, true},
					{true, true, true, false, false, false, true, true, false, false, false, false, false, false, true},
					{false, false, false, false, false, true, false, false, false, false, true, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, true, false, false, false, false, false},
					{false, false, false, true, false, false, false, false, false, true, false, false, false, false, false},
				}
				assert.Equal(t, expected, puzzle.CellBlocks)
			},
		},
		{
			name:  "cell clue numbers",
			input: load(t, "xwordinfo-nyt-20181231.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := [][]int{
					{1, 2, 3, 4, 5, 0, 6, 7, 8, 9, 10, 0, 11, 12, 13},
					{14, 0, 0, 0, 0, 0, 15, 0, 0, 0, 0, 16, 0, 0, 0},
					{17, 0, 0, 0, 0, 18, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					{19, 0, 0, 0, 0, 20, 0, 0, 0, 0, 21, 0, 0, 0, 0},
					{0, 22, 0, 0, 23, 0, 0, 0, 0, 24, 0, 0, 0, 0, 0},
					{0, 0, 0, 25, 0, 0, 0, 26, 27, 0, 0, 0, 28, 29, 0},
					{30, 31, 32, 0, 0, 0, 0, 33, 0, 0, 0, 0, 0, 0, 34},
					{35, 0, 0, 0, 0, 36, 37, 0, 0, 0, 0, 0, 38, 0, 0},
					{39, 0, 0, 40, 41, 0, 0, 0, 0, 0, 42, 43, 0, 0, 0},
					{0, 44, 0, 0, 0, 0, 0, 0, 45, 46, 0, 0, 0, 0, 0},
					{0, 0, 0, 47, 0, 0, 0, 0, 48, 0, 0, 0, 49, 50, 0},
					{51, 52, 53, 0, 0, 0, 54, 55, 0, 0, 0, 56, 0, 0, 57},
					{58, 0, 0, 0, 0, 59, 0, 0, 0, 0, 60, 0, 0, 0, 0},
					{61, 0, 0, 0, 0, 0, 0, 0, 0, 0, 62, 0, 0, 0, 0},
					{63, 0, 0, 0, 64, 0, 0, 0, 0, 0, 65, 0, 0, 0, 0},
				}
				assert.Equal(t, expected, puzzle.CellClueNumbers)
			},
		},
		{
			name:  "cell circles (none)",
			input: load(t, "xwordinfo-nyt-20181231.json"),
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
			input: load(t, "xwordinfo-nyt-20001031-circles.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := [][]bool{
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, true, true, true, false, false, false, false, false, false},
					{false, false, false, false, false, true, false, false, false, true, false, false, false, false, false},
					{false, false, false, false, false, false, true, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, true, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, true, false, false, false, false, false, false},
					{false, false, false, false, false, true, false, false, false, true, false, false, false, false, false},
					{false, false, false, false, false, false, true, true, true, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
				}
				assert.Equal(t, expected, puzzle.CellCircles)
			},
		},
		{
			name:  "cell shades (none)",
			input: load(t, "xwordinfo-nyt-20181231.json"),
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
				assert.Equal(t, expected, puzzle.CellShades)
			},
		},
		{
			name:  "cell shades",
			input: load(t, "xwordinfo-nyt-20181216-shades.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := [][]bool{
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, true, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, true, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, true, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, true, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, true, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
				}
				assert.Equal(t, expected, puzzle.CellShades)
			},
		},
		{
			name:  "across clues",
			input: load(t, "xwordinfo-nyt-20181231.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := map[int]string{
					1:  "Exchange after a lecture, informally",
					6:  "Room just under the roof",
					11: "Sweetheart",
					14: "Base just before home base",
					15: "Postponed for later consideration",
					17: `"You young people go ahead!"`,
					19: "Country between Ecuador and Bolivia",
					20: "Part of a tree or a book",
					21: "Lowest workers",
					22: "G.I.'s ID",
					24: `"That's so funny," in a text`,
					25: "Lack in energy",
					30: "Dull, as a finish",
					33: "Begged earnestly",
					35: "Make a goof",
					36: "Free-___ (like some chickens)",
					38: "Punk offshoot",
					39: `"Don't leave this spot"`,
					42: "Cairo's land",
					44: "Force to exit, as a performer",
					47: "Hosp. trauma centers",
					48: "Broadway's ___ O'Neill Theater",
					51: "Puppeteer Lewis",
					54: "___ Fein (Irish political party)",
					56: "Either side of an airplane",
					58: "Traffic reporter's comment",
					61: "Plant-eating dino with spikes on its back",
					62: "Discover almost by chance, as a solution",
					63: "Hoppy brew, for short",
					64: "Helper in an operating room",
					65: "Another name for O3 (as appropriate to 17-, 25-, 44- and 58-Across?)",
				}
				assert.Equal(t, expected, puzzle.CluesAcross)
			},
		},
		{
			name:  "down clues",
			input: load(t, "xwordinfo-nyt-20181231.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := map[int]string{
					1:  `Brand of swabs`,
					2:  `Man's name related to the name of Islam's founder`,
					3:  `Lead-in to glycerin`,
					4:  `Prolonged dry spell`,
					5:  `"Much ___ About Nothing"`,
					6:  `Assert without proof`,
					7:  `Cry of triumph`,
					8:  `Spat`,
					9:  `Last words before being pronounced husband and wife`,
					10: `Not drive by oneself to work`,
					11: `Cheery greeting`,
					12: `Ares : Greek :: ___ : Norse`,
					13: `Loch ___ monster`,
					16: `Patron of sailors`,
					18: `Kingly name in Norway`,
					23: `___ Bo (exercise system)`,
					24: `Make great strides?`,
					26: `Highest digits in sudoku`,
					27: `"Holy cow!," in a text`,
					28: `Quarry`,
					29: `Plant supplying burlap fiber`,
					30: `Kitten's sound`,
					31: `Spirited horse`,
					32: `Sextet halved`,
					34: `"i" or "j" topper`,
					36: `Dictionaries, almanacs, etc., in brief`,
					37: `Poodle's sound`,
					40: `Scoundrel, in British slang`,
					41: `What a setting sun dips below`,
					42: `Urge (on)`,
					43: `"Who'da thunk it?!"`,
					45: `Professor's goal, one day`,
					46: `___ Jemima`,
					49: `Mexican president Enrique Peña ___`,
					50: `Company in a 2001-02 business scandal`,
					51: `Enthusiastic assent in Mexico`,
					52: `Web address starter`,
					53: `On the waves`,
					54: `Fly high`,
					55: `Notes from players who can't pay`,
					57: `Bit of inheritance?`,
					59: `The Buckeyes of the Big Ten, for short`,
					60: `However, briefly`,
				}
				assert.Equal(t, expected, puzzle.CluesDown)
			},
		},
		{
			name:  "notes",
			input: load(t, "xwordinfo-nyt-20181231.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, "", puzzle.Notes)
			},
		},
		{
			name:  "rebus",
			input: load(t, "xwordinfo-nyt-20181227-rebus.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, "CON", puzzle.Cells[6][8])
				assert.Equal(t, "CON", puzzle.Cells[7][7])
				assert.Equal(t, "CON", puzzle.Cells[8][6])
				assert.Equal(t, "CON", puzzle.Cells[9][5])
				assert.Equal(t, "CON", puzzle.Cells[10][4])
			},
		},
		{
			name:  "non-square dimensions",
			input: load(t, "xwordinfo-nyt-20180621-nonsquare.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, 13, puzzle.Rows)
				assert.Equal(t, 17, puzzle.Cols)
			},
		},
		{
			name:  "non-square cells",
			input: load(t, "xwordinfo-nyt-20180621-nonsquare.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, 13, puzzle.Rows)
				assert.Equal(t, 17, puzzle.Cols)

				expected := [][]string{
					{"T", "H", "R", "U", "", "S", "T", "A", "M", "E", "N", "", "", "P", "E", "R", "I"},
					{"E", "Y", "E", "S", "", "C", "A", "R", "P", "E", "T", "", "B", "A", "R", "O", "N"},
					{"A", "D", "I", "O", "S", "A", "M", "I", "G", "O", "S", "", "A", "S", "I", "A", "N"},
					{"M", "E", "N", "", "T", "R", "E", "E", "", "", "B", "I", "T", "T", "E", "R", "S"},
					{"", "", "", "A", "Y", "E", "", "S", "A", "G", "", "C", "I", "O", "", "", ""},
					{"P", "E", "P", "P", "E", "R", "S", "", "B", "R", "I", "C", "K", "R", "O", "A", "D"},
					{"U", "R", "L", "S", "", "", "A", "F", "O", "U", "L", "", "", "A", "R", "I", "A"},
					{"B", "A", "Y", "P", "A", "C", "K", "E", "R", "", "L", "A", "N", "T", "E", "R", "N"},
					{"", "", "", "A", "R", "E", "", "E", "T", "A", "", "L", "E", "E", "", "", ""},
					{"M", "E", "A", "N", "I", "E", "S", "", "", "L", "A", "W", "S", "", "H", "E", "N"},
					{"A", "U", "D", "I", "S", "", "O", "F", "F", "O", "N", "E", "S", "G", "A", "M", "E"},
					{"P", "R", "O", "S", "E", "", "H", "E", "A", "R", "T", "S", "", "R", "A", "I", "N"},
					{"P", "O", "S", "H", "", "", "O", "W", "N", "S", "I", "T", "", "O", "G", "L", "E"},
				}
				assert.Equal(t, expected, puzzle.Cells)
			},
		},
		{
			name:  "notepad",
			input: load(t, "xwordinfo-nyt-20180119-notepad.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.NotEmpty(t, puzzle.Notes)
			},
		},
		{
			name:  "notepad + jnotes",
			input: load(t, "xwordinfo-nyt-20110513-notepad-and-jnotes.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.NotEmpty(t, puzzle.Notes)

				// Beginning of notepad
				assert.True(t, strings.Contains(puzzle.Notes, "Every length of answer"))

				// Part of jnotes
				assert.True(t, strings.Contains(puzzle.Notes, "print version"))

				// Notepad and jnotes should be joined with a line break
				assert.True(t, strings.Contains(puzzle.Notes, "<br/>"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer test.input.Close()

			puzzle, err := ParseXWordInfoResponse(test.input)
			require.NoError(t, err)
			test.verify(t, puzzle)
		})
	}
}

func TestParseXWordInfoResponse_Error(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "malformed response",
			input: `{true}`,
		},
		{
			name:  "empty response",
			input: ``,
		},
		{
			name:  "empty puzzle",
			input: `{}`,
		},
		{
			name:  "malformed published date",
			input: `{"grid":["a","b","c","d"], "date":"hello world"}`,
		},
		{
			name: "malformed across clue",
			input: `{
								"date": "01/01/2019",
								"grid": ["a","b","c","d"],
								"clues": {
									"across": [
                    "first clue",
                    "2. second clue"
                  ],
									"down": [
                    "1. first clue",
                    "2. second clue"
                  ]
								}
							}`,
		},
		{
			name: "malformed down clue",
			input: `{
								"date": "01/01/2019",
								"grid": ["a","b","c","d"],
								"clues": {
									"across": [
                    "1. first clue",
                    "2. second clue"
                  ],
									"down": [
                    "1. first clue",
                    "second clue"
                  ]
								}
							}`,
		},
		{
			name:  "missing puzzle response",
			input: toString(t, load(t, "xwordinfo-nyt-19000513-failure.json")),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseXWordInfoResponse(strings.NewReader(test.input))
			require.Error(t, err)
		})
	}
}

func TestLoadAvailableNYTDates(t *testing.T) {
	tests := []struct {
		name     string
		expected time.Time
	}{
		{
			name:     "first puzzle date",
			expected: NYTFirstPuzzleDate,
		},
		{
			name:     "first non-sunday puzzle",
			expected: NYTSwitchToDailyDate,
		},
		{
			name:     "1943-01-03",
			expected: time.Date(1943, time.January, 3, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1944-01-02",
			expected: time.Date(1944, time.January, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1945-01-07",
			expected: time.Date(1945, time.January, 7, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1946-01-06",
			expected: time.Date(1946, time.January, 6, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1947-01-05",
			expected: time.Date(1947, time.January, 5, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1948-01-04",
			expected: time.Date(1948, time.January, 4, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1949-01-02",
			expected: time.Date(1949, time.January, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1950-01-01",
			expected: time.Date(1950, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1951-01-01",
			expected: time.Date(1951, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1952-01-01",
			expected: time.Date(1952, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1953-01-01",
			expected: time.Date(1953, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1954-01-01",
			expected: time.Date(1954, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1955-01-01",
			expected: time.Date(1955, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1956-01-01",
			expected: time.Date(1956, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1957-01-01",
			expected: time.Date(1957, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1958-01-01",
			expected: time.Date(1958, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1959-01-01",
			expected: time.Date(1959, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1960-01-01",
			expected: time.Date(1960, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1961-01-01",
			expected: time.Date(1961, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1962-01-01",
			expected: time.Date(1962, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1963-01-01",
			expected: time.Date(1963, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1964-01-01",
			expected: time.Date(1964, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1965-01-01",
			expected: time.Date(1965, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1966-01-01",
			expected: time.Date(1966, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1967-01-01",
			expected: time.Date(1967, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1968-01-01",
			expected: time.Date(1968, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1969-01-01",
			expected: time.Date(1969, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1970-01-01",
			expected: time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1971-01-01",
			expected: time.Date(1971, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1972-01-01",
			expected: time.Date(1972, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1973-01-01",
			expected: time.Date(1973, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1974-01-01",
			expected: time.Date(1974, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1975-01-01",
			expected: time.Date(1975, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1976-01-01",
			expected: time.Date(1976, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1977-01-01",
			expected: time.Date(1977, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1978-01-01",
			expected: time.Date(1978, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1979-01-01",
			expected: time.Date(1979, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1980-01-01",
			expected: time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1981-01-01",
			expected: time.Date(1981, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1982-01-01",
			expected: time.Date(1982, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1983-01-01",
			expected: time.Date(1983, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1984-01-01",
			expected: time.Date(1984, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1985-01-01",
			expected: time.Date(1985, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1986-01-01",
			expected: time.Date(1986, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1987-01-01",
			expected: time.Date(1987, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1988-01-01",
			expected: time.Date(1988, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1989-01-01",
			expected: time.Date(1989, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1990-01-01",
			expected: time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1991-01-01",
			expected: time.Date(1991, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1992-01-01",
			expected: time.Date(1992, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1993-01-01",
			expected: time.Date(1993, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1994-01-01",
			expected: time.Date(1994, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1995-01-01",
			expected: time.Date(1995, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1996-01-01",
			expected: time.Date(1996, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1997-01-01",
			expected: time.Date(1997, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1998-01-01",
			expected: time.Date(1998, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "1999-01-01",
			expected: time.Date(1999, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "2000-01-01",
			expected: time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "2001-01-01",
			expected: time.Date(2001, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "2002-01-01",
			expected: time.Date(2002, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "2003-01-01",
			expected: time.Date(2003, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "2004-01-01",
			expected: time.Date(2004, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "2005-01-01",
			expected: time.Date(2005, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "2006-01-01",
			expected: time.Date(2006, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "2007-01-01",
			expected: time.Date(2007, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "2008-01-01",
			expected: time.Date(2008, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "2009-01-01",
			expected: time.Date(2009, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "2010-01-01",
			expected: time.Date(2010, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "2011-01-01",
			expected: time.Date(2011, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "2012-01-01",
			expected: time.Date(2012, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "2013-01-01",
			expected: time.Date(2013, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "2014-01-01",
			expected: time.Date(2014, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "2015-01-01",
			expected: time.Date(2015, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "2016-01-01",
			expected: time.Date(2016, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "2017-01-01",
			expected: time.Date(2017, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "2018-01-01",
			expected: time.Date(2018, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "2019-01-01",
			expected: time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "2020-01-01",
			expected: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "today",
			expected: time.Now().UTC().Truncate(24 * time.Hour),
		},
	}

	dates := LoadAvailableNYTDates()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.True(t, sort.SliceIsSorted(dates, func(i, j int) bool {
				return dates[i].Before(dates[j])
			}))

			index := sort.Search(len(dates), func(i int) bool {
				return dates[i].Equal(test.expected) || dates[i].After(test.expected)
			})
			assert.Equal(t, test.expected, dates[index])
		})
	}
}

func toString(t *testing.T, r io.ReadCloser) string {
	t.Helper()
	defer r.Close()

	buf := bytes.NewBuffer(nil)
	_, err := io.Copy(buf, r)
	require.NoError(t, err)

	return string(buf.Bytes())
}
