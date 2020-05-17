package crossword

import (
	"fmt"
	"time"
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
	published, err := time.Parse("2006-01-02", date)
	if err != nil {
		err = fmt.Errorf("unable to parse date %s: %+v", date, err)
		return nil, err
	}

	// Download the .puz file from the herbach.dnsalias.com site.
	url := fmt.Sprintf("http://herbach.dnsalias.com/wsj/wsj%02d%02d%02d.puz", published.Year()-2000, published.Month(), published.Day())
	puzzle, err := LoadFromPuzFileURL(url)
	if err != nil {
		return nil, err
	}

	puzzle.Description = fmt.Sprintf("Wall Street Journal puzzle from %s", published.Format("2006-01-02"))

	// Normally .puz files don't have puzzle dates recorded in them, but we
	// happen to know the date for this puzzle, so fill it in.
	puzzle.PublishedDate = published
	puzzle.Publisher = "The Wall Street Journal"

	return puzzle, nil
}
