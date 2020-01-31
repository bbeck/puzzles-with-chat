package crossword

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestState_ApplyAnswer_Cells(t *testing.T) {
	tests := []struct {
		name   string
		puzzle string
		clue   string
		answer string
		verify func(*testing.T, *State)
	}{
		{
			name:   "across answer",
			puzzle: "xwordinfo-success-20181231.json",
			clue:   "1a",
			answer: "Q AND A",
			verify: func(t *testing.T, state *State) {
				assert.Equal(t, "Q", state.Cells[0][0])
				assert.Equal(t, "A", state.Cells[0][1])
				assert.Equal(t, "N", state.Cells[0][2])
				assert.Equal(t, "D", state.Cells[0][3])
				assert.Equal(t, "A", state.Cells[0][4])
			},
		},
		{
			name:   "down answer",
			puzzle: "xwordinfo-success-20181231.json",
			clue:   "1d",
			answer: "QTIP",
			verify: func(t *testing.T, state *State) {
				assert.Equal(t, "Q", state.Cells[0][0])
				assert.Equal(t, "T", state.Cells[1][0])
				assert.Equal(t, "I", state.Cells[2][0])
				assert.Equal(t, "P", state.Cells[3][0])
			},
		},
		{
			name:   "overwriting across answer",
			puzzle: "xwordinfo-success-20181231.json",
			clue:   "1a",
			answer: "XXXXX",
			verify: func(t *testing.T, state *State) {
				err := state.ApplyAnswer("1a", "Q AND A", false)
				require.NoError(t, err)

				assert.Equal(t, "Q", state.Cells[0][0])
				assert.Equal(t, "A", state.Cells[0][1])
				assert.Equal(t, "N", state.Cells[0][2])
				assert.Equal(t, "D", state.Cells[0][3])
				assert.Equal(t, "A", state.Cells[0][4])
			},
		},
		{
			name:   "overwriting down answer",
			puzzle: "xwordinfo-success-20181231.json",
			clue:   "1d",
			answer: "XXXX",
			verify: func(t *testing.T, state *State) {
				err := state.ApplyAnswer("1d", "QTIP", false)
				require.NoError(t, err)

				assert.Equal(t, "Q", state.Cells[0][0])
				assert.Equal(t, "T", state.Cells[1][0])
				assert.Equal(t, "I", state.Cells[2][0])
				assert.Equal(t, "P", state.Cells[3][0])
			},
		},
		{
			name:   "unknown letters",
			puzzle: "xwordinfo-success-20181231.json",
			clue:   "1a",
			answer: ". AND .",
			verify: func(t *testing.T, state *State) {
				assert.Equal(t, "", state.Cells[0][0])
				assert.Equal(t, "A", state.Cells[0][1])
				assert.Equal(t, "N", state.Cells[0][2])
				assert.Equal(t, "D", state.Cells[0][3])
				assert.Equal(t, "", state.Cells[0][4])
			},
		},
		{
			name:   "delete part of answer",
			puzzle: "xwordinfo-success-20181231.json",
			clue:   "1a",
			answer: "A AND Q",
			verify: func(t *testing.T, state *State) {
				err := state.ApplyAnswer("1a", ".AND.", false)
				require.NoError(t, err)

				assert.Equal(t, "", state.Cells[0][0])
				assert.Equal(t, "A", state.Cells[0][1])
				assert.Equal(t, "N", state.Cells[0][2])
				assert.Equal(t, "D", state.Cells[0][3])
				assert.Equal(t, "", state.Cells[0][4])
			},
		},
		{
			name:   "rebus",
			puzzle: "xwordinfo-success-20181231.json",
			clue:   "1a",
			answer: "(RED)AND(BLUE)",
			verify: func(t *testing.T, state *State) {
				assert.Equal(t, "RED", state.Cells[0][0])
				assert.Equal(t, "A", state.Cells[0][1])
				assert.Equal(t, "N", state.Cells[0][2])
				assert.Equal(t, "D", state.Cells[0][3])
				assert.Equal(t, "BLUE", state.Cells[0][4])
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := newState(t, test.puzzle)
			err := s.ApplyAnswer(test.clue, test.answer, false)
			require.NoError(t, err)
			test.verify(t, s)
		})
	}
}

func TestState_ApplyAnswer_Cells_CorrectOnly(t *testing.T) {
	tests := []struct {
		name   string
		puzzle string
		clue   string
		answer string
		verify func(*testing.T, *State)
	}{
		{
			name:   "across answer",
			puzzle: "xwordinfo-success-20181231.json",
			clue:   "1a",
			answer: "Q AND A",
		},
		{
			name:   "down answer",
			puzzle: "xwordinfo-success-20181231.json",
			clue:   "1d",
			answer: "QTIP",
		},
		{
			name:   "unknown letters",
			puzzle: "xwordinfo-success-20181231.json",
			clue:   "1a",
			answer: ". AND .",
		},
		{
			name:   "can specify unknown letters",
			puzzle: "xwordinfo-success-20181231.json",
			clue:   "1a",
			answer: ". AND .",
			verify: func(t *testing.T, state *State) {
				assert.NoError(t, state.ApplyAnswer("1a", "Q AND .", true))
				assert.NoError(t, state.ApplyAnswer("1a", "Q AND A", true))
			},
		},
		{
			name:   "rebus",
			puzzle: "xwordinfo-rebus-20181227.json",
			clue:   "30a",
			answer: "AERIAL RE(CON)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := newState(t, test.puzzle)
			err := s.ApplyAnswer(test.clue, test.answer, true)
			require.NoError(t, err)

			if test.verify != nil {
				test.verify(t, s)
			}
		})
	}
}

func TestState_ApplyAnswer_Cells_CorrectOnly_Error(t *testing.T) {
	tests := []struct {
		name   string
		puzzle string
		clue   string
		answer string
		verify func(*testing.T, *State, error)
	}{
		{
			name:   "cannot specify incorrect cell",
			puzzle: "xwordinfo-success-20181231.json",
			clue:   "1a",
			answer: "R AND A",
		},
		{
			name:   "cannot change correct cell",
			puzzle: "xwordinfo-success-20181231.json",
			clue:   "1a",
			answer: "Q AND A",
			verify: func(t *testing.T, state *State, err error) {
				require.NoError(t, err)
				assert.Error(t, state.ApplyAnswer("1a", "R AND A", true))
			},
		},
		{
			name:   "cannot clear correct cell",
			puzzle: "xwordinfo-success-20181231.json",
			clue:   "1a",
			answer: "Q AND A",
			verify: func(t *testing.T, state *State, err error) {
				require.NoError(t, err)
				assert.Error(t, state.ApplyAnswer("1a", ". AND A", true))
			},
		},
		{
			name:   "cannot incorrectly specify missing cell",
			puzzle: "xwordinfo-success-20181231.json",
			clue:   "1a",
			answer: ". AND A",
			verify: func(t *testing.T, state *State, err error) {
				require.NoError(t, err)
				assert.Error(t, state.ApplyAnswer("1a", "R AND A", true))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := newState(t, test.puzzle)
			err := s.ApplyAnswer(test.clue, test.answer, true)

			if test.verify == nil {
				assert.Error(t, err)
			} else {
				test.verify(t, s, err)
			}
		})
	}
}

func TestState_ApplyAnswer_Filled(t *testing.T) {
	tests := []struct {
		name   string
		puzzle string
		clue   string
		answer string
		verify func(*testing.T, *State)
	}{
		{
			name:   "across answer",
			puzzle: "xwordinfo-success-20181231.json",
			clue:   "1a",
			answer: "Q AND A",
			verify: func(t *testing.T, state *State) {
				assert.True(t, state.AcrossCluesFilled[1])

				// Everything else should be unfilled
				for num, filled := range state.AcrossCluesFilled {
					if num != 1 {
						assert.False(t, filled)
					}
				}
				for _, filled := range state.DownCluesFilled {
					assert.False(t, filled)
				}
			},
		},
		{
			name:   "down answer",
			puzzle: "xwordinfo-success-20181231.json",
			clue:   "1d",
			answer: "QTIP",
			verify: func(t *testing.T, state *State) {
				assert.True(t, state.DownCluesFilled[1])

				// Everything else should be unfilled
				for _, filled := range state.AcrossCluesFilled {
					assert.False(t, filled)
				}
				for num, filled := range state.DownCluesFilled {
					if num != 1 {
						assert.False(t, filled)
					}
				}
			},
		},
		{
			name:   "across answer completes multiple down clues",
			puzzle: "xwordinfo-success-20181231.json",
			clue:   "1a",
			answer: "Q AND A",
			verify: func(t *testing.T, state *State) {
				require.NoError(t, state.ApplyAnswer("14a", "THIRD", false))
				require.NoError(t, state.ApplyAnswer("17a", "IM TOO OLD FOR THIS", false))
				require.NoError(t, state.ApplyAnswer("19a", "PERU", false))
				require.NoError(t, state.ApplyAnswer("22a", "DOG TAG", false))

				assert.True(t, state.DownCluesFilled[1])
				assert.True(t, state.DownCluesFilled[2])
				assert.True(t, state.DownCluesFilled[3])
				assert.False(t, state.DownCluesFilled[4])
				assert.True(t, state.DownCluesFilled[5])
				assert.False(t, state.DownCluesFilled[6])
				assert.False(t, state.DownCluesFilled[7])
				assert.False(t, state.DownCluesFilled[8])
				assert.False(t, state.DownCluesFilled[9])
				assert.False(t, state.DownCluesFilled[10])
				assert.False(t, state.DownCluesFilled[11])
				assert.False(t, state.DownCluesFilled[12])
				assert.False(t, state.DownCluesFilled[13])
			},
		},
		{
			name:   "down answer completes multiple across clues",
			puzzle: "xwordinfo-success-20181231.json",
			clue:   "1d",
			answer: "QTIP",
			verify: func(t *testing.T, state *State) {
				require.NoError(t, state.ApplyAnswer("2d", "AHMED", false))
				require.NoError(t, state.ApplyAnswer("3d", "NITRO", false))
				require.NoError(t, state.ApplyAnswer("4d", "DROUGHT", false))
				require.NoError(t, state.ApplyAnswer("5d", "ADO", false))

				assert.True(t, state.AcrossCluesFilled[1])
				assert.True(t, state.AcrossCluesFilled[14])
				assert.False(t, state.AcrossCluesFilled[17])
				assert.True(t, state.AcrossCluesFilled[19])
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := newState(t, test.puzzle)
			err := s.ApplyAnswer(test.clue, test.answer, false)
			require.NoError(t, err)
			test.verify(t, s)
		})
	}
}

func TestState_ApplyAnswer_Status(t *testing.T) {
	tests := []struct {
		name   string
		puzzle string
		clue   string
		answer string
		verify func(*testing.T, *State)
	}{
		{
			name:   "single answer",
			puzzle: "xwordinfo-success-20181231.json",
			clue:   "1a",
			answer: "Q AND A",
			verify: func(t *testing.T, state *State) {
				assert.Equal(t, StatusSolving, state.Status)
			},
		},
		{
			name:   "complete and correct puzzle",
			puzzle: "xwordinfo-success-20181231.json",
			clue:   "1a",
			answer: "Q AND A",
			verify: func(t *testing.T, state *State) {
				require.NoError(t, state.ApplyAnswer("1a", "Q AND A", false))
				require.NoError(t, state.ApplyAnswer("6a", "ATTIC", false))
				require.NoError(t, state.ApplyAnswer("11a", "HON", false))
				require.NoError(t, state.ApplyAnswer("14a", "THIRD", false))
				require.NoError(t, state.ApplyAnswer("15a", "LAID ASIDE", false))
				require.NoError(t, state.ApplyAnswer("17a", "IM TOO OLD FOR THIS", false))
				require.NoError(t, state.ApplyAnswer("19a", "PERU", false))
				require.NoError(t, state.ApplyAnswer("20a", "LEAF", false))
				require.NoError(t, state.ApplyAnswer("21a", "PEONS", false))
				require.NoError(t, state.ApplyAnswer("22a", "DOG TAG", false))
				require.NoError(t, state.ApplyAnswer("24a", "LOL", false))
				require.NoError(t, state.ApplyAnswer("25a", "HAVE NO OOMPH", false))
				require.NoError(t, state.ApplyAnswer("30a", "MATTE", false))
				require.NoError(t, state.ApplyAnswer("33a", "IMPLORED", false))
				require.NoError(t, state.ApplyAnswer("35a", "ERR", false))
				require.NoError(t, state.ApplyAnswer("36a", "RANGE", false))
				require.NoError(t, state.ApplyAnswer("38a", "EMO", false))
				require.NoError(t, state.ApplyAnswer("39a", "WAIT HERE", false))
				require.NoError(t, state.ApplyAnswer("42a", "EGYPT", false))
				require.NoError(t, state.ApplyAnswer("44a", "BOO OFF STAGE", false))
				require.NoError(t, state.ApplyAnswer("47a", "ERS", false))
				require.NoError(t, state.ApplyAnswer("48a", "EUGENE", false))
				require.NoError(t, state.ApplyAnswer("51a", "SHARI", false))
				require.NoError(t, state.ApplyAnswer("54a", "SINN", false))
				require.NoError(t, state.ApplyAnswer("56a", "WING", false))
				require.NoError(t, state.ApplyAnswer("58a", "ITS A ZOO OUT THERE", false))
				require.NoError(t, state.ApplyAnswer("61a", "STEGOSAUR", false))
				require.NoError(t, state.ApplyAnswer("62a", "HIT ON", false))
				require.NoError(t, state.ApplyAnswer("63a", "IPA", false))
				require.NoError(t, state.ApplyAnswer("64a", "NURSE", false))
				require.NoError(t, state.ApplyAnswer("65a", "OZONE", false))

				assert.Equal(t, StatusComplete, state.Status)
			},
		},
		{
			name:   "complete and incorrect puzzle",
			puzzle: "xwordinfo-success-20181231.json",
			clue:   "1a",
			answer: "Q AND A",
			verify: func(t *testing.T, state *State) {
				require.NoError(t, state.ApplyAnswer("1a", "XXXXX", false))
				require.NoError(t, state.ApplyAnswer("6a", "XXXXX", false))
				require.NoError(t, state.ApplyAnswer("11a", "XXX", false))
				require.NoError(t, state.ApplyAnswer("14a", "XXXXX", false))
				require.NoError(t, state.ApplyAnswer("15a", "XXXXXXXXX", false))
				require.NoError(t, state.ApplyAnswer("17a", "XXXXXXXXXXXXXXX", false))
				require.NoError(t, state.ApplyAnswer("19a", "XXXX", false))
				require.NoError(t, state.ApplyAnswer("20a", "XXXX", false))
				require.NoError(t, state.ApplyAnswer("21a", "XXXXX", false))
				require.NoError(t, state.ApplyAnswer("22a", "XXXXXX", false))
				require.NoError(t, state.ApplyAnswer("24a", "XXX", false))
				require.NoError(t, state.ApplyAnswer("25a", "XXXXXXXXXXX", false))
				require.NoError(t, state.ApplyAnswer("30a", "XXXXX", false))
				require.NoError(t, state.ApplyAnswer("33a", "XXXXXXXX", false))
				require.NoError(t, state.ApplyAnswer("35a", "XXX", false))
				require.NoError(t, state.ApplyAnswer("36a", "XXXXX", false))
				require.NoError(t, state.ApplyAnswer("38a", "XXX", false))
				require.NoError(t, state.ApplyAnswer("39a", "XXXXXXXX", false))
				require.NoError(t, state.ApplyAnswer("42a", "XXXXX", false))
				require.NoError(t, state.ApplyAnswer("44a", "XXXXXXXXXXX", false))
				require.NoError(t, state.ApplyAnswer("47a", "XXX", false))
				require.NoError(t, state.ApplyAnswer("48a", "XXXXXX", false))
				require.NoError(t, state.ApplyAnswer("51a", "XXXXX", false))
				require.NoError(t, state.ApplyAnswer("54a", "XXXX", false))
				require.NoError(t, state.ApplyAnswer("56a", "XXXX", false))
				require.NoError(t, state.ApplyAnswer("58a", "XXXXXXXXXXXXXXX", false))
				require.NoError(t, state.ApplyAnswer("61a", "XXXXXXXXX", false))
				require.NoError(t, state.ApplyAnswer("62a", "XXXXX", false))
				require.NoError(t, state.ApplyAnswer("63a", "XXX", false))
				require.NoError(t, state.ApplyAnswer("64a", "XXXXX", false))
				require.NoError(t, state.ApplyAnswer("65a", "XXXXX", false))

				assert.Equal(t, StatusSolving, state.Status)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := newState(t, test.puzzle)
			err := s.ApplyAnswer(test.clue, test.answer, false)
			require.NoError(t, err)
			test.verify(t, s)
		})
	}
}

func TestState_ApplyAnswer_Error(t *testing.T) {
	tests := []struct {
		name   string
		puzzle string
		clue   string
		answer string
	}{
		{
			name:   "bad clue",
			puzzle: "xwordinfo-success-20181231.json",
			clue:   "xyz",
			answer: "ABC",
		},
		{
			name:   "invalid clue",
			puzzle: "xwordinfo-success-20181231.json",
			clue:   "199a",
			answer: "ABC",
		},
		{
			name:   "bad answer",
			puzzle: "xwordinfo-success-20181231.json",
			clue:   "1a",
			answer: ")Q AND A",
		},
		{
			name:   "answer too short",
			puzzle: "xwordinfo-success-20181231.json",
			clue:   "1a",
			answer: "ABC",
		},
		{
			name:   "answer too long",
			puzzle: "xwordinfo-success-20181231.json",
			clue:   "1a",
			answer: "ABCDEFGHIJKLMNOPQRSTUVWXYZ",
		},
		{
			name:   "answer too short (rebus)",
			puzzle: "xwordinfo-success-20181231.json",
			clue:   "1a",
			answer: "(Q AND A)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := newState(t, test.puzzle)
			err := s.ApplyAnswer(test.clue, test.answer, false)
			assert.Error(t, err)
		})
	}
}

func TestState_ClearIncorrectCells(t *testing.T) {
	tests := []struct {
		name    string
		puzzle  string
		answers map[string]string
		verify  func(*testing.T, *State)
	}{
		{
			name:   "correct across answer",
			puzzle: "xwordinfo-success-20181231.json",
			answers: map[string]string{
				"1a": "QANDA",
			},
			verify: func(t *testing.T, state *State) {
				assert.Equal(t, "Q", state.Cells[0][0])
				assert.Equal(t, "A", state.Cells[0][1])
				assert.Equal(t, "N", state.Cells[0][2])
				assert.Equal(t, "D", state.Cells[0][3])
				assert.Equal(t, "A", state.Cells[0][4])
				assert.True(t, state.AcrossCluesFilled[1])
			},
		},
		{
			name:   "correct down answer",
			puzzle: "xwordinfo-success-20181231.json",
			answers: map[string]string{
				"1d": "QTIP",
			},
			verify: func(t *testing.T, state *State) {
				assert.Equal(t, "Q", state.Cells[0][0])
				assert.Equal(t, "T", state.Cells[1][0])
				assert.Equal(t, "I", state.Cells[2][0])
				assert.Equal(t, "P", state.Cells[3][0])
				assert.True(t, state.DownCluesFilled[1])
			},
		},
		{
			name:   "partially incorrect across answer",
			puzzle: "xwordinfo-success-20181231.json",
			answers: map[string]string{
				"1a": "QNORA",
			},
			verify: func(t *testing.T, state *State) {
				assert.Equal(t, "Q", state.Cells[0][0])
				assert.Equal(t, "", state.Cells[0][1])
				assert.Equal(t, "", state.Cells[0][2])
				assert.Equal(t, "", state.Cells[0][3])
				assert.Equal(t, "A", state.Cells[0][4])
				assert.False(t, state.AcrossCluesFilled[1])
			},
		},
		{
			name:   "partially incorrect down answer",
			puzzle: "xwordinfo-success-20181231.json",
			answers: map[string]string{
				"1d": "QTOP",
			},
			verify: func(t *testing.T, state *State) {
				assert.Equal(t, "Q", state.Cells[0][0])
				assert.Equal(t, "T", state.Cells[1][0])
				assert.Equal(t, "", state.Cells[2][0])
				assert.Equal(t, "P", state.Cells[3][0])
				assert.False(t, state.DownCluesFilled[1])
			},
		},
		{
			name:   "completely incorrect across answer",
			puzzle: "xwordinfo-success-20181231.json",
			answers: map[string]string{
				"1a": "XXXXX",
			},
			verify: func(t *testing.T, state *State) {
				assert.Equal(t, "", state.Cells[0][0])
				assert.Equal(t, "", state.Cells[0][1])
				assert.Equal(t, "", state.Cells[0][2])
				assert.Equal(t, "", state.Cells[0][3])
				assert.Equal(t, "", state.Cells[0][4])
				assert.False(t, state.AcrossCluesFilled[1])
			},
		},
		{
			name:   "completely incorrect down answer",
			puzzle: "xwordinfo-success-20181231.json",
			answers: map[string]string{
				"1d": "XXXX",
			},
			verify: func(t *testing.T, state *State) {
				assert.Equal(t, "", state.Cells[0][0])
				assert.Equal(t, "", state.Cells[1][0])
				assert.Equal(t, "", state.Cells[2][0])
				assert.Equal(t, "", state.Cells[3][0])
				assert.False(t, state.DownCluesFilled[1])
			},
		},
		{
			name:   "incorrect across answer clears completed down clue",
			puzzle: "xwordinfo-success-20181231.json",
			answers: map[string]string{
				"1a": "XXXXX",
				"1d": "XTIP",
			},
			verify: func(t *testing.T, state *State) {
				assert.False(t, state.DownCluesFilled[1])
			},
		},
		{
			name:   "incorrect down answer clears completed across clue",
			puzzle: "xwordinfo-success-20181231.json",
			answers: map[string]string{
				"1a": "XANDA",
				"1d": "XXXX",
			},
			verify: func(t *testing.T, state *State) {
				assert.False(t, state.AcrossCluesFilled[1])
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := newState(t, test.puzzle)

			for clue, answer := range test.answers {
				err := s.ApplyAnswer(clue, answer, false)
				require.NoError(t, err)
			}

			err := s.ClearIncorrectCells()
			assert.NoError(t, err)

			test.verify(t, s)
		})
	}
}

func TestParseClue(t *testing.T) {
	tests := []struct {
		clue        string
		expectedNum int
		expectedDir string
	}{
		{
			clue:        "1a",
			expectedNum: 1,
			expectedDir: "a",
		},
		{
			clue:        "10a",
			expectedNum: 10,
			expectedDir: "a",
		},
		{
			clue:        "100a",
			expectedNum: 100,
			expectedDir: "a",
		},
		{
			clue:        "1d",
			expectedNum: 1,
			expectedDir: "d",
		},
		{
			clue:        "10d",
			expectedNum: 10,
			expectedDir: "d",
		},
		{
			clue:        "100d",
			expectedNum: 100,
			expectedDir: "d",
		},
	}

	for _, test := range tests {
		t.Run(test.clue, func(t *testing.T) {
			num, dir, err := ParseClue(test.clue)
			require.NoError(t, err)
			assert.Equal(t, test.expectedNum, num)
			assert.Equal(t, test.expectedDir, dir)
		})
	}
}

func TestParseClue_Error(t *testing.T) {
	tests := []string{
		"",
		"1x",
		"1ad",
		"1a2",
		"1",
		"a",
	}

	for _, clue := range tests {
		t.Run(clue, func(t *testing.T) {
			_, _, err := ParseClue(clue)
			assert.Error(t, err)
		})
	}
}

func TestParseAnswer(t *testing.T) {
	tests := []struct {
		answer   string
		expected []string
	}{
		{
			answer:   "ABCDE",
			expected: []string{"A", "B", "C", "D", "E"},
		},
		{
			answer:   "abcde",
			expected: []string{"A", "B", "C", "D", "E"},
		},
		{
			answer:   " ABCDE",
			expected: []string{"A", "B", "C", "D", "E"},
		},
		{
			answer:   "ABCDE ",
			expected: []string{"A", "B", "C", "D", "E"},
		},
		{
			answer:   "ABC DE",
			expected: []string{"A", "B", "C", "D", "E"},
		},
		{
			answer:   "....S",
			expected: []string{"", "", "", "", "S"},
		},
		{
			answer:   "(RED) VELVET CAKE",
			expected: []string{"RED", "V", "E", "L", "V", "E", "T", "C", "A", "K", "E"},
		},
	}

	for _, test := range tests {
		t.Run(test.answer, func(t *testing.T) {
			actual, err := ParseAnswer(test.answer)
			require.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestParseAnswer_Error(t *testing.T) {
	tests := []string{
		"",
		"((red) velvet cake)",
		"(red velvet cake",
		"((red) velvet cake",
		"red velvet cake)",
		")red velvet cake",
	}

	for _, answer := range tests {
		t.Run(answer, func(t *testing.T) {
			_, err := ParseAnswer(answer)
			assert.Error(t, err)
		})
	}
}

func TestStatus_String(t *testing.T) {
	tests := []struct {
		name     string
		state    Status
		expected string
	}{
		{
			name:     "created",
			state:    StatusCreated,
			expected: "created",
		},
		{
			name:     "paused",
			state:    StatusPaused,
			expected: "paused",
		},
		{
			name:     "solving",
			state:    StatusSolving,
			expected: "solving",
		},
		{
			name:     "complete",
			state:    StatusComplete,
			expected: "complete",
		},
		{
			name:     "invalid",
			state:    Status(17),
			expected: "unknown",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := test.state.String()
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestStatus_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		state    Status
		expected []byte
	}{
		{
			name:     "created",
			state:    StatusCreated,
			expected: []byte(`"created"`),
		},
		{
			name:     "paused",
			state:    StatusPaused,
			expected: []byte(`"paused"`),
		},
		{
			name:     "solving",
			state:    StatusSolving,
			expected: []byte(`"solving"`),
		},
		{
			name:     "complete",
			state:    StatusComplete,
			expected: []byte(`"complete"`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bs, err := json.Marshal(test.state)
			require.NoError(t, err)
			assert.Equal(t, test.expected, bs)
		})
	}
}

func TestStatus_MarshalJSON_Error(t *testing.T) {
	tests := []struct {
		name  string
		state Status
	}{
		{
			name:  "invalid",
			state: Status(17),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := json.Marshal(test.state)
			assert.Error(t, err)
		})
	}
}

func TestStatus_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		bs       []byte
		expected Status
	}{
		{
			name:     "created",
			bs:       []byte(`"created"`),
			expected: StatusCreated,
		},
		{
			name:     "paused",
			bs:       []byte(`"paused"`),
			expected: StatusPaused,
		},
		{
			name:     "solving",
			bs:       []byte(`"solving"`),
			expected: StatusSolving,
		},
		{
			name:     "complete",
			bs:       []byte(`"complete"`),
			expected: StatusComplete,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var actual Status

			err := json.Unmarshal(test.bs, &actual)
			require.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestStatus_UnmarshalJSON_Error(t *testing.T) {
	tests := []struct {
		name string
		bs   []byte
	}{
		{
			name: "wrong type",
			bs:   []byte(`true`),
		},
		{
			name: "empty value",
			bs:   []byte(`""`),
		},
		{
			name: "invalid value",
			bs:   []byte(`"asdf"`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var actual Status

			err := json.Unmarshal(test.bs, &actual)
			assert.Error(t, err)
		})
	}
}

func TestDuration_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		duration Duration
		expected []byte
	}{
		{
			name:     "empty",
			duration: Duration{},
			expected: []byte(`"0s"`),
		},
		{
			name:     "1 second",
			duration: Duration{time.Second},
			expected: []byte(`"1s"`),
		},
		{
			name:     "1 minute",
			duration: Duration{time.Minute},
			expected: []byte(`"1m0s"`),
		},
		{
			name:     "1 hour",
			duration: Duration{time.Hour},
			expected: []byte(`"1h0m0s"`),
		},
		{
			name:     "24 hours",
			duration: Duration{24 * time.Hour},
			expected: []byte(`"24h0m0s"`),
		},
		{
			name:     "2 hours 12 minutes 9 seconds",
			duration: Duration{2*time.Hour + 12*time.Minute + 9*time.Second},
			expected: []byte(`"2h12m9s"`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := json.Marshal(test.duration)
			require.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestDuration_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		bs       []byte
		expected Duration
	}{
		{
			name:     "empty",
			bs:       []byte(`"0s"`),
			expected: Duration{},
		},
		{
			name:     "1 second",
			bs:       []byte(`"1s"`),
			expected: Duration{time.Second},
		},
		{
			name:     "1 minute",
			bs:       []byte(`"1m0s"`),
			expected: Duration{time.Minute},
		},
		{
			name:     "1 hour",
			bs:       []byte(`"1h0m0s"`),
			expected: Duration{time.Hour},
		},
		{
			name:     "24 hours",
			bs:       []byte(`"24h0m0s"`),
			expected: Duration{24 * time.Hour},
		},
		{
			name:     "2 hours 12 minutes 9 seconds",
			bs:       []byte(`"2h12m9s"`),
			expected: Duration{2*time.Hour + 12*time.Minute + 9*time.Second},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var actual Duration

			err := json.Unmarshal(test.bs, &actual)
			require.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestDuration_UnmarshalJSON_Error(t *testing.T) {
	tests := []struct {
		name string
		bs   []byte
	}{
		{
			name: "invalid type",
			bs:   []byte(`true`),
		},
		{
			name: "empty value",
			bs:   []byte(`""`),
		},
		{
			name: "incorrect value",
			bs:   []byte(`"1x2y"`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var actual Duration

			err := json.Unmarshal(test.bs, &actual)
			assert.Error(t, err)
		})
	}
}

func newState(t *testing.T, filename string) *State {
	var s State
	s.Puzzle = LoadTestPuzzle(t, filename)

	for y := 0; y < s.Puzzle.Rows; y++ {
		s.Cells = append(s.Cells, make([]string, s.Puzzle.Cols))
	}

	s.AcrossCluesFilled = make(map[int]bool)
	s.DownCluesFilled = make(map[int]bool)
	s.Status = StatusSolving

	return &s
}

// TODO: Come back to these test cases and refactor them to make the verify
// method more clear.  Right now it implies that all it's doing it verifying
// the expected result of the test, but in reality a lot of them actually
// perform the meaningful parts of the test.  This is misleading and not very
// intuitive.
