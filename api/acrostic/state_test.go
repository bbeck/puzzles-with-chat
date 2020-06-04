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
