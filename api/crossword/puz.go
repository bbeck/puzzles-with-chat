package crossword

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bbeck/twitch-plays-crosswords/api/web"
	"io"
	"net/http"
	"os"
	"time"
)

// The HTTP client to use when communicating with the converter service.
var ConverterServiceHTTPClient = &http.Client{
	Timeout: 10 * time.Second, // puzpy is a bit slow on some .puz files
}

// LoadFromEncodedPuzFile loads a crossword puzzle from the base64 encoded bytes
// of the .puz file and uses the converter service to convert it into a puzzle
// object.
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

	return ConvertPuzBytes(bs)
}

// ConvertPuzBytes takes the bytes of a .puz file and uses the converter service
// to convert it into a puzzle object.
func ConvertPuzBytes(bs []byte) (*Puzzle, error) {
	host, ok := os.LookupEnv("CONVERTER_HOST")
	if !ok {
		return nil, errors.New("unable to determine converter service hostname")
	}

	url := fmt.Sprintf("http://%s/puz", host)
	response, err := web.PostWithClient(ConverterServiceHTTPClient, url, bytes.NewReader(bs))
	if response != nil {
		defer func() { _ = response.Body.Close() }()
	}
	if err != nil {
		return nil, err
	}

	return ParseConverterResponse(response.Body)
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
