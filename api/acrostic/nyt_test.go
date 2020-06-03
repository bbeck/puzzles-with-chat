package acrostic

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"strings"
	"testing"
	"time"
)

func TestGetClueLetter(t *testing.T) {
	tests := []struct {
		name     string
		index    int
		expected string
	}{
		{name: "A", index: 0, expected: "A"},
		{name: "B", index: 1, expected: "B"},
		{name: "C", index: 2, expected: "C"},
		{name: "D", index: 3, expected: "D"},
		{name: "E", index: 4, expected: "E"},
		{name: "F", index: 5, expected: "F"},
		{name: "G", index: 6, expected: "G"},
		{name: "H", index: 7, expected: "H"},
		{name: "I", index: 8, expected: "I"},
		{name: "J", index: 9, expected: "J"},
		{name: "K", index: 10, expected: "K"},
		{name: "L", index: 11, expected: "L"},
		{name: "M", index: 12, expected: "M"},
		{name: "N", index: 13, expected: "N"},
		{name: "O", index: 14, expected: "O"},
		{name: "P", index: 15, expected: "P"},
		{name: "Q", index: 16, expected: "Q"},
		{name: "R", index: 17, expected: "R"},
		{name: "S", index: 18, expected: "S"},
		{name: "T", index: 19, expected: "T"},
		{name: "U", index: 20, expected: "U"},
		{name: "V", index: 21, expected: "V"},
		{name: "W", index: 22, expected: "W"},
		{name: "X", index: 23, expected: "X"},
		{name: "Y", index: 24, expected: "Y"},
		{name: "Z", index: 25, expected: "Z"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := GetClueLetter(test.index)
			require.NoError(t, err)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetClueLetter_Error(t *testing.T) {
	tests := []struct {
		name  string
		index int
	}{
		{name: "-1", index: -1},
		{name: "26", index: 26},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := GetClueLetter(test.index)
			require.Error(t, err)
		})
	}
}

func TestParseInts(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []int
	}{
		{
			name:     "one int",
			input:    "1",
			expected: []int{1},
		},
		{
			name:     "multiple ints",
			input:    "1,2,3,12",
			expected: []int{1, 2, 3, 12},
		},
		{
			name:     "non-sorted ints",
			input:    "12,3,1,2",
			expected: []int{12, 3, 1, 2},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := ParseInts(test.input)
			require.NoError(t, err)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestParseInts_Error(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "non integer",
			input: "a",
		},
		{
			name:  "non integer embedded within integers",
			input: "1,2,3,a,5",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseInts(test.input)
			require.Error(t, err)
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
			input: load(t, "xwordinfo-nyt-20200524.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := "New York Times puzzle from 2020-05-24"
				assert.Equal(t, expected, puzzle.Description)
			},
		},
		{
			name:  "size",
			input: load(t, "xwordinfo-nyt-20200524.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, 27, puzzle.Cols)
				assert.Equal(t, 8, puzzle.Rows)
			},
		},
		{
			name:  "publisher",
			input: load(t, "xwordinfo-nyt-20200524.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, "The New York Times", puzzle.Publisher)
			},
		},
		{
			name:  "published date",
			input: load(t, "xwordinfo-nyt-20200524.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, 2020, puzzle.PublishedDate.Year())
				assert.Equal(t, time.May, puzzle.PublishedDate.Month())
				assert.Equal(t, 24, puzzle.PublishedDate.Day())
			},
		},
		{
			name:  "cells",
			input: load(t, "xwordinfo-nyt-20200524.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := [][]string{
					{"P", "E", "O", "P", "L", "E", "", "S", "E", "L", "D", "O", "M", "", "A", "P", "P", "R", "E", "C", "I", "A", "T", "E", "", "T", "H"},
					{"E", "", "V", "A", "S", "T", "", "K", "N", "O", "W", "L", "E", "D", "G", "E", "", "W", "H", "I", "C", "H", "", "O", "R", "C", "H"},
					{"E", "S", "T", "R", "A", "", "P", "L", "A", "Y", "E", "R", "S", "", "P", "O", "S", "S", "E", "S", "S", "", "M", "O", "S", "T", ""},
					{"O", "F", "", "T", "H", "E", "M", "", "P", "L", "A", "Y", "", "S", "E", "V", "E", "R", "A", "L", "", "I", "N", "S", "T", "R", "U"},
					{"M", "E", "N", "T", "S", "", "A", "N", "D", "", "T", "H", "E", "Y", "", "A", "L", "L", "", "H", "O", "L", "D", "", "A", "S", ""},
					{"A", "", "C", "R", "E", "E", "D", "", "T", "H", "A", "T", "", "A", "", "F", "A", "L", "S", "E", "", "N", "O", "T", "E", "", "I"},
					{"S", "", "A", "", "S", "I", "N", "", "A", "N", "D", "", "A", "", "V", "A", "R", "I", "A", "T", "I", "O", "N", "", "I", "N", ""},
					{"R", "H", "Y", "T", "H", "M", "", "I", "S", "", "A", "", "F", "A", "L", "L", "", "F", "R", "O", "M", "", "G", "R", "A", "C", "E"},
				}
				assert.Equal(t, expected, puzzle.Cells)
			},
		},
		{
			name:  "blocks",
			input: load(t, "xwordinfo-nyt-20200524.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := [][]bool{
					{false, false, false, false, false, false, true, false, false, false, false, false, false, true, false, false, false, false, false, false, false, false, false, false, true, false, false},
					{false, true, false, false, false, false, true, false, false, false, false, false, false, false, false, false, true, false, false, false, false, false, true, false, false, false, false},
					{false, false, false, false, false, true, false, false, false, false, false, false, false, true, false, false, false, false, false, false, false, true, false, false, false, false, true},
					{false, false, true, false, false, false, false, true, false, false, false, false, true, false, false, false, false, false, false, false, true, false, false, false, false, false, false},
					{false, false, false, false, false, true, false, false, false, true, false, false, false, false, true, false, false, false, true, false, false, false, false, true, false, false, true},
					{false, true, false, false, false, false, false, true, false, false, false, false, true, false, true, false, false, false, false, false, true, false, false, false, false, true, false},
					{false, true, false, true, false, false, false, true, false, false, false, true, false, true, false, false, false, false, false, false, false, false, false, true, false, false, true},
					{false, false, false, false, false, false, true, false, false, true, false, true, false, false, false, false, true, false, false, false, false, true, false, false, false, false, false},
				}
				assert.Equal(t, expected, puzzle.CellBlocks)
			},
		},
		{
			name:  "cell numbers",
			input: load(t, "xwordinfo-nyt-20200524.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := [][]int{
					{1, 2, 3, 4, 5, 6, 0, 7, 8, 9, 10, 11, 12, 0, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 0, 23, 24},
					{25, 0, 26, 27, 28, 29, 0, 30, 31, 32, 33, 34, 35, 36, 37, 38, 0, 39, 40, 41, 42, 43, 0, 44, 45, 46, 47},
					{48, 49, 50, 51, 52, 0, 53, 54, 55, 56, 57, 58, 59, 0, 60, 61, 62, 63, 64, 65, 66, 0, 67, 68, 69, 70, 0},
					{71, 72, 0, 73, 74, 75, 76, 0, 77, 78, 79, 80, 0, 81, 82, 83, 84, 85, 86, 87, 0, 88, 89, 90, 91, 92, 93},
					{94, 95, 96, 97, 98, 0, 99, 100, 101, 0, 102, 103, 104, 105, 0, 106, 107, 108, 0, 109, 110, 111, 112, 0, 113, 114, 0},
					{115, 0, 116, 117, 118, 119, 120, 0, 121, 122, 123, 124, 0, 125, 0, 126, 127, 128, 129, 130, 0, 131, 132, 133, 134, 0, 135},
					{136, 0, 137, 0, 138, 139, 140, 0, 141, 142, 143, 0, 144, 0, 145, 146, 147, 148, 149, 150, 151, 152, 153, 0, 154, 155, 0},
					{156, 157, 158, 159, 160, 161, 0, 162, 163, 0, 164, 0, 165, 166, 167, 168, 0, 169, 170, 171, 172, 0, 173, 174, 175, 176, 177},
				}
				assert.Equal(t, expected, puzzle.CellNumbers)
			},
		},
		{
			name:  "cell clue letters",
			input: load(t, "xwordinfo-nyt-20200524.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := [][]string{
					{"T", "I", "K", "V", "E", "R", "", "D", "O", "F", "U", "G", "B", "", "N", "Q", "M", "K", "A", "V", "P", "E", "R", "L", "", "S", "H"},
					{"F", "", "D", "U", "W", "Q", "", "T", "M", "B", "A", "I", "J", "P", "C", "L", "", "O", "N", "G", "S", "V", "", "T", "R", "K", "B"},
					{"I", "W", "P", "U", "A", "", "L", "D", "E", "M", "F", "J", "O", "", "H", "K", "V", "Q", "B", "W", "S", "", "E", "N", "A", "I", ""},
					{"F", "O", "", "R", "J", "D", "G", "", "C", "E", "V", "M", "", "W", "S", "T", "Q", "F", "O", "P", "", "K", "D", "H", "L", "V", "U"},
					{"M", "G", "R", "J", "I", "", "N", "W", "E", "", "F", "P", "T", "C", "", "L", "O", "G", "", "M", "S", "Q", "V", "", "D", "B", ""},
					{"J", "", "R", "L", "H", "U", "K", "", "G", "A", "W", "I", "", "P", "", "T", "S", "D", "O", "R", "", "F", "M", "B", "E", "", "G"},
					{"W", "", "P", "", "C", "T", "K", "", "Q", "J", "H", "", "R", "", "S", "B", "U", "D", "I", "G", "V", "T", "P", "", "W", "E", ""},
					{"H", "Q", "C", "F", "D", "N", "", "B", "M", "", "K", "", "G", "W", "A", "O", "", "Q", "T", "V", "U", "", "F", "B", "H", "K", "E"},
				}
				assert.Equal(t, expected, puzzle.CellClueLetters)
			},
		},
		{
			name:  "clues",
			input: load(t, "xwordinfo-nyt-20200524.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := map[string]string{
					"A": "Sources of nonhuman &quot;songs&quot; on a 1970s album with more than 10 million copies",
					"B": "Rock band called the &quot;Bad Boys from Boston&quot;",
					"C": "Hit Broadway musical with the songs &quot;Let Me Entertain You&quot; and &quot;You Gotta Get a Gimmick&quot;",
					"D": "Country Music Hall of Fame site",
					"E": "Turn in square dancing",
					"F": "Prop whose name comes from a French word meaning &quot;squinting&quot;",
					"G": "Recurrent theme",
					"H": "Taken from a B to a C, say",
					"I": "Fertile area for 1990s grunge music",
					"J": "Capital hosting the Fajr Music Festival ",
					"K": "Equipment for a busker, maybe",
					"L": "What &quot;da capo&quot; tells you to do",
					"M": "Work for a pit crew?",
					"N": "City that hosted jazz&#39;s storied Dreamland Ballroom",
					"O": "Brought off without a single fluff",
					"P": "Nation whose country music is &quot;luk thung&quot;",
					"Q": "Interval from ti to do (2 wds.)",
					"R": "Pause for a change of scenery, perhaps",
					"S": "Jumps a prima donna may make",
					"T": "&quot;Peter and the Wolf&quot; composer",
					"U": "Catcher of some waves",
					"V": "Affected by &quot;Ode to Joy,&quot; perhaps",
					"W": "Killer musical by Stephen Sondheim",
				}
				assert.Equal(t, expected, puzzle.Clues)
			},
		},
		{
			name:  "clue numbers",
			input: load(t, "xwordinfo-nyt-20200524.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := map[string][]int{
					"A": {33, 122, 52, 167, 17, 69},
					"B": {146, 64, 174, 32, 114, 12, 162, 133, 47},
					"C": {37, 105, 77, 138, 158},
					"D": {89, 113, 7, 160, 26, 148, 54, 128, 75},
					"E": {55, 5, 78, 134, 67, 20, 155, 101, 177},
					"F": {9, 71, 85, 173, 131, 25, 102, 159, 57},
					"G": {108, 95, 135, 150, 76, 11, 121, 41, 165},
					"H": {90, 24, 175, 156, 60, 118, 143},
					"I": {98, 48, 149, 70, 124, 34, 2},
					"J": {97, 35, 74, 58, 115, 142},
					"K": {164, 176, 46, 61, 16, 120, 88, 3, 140},
					"L": {117, 22, 53, 38, 106, 91},
					"M": {163, 80, 94, 15, 109, 132, 31, 56},
					"N": {68, 161, 13, 40, 99},
					"O": {72, 107, 86, 39, 168, 8, 129, 59},
					"P": {50, 103, 125, 19, 87, 137, 153, 36},
					"Q": {157, 141, 111, 169, 63, 29, 84, 14},
					"R": {130, 96, 21, 45, 144, 116, 73, 6},
					"S": {110, 42, 23, 127, 145, 82, 66},
					"T": {1, 170, 152, 30, 44, 126, 139, 104, 83},
					"U": {119, 27, 51, 10, 147, 93, 172},
					"V": {92, 43, 79, 4, 62, 171, 112, 151, 18},
					"W": {123, 136, 49, 166, 81, 65, 154, 100, 28},
				}
				assert.Equal(t, expected, puzzle.ClueNumbers)
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
			input: `{"date":"1/1/2001"}`,
		},
		{
			name:  "malformed published date",
			input: `{"date":"hello world"}`,
		},
		{
			name: "too many clues",
			input: `{
                "date": "1/1/2001",
                "answerKey": "answers go here",
								"clues": [
                  "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", 
                  "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", 
                  "Y", "Z", "extra"
                ],
                "clueData": [
									"1", "1", "1", "1", "1", "1", "1", "1", "1", "1", "1", "1", 
                  "1", "1", "1", "1", "1", "1", "1", "1", "1", "1", "1", "1", 
                  "1", "1", "1"
                ]
							}`,
		},
		{
			name: "bad clue data (numbers)",
			input: `{
                "date": "1/1/2001",
                "answerKey": "answers go here",
								"clues": ["A"],
                "clueData": ["1,2,a,3"]
							}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseXWordInfoResponse(strings.NewReader(test.input))
			require.Error(t, err)
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
