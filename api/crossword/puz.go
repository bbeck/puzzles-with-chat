package crossword

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/bbeck/twitch-plays-crosswords/api/web"
	"io"
)

// The URL to the converter service.
const ConverterURL = "http://converter:5001/puz" // TODO: Environment variable?

// LoadFromEncodedPuzFile loads a crossword puzzle from the base64 encoded bytes
// of the .puz file using the converter service.
//
// If the puzzle cannot be loaded or parsed then an error is returned.
func LoadFromEncodedPuzFile(encoded string) (*Puzzle, error) {
	if testCachedPuzzle != nil {
		return testCachedPuzzle, nil
	}

	if testCachedError != nil {
		return nil, testCachedError
	}

	bs, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		err = fmt.Errorf("unable to base64 decode .puz bytes: %+v", err)
		return nil, err
	}

	response, err := web.Post(ConverterURL, bytes.NewReader(bs))
	if response != nil {
		defer func() { _ = response.Body.Close() }()
	}
	if err != nil {
		return nil, err
	}

	var puzzle Puzzle
	if err := json.NewDecoder(response.Body).Decode(&puzzle); err != nil {
		return nil, fmt.Errorf("unable to parse JSON response: %v", err)
	}

	return &puzzle, nil
}

// ParseConverterResponse converts a JSON response from the puzzle converter
// service into a puzzle object.
func ParseConverterResponse(in io.Reader) (*Puzzle, error) {
	var puzzle *Puzzle
	if err := json.NewDecoder(in).Decode(&puzzle); err != nil {
		return nil, fmt.Errorf("unable to parse JSON response: %v", err)
	}

	return puzzle, nil
}
