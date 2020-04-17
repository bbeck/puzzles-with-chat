package spellingbee

import (
	"github.com/bbeck/twitch-plays-crosswords/api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestState_ApplyAnswer_Words(t *testing.T) {
	tests := []struct {
		name            string
		puzzle          *Puzzle
		initialWords    []string
		answer          string
		allowUnofficial bool
		expectedWords   []string
	}{
		{
			name:          "answer from official list",
			puzzle:        LoadTestPuzzle(t, "nytbee-20200408.html"),
			answer:        "COCONUT",
			expectedWords: []string{"COCONUT"},
		},
		{
			name:            "answer from unofficial list",
			puzzle:          LoadTestPuzzle(t, "nytbee-20200408.html"),
			answer:          "CONCOCTOR",
			allowUnofficial: true,
			expectedWords:   []string{"CONCOCTOR"},
		},
		{
			name:          "lowercase answer",
			puzzle:        LoadTestPuzzle(t, "nytbee-20200408.html"),
			answer:        "coconut",
			expectedWords: []string{"COCONUT"},
		},
		{
			name:          "words stay sorted",
			puzzle:        LoadTestPuzzle(t, "nytbee-20200408.html"),
			initialWords:  []string{"COUNTY", "CROUTON"},
			answer:        "COURT",
			expectedWords: []string{"COUNTY", "COURT", "CROUTON"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := newState(test.puzzle)
			for _, word := range test.initialWords {
				state.Words = append(state.Words, word)
			}

			err := state.ApplyAnswer(test.answer, test.allowUnofficial)
			require.NoError(t, err)
			assert.Equal(t, test.expectedWords, state.Words)
		})
	}
}

func TestState_ApplyAnswer_Status(t *testing.T) {
	tests := []struct {
		name            string
		puzzle          *Puzzle
		answers         []string
		allowUnofficial bool
		expectedStatus  model.Status
	}{
		{
			name:           "single answer",
			puzzle:         LoadTestPuzzle(t, "nytbee-20200408.html"),
			answers:        []string{"COCONUT"},
			expectedStatus: model.StatusSolving,
		},
		{
			name:   "all official answers (unofficial not allowed)",
			puzzle: LoadTestPuzzle(t, "nytbee-20200408.html"),
			answers: []string{
				"COCONUT",
				"CONCOCT",
				"CONTORT",
				"CONTOUR",
				"COOT",
				"COTTON",
				"COTTONY",
				"COUNT",
				"COUNTRY",
				"COUNTY",
				"COURT",
				"CROUTON",
				"CURT",
				"CUTOUT",
				"NUTTY",
				"ONTO",
				"OUTCRY",
				"OUTRO",
				"OUTRUN",
				"ROOT",
				"ROTO",
				"ROTOR",
				"ROUT",
				"RUNOUT",
				"RUNT",
				"RUNTY",
				"RUTTY",
				"TONY",
				"TOON",
				"TOOT",
				"TORN",
				"TORO",
				"TORT",
				"TOUR",
				"TOUT",
				"TROT",
				"TROUT",
				"TROY",
				"TRYOUT",
				"TURN",
				"TURNOUT",
				"TUTOR",
				"TUTU",
				"TYCOON",
				"TYRO",
				"UNCUT",
				"UNTO",
				"YURT",
			},
			allowUnofficial: false,
			expectedStatus:  model.StatusComplete,
		},
		{
			name:   "all official answers (unofficial allowed)",
			puzzle: LoadTestPuzzle(t, "nytbee-20200408.html"),
			answers: []string{
				"COCONUT",
				"CONCOCT",
				"CONTORT",
				"CONTOUR",
				"COOT",
				"COTTON",
				"COTTONY",
				"COUNT",
				"COUNTRY",
				"COUNTY",
				"COURT",
				"CROUTON",
				"CURT",
				"CUTOUT",
				"NUTTY",
				"ONTO",
				"OUTCRY",
				"OUTRO",
				"OUTRUN",
				"ROOT",
				"ROTO",
				"ROTOR",
				"ROUT",
				"RUNOUT",
				"RUNT",
				"RUNTY",
				"RUTTY",
				"TONY",
				"TOON",
				"TOOT",
				"TORN",
				"TORO",
				"TORT",
				"TOUR",
				"TOUT",
				"TROT",
				"TROUT",
				"TROY",
				"TRYOUT",
				"TURN",
				"TURNOUT",
				"TUTOR",
				"TUTU",
				"TYCOON",
				"TYRO",
				"UNCUT",
				"UNTO",
				"YURT",
			},
			allowUnofficial: true,
			expectedStatus:  model.StatusSolving,
		},
		{
			name:   "all answers (unofficial allowed)",
			puzzle: LoadTestPuzzle(t, "nytbee-20200408.html"),
			answers: []string{
				"COCONUT",
				"CONCOCT",
				"CONTORT",
				"CONTOUR",
				"COOT",
				"COTTON",
				"COTTONY",
				"COUNT",
				"COUNTRY",
				"COUNTY",
				"COURT",
				"CROUTON",
				"CURT",
				"CUTOUT",
				"NUTTY",
				"ONTO",
				"OUTCRY",
				"OUTRO",
				"OUTRUN",
				"ROOT",
				"ROTO",
				"ROTOR",
				"ROUT",
				"RUNOUT",
				"RUNT",
				"RUNTY",
				"RUTTY",
				"TONY",
				"TOON",
				"TOOT",
				"TORN",
				"TORO",
				"TORT",
				"TOUR",
				"TOUT",
				"TROT",
				"TROUT",
				"TROY",
				"TRYOUT",
				"TURN",
				"TURNOUT",
				"TUTOR",
				"TUTU",
				"TYCOON",
				"TYRO",
				"UNCUT",
				"UNTO",
				"YURT",
				"CONCOCTOR",
				"CONTO",
				"CORNUTO",
				"CROTON",
				"CRYOTRON",
				"CUNT",
				"CUTTY",
				"CYTON",
				"NOCTURN",
				"NONCOUNT",
				"NONCOUNTRY",
				"NONCOUNTY",
				"NOTTURNO",
				"OCTOROON",
				"OTTO",
				"OUTCOUNT",
				"OUTROOT",
				"OUTTROT",
				"OUTTURN",
				"ROOTY",
				"RYOT",
				"TOCO",
				"TORC",
				"TOROT",
				"TORR",
				"TORY",
				"TOTTY",
				"TOUTON",
				"TOYO",
				"TOYON",
				"TROU",
				"TROUTY",
				"TUNNY",
				"TURNON",
				"TURR",
				"TUTTY",
				"UNROOT",
				"UNTORN",
			},
			allowUnofficial: true,
			expectedStatus:  model.StatusComplete,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := newState(test.puzzle)
			for _, answer := range test.answers {
				require.NoError(t, state.ApplyAnswer(answer, test.allowUnofficial))
			}

			assert.Equal(t, test.expectedStatus, state.Status)
		})
	}
}

func TestState_ApplyAnswer_Error(t *testing.T) {
	tests := []struct {
		name            string
		puzzle          *Puzzle
		initialWords    []string
		answer          string
		allowUnofficial bool
	}{
		{
			name:   "not allowed letter",
			puzzle: LoadTestPuzzle(t, "nytbee-20200408.html"),
			answer: "WXYZ",
		},
		{
			name:         "already given answer",
			puzzle:       LoadTestPuzzle(t, "nytbee-20200408.html"),
			initialWords: []string{"COCONUT"},
			answer:       "COCONUT",
		},
		{
			name:   "answer not in official list",
			puzzle: LoadTestPuzzle(t, "nytbee-20200408.html"),
			answer: "CONCOCTOR",
		},
		{
			name:   "answer from unofficial list, not allowed",
			puzzle: LoadTestPuzzle(t, "nytbee-20200408.html"),
			answer: "CONCOCTOR",
		},
		{
			name:            "answer not in either list",
			puzzle:          LoadTestPuzzle(t, "nytbee-20200408.html"),
			answer:          "CCCC",
			allowUnofficial: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := newState(test.puzzle)
			for _, word := range test.initialWords {
				state.Words = append(state.Words, word)
			}

			err := state.ApplyAnswer(test.answer, test.allowUnofficial)
			assert.Error(t, err)
		})
	}
}

func newState(puzzle *Puzzle) *State {
	return &State{
		Status: model.StatusSolving,
		Puzzle: puzzle,
		Words:  make([]string, 0),
	}
}
