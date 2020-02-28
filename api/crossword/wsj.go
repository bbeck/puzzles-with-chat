package crossword

import (
	"fmt"
	"github.com/bbeck/twitch-plays-crosswords/api/web"
	"io/ioutil"
	"strings"
)

// LoadFromWallStreetJournal loads a crossword puzzle from the Wall Street
// Journal for a particular date.
//
// This method downloads a .puz file and loads it into a Puzzle object.  We
// do this in particular on the server side instead of within the client because
// the herbach.dnsalias.com site unfortunately is only HTTP and we can't load
// resources from a non-HTTPS site in the browser.
//
// If the puzzle cannot be loaded or parsed then an error is returned.
func LoadFromWallStreetJournal(date string) (*Puzzle, error) {
	if testCachedPuzzle != nil {
		return testCachedPuzzle, nil
	}

	if testCachedError != nil {
		return nil, testCachedError
	}

	parts := strings.Split(date, "-")
	year := parts[0][2:]
	month := parts[1]
	day := parts[2]

	// First, download the .puz file from the herbach.dnsalias.com site.
	url := fmt.Sprintf("http://herbach.dnsalias.com/wsj/wsj%s%s%s.puz", year, month, day)
	response, err := web.Get(url)
	if response != nil {
		defer func() { _ = response.Body.Close() }()
	}
	if err != nil {
		return nil, err
	}

	bs, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// Next, convert the .puz file to a puzzle using the .puz converter.
	return ConvertPuzBytes(bs)
}
