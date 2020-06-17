package spellingbee

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"sort"
	"strings"
	"testing"
	"testing/iotest"
	"time"
)

func TestInferPuzzle(t *testing.T) {
	tests := []struct {
		name       string
		official   []string
		unofficial []string
		center     string
		letters    []string
	}{
		{
			name: "nytbee-20200408",
			official: []string{
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
			unofficial: []string{
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
			center:  "T",
			letters: []string{"C", "N", "O", "R", "U", "Y"},
		},
		{
			name:       "multiple options for center letter",
			official:   []string{"ABCYZ"},
			unofficial: []string{"DEYZ"},
			center:     "Y", // We take the center candidate that's first alphabetically
			letters:    []string{"A", "B", "C", "D", "E", "Z"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			puzzle, err := InferPuzzle(test.official, test.unofficial, false)
			require.NoError(t, err)

			assert.ElementsMatch(t, test.official, puzzle.OfficialAnswers)
			assert.ElementsMatch(t, test.unofficial, puzzle.UnofficialAnswers)
			assert.Equal(t, test.center, puzzle.CenterLetter)
			assert.ElementsMatch(t, test.letters, puzzle.Letters)
		})
	}
}

func TestInferPuzzle_Error(t *testing.T) {
	official := []string{
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

	unofficial := []string{
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

	tests := []struct {
		name       string
		official   []string
		unofficial []string
	}{
		{
			name:       "no official words",
			unofficial: unofficial,
		},
		{
			name:     "no unofficial words",
			official: official,
		},
		{
			name:       "official word too short",
			official:   append(official, "RUT"),
			unofficial: unofficial,
		},
		{
			name:       "unofficial word too short",
			official:   official,
			unofficial: append(unofficial, "RUT"),
		},
		{
			name: "no options for center letter",
			official: []string{
				"ABCDE",
				"FGHIJ",
			},
			unofficial: []string{
				"ABCDE",
				"FGHIJ",
			},
		},
		{
			name: "too many possible letters",
			official: []string{
				"ABCD",
				"AFGH",
				"AIJK",
			},
			unofficial: []string{
				"ABCD",
				"AFGH",
				"AIJK",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := InferPuzzle(test.official, test.unofficial, true)
			assert.Error(t, err)
		})
	}
}

func TestParseNYTBeeResponse(t *testing.T) {
	tests := []struct {
		name   string
		input  io.ReadCloser
		verify func(t *testing.T, puzzle *Puzzle)
	}{
		{
			name:  "official answers",
			input: load(t, "nytbee-20200408.html"),
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
			input: load(t, "nytbee-20200408.html"),
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
			name:  "letters",
			input: load(t, "nytbee-20200408.html"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, "T", puzzle.CenterLetter)

				expected := []string{"C", "N", "O", "R", "U", "Y"}
				assert.ElementsMatch(t, expected, puzzle.Letters)
			},
		},
		{
			name:  "maximum official score",
			input: load(t, "nytbee-20200408.html"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, 183, puzzle.MaximumOfficialScore)
			},
		},
		{
			name:  "maximum unofficial score",
			input: load(t, "nytbee-20200408.html"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, 384, puzzle.MaximumUnofficialScore)
			},
		},
		{
			name:  "num official answers",
			input: load(t, "nytbee-20200408.html"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, 48, puzzle.NumOfficialAnswers)
			},
		},
		{
			name:  "num unofficial answers",
			input: load(t, "nytbee-20200408.html"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, 38, puzzle.NumUnofficialAnswers)
			},
		},

		{
			name:  "multiple options for center letter",
			input: load(t, "nytbee-20200424-multiple-centers.html"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, "M", puzzle.CenterLetter)

				expected := []string{"N", "O", "P", "R", "T", "Y"}
				assert.ElementsMatch(t, expected, puzzle.Letters)
			},
		},
		{
			name:  "legacy format 1 (answers)",
			input: load(t, "nytbee-20180729.html"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := []string{
					"ACYCLIC",
					"ALFALFA",
					"ALLAY",
					"ALLY",
					"ANAL",
					"ANALLY",
					"CALCIFY",
					"CALF",
					"CALL",
					"CALLA",
					"CANAL",
					"CANNILY",
					"CILIA",
					"CLAN",
					"CLAY",
					"CLIFF",
					"CLINIC",
					"CLINICAL",
					"CLINICALLY",
					"CLINICIAN",
					"CYCLIC",
					"CYCLICAL",
					"CYCLICALLY",
					"CYNICAL",
					"CYNICALLY",
					"FACIAL",
					"FACIALLY",
					"FAIL",
					"FALL",
					"FALLACY",
					"FANCILY",
					"FILIAL",
					"FILIALLY",
					"FILL",
					"FILLY",
					"FINAL",
					"FINALLY",
					"FINANCIAL",
					"FINANCIALLY",
					"FLAIL",
					"FLAN",
					"FLAY",
					"ICILY",
					"ILLY",
					"INFILL",
					"INLAY",
					"LACY",
					"LAIC",
					"LAICAL",
					"LAICALLY",
					"LAIN",
					"LANAI",
					"LILAC",
					"LILY",
					"NAIL",
				}
				assert.ElementsMatch(t, expected, puzzle.OfficialAnswers)
			},
		},
		{
			name:  "legacy format 1 (letters)",
			input: load(t, "nytbee-20180729.html"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, "L", puzzle.CenterLetter)

				expected := []string{"A", "C", "F", "I", "N", "Y"}
				assert.ElementsMatch(t, expected, puzzle.Letters)
			},
		},
		{
			name:  "legacy format 2 (answers)",
			input: load(t, "nytbee-20180731.html"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := []string{
					"AIRMAN",
					"ANIMA",
					"APIAN",
					"ARIA",
					"IMAM",
					"IMPAIR",
					"MAIM",
					"MAIN",
					"MANIA",
					"MARINA",
					"MARINARA",
					"MARZIPAN",
					"MINI",
					"MINIM",
					"MINIMA",
					"PAIN",
					"PAIR",
					"PANINI",
					"PAPARAZZI",
					"PIAZZA",
					"PIMP",
					"PIZAZZ",
					"PIZZA",
					"PIZZAZZ",
					"PRIM",
					"PRIMP",
					"RAIN",
					"RAPINI",
					"RIPARIAN",
					"ZINNIA",
				}
				assert.ElementsMatch(t, expected, puzzle.OfficialAnswers)
			},
		},
		{
			name:  "legacy format 2 (letters)",
			input: load(t, "nytbee-20180731.html"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, "I", puzzle.CenterLetter)

				expected := []string{"A", "M", "N", "P", "R", "Z"}
				assert.ElementsMatch(t, expected, puzzle.Letters)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer test.input.Close()

			puzzle, err := ParseNYTBeeResponse(test.input)
			require.NoError(t, err)
			test.verify(t, puzzle)
		})
	}
}

func TestParseNYTBeeResponse_Error(t *testing.T) {
	tests := []struct {
		name  string
		input io.Reader
	}{
		{
			name:  "reader returning error",
			input: iotest.TimeoutReader(strings.NewReader("random input")),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseNYTBeeResponse(test.input)
			require.Error(t, err)
		})
	}
}

func TestLoadAvailableNYTBeeDates(t *testing.T) {
	tests := []struct {
		name     string
		expected time.Time
	}{
		{
			name:     "first puzzle date",
			expected: NYTBeeFirstPuzzleDate,
		},
		{
			name:     "today",
			expected: time.Now().UTC().Truncate(24 * time.Hour),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dates := LoadAvailableNYTBeeDates()

			index := sort.Search(len(dates), func(i int) bool {
				return dates[i].Equal(test.expected) || dates[i].After(test.expected)
			})
			assert.Equal(t, test.expected, dates[index])
		})
	}
}
