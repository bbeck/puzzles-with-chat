package acrostic

import (
	"github.com/bbeck/puzzles-with-chat/api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestState_ApplyClueAnswer_Cells(t *testing.T) {
	tests := []struct {
		name     string
		filename string // The test puzzle to load
		setup    string // initial answer applied before the desired answer
		clue     string
		answer   string
		verify   func(*testing.T, State)
	}{
		{
			name:     "A",
			filename: "xwordinfo-nyt-20200524.json",
			clue:     "A",
			answer:   "WHALES",
			verify: func(t *testing.T, state State) {
				assert.Equal(t, "W", state.Cells[1][10])
				assert.Equal(t, "H", state.Cells[5][9])
				assert.Equal(t, "A", state.Cells[2][4])
				assert.Equal(t, "L", state.Cells[7][14])
				assert.Equal(t, "E", state.Cells[0][18])
				assert.Equal(t, "S", state.Cells[2][24])
			},
		},
		{
			name:     "Q (space in answer)",
			filename: "xwordinfo-nyt-20200524.json",
			clue:     "Q",
			answer:   "HALF STEP",
			verify: func(t *testing.T, state State) {
				assert.Equal(t, "H", state.Cells[7][1])
				assert.Equal(t, "A", state.Cells[6][8])
				assert.Equal(t, "L", state.Cells[4][21])
				assert.Equal(t, "F", state.Cells[7][17])
				assert.Equal(t, "S", state.Cells[2][17])
				assert.Equal(t, "T", state.Cells[1][5])
				assert.Equal(t, "E", state.Cells[3][16])
				assert.Equal(t, "P", state.Cells[0][15])
			},
		},
		{
			name:     "overwriting values",
			filename: "xwordinfo-nyt-20200524.json",
			setup:    "WHALES",
			clue:     "A",
			answer:   "ABCDEF",
			verify: func(t *testing.T, state State) {
				assert.Equal(t, "A", state.Cells[1][10])
				assert.Equal(t, "B", state.Cells[5][9])
				assert.Equal(t, "C", state.Cells[2][4])
				assert.Equal(t, "D", state.Cells[7][14])
				assert.Equal(t, "E", state.Cells[0][18])
				assert.Equal(t, "F", state.Cells[2][24])
			},
		},
		{
			name:     "unknown letters",
			filename: "xwordinfo-nyt-20200524.json",
			clue:     "A",
			answer:   "WHALE.",
			verify: func(t *testing.T, state State) {
				assert.Equal(t, "W", state.Cells[1][10])
				assert.Equal(t, "H", state.Cells[5][9])
				assert.Equal(t, "A", state.Cells[2][4])
				assert.Equal(t, "L", state.Cells[7][14])
				assert.Equal(t, "E", state.Cells[0][18])
				assert.Equal(t, "", state.Cells[2][24])
			},
		},
		{
			name:     "changing answer",
			filename: "xwordinfo-nyt-20200524.json",
			setup:    "ABCDEF",
			clue:     "A",
			answer:   "WHALES",
			verify: func(t *testing.T, state State) {
				assert.Equal(t, "W", state.Cells[1][10])
				assert.Equal(t, "H", state.Cells[5][9])
				assert.Equal(t, "A", state.Cells[2][4])
				assert.Equal(t, "L", state.Cells[7][14])
				assert.Equal(t, "E", state.Cells[0][18])
				assert.Equal(t, "S", state.Cells[2][24])
			},
		},
		{
			name:     "removing letters",
			filename: "xwordinfo-nyt-20200524.json",
			setup:    "WHALES",
			clue:     "A",
			answer:   "WHALE.",
			verify: func(t *testing.T, state State) {
				assert.Equal(t, "W", state.Cells[1][10])
				assert.Equal(t, "H", state.Cells[5][9])
				assert.Equal(t, "A", state.Cells[2][4])
				assert.Equal(t, "L", state.Cells[7][14])
				assert.Equal(t, "E", state.Cells[0][18])
				assert.Equal(t, "", state.Cells[2][24])
			},
		},
		{
			name:     "lowercase letters",
			filename: "xwordinfo-nyt-20200524.json",
			clue:     "A",
			answer:   "whales",
			verify: func(t *testing.T, state State) {
				assert.Equal(t, "W", state.Cells[1][10])
				assert.Equal(t, "H", state.Cells[5][9])
				assert.Equal(t, "A", state.Cells[2][4])
				assert.Equal(t, "L", state.Cells[7][14])
				assert.Equal(t, "E", state.Cells[0][18])
				assert.Equal(t, "S", state.Cells[2][24])
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := NewState(t, test.filename)
			if test.setup != "" {
				require.NoError(t, state.ApplyClueAnswer(test.clue, test.setup, false))
			}

			err := state.ApplyClueAnswer(test.clue, test.answer, false)
			require.NoError(t, err)
			test.verify(t, state)
		})
	}
}

func TestState_ApplyClueAnswer_Cells_CorrectOnly(t *testing.T) {
	tests := []struct {
		name     string
		filename string // The test puzzle to load
		setup    string // initial answer applied before the desired answer
		clue     string
		answer   string
		verify   func(*testing.T, State)
	}{
		{
			name:     "A",
			filename: "xwordinfo-nyt-20200524.json",
			clue:     "A",
			answer:   "WHALES",
			verify: func(t *testing.T, state State) {
				assert.Equal(t, "W", state.Cells[1][10])
				assert.Equal(t, "H", state.Cells[5][9])
				assert.Equal(t, "A", state.Cells[2][4])
				assert.Equal(t, "L", state.Cells[7][14])
				assert.Equal(t, "E", state.Cells[0][18])
				assert.Equal(t, "S", state.Cells[2][24])
			},
		},
		{
			name:     "Q (space in answer)",
			filename: "xwordinfo-nyt-20200524.json",
			clue:     "Q",
			answer:   "HALF STEP",
			verify: func(t *testing.T, state State) {
				assert.Equal(t, "H", state.Cells[7][1])
				assert.Equal(t, "A", state.Cells[6][8])
				assert.Equal(t, "L", state.Cells[4][21])
				assert.Equal(t, "F", state.Cells[7][17])
				assert.Equal(t, "S", state.Cells[2][17])
				assert.Equal(t, "T", state.Cells[1][5])
				assert.Equal(t, "E", state.Cells[3][16])
				assert.Equal(t, "P", state.Cells[0][15])
			},
		},
		{
			name:     "can fill in some unknown letters",
			filename: "xwordinfo-nyt-20200524.json",
			setup:    ".... STEP",
			clue:     "Q",
			answer:   "H..F STEP",
			verify: func(t *testing.T, state State) {
				assert.Equal(t, "H", state.Cells[7][1])
				assert.Equal(t, "", state.Cells[6][8])
				assert.Equal(t, "", state.Cells[4][21])
				assert.Equal(t, "F", state.Cells[7][17])
				assert.Equal(t, "S", state.Cells[2][17])
				assert.Equal(t, "T", state.Cells[1][5])
				assert.Equal(t, "E", state.Cells[3][16])
				assert.Equal(t, "P", state.Cells[0][15])
			},
		},
		{
			name:     "can fill in all unknown letters",
			filename: "xwordinfo-nyt-20200524.json",
			setup:    ".... STEP",
			clue:     "Q",
			answer:   "HALF STEP",
			verify: func(t *testing.T, state State) {
				assert.Equal(t, "H", state.Cells[7][1])
				assert.Equal(t, "A", state.Cells[6][8])
				assert.Equal(t, "L", state.Cells[4][21])
				assert.Equal(t, "F", state.Cells[7][17])
				assert.Equal(t, "S", state.Cells[2][17])
				assert.Equal(t, "T", state.Cells[1][5])
				assert.Equal(t, "E", state.Cells[3][16])
				assert.Equal(t, "P", state.Cells[0][15])
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := NewState(t, test.filename)
			if test.setup != "" {
				require.NoError(t, state.ApplyClueAnswer(test.clue, test.setup, true))
			}

			err := state.ApplyClueAnswer(test.clue, test.answer, true)
			require.NoError(t, err)
			test.verify(t, state)
		})
	}
}

func TestState_ApplyClueAnswer_Cells_CorrectOnly_Error(t *testing.T) {
	tests := []struct {
		name     string
		filename string // The test puzzle to load
		setup    string // initial answer applied before the desired answer
		clue     string
		answer   string
	}{
		{
			name:     "change correct value to incorrect one",
			filename: "xwordinfo-nyt-20200524.json",
			setup:    "WHALES",
			clue:     "A",
			answer:   "WHALER",
		},
		{
			name:     "incorrect answer",
			filename: "xwordinfo-nyt-20200524.json",
			clue:     "A",
			answer:   "WHALER",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := NewState(t, test.filename)
			if test.setup != "" {
				require.NoError(t, state.ApplyClueAnswer(test.clue, test.setup, true))
			}

			err := state.ApplyClueAnswer(test.clue, test.answer, true)
			require.Error(t, err)
		})
	}
}

func TestState_ApplyClueAnswer_CluesFilled(t *testing.T) {
	tests := []struct {
		name     string
		filename string // The test puzzle to load
		setup    string // initial answer applied before the desired answer
		clue     string
		answer   string
		verify   func(*testing.T, State)
	}{
		{
			name:     "A",
			filename: "xwordinfo-nyt-20200524.json",
			clue:     "A",
			answer:   "WHALES",
			verify: func(t *testing.T, state State) {
				assert.True(t, state.CluesFilled["A"])
			},
		},
		{
			name:     "incorrect answer",
			filename: "xwordinfo-nyt-20200524.json",
			clue:     "A",
			answer:   "ABCDEF",
			verify: func(t *testing.T, state State) {
				assert.True(t, state.CluesFilled["A"])
			},
		},
		{
			name:     "partial answer",
			filename: "xwordinfo-nyt-20200524.json",
			clue:     "A",
			answer:   "WHALE.",
			verify: func(t *testing.T, state State) {
				assert.False(t, state.CluesFilled["A"])
			},
		},
		{
			name:     "removing letters",
			filename: "xwordinfo-nyt-20200524.json",
			setup:    "WHALES",
			clue:     "A",
			answer:   "WHALE.",
			verify: func(t *testing.T, state State) {
				assert.False(t, state.CluesFilled["A"])
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := NewState(t, test.filename)
			if test.setup != "" {
				require.NoError(t, state.ApplyClueAnswer(test.clue, test.setup, false))
			}

			err := state.ApplyClueAnswer(test.clue, test.answer, false)
			require.NoError(t, err)
			test.verify(t, state)
		})
	}
}

func TestState_ApplyClueAnswer_Status(t *testing.T) {
	tests := []struct {
		name     string
		filename string // The test puzzle to load
		answers  map[string]string
		expected model.Status
	}{
		{
			name:     "no answers",
			filename: "xwordinfo-nyt-20200524.json",
			expected: model.StatusSolving,
		},
		{
			name:     "one answer",
			filename: "xwordinfo-nyt-20200524.json",
			answers: map[string]string{
				"A": "WHALES",
			},
			expected: model.StatusSolving,
		},
		{
			name:     "complete and correct puzzle",
			filename: "xwordinfo-nyt-20200524.json",
			answers: map[string]string{
				"A": "WHALES",
				"B": "AEROSMITH",
				"C": "GYPSY",
				"D": "NASHVILLE",
				"E": "ALLEMANDE",
				"F": "LORGNETTE",
				"G": "LEITMOTIF",
				"H": "SHARPED",
				"I": "SEATTLE",
				"J": "TEHRAN",
				"K": "ACCORDION",
				"L": "REPEAT",
				"M": "SYMPHONY",
				"N": "OMAHA",
				"O": "FLAWLESS",
				"P": "THAILAND",
				"Q": "HALF STEP",
				"R": "ENTRACTE",
				"S": "OCTAVES",
				"T": "PROKOFIEV",
				"U": "EARDRUM",
				"V": "RHAPSODIC",
				"W": "ASSASSINS",
			},
			expected: model.StatusComplete,
		},
		{
			name:     "complete and incorrect puzzle",
			filename: "xwordinfo-nyt-20200524.json",
			answers: map[string]string{
				"A": "XXXXXX",
				"B": "XXXXXXXXX",
				"C": "XXXXX",
				"D": "XXXXXXXXX",
				"E": "XXXXXXXXX",
				"F": "XXXXXXXXX",
				"G": "XXXXXXXXX",
				"H": "XXXXXXX",
				"I": "XXXXXXX",
				"J": "XXXXXX",
				"K": "XXXXXXXXX",
				"L": "XXXXXX",
				"M": "XXXXXXXX",
				"N": "XXXXX",
				"O": "XXXXXXXX",
				"P": "XXXXXXXX",
				"Q": "XXXX XXXX",
				"R": "XXXXXXXX",
				"S": "XXXXXXX",
				"T": "XXXXXXXXX",
				"U": "XXXXXXX",
				"V": "XXXXXXXXX",
				"W": "XXXXXXXXX",
			},
			expected: model.StatusSolving,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := NewState(t, test.filename)
			state.Status = model.StatusSolving

			for clue, answer := range test.answers {
				require.NoError(t, state.ApplyClueAnswer(clue, answer, false))
			}

			assert.Equal(t, test.expected, state.Status)
		})
	}
}

func TestState_ApplyClueAnswer_Error(t *testing.T) {
	tests := []struct {
		name        string
		filename    string // The test puzzle to load
		setup       func(*testing.T, State)
		clue        string
		answer      string
		onlyCorrect bool
	}{
		{
			name:     "invalid clue",
			filename: "xwordinfo-nyt-20200524.json",
			clue:     "abc",
		},
		{
			name:     "answer too short",
			filename: "xwordinfo-nyt-20200524.json",
			clue:     "A",
			answer:   "ABC",
		},
		{
			name:     "answer too long",
			filename: "xwordinfo-nyt-20200524.json",
			clue:     "A",
			answer:   "ABCDEFGHIJKLMNOPQRSTUVWXYZ",
		},
		{
			name:     "unable to determine cell coordinates",
			filename: "xwordinfo-nyt-20200524.json",
			setup: func(t *testing.T, state State) {
				// Change the underlying puzzle to reference an invalid cell number
				// for answer A.
				state.Puzzle.ClueNumbers["A"][0] = 999
			},
			clue:   "A",
			answer: "WHALES",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := NewState(t, test.filename)
			if test.setup != nil {
				test.setup(t, state)
			}

			err := state.ApplyClueAnswer(test.clue, test.answer, test.onlyCorrect)
			assert.Error(t, err)
		})
	}
}

func TestState_ApplyCellAnswer_Cells(t *testing.T) {
	tests := []struct {
		name     string
		filename string // The test puzzle to load
		setup    string // initial answer applied before the desired answer
		start    int
		answer   string
		verify   func(*testing.T, State)
	}{
		{
			name:     "first word",
			filename: "xwordinfo-nyt-20200524.json",
			start:    1,
			answer:   "PEOPLE",
			verify: func(t *testing.T, state State) {
				assert.Equal(t, "P", state.Cells[0][0])
				assert.Equal(t, "E", state.Cells[0][1])
				assert.Equal(t, "O", state.Cells[0][2])
				assert.Equal(t, "P", state.Cells[0][3])
				assert.Equal(t, "L", state.Cells[0][4])
				assert.Equal(t, "E", state.Cells[0][5])
			},
		},
		{
			name:     "last word",
			filename: "xwordinfo-nyt-20200524.json",
			start:    173,
			answer:   "GRACE",
			verify: func(t *testing.T, state State) {
				assert.Equal(t, "G", state.Cells[7][22])
				assert.Equal(t, "R", state.Cells[7][23])
				assert.Equal(t, "A", state.Cells[7][24])
				assert.Equal(t, "C", state.Cells[7][25])
				assert.Equal(t, "E", state.Cells[7][26])
			},
		},
		{
			name:     "multiple words",
			filename: "xwordinfo-nyt-20200524.json",
			start:    7,
			answer:   "SELDOM APPRECIATE",
			verify: func(t *testing.T, state State) {
				assert.Equal(t, "S", state.Cells[0][7])
				assert.Equal(t, "E", state.Cells[0][8])
				assert.Equal(t, "L", state.Cells[0][9])
				assert.Equal(t, "D", state.Cells[0][10])
				assert.Equal(t, "O", state.Cells[0][11])
				assert.Equal(t, "M", state.Cells[0][12])
				assert.Equal(t, "A", state.Cells[0][14])
				assert.Equal(t, "P", state.Cells[0][15])
				assert.Equal(t, "P", state.Cells[0][16])
				assert.Equal(t, "R", state.Cells[0][17])
				assert.Equal(t, "E", state.Cells[0][18])
				assert.Equal(t, "C", state.Cells[0][19])
				assert.Equal(t, "I", state.Cells[0][20])
				assert.Equal(t, "A", state.Cells[0][21])
				assert.Equal(t, "T", state.Cells[0][22])
				assert.Equal(t, "E", state.Cells[0][23])
			},
		},
		{
			name:     "wraps line",
			filename: "xwordinfo-nyt-20200524.json",
			start:    23,
			answer:   "THE",
			verify: func(t *testing.T, state State) {
				assert.Equal(t, "T", state.Cells[0][25])
				assert.Equal(t, "H", state.Cells[0][26])
				assert.Equal(t, "E", state.Cells[1][0])
			},
		},
		{
			name:     "unknown letters",
			filename: "xwordinfo-nyt-20200524.json",
			start:    1,
			answer:   "PE..LE",
			verify: func(t *testing.T, state State) {
				assert.Equal(t, "P", state.Cells[0][0])
				assert.Equal(t, "E", state.Cells[0][1])
				assert.Equal(t, "", state.Cells[0][2])
				assert.Equal(t, "", state.Cells[0][3])
				assert.Equal(t, "L", state.Cells[0][4])
				assert.Equal(t, "E", state.Cells[0][5])
			},
		},
		{
			name:     "changing answer",
			filename: "xwordinfo-nyt-20200524.json",
			start:    1,
			setup:    "ABCDEF",
			answer:   "PEOPLE",
			verify: func(t *testing.T, state State) {
				assert.Equal(t, "P", state.Cells[0][0])
				assert.Equal(t, "E", state.Cells[0][1])
				assert.Equal(t, "O", state.Cells[0][2])
				assert.Equal(t, "P", state.Cells[0][3])
				assert.Equal(t, "L", state.Cells[0][4])
				assert.Equal(t, "E", state.Cells[0][5])
			},
		},
		{
			name:     "removing letters",
			filename: "xwordinfo-nyt-20200524.json",
			start:    1,
			setup:    "PURPLE",
			answer:   "P..PLE",
			verify: func(t *testing.T, state State) {
				assert.Equal(t, "P", state.Cells[0][0])
				assert.Equal(t, "", state.Cells[0][1])
				assert.Equal(t, "", state.Cells[0][2])
				assert.Equal(t, "P", state.Cells[0][3])
				assert.Equal(t, "L", state.Cells[0][4])
				assert.Equal(t, "E", state.Cells[0][5])
			},
		},
		{
			name:     "lowercase",
			filename: "xwordinfo-nyt-20200524.json",
			start:    1,
			answer:   "people",
			verify: func(t *testing.T, state State) {
				assert.Equal(t, "P", state.Cells[0][0])
				assert.Equal(t, "E", state.Cells[0][1])
				assert.Equal(t, "O", state.Cells[0][2])
				assert.Equal(t, "P", state.Cells[0][3])
				assert.Equal(t, "L", state.Cells[0][4])
				assert.Equal(t, "E", state.Cells[0][5])
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := NewState(t, test.filename)
			if test.setup != "" {
				require.NoError(t, state.ApplyCellAnswer(test.start, test.answer, false))
			}

			err := state.ApplyCellAnswer(test.start, test.answer, false)
			require.NoError(t, err)
			test.verify(t, state)
		})
	}
}

func TestState_ApplyCellAnswer_Cells_CorrectOnly(t *testing.T) {
	tests := []struct {
		name     string
		filename string // The test puzzle to load
		setup    string // initial answer applied before the desired answer
		start    int
		answer   string
		verify   func(*testing.T, State)
	}{
		{
			name:     "first word",
			filename: "xwordinfo-nyt-20200524.json",
			start:    1,
			answer:   "PEOPLE",
			verify: func(t *testing.T, state State) {
				assert.Equal(t, "P", state.Cells[0][0])
				assert.Equal(t, "E", state.Cells[0][1])
				assert.Equal(t, "O", state.Cells[0][2])
				assert.Equal(t, "P", state.Cells[0][3])
				assert.Equal(t, "L", state.Cells[0][4])
				assert.Equal(t, "E", state.Cells[0][5])
			},
		},
		{
			name:     "last word",
			filename: "xwordinfo-nyt-20200524.json",
			start:    173,
			answer:   "GRACE",
			verify: func(t *testing.T, state State) {
				assert.Equal(t, "G", state.Cells[7][22])
				assert.Equal(t, "R", state.Cells[7][23])
				assert.Equal(t, "A", state.Cells[7][24])
				assert.Equal(t, "C", state.Cells[7][25])
				assert.Equal(t, "E", state.Cells[7][26])
			},
		},
		{
			name:     "multiple words",
			filename: "xwordinfo-nyt-20200524.json",
			start:    7,
			answer:   "SELDOM APPRECIATE",
			verify: func(t *testing.T, state State) {
				assert.Equal(t, "S", state.Cells[0][7])
				assert.Equal(t, "E", state.Cells[0][8])
				assert.Equal(t, "L", state.Cells[0][9])
				assert.Equal(t, "D", state.Cells[0][10])
				assert.Equal(t, "O", state.Cells[0][11])
				assert.Equal(t, "M", state.Cells[0][12])
				assert.Equal(t, "A", state.Cells[0][14])
				assert.Equal(t, "P", state.Cells[0][15])
				assert.Equal(t, "P", state.Cells[0][16])
				assert.Equal(t, "R", state.Cells[0][17])
				assert.Equal(t, "E", state.Cells[0][18])
				assert.Equal(t, "C", state.Cells[0][19])
				assert.Equal(t, "I", state.Cells[0][20])
				assert.Equal(t, "A", state.Cells[0][21])
				assert.Equal(t, "T", state.Cells[0][22])
				assert.Equal(t, "E", state.Cells[0][23])
			},
		},
		{
			name:     "wraps line",
			filename: "xwordinfo-nyt-20200524.json",
			start:    23,
			answer:   "THE",
			verify: func(t *testing.T, state State) {
				assert.Equal(t, "T", state.Cells[0][25])
				assert.Equal(t, "H", state.Cells[0][26])
				assert.Equal(t, "E", state.Cells[1][0])
			},
		},
		{
			name:     "unknown letters",
			filename: "xwordinfo-nyt-20200524.json",
			start:    1,
			answer:   "PE..LE",
			verify: func(t *testing.T, state State) {
				assert.Equal(t, "P", state.Cells[0][0])
				assert.Equal(t, "E", state.Cells[0][1])
				assert.Equal(t, "", state.Cells[0][2])
				assert.Equal(t, "", state.Cells[0][3])
				assert.Equal(t, "L", state.Cells[0][4])
				assert.Equal(t, "E", state.Cells[0][5])
			},
		},
		{
			name:     "can fill in some unknown letters",
			filename: "xwordinfo-nyt-20200524.json",
			start:    1,
			setup:    "PE..LE",
			answer:   "PEO.LE",
			verify: func(t *testing.T, state State) {
				assert.Equal(t, "P", state.Cells[0][0])
				assert.Equal(t, "E", state.Cells[0][1])
				assert.Equal(t, "O", state.Cells[0][2])
				assert.Equal(t, "", state.Cells[0][3])
				assert.Equal(t, "L", state.Cells[0][4])
				assert.Equal(t, "E", state.Cells[0][5])
			},
		},
		{
			name:     "can fill in all unknown letters",
			filename: "xwordinfo-nyt-20200524.json",
			start:    1,
			setup:    "PE..LE",
			answer:   "PEOPLE",
			verify: func(t *testing.T, state State) {
				assert.Equal(t, "P", state.Cells[0][0])
				assert.Equal(t, "E", state.Cells[0][1])
				assert.Equal(t, "O", state.Cells[0][2])
				assert.Equal(t, "P", state.Cells[0][3])
				assert.Equal(t, "L", state.Cells[0][4])
				assert.Equal(t, "E", state.Cells[0][5])
			},
		},
		{
			name:     "lowercase",
			filename: "xwordinfo-nyt-20200524.json",
			start:    1,
			answer:   "people",
			verify: func(t *testing.T, state State) {
				assert.Equal(t, "P", state.Cells[0][0])
				assert.Equal(t, "E", state.Cells[0][1])
				assert.Equal(t, "O", state.Cells[0][2])
				assert.Equal(t, "P", state.Cells[0][3])
				assert.Equal(t, "L", state.Cells[0][4])
				assert.Equal(t, "E", state.Cells[0][5])
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := NewState(t, test.filename)
			if test.setup != "" {
				require.NoError(t, state.ApplyCellAnswer(test.start, test.answer, true))
			}

			err := state.ApplyCellAnswer(test.start, test.answer, true)
			require.NoError(t, err)
			test.verify(t, state)
		})
	}
}

func TestState_ApplyCellAnswer_Cells_CorrectOnly_Error(t *testing.T) {
	tests := []struct {
		name     string
		filename string // The test puzzle to load
		setup    string // initial answer applied before the desired answer
		start    int
		answer   string
	}{
		{
			name:     "change correct value to incorrect one",
			filename: "xwordinfo-nyt-20200524.json",
			start:    1,
			setup:    "PEOPLE",
			answer:   "PURPLE",
		},
		{
			name:     "incorrect answer",
			filename: "xwordinfo-nyt-20200524.json",
			start:    1,
			answer:   "PURPLE",
		},
		{
			name:     "remove correct letters",
			filename: "xwordinfo-nyt-20200524.json",
			start:    1,
			setup:    "PEOPLE",
			answer:   "P..PLE",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := NewState(t, test.filename)
			if test.setup != "" {
				require.NoError(t, state.ApplyCellAnswer(test.start, test.setup, true))
			}

			err := state.ApplyCellAnswer(test.start, test.answer, true)
			require.Error(t, err)
		})
	}
}

func TestState_ApplyCellAnswer_CluesFilled(t *testing.T) {
	tests := []struct {
		name     string
		filename string         // The test puzzle to load
		setup    map[int]string // initial answers applied before the desired answer
		start    int
		answer   string
		verify   func(*testing.T, State)
	}{
		{
			name:     "A",
			filename: "xwordinfo-nyt-20200524.json",
			setup: map[int]string{
				33:  "W",
				122: "H",
				52:  "A",
				167: "L",
				17:  "E",
			},
			start:  69,
			answer: "S",
			verify: func(t *testing.T, state State) {
				assert.True(t, state.CluesFilled["A"])
			},
		},
		{
			name:     "incorrect answer",
			filename: "xwordinfo-nyt-20200524.json",
			setup: map[int]string{
				33:  "A",
				122: "B",
				52:  "C",
				167: "D",
				17:  "E",
			},
			start:  69,
			answer: "F",
			verify: func(t *testing.T, state State) {
				assert.True(t, state.CluesFilled["A"])
			},
		},
		{
			name:     "partial answer",
			filename: "xwordinfo-nyt-20200524.json",
			setup: map[int]string{
				122: "H",
				52:  "A",
				167: "L",
				17:  "E",
			},
			start:  69,
			answer: "S",
			verify: func(t *testing.T, state State) {
				assert.False(t, state.CluesFilled["A"])
			},
		},
		{
			name:     "removing letters",
			filename: "xwordinfo-nyt-20200524.json",
			setup: map[int]string{
				33:  "W",
				122: "H",
				52:  "A",
				167: "L",
				17:  "E",
				69:  "S",
			},
			start:  69,
			answer: ".",
			verify: func(t *testing.T, state State) {
				assert.False(t, state.CluesFilled["A"])
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := NewState(t, test.filename)
			for start, answer := range test.setup {
				require.NoError(t, state.ApplyCellAnswer(start, answer, false))
			}

			err := state.ApplyCellAnswer(test.start, test.answer, false)
			require.NoError(t, err)
			test.verify(t, state)
		})
	}
}

func TestState_ApplyCellAnswer_Status(t *testing.T) {
	tests := []struct {
		name     string
		filename string // The test puzzle to load
		answers  map[int]string
		expected model.Status
	}{
		{
			name:     "no answers",
			filename: "xwordinfo-nyt-20200524.json",
			expected: model.StatusSolving,
		},
		{
			name:     "one answer",
			filename: "xwordinfo-nyt-20200524.json",
			answers: map[int]string{
				1: "PEOPLE",
			},
			expected: model.StatusSolving,
		},
		{
			name:     "complete and correct puzzle",
			filename: "xwordinfo-nyt-20200524.json",
			answers: map[int]string{
				1:   "PEOPLE",
				7:   "SELDOM",
				13:  "APPRECIATE",
				23:  "THE",
				26:  "VAST",
				30:  "KNOWLEDGE",
				39:  "WHICH",
				44:  "ORCHESTRA",
				53:  "PLAYERS",
				60:  "POSSESS",
				67:  "MOST",
				71:  "OF",
				73:  "THEM",
				77:  "PLAY",
				81:  "SEVERAL",
				88:  "INSTRUMENTS",
				99:  "AND",
				102: "THEY",
				106: "ALL",
				109: "HOLD",
				113: "AS",
				115: "A",
				116: "CREED",
				121: "THAT",
				125: "A",
				126: "FALSE",
				131: "NOTE",
				135: "IS",
				137: "A",
				138: "SIN",
				141: "AND",
				144: "A",
				145: "VARIATION",
				154: "IN",
				156: "RHYTHM",
				162: "IS",
				164: "A",
				165: "FALL",
				169: "FROM",
				173: "GRACE",
			},
			expected: model.StatusComplete,
		},
		{
			name:     "complete and incorrect puzzle",
			filename: "xwordinfo-nyt-20200524.json",
			answers: map[int]string{
				1:   "XXXXXX",
				7:   "XXXXXX",
				13:  "XXXXXXXXXX",
				23:  "XXX",
				26:  "XXXX",
				30:  "XXXXXXXXX",
				39:  "XXXXX",
				44:  "XXXXXXXXX",
				53:  "XXXXXXX",
				60:  "XXXXXXX",
				67:  "XXXX",
				71:  "XX",
				73:  "XXXX",
				77:  "XXXX",
				81:  "XXXXXXX",
				88:  "XXXXXXXXXXX",
				99:  "XXX",
				102: "XXXX",
				106: "XXX",
				109: "XXXX",
				113: "XX",
				115: "X",
				116: "XXXXX",
				121: "XXXX",
				125: "X",
				126: "XXXXX",
				131: "XXXX",
				135: "XX",
				137: "X",
				138: "XXX",
				141: "XXX",
				144: "X",
				145: "XXXXXXXXX",
				154: "XX",
				156: "XXXXXX",
				162: "XX",
				164: "X",
				165: "XXXX",
				169: "XXXX",
				173: "XXXXX",
			},
			expected: model.StatusSolving,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := NewState(t, test.filename)
			state.Status = model.StatusSolving

			for clue, answer := range test.answers {
				require.NoError(t, state.ApplyCellAnswer(clue, answer, false))
			}

			assert.Equal(t, test.expected, state.Status)
		})
	}
}

func TestState_ApplyCellAnswer_Error(t *testing.T) {
	tests := []struct {
		name        string
		filename    string // The test puzzle to load
		setup       func(*testing.T, State)
		start       int
		answer      string
		onlyCorrect bool
	}{
		{
			name:     "start id negative",
			filename: "xwordinfo-nyt-20200524.json",
			start:    0,
			answer:   "PEOPLE",
		},
		{
			name:     "start id zero",
			filename: "xwordinfo-nyt-20200524.json",
			start:    0,
			answer:   "PEOPLE",
		},
		{
			name:     "start id too large",
			filename: "xwordinfo-nyt-20200524.json",
			start:    178,
			answer:   "PEOPLE",
		},
		{
			name:     "empty answer",
			filename: "xwordinfo-nyt-20200524.json",
			start:    1,
			answer:   "",
		},
		{
			name:     "answer overruns last cell",
			filename: "xwordinfo-nyt-20200524.json",
			start:    173,
			answer:   "ABCDEF",
		},
		{
			name:     "only spaces",
			filename: "xwordinfo-nyt-20200524.json",
			start:    1,
			answer:   "      ",
		},
		{
			name:     "unable to determine cell coordinates",
			filename: "xwordinfo-nyt-20200524.json",
			setup: func(t *testing.T, state State) {
				// Change the underlying puzzle to reference an invalid cell number
				// for answer A.
				state.Puzzle.ClueNumbers["A"][0] = 999
			},
			start:  1,
			answer: "PEOPLE",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := NewState(t, test.filename)
			if test.setup != nil {
				test.setup(t, state)
			}

			err := state.ApplyCellAnswer(test.start, test.answer, false)
			assert.Error(t, err)
		})
	}
}

func TestState_ClearIncorrectCells(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		setup    map[string]string
		verify   func(*testing.T, State)
	}{
		{
			name:     "correct answer",
			filename: "xwordinfo-nyt-20200524.json",
			setup: map[string]string{
				"A": "WHALES",
			},
			verify: func(t *testing.T, state State) {
				assert.Equal(t, "W", state.Cells[1][10])
				assert.Equal(t, "H", state.Cells[5][9])
				assert.Equal(t, "A", state.Cells[2][4])
				assert.Equal(t, "L", state.Cells[7][14])
				assert.Equal(t, "E", state.Cells[0][18])
				assert.Equal(t, "S", state.Cells[2][24])
				assert.True(t, state.CluesFilled["A"])
			},
		},
		{
			name:     "partially incorrect answer",
			filename: "xwordinfo-nyt-20200524.json",
			setup: map[string]string{
				"A": "WHARFS",
			},
			verify: func(t *testing.T, state State) {
				assert.Equal(t, "W", state.Cells[1][10])
				assert.Equal(t, "H", state.Cells[5][9])
				assert.Equal(t, "A", state.Cells[2][4])
				assert.Equal(t, "", state.Cells[7][14])
				assert.Equal(t, "", state.Cells[0][18])
				assert.Equal(t, "S", state.Cells[2][24])
				assert.False(t, state.CluesFilled["A"])
			},
		},
		{
			name:     "completely incorrect answer",
			filename: "xwordinfo-nyt-20200524.json",
			setup: map[string]string{
				"A": "XXXXXX",
			},
			verify: func(t *testing.T, state State) {
				assert.Equal(t, "", state.Cells[1][10])
				assert.Equal(t, "", state.Cells[5][9])
				assert.Equal(t, "", state.Cells[2][4])
				assert.Equal(t, "", state.Cells[7][14])
				assert.Equal(t, "", state.Cells[0][18])
				assert.Equal(t, "", state.Cells[2][24])
				assert.False(t, state.CluesFilled["A"])
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := NewState(t, test.filename)
			for clue, answer := range test.setup {
				require.NoError(t, state.ApplyClueAnswer(clue, answer, false))
			}

			err := state.ClearIncorrectCells()
			assert.NoError(t, err)
			test.verify(t, state)
		})
	}
}
