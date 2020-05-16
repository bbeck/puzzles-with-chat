package spellingbee

import (
	"errors"
	"fmt"
	"github.com/bbeck/puzzles-with-chat/api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestState_ApplyAnswer_Words(t *testing.T) {
	tests := []struct {
		name            string
		filename        string
		initialWords    []string
		answer          string
		allowUnofficial bool
		expectedWords   []string
	}{
		{
			name:          "answer from official list",
			filename:      "nytbee-20200408.html",
			answer:        "COCONUT",
			expectedWords: []string{"COCONUT"},
		},
		{
			name:            "answer from unofficial list",
			filename:        "nytbee-20200408.html",
			answer:          "CONCOCTOR",
			allowUnofficial: true,
			expectedWords:   []string{"CONCOCTOR"},
		},
		{
			name:          "lowercase answer",
			filename:      "nytbee-20200408.html",
			answer:        "coconut",
			expectedWords: []string{"COCONUT"},
		},
		{
			name:          "words stay sorted",
			filename:      "nytbee-20200408.html",
			initialWords:  []string{"COUNTY", "CROUTON"},
			answer:        "COURT",
			expectedWords: []string{"COUNTY", "COURT", "CROUTON"},
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
		initialWords    []string
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
			initialWords: []string{"COCONUT"},
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

func TestState_ClearUnofficialAnswers_Words(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		answers  []string // The answers already given
		expected []string // The expected answers
	}{
		{
			name:     "no answers",
			filename: "nytbee-20200408.html",
		},
		{
			name:     "no unofficial answers",
			filename: "nytbee-20200408.html",
			answers: []string{
				"COCONUT",
				"CONCOCT",
			},
			expected: []string{
				"COCONUT",
				"CONCOCT",
			},
		},
		{
			name:     "one unofficial answer",
			filename: "nytbee-20200408.html",
			answers: []string{
				"CONCOCTOR",
			},
		},
		{
			name:     "multiple unofficial answers",
			filename: "nytbee-20200408.html",
			answers: []string{
				"CONCOCTOR",
				"CONTO",
			},
		},
		{
			name:     "mixed unofficial answers",
			filename: "nytbee-20200408.html",
			answers: []string{
				"COCONUT",
				"CONCOCT",
				"CONCOCTOR",
				"CONTO",
			},
			expected: []string{
				"COCONUT",
				"CONCOCT",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := NewState(t, test.filename)
			state.Words = test.answers

			state.ClearUnofficialAnswers()

			assert.ElementsMatch(t, test.expected, state.Words)
		})
	}
}

func TestState_ClearUnofficialAnswers_Score(t *testing.T) {
	tests := []struct {
		name          string
		filename      string
		answers       []string // The answers already given
		expectedScore int      // The expected score
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
			name:     "mixed unofficial answers",
			filename: "nytbee-20200408.html",
			answers: []string{
				"COCONUT",
				"CONCOCT",
				"CONCOCTOR",
				"CONTO",
			},
			expectedScore: 14,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := NewState(t, test.filename)
			state.Words = test.answers

			state.ClearUnofficialAnswers()

			assert.Equal(t, test.expectedScore, state.Score)
		})
	}
}

func TestState_ClearUnofficialAnswers_Status(t *testing.T) {
	tests := []struct {
		name           string
		filename       string
		answers        []string     // The answers already given
		expectedStatus model.Status // The expected status
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
			state.Words = test.answers

			state.ClearUnofficialAnswers()

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
					filename: "nytbee-20200408.html",
					status:   model.StatusSolving,
				},
			},
			expected: []model.Channel{
				{
					Name:   "channel",
					Status: model.StatusSolving,
				},
			},
		},
		{
			name: "multiple channels",
			channels: []ChannelToCreate{
				{
					name:     "channel1",
					filename: "nytbee-20180729.html",
					status:   model.StatusPaused,
				},
				{
					name:     "channel2",
					filename: "nytbee-20200408.html",
					status:   model.StatusSolving,
				},
			},
			expected: []model.Channel{
				{
					Name:   "channel1",
					Status: model.StatusPaused,
				},
				{
					Name:   "channel2",
					Status: model.StatusSolving,
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
