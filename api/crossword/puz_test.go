package crossword

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func Test_LoadFromEncodedPuzFile_Error(t *testing.T) {
	tests := []struct {
		name    string
		encoded string
	}{
		{
			name:    "non-base64 encoded input",
			encoded: "not base64",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := LoadFromEncodedPuzFile(test.encoded)
			require.Error(t, err)
		})
	}
}

func Test_ConvertPuzBytes(t *testing.T) {
	tests := []struct {
		name     string
		response io.ReadCloser
		verify   func(t *testing.T, puzzle *Puzzle)
	}{
		{
			name:     "size",
			response: load(t, "converter-nyt-20080912-notes.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, 15, puzzle.Cols)
				assert.Equal(t, 15, puzzle.Rows)
			},
		},
		{
			name:     "title",
			response: load(t, "converter-nyt-20080912-notes.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, "NY Times, Fri, Sep 12, 2008", puzzle.Title)
			},
		},
		{
			name:     "publisher",
			response: load(t, "converter-nyt-20080912-notes.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, "", puzzle.Publisher) // .puz files don't contain a publisher field
			},
		},
		{
			name:     "published date",
			response: load(t, "converter-nyt-20080912-notes.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, time.Time{}, puzzle.PublishedDate) // .puz files don't contain a published date field
			},
		},
		{
			name:     "author",
			response: load(t, "converter-nyt-20080912-notes.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, "Natan Last / Will Shortz", puzzle.Author)
			},
		},
		{
			name:     "cells",
			response: load(t, "converter-nyt-20080912-notes.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := [][]string{
					{"P", "A", "S", "O", "D", "O", "B", "L", "E", "", "", "G", "R", "U", "B"},
					{"I", "S", "T", "H", "I", "S", "L", "O", "V", "E", "", "A", "H", "S", "O"},
					{"C", "H", "O", "C", "O", "H", "O", "L", "I", "C", "", "Z", "E", "E", "S"},
					{"T", "Y", "P", "O", "", "K", "O", "A", "L", "A", "B", "E", "A", "R", "S"},
					{"", "", "", "M", "O", "O", "D", "", "D", "R", "U", "B", "", "", ""},
					{"", "A", "M", "E", "N", "S", "", "G", "O", "T", "L", "O", "O", "S", "E"},
					{"S", "W", "O", "O", "S", "H", "", "R", "E", "E", "L", "", "P", "E", "S"},
					{"H", "A", "U", "N", "T", "", "C", "A", "R", "", "H", "B", "E", "A", "M"},
					{"O", "R", "R", "", "R", "O", "O", "M", "", "P", "O", "O", "D", "L", "E"},
					{"D", "E", "N", "T", "I", "S", "T", "S", "", "A", "R", "O", "S", "E", ""},
					{"", "", "", "S", "K", "I", "T", "", "S", "I", "N", "G", "", "", ""},
					{"Q", "U", "A", "K", "E", "R", "O", "A", "T", "S", "", "A", "P", "S", "E"},
					{"T", "N", "U", "T", "", "I", "N", "L", "A", "L", "A", "L", "A", "N", "D"},
					{"I", "D", "E", "S", "", "S", "T", "E", "V", "E", "D", "O", "R", "E", "D"},
					{"P", "O", "R", "K", "", "", "O", "K", "E", "Y", "D", "O", "K", "E", "Y"},
				}
				assert.Equal(t, expected, puzzle.Cells)
			},
		},
		{
			name:     "blocks",
			response: load(t, "converter-nyt-20080912-notes.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := [][]bool{
					{false, false, false, false, false, false, false, false, false, true, true, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, true, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, true, false, false, false, false},
					{false, false, false, false, true, false, false, false, false, false, false, false, false, false, false},
					{true, true, true, false, false, false, false, true, false, false, false, false, true, true, true},
					{true, false, false, false, false, false, true, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, true, false, false, false, false, true, false, false, false},
					{false, false, false, false, false, true, false, false, false, true, false, false, false, false, false},
					{false, false, false, true, false, false, false, false, true, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, true, false, false, false, false, false, true},
					{true, true, true, false, false, false, false, true, false, false, false, false, true, true, true},
					{false, false, false, false, false, false, false, false, false, false, true, false, false, false, false},
					{false, false, false, false, true, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, true, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, true, true, false, false, false, false, false, false, false, false, false},
				}
				assert.Equal(t, expected, puzzle.CellBlocks)
			},
		},
		{
			name:     "cell clue numbers",
			response: load(t, "converter-nyt-20080912-notes.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := [][]int{
					{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 0, 10, 11, 12, 13},
					{14, 0, 0, 0, 0, 0, 0, 0, 0, 15, 0, 16, 0, 0, 0},
					{17, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 18, 0, 0, 0},
					{19, 0, 0, 0, 0, 20, 0, 0, 0, 0, 21, 0, 0, 0, 0},
					{0, 0, 0, 22, 23, 0, 0, 0, 24, 0, 0, 0, 0, 0, 0},
					{0, 25, 26, 0, 0, 0, 0, 27, 0, 0, 0, 0, 28, 29, 30},
					{31, 0, 0, 0, 0, 0, 0, 32, 0, 0, 0, 0, 33, 0, 0},
					{34, 0, 0, 0, 0, 0, 35, 0, 0, 0, 36, 37, 0, 0, 0},
					{38, 0, 0, 0, 39, 40, 0, 0, 0, 41, 0, 0, 0, 0, 0},
					{42, 0, 0, 43, 0, 0, 0, 0, 0, 44, 0, 0, 0, 0, 0},
					{0, 0, 0, 45, 0, 0, 0, 0, 46, 0, 0, 0, 0, 0, 0},
					{47, 48, 49, 0, 0, 0, 0, 50, 0, 0, 0, 51, 52, 53, 54},
					{55, 0, 0, 0, 0, 56, 0, 0, 0, 0, 57, 0, 0, 0, 0},
					{58, 0, 0, 0, 0, 59, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					{60, 0, 0, 0, 0, 0, 61, 0, 0, 0, 0, 0, 0, 0, 0},
				}
				assert.Equal(t, expected, puzzle.CellClueNumbers)
			},
		},
		{
			name:     "cell circles (none)",
			response: load(t, "converter-nyt-20080912-notes.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := [][]bool{
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
				}
				assert.Equal(t, expected, puzzle.CellCircles)
			},
		},
		{
			name:     "cell circles",
			response: load(t, "converter-nyt-20081006-nonsquare-with-circles.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := [][]bool{
					{true, true, true, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, true, true, true},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
					{true, true, true, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, true, true, true},
				}
				assert.Equal(t, expected, puzzle.CellCircles)
			},
		},
		{
			name:     "across clues",
			response: load(t, "converter-nyt-20080912-notes.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := map[int]string{
					1:  `Dance that simulates the drama of a bullfight`,
					10: `Chuck wagon fare`,
					14: `1978 Bob Marley hit whose title words are sung four times before "... that I'm feelin'"`,
					16: `Faux Japanese reply`,
					17: `One needing kisses, say`,
					18: `Jazz duo?`,
					19: `Nooks for books, maybe`,
					20: `Furry folivores`,
					22: `It may be set with music`,
					24: `Cudgel`,
					25: `Believers' comments`,
					27: `Escaped`,
					31: `Sound at an auto race`,
					32: `It holds the line`,
					33: `Foot of the Appian Way?`,
					34: `Trouble, in a way`,
					35: `Locale of some mirrors`,
					36: `Letter-shaped girder`,
					38: `Lord John Boyd ___, winner of the 1949 Nobel Peace Prize`,
					39: `Study, say`,
					41: `Winston Churchill's Rufus, for one`,
					42: `They know the drill`,
					44: `Turned up`,
					45: `Child's play, perhaps`,
					46: `Snitch`,
					47: `Company that makes Aunt Jemima syrup`,
					51: `Area next to an ambulatory`,
					55: `Letter-shaped fastener`,
					56: `Daydreaming`,
					58: `Days of old`,
					59: `Worked the docks`,
					60: `Waste of Congress?`,
					61: `"You got it!"`,
				}
				assert.Equal(t, expected, puzzle.CluesAcross)
			},
		},
		{
			name:     "down clues",
			response: load(t, "converter-nyt-20080912-notes.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := map[int]string{
					1:  `Early Inverness resident`,
					2:  `Cadaverous`,
					3:  `Ticklee's cry`,
					4:  `"You have got to be kidding!"`,
					5:  `The Divine, to da Vinci`,
					6:  `City at the mouth of the Fox River`,
					7:  `Shade of red`,
					8:  `"She was ___ in slacks" (part of an opening soliloquy by Humbert Humbert)`,
					9:  `Baddie`,
					10: `Shady spot in a 52-Down`,
					11: `Cousin of a cassowary`,
					12: `___ fee`,
					13: `One with fire power?`,
					15: `Trick-taking game`,
					21: `March instrument?`,
					23: `Out`,
					25: `Au courant`,
					26: `Keen`,
					27: `Nutrition units`,
					28: `Some essays`,
					29: `"A Lonely Rage" autobiographer Bobby`,
					30: `The farmer's wife in "Babe"`,
					31: `Did a farrier's work`,
					35: `Start to like`,
					37: `Energetic 1960s dance with swiveling and shuffling`,
					40: `God of life, death and fertility who underwent resurrection`,
					41: `Pattern sometimes called "Persian pickles"`,
					43: `"I'm very disappointed in you"`,
					46: `Song verse`,
					47: `Canal cleaner`,
					48: `Menu option`,
					49: `Teacher of Heifetz`,
					50: `Fashion model Wek`,
					52: `See 10-Down`,
					53: `Ko-Ko's dagger in "The Mikado"`,
					54: `Current happening?`,
					57: `Kick in`,
				}
				assert.Equal(t, expected, puzzle.CluesDown)
			},
		},
		{
			name:     "notes",
			response: load(t, "converter-nyt-20080912-notes.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.True(t, strings.HasPrefix(puzzle.Notes, "TEEN PUZZLEMAKER WEEK"))
			},
		},
		{
			name:     "rebus",
			response: load(t, "converter-nyt-20080914-rebus.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, "JAN", puzzle.Cells[3][6])
				assert.Equal(t, "FEB", puzzle.Cells[2][10])
				assert.Equal(t, "MAR", puzzle.Cells[3][14])
				assert.Equal(t, "APR", puzzle.Cells[6][17])
				assert.Equal(t, "MAY", puzzle.Cells[10][18])
				assert.Equal(t, "JUN", puzzle.Cells[14][17])
				assert.Equal(t, "JUL", puzzle.Cells[17][14])
				assert.Equal(t, "AUG", puzzle.Cells[18][10])
				assert.Equal(t, "SEP", puzzle.Cells[17][6])
				assert.Equal(t, "OCT", puzzle.Cells[14][3])
				assert.Equal(t, "NOV", puzzle.Cells[10][2])
				assert.Equal(t, "DEC", puzzle.Cells[6][3])
			},
		},
		{
			name:     "non-square dimensions",
			response: load(t, "converter-nyt-20081006-nonsquare-with-circles.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, 9, puzzle.Rows)
				assert.Equal(t, 24, puzzle.Cols)
			},
		},
		{
			name:     "non-square cells",
			response: load(t, "converter-nyt-20081006-nonsquare-with-circles.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := [][]string{
					{"O", "N", "E", "G", "", "L", "E", "S", "", "D", "O", "L", "L", "A", "R", "", "H", "B", "O", "", "Z", "O", "N", "E"},
					{"P", "O", "T", "R", "O", "A", "S", "T", "", "E", "U", "G", "E", "N", "E", "", "E", "A", "R", "L", "O", "B", "E", "S"},
					{"T", "H", "E", "U", "N", "I", "T", "E", "D", "S", "T", "A", "T", "E", "S", "O", "F", "A", "M", "E", "R", "I", "C", "A"},
					{"", "", "", "N", "O", "R", "", "R", "I", "P", "", "", "", "", "O", "S", "T", "", "E", "B", "B", "", "", ""},
					{"", "S", "S", "T", "", "", "I", "N", "G", "O", "D", "W", "E", "T", "R", "U", "S", "T", "", "", "A", "M", "S", ""},
					{"E", "A", "U", "", "E", "S", "C", "", "", "T", "E", "A", "A", "C", "T", "", "", "Y", "O", "W", "", "A", "A", "H"},
					{"G", "R", "E", "A", "T", "S", "E", "A", "L", "", "W", "I", "S", "E", "", "B", "A", "L", "D", "E", "A", "G", "L", "E"},
					{"G", "A", "M", "M", "A", "R", "A", "Y", "S", "", "E", "V", "E", "L", "", "S", "T", "E", "E", "L", "D", "O", "O", "R"},
					{"O", "N", "E", "A", "L", "", "X", "E", "D", "", "Y", "E", "L", "L", "", "A", "E", "R", "", "D", "O", "O", "N", "E"}}
				assert.Equal(t, expected, puzzle.Cells)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// These tests manipulate the environment to let the ConvertPuzBytes
			// method know where the test HTTP server is at, so we save a copy before
			// each test to ensure that it doesn't get permanently changed by the test
			// case.
			saved := SaveEnvironmentVars()
			defer RestoreEnvironmentVars(saved)

			// Get the bytes of the response
			bs, err := ioutil.ReadAll(test.response)
			test.response.Close()
			require.NoError(t, err)

			// Setup a test server that returns our desired input.
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				w.Write(bs)
			}))
			defer server.Close()

			// Tell the ConvertPuzBytes method the host/port of the server.
			os.Setenv("CONVERTER_HOST", server.Listener.Addr().String())

			puzzle, err := ConvertPuzBytes([]byte("unused"))
			require.NoError(t, err)
			test.verify(t, puzzle)
		})
	}
}

func Test_ConvertPuzBytes_Error(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*httptest.Server)
		respond func(http.ResponseWriter)
	}{
		{
			name: "no CONVERTER_HOST environment variable defined",
		},
		{
			name: "service returns error",
			setup: func(server *httptest.Server) {
				os.Setenv("CONVERTER_HOST", server.Listener.Addr().String())
			},
			respond: func(w http.ResponseWriter) {
				w.WriteHeader(400)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// These tests manipulate the environment to let the ConvertPuzBytes
			// method know where the test HTTP server is at, so we save a copy before
			// each test to ensure that it doesn't get permanently changed by the test
			// case.
			saved := SaveEnvironmentVars()
			defer RestoreEnvironmentVars(saved)

			// Setup a test server that returns our desired input.
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				test.respond(w)
			}))
			defer server.Close()

			if test.setup != nil {
				test.setup(server)
			}

			_, err := ConvertPuzBytes([]byte("unused"))
			require.Error(t, err)
		})
	}
}

func Test_ParseConverterResponse_Error(t *testing.T) {
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseConverterResponse(strings.NewReader(test.input))
			require.Error(t, err)
		})
	}
}
