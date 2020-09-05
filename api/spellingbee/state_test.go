package spellingbee

import (
	"errors"
	"fmt"
	"github.com/bbeck/puzzles-with-chat/api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestState_ApplyAnswer_Words(t *testing.T) {
	tests := []struct {
		name            string
		filename        string
		initialWords    map[string]int
		answer          string
		allowUnofficial bool
		expectedWords   map[string]int
	}{
		{
			name:          "answer from official list",
			filename:      "nytbee-20200408.html",
			initialWords:  make(map[string]int),
			answer:        "COCONUT",
			expectedWords: map[string]int{"COCONUT": 0},
		},
		{
			name:            "answer from unofficial list",
			filename:        "nytbee-20200408.html",
			initialWords:    make(map[string]int),
			answer:          "CONCOCTOR",
			allowUnofficial: true,
			expectedWords:   map[string]int{"CONCOCTOR": 2},
		},
		{
			name:          "lowercase answer",
			filename:      "nytbee-20200408.html",
			initialWords:  make(map[string]int),
			answer:        "coconut",
			expectedWords: map[string]int{"COCONUT": 0},
		},
		{
			name:     "existing indices preserved",
			filename: "nytbee-20200408.html",
			initialWords: map[string]int{
				"COUNTY":  9,
				"CROUTON": 11,
			},
			answer: "COURT",
			expectedWords: map[string]int{
				"COUNTY":  9,
				"COURT":   10,
				"CROUTON": 11,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := NewState(t, test.filename)
			state.Words = test.initialWords

			err := state.ApplyAnswer(test.answer, test.allowUnofficial)
			require.NoError(t, err)
			assert.Equal(t, test.expectedWords, state.Words)
		})
	}
}

func TestState_ApplyAnswer_Status(t *testing.T) {
	tests := []struct {
		name            string
		filename        string
		answers         []string
		allowUnofficial bool
		expectedStatus  model.Status
	}{
		{
			name:           "single answer",
			filename:       "nytbee-20200408.html",
			answers:        []string{"COCONUT"},
			expectedStatus: model.StatusSolving,
		},
		{
			name:     "all official answers (unofficial not allowed)",
			filename: "nytbee-20200408.html",
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
			name:     "all official answers (unofficial allowed)",
			filename: "nytbee-20200408.html",
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
			name:     "all answers (unofficial allowed)",
			filename: "nytbee-20200408.html",
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
			state := NewState(t, test.filename)
			state.Status = model.StatusSolving

			for _, answer := range test.answers {
				require.NoError(t, state.ApplyAnswer(answer, test.allowUnofficial))
			}

			assert.Equal(t, test.expectedStatus, state.Status)
		})
	}
}

func TestState_ApplyAnswer_Score(t *testing.T) {
	tests := []struct {
		name            string
		filename        string
		answers         []string
		allowUnofficial bool
		expectedScore   int
	}{
		{
			name:          "four letter answer",
			filename:      "nytbee-20200408.html",
			answers:       []string{"COOT"},
			expectedScore: 1,
		},
		{
			name:          "long answer",
			filename:      "nytbee-20200408.html",
			answers:       []string{"COCONUT"},
			expectedScore: 7,
		},
		{
			name:          "pangram",
			filename:      "nytbee-20200408.html",
			answers:       []string{"COUNTRY"},
			expectedScore: 14,
		},
		{
			name:     "all official answers",
			filename: "nytbee-20200408.html",
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
			expectedScore:   183,
		},
		{
			name:     "all answers (unofficial allowed)",
			filename: "nytbee-20200408.html",
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
			expectedScore:   384,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := NewState(t, test.filename)
			state.Status = model.StatusSolving

			for _, answer := range test.answers {
				require.NoError(t, state.ApplyAnswer(answer, test.allowUnofficial))
			}

			assert.Equal(t, test.expectedScore, state.Score)
		})
	}
}

func TestState_ApplyAnswer_Error(t *testing.T) {
	tests := []struct {
		name            string
		filename        string
		initialWords    map[string]int
		answer          string
		allowUnofficial bool
	}{
		{
			name:     "not allowed letter",
			filename: "nytbee-20200408.html",
			answer:   "WXYZ",
		},
		{
			name:         "already given answer",
			filename:     "nytbee-20200408.html",
			initialWords: map[string]int{"COCONUT": 0},
			answer:       "COCONUT",
		},
		{
			name:     "answer not in official list",
			filename: "nytbee-20200408.html",
			answer:   "CONCOCTOR",
		},
		{
			name:     "answer from unofficial list, not allowed",
			filename: "nytbee-20200408.html",
			answer:   "CONCOCTOR",
		},
		{
			name:            "answer not in either list",
			filename:        "nytbee-20200408.html",
			answer:          "CCCC",
			allowUnofficial: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := NewState(t, test.filename)
			state.Words = test.initialWords

			err := state.ApplyAnswer(test.answer, test.allowUnofficial)
			assert.Error(t, err)
		})
	}
}

func TestState_RebuildWordMap(t *testing.T) {
	tests := []struct {
		name            string
		filename        string
		allowUnofficial bool
		answers         []string       // The answers already given
		expected        map[string]int // The expected answers after rebuilding
	}{
		{
			name:     "no answers",
			filename: "nytbee-20200408.html",
			expected: map[string]int{},
		},
		{
			name:     "no unofficial answers",
			filename: "nytbee-20200408.html",
			answers:  []string{"COCONUT", "CONCOCT"},
			expected: map[string]int{
				"COCONUT": 0,
				"CONCOCT": 1,
			},
		},
		{
			name:            "one unofficial answer, no unofficial allowed",
			filename:        "nytbee-20200408.html",
			allowUnofficial: false,
			answers:         []string{"CONCOCTOR"},
			expected:        map[string]int{},
		},
		{
			name:            "multiple unofficial answers, no unofficial allowed",
			filename:        "nytbee-20200408.html",
			allowUnofficial: false,
			answers:         []string{"CONCOCTOR", "CONTO"},
			expected:        map[string]int{},
		},
		{
			name:            "mixed answers, no unofficial allowed",
			filename:        "nytbee-20200408.html",
			allowUnofficial: false,
			answers:         []string{"COCONUT", "CONCOCT", "CONCOCTOR", "CONTO"},
			expected: map[string]int{
				"COCONUT": 0,
				"CONCOCT": 1,
			},
		},
		{
			name:            "one unofficial answer, unofficial allowed",
			filename:        "nytbee-20200408.html",
			allowUnofficial: true,
			answers:         []string{"CONCOCTOR"},
			expected: map[string]int{
				"CONCOCTOR": 2,
			},
		},
		{
			name:            "multiple unofficial answers, no unofficial allowed",
			filename:        "nytbee-20200408.html",
			allowUnofficial: true,
			answers:         []string{"CONCOCTOR", "CONTO"},
			expected: map[string]int{
				"CONCOCTOR": 2,
				"CONTO":     3,
			},
		},
		{
			name:            "mixed answers, unofficial allowed",
			filename:        "nytbee-20200408.html",
			allowUnofficial: true,
			answers:         []string{"COCONUT", "CONCOCT", "CONCOCTOR", "CONTO"},
			expected: map[string]int{
				"COCONUT":   0,
				"CONCOCT":   1,
				"CONCOCTOR": 2,
				"CONTO":     3,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := NewState(t, test.filename)
			for i, word := range test.answers {
				state.Words[word] = i
			}

			state.RebuildWordMap(test.allowUnofficial)
			assert.Equal(t, test.expected, state.Words)
		})
	}
}

func TestState_RebuildWordMap_Score(t *testing.T) {
	tests := []struct {
		name            string
		filename        string
		allowUnofficial bool
		answers         []string // The answers already given
		expectedScore   int      // The expected score after rebuilding
	}{
		{
			name:          "no answers",
			filename:      "nytbee-20200408.html",
			expectedScore: 0,
		},
		{
			name:     "no unofficial answers",
			filename: "nytbee-20200408.html",
			answers: []string{
				"COCONUT",
				"CONCOCT",
			},
			expectedScore: 14,
		},
		{
			name:     "one unofficial answer",
			filename: "nytbee-20200408.html",
			answers: []string{
				"CONCOCTOR",
			},
			expectedScore: 0,
		},
		{
			name:     "multiple unofficial answers",
			filename: "nytbee-20200408.html",
			answers: []string{
				"CONCOCTOR",
				"CONTO",
			},
			expectedScore: 0,
		},
		{
			name:     "mixed unofficial answers, no unofficial allowed",
			filename: "nytbee-20200408.html",
			answers: []string{
				"COCONUT",
				"CONCOCT",
				"CONCOCTOR",
				"CONTO",
			},
			expectedScore: 14,
		},
		{
			name:            "mixed unofficial answers, unofficial allowed",
			filename:        "nytbee-20200408.html",
			allowUnofficial: true,
			answers: []string{
				"COCONUT",
				"CONCOCT",
				"CONCOCTOR",
				"CONTO",
			},
			expectedScore: 28,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := NewState(t, test.filename)
			for i, word := range test.answers {
				state.Words[word] = i
			}

			state.RebuildWordMap(test.allowUnofficial)
			assert.Equal(t, test.expectedScore, state.Score)
		})
	}
}

func TestState_RebuildWordMap_Status(t *testing.T) {
	tests := []struct {
		name            string
		filename        string
		allowUnofficial bool
		answers         []string     // The answers already given
		expectedStatus  model.Status // The expected status after rebuilding
	}{
		{
			name:           "no answers",
			filename:       "nytbee-20200408.html",
			expectedStatus: model.StatusSelected,
		},
		{
			name:     "no unofficial answers",
			filename: "nytbee-20200408.html",
			answers: []string{
				"COCONUT",
				"CONCOCT",
			},
			expectedStatus: model.StatusSelected,
		},
		{
			name:     "one unofficial answer",
			filename: "nytbee-20200408.html",
			answers: []string{
				"CONCOCTOR",
			},
			expectedStatus: model.StatusSelected,
		},
		{
			name:     "all official answers, no unofficial ones",
			filename: "nytbee-20200408.html",
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
			expectedStatus: model.StatusComplete,
		},
		{
			name:     "all official answers, some unofficial answers",
			filename: "nytbee-20200408.html",
			answers: []string{
				"COCONUT",
				"CONCOCT",
				"CONCOCTOR",
				"CONTORT",
				"CONTO",
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
			expectedStatus: model.StatusComplete,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := NewState(t, test.filename)
			for i, word := range test.answers {
				state.Words[word] = i
			}

			state.RebuildWordMap(test.allowUnofficial)
			assert.Equal(t, test.expectedStatus, state.Status)
		})
	}
}

func TestGetAllChannels(t *testing.T) {
	type ChannelToCreate struct {
		name     string
		filename string
		status   model.Status
	}

	tests := []struct {
		name     string
		channels []ChannelToCreate
		expected []model.Channel
	}{
		{
			name: "no channels",
		},
		{
			name: "no selected puzzle",
			channels: []ChannelToCreate{
				{
					name:   "channel",
					status: model.StatusCreated,
				},
			},
			expected: []model.Channel{
				{
					Name:   "channel",
					Status: model.StatusCreated,
				},
			},
		},
		{
			name: "channel with nyt puzzle",
			channels: []ChannelToCreate{
				{
					name:     "channel",
					filename: "nytbee-20200408.json",
					status:   model.StatusSolving,
				},
			},
			expected: []model.Channel{
				{
					Name:        "channel",
					Status:      model.StatusSolving,
					Description: "New York Times puzzle from 2020-04-08",
					Puzzle: model.PuzzleSource{
						Publisher:     "The New York Times",
						PublishedDate: time.Date(2020, time.April, 8, 0, 0, 0, 0, time.UTC),
					},
				},
			},
		},
		{
			name: "multiple channels",
			channels: []ChannelToCreate{
				{
					name:     "channel1",
					filename: "nytbee-20180729.json",
					status:   model.StatusPaused,
				},
				{
					name:     "channel2",
					filename: "nytbee-20200408.json",
					status:   model.StatusSolving,
				},
			},
			expected: []model.Channel{
				{
					Name:        "channel1",
					Status:      model.StatusPaused,
					Description: "New York Times puzzle from 2018-07-29",
					Puzzle: model.PuzzleSource{
						Publisher:     "The New York Times",
						PublishedDate: time.Date(2018, time.July, 29, 0, 0, 0, 0, time.UTC),
					},
				},
				{
					Name:        "channel2",
					Status:      model.StatusSolving,
					Description: "New York Times puzzle from 2020-04-08",
					Puzzle: model.PuzzleSource{
						Publisher:     "The New York Times",
						PublishedDate: time.Date(2020, time.April, 8, 0, 0, 0, 0, time.UTC),
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, pool, _ := NewTestRouter(t)
			conn := NewRedisConnection(t, pool)

			// Create a state for each channel
			for _, create := range test.channels {
				var state State
				if create.filename != "" {
					state = NewState(t, create.filename)
				}

				state.Status = create.status
				require.NoError(t, SetState(conn, create.name, state))
			}

			channels, err := GetAllChannels(conn)
			require.NoError(t, err)
			assert.ElementsMatch(t, test.expected, channels)
		})
	}
}

func TestGetAllChannels_Error(t *testing.T) {
	tests := []struct {
		name       string
		connection ConnectionFunc
	}{
		{
			name: "db.ScanKeys error",
			connection: func(command string, args ...interface{}) (interface{}, error) {
				switch command {
				case "SCAN":
					return nil, errors.New("forced error")
				default:
					return nil, fmt.Errorf("unrecognized command: %s", command)
				}
			},
		},
		{
			name: "db.GetAll error",
			connection: func(command string, args ...interface{}) (interface{}, error) {
				switch command {
				case "SCAN":
					values := []interface{}{int64(0), []interface{}{"channel"}}
					return values, nil
				case "MGET":
					return nil, errors.New("forced error")
				default:
					return nil, fmt.Errorf("unrecognized command: %s", command)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := GetAllChannels(test.connection)
			assert.Error(t, err)
			assert.Equal(t, "forced error", err.Error())
		})
	}
}

type ConnectionFunc func(command string, args ...interface{}) (interface{}, error)

func (cf ConnectionFunc) Do(command string, args ...interface{}) (interface{}, error) {
	return cf(command, args...)
}
