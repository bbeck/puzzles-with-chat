package spellingbee

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPuzzle_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name   string
		input  io.ReadCloser
		verify func(t *testing.T, puzzle *Puzzle)
	}{
		{
			name:  "published date",
			input: load(t, "nytbee-20200408.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := time.Date(2020, time.April, 8, 0, 0, 0, 0, time.UTC)
				assert.Equal(t, expected, puzzle.PublishedDate)
			},
		},
		{
			name:  "letters",
			input: load(t, "nytbee-20200408.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, "T", puzzle.CenterLetter)

				expected := []string{"C", "N", "O", "R", "U", "Y"}
				assert.ElementsMatch(t, expected, puzzle.Letters)
			},
		},
		{
			name:  "official answers",
			input: load(t, "nytbee-20200408.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := []string{
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
				}
				assert.ElementsMatch(t, expected, puzzle.OfficialAnswers)
			},
		},
		{
			name:  "unofficial answers",
			input: load(t, "nytbee-20200408.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := []string{
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
				}
				assert.ElementsMatch(t, expected, puzzle.UnofficialAnswers)
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

func TestPuzzle_WithoutAnswers(t *testing.T) {
	tests := []struct {
		name       string
		official   []string
		unofficial []string
	}{
		{
			name: "nil answers",
		},
		{
			name:       "empty answers",
			official:   []string{},
			unofficial: []string{},
		},
		{
			name:       "non-empty answers",
			official:   []string{"official"},
			unofficial: []string{"unofficial"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			puzzle := &Puzzle{OfficialAnswers: test.official, UnofficialAnswers: test.unofficial}
			assert.Nil(t, puzzle.WithoutAnswers().OfficialAnswers)
			assert.Nil(t, puzzle.WithoutAnswers().UnofficialAnswers)
		})
	}
}
