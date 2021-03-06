package crossword

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"
)

var puzFileCases = []struct {
	puzFilename  string // relative to the testdata/puz directory
	jsonFilename string // relative to the testdata/puz directory
}{
	{
		puzFilename:  "nyt-20081006-nonsquare.puz",
		jsonFilename: "nyt-20081006-nonsquare.json",
	},
	{
		puzFilename:  "puzpy-20080911-nyt-rebus-with-notes-and-shape.puz",
		jsonFilename: "puzpy-20080911-nyt-rebus-with-notes-and-shape.json",
	},
	{
		puzFilename:  "puzpy-avclub-20110622.puz",
		jsonFilename: "puzpy-avclub-20110622.json",
	},
	{
		puzFilename:  "puzpy-crossynergy-20080904.puz",
		jsonFilename: "puzpy-crossynergy-20080904.json",
	},
	{
		puzFilename:  "puzpy-nyt-20080203-odd-numbering.puz",
		jsonFilename: "puzpy-nyt-20080203-odd-numbering.json",
	},
	{
		puzFilename:  "puzpy-nyt-20080224-diagramless.puz",
		jsonFilename: "puzpy-nyt-20080224-diagramless.json",
	},
	{
		puzFilename:  "puzpy-nyt-20080310-partly-filled.puz",
		jsonFilename: "puzpy-nyt-20080310-partly-filled.json",
	},
	{
		puzFilename:  "puzpy-nyt-20080720-shape.puz",
		jsonFilename: "puzpy-nyt-20080720-shape.json",
	},
	{
		puzFilename:  "puzpy-nyt-20080912-weekday-with-notes.puz",
		jsonFilename: "puzpy-nyt-20080912-weekday-with-notes.json",
	},
	{
		puzFilename:  "puzpy-nyt-20080914-sunday-rebus.puz",
		jsonFilename: "puzpy-nyt-20080914-sunday-rebus.json",
	},
	{
		puzFilename:  "puzpy-nyt-20080919-locked.puz",
		jsonFilename: "puzpy-nyt-20080919-locked.json",
	},
	{
		puzFilename:  "puzpy-washpost-20051206.puz",
		jsonFilename: "puzpy-washpost-20051206.json",
	},
	{
		puzFilename:  "puzpy-wsj-20110624.puz",
		jsonFilename: "puzpy-wsj-20110624.json",
	},
	{
		puzFilename:  "cru-cryptic-20010201.puz",
		jsonFilename: "cru-cryptic-20010201.json",
	},
}

func TestLoadPuzFile_EquivalenceWithKnownGood(t *testing.T) {
	for _, test := range puzFileCases {
		test := test
		t.Run(test.puzFilename, func(t *testing.T) {
			t.Parallel()

			puzPuzzle := loadPuz(t, test.puzFilename)
			jsonPuzzle := loadJson(t, test.jsonFilename)
			assert.Equal(t, jsonPuzzle, puzPuzzle)
		})
	}
}

func TestLoadFromPuzFileURL_EquivalenceWithKnownGood(t *testing.T) {
	for _, test := range puzFileCases {
		test := test
		t.Run(test.puzFilename, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				w.WriteHeader(200)

				reader := load(t, path.Join("puz", test.puzFilename))
				_, err := io.Copy(w, reader)
				require.NoError(t, err)
			}))
			defer server.Close()

			puzPuzzle, err := LoadFromPuzFileURL(server.URL)
			assert.NoError(t, err)

			jsonPuzzle := loadJson(t, test.jsonFilename)
			assert.Equal(t, jsonPuzzle, puzPuzzle)
		})
	}
}

func loadPuz(t *testing.T, filename string) *Puzzle {
	t.Helper()

	reader := load(t, path.Join("puz", filename))
	defer reader.Close()

	puzzle, err := LoadPuzFile(reader)
	require.NoError(t, err)

	return puzzle
}

func loadJson(t *testing.T, filename string) *Puzzle {
	t.Helper()

	reader := load(t, path.Join("puz", filename))
	defer reader.Close()

	var puzzle Puzzle
	err := json.NewDecoder(reader).Decode(&puzzle)
	require.NoError(t, err)

	return &puzzle
}
