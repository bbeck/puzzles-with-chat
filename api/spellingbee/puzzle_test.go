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
			name:  "description",
			input: load(t, "nytbee-20200408.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := "New York Times puzzle from 2020-04-08"
				assert.Equal(t, expected, puzzle.Description)
			},
		},
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
		{
			name:  "maximum official score",
			input: load(t, "nytbee-20200408.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, 183, puzzle.MaximumOfficialScore)
			},
		},
		{
			name:  "maximum unofficial score",
			input: load(t, "nytbee-20200408.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, 384, puzzle.MaximumUnofficialScore)
			},
		},
		{
			name:  "num official answers",
			input: load(t, "nytbee-20200408.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, 48, puzzle.NumOfficialAnswers)
			},
		},
		{
			name:  "num unofficial answers",
			input: load(t, "nytbee-20200408.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, 38, puzzle.NumUnofficialAnswers)
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
		name     string
		filename string
	}{
		{
			name:     "nytbee-20200408",
			filename: "nytbee-20200408.html",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			puzzle := LoadTestPuzzle(t, test.filename)
			without := puzzle.WithoutAnswers()

			assert.Equal(t, puzzle.Description, without.Description)
			assert.Equal(t, puzzle.PublishedDate, without.PublishedDate)
			assert.Equal(t, puzzle.CenterLetter, without.CenterLetter)
			assert.Equal(t, puzzle.Letters, without.Letters)
			assert.Equal(t, puzzle.MaximumOfficialScore, without.MaximumOfficialScore)
			assert.Equal(t, puzzle.MaximumUnofficialScore, without.MaximumUnofficialScore)
			assert.Equal(t, puzzle.NumOfficialAnswers, without.NumOfficialAnswers)
			assert.Equal(t, puzzle.NumUnofficialAnswers, without.NumUnofficialAnswers)
			assert.Nil(t, without.OfficialAnswers)
			assert.Nil(t, without.UnofficialAnswers)
		})
	}
}

func TestPuzzle_ComputeScore(t *testing.T) {
	tests := []struct {
		name     string
		center   string
		letters  []string
		words    []string
		expected int
	}{
		{
			name:     "no words",
			center:   "T",
			letters:  []string{"C", "N", "O", "R", "U", "Y"},
			words:    []string{},
			expected: 0,
		},
		{
			name:     "length 4",
			center:   "T",
			letters:  []string{"C", "N", "O", "R", "U", "Y"},
			words:    []string{"RUNT"},
			expected: 1,
		},
		{
			name:     "length 5",
			center:   "T",
			letters:  []string{"C", "N", "O", "R", "U", "Y"},
			words:    []string{"COUNT"},
			expected: 5,
		},
		{
			name:     "pangram",
			center:   "T",
			letters:  []string{"C", "N", "O", "R", "U", "Y"},
			words:    []string{"COUNTRY"},
			expected: 14,
		},
		{
			name:     "multiple words",
			center:   "T",
			letters:  []string{"C", "N", "O", "R", "U", "Y"},
			words:    []string{"RUNT", "COUNT", "COUNTRY"},
			expected: 20,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			puzzle := &Puzzle{
				CenterLetter: test.center,
				Letters:      test.letters,
			}

			assert.Equal(t, test.expected, puzzle.ComputeScore(test.words))
		})
	}
}
