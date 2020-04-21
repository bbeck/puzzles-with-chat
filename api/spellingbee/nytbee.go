package spellingbee

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/bbeck/twitch-plays-crosswords/api/web"
	"io"
	"strings"
	"time"
)

// LoadFromNYTBee loads a spelling bee puzzle from the NYTBee website for a
// particular date.
//
// This method loads the HTML of the spelling bee page for a desired date and
// parses it to obtain the answer list(s) from it.
//
// If the puzzle cannot be loaded or the HTML properly parsed then an error is
// returned.
func LoadFromNYTBee(date string) (*Puzzle, error) {
	if testPuzzle != nil {
		return testPuzzle, nil
	}

	if testPuzzleLoadError != nil {
		return nil, testPuzzleLoadError
	}

	published, err := time.Parse("2006-01-02", date)
	if err != nil {
		err = fmt.Errorf("unable to parse date %s: %+v", date, err)
		return nil, err
	}

	// Load the HTML page for this date from nytbee.com.
	url := fmt.Sprintf("https://nytbee.com/Bee_%04d%02d%02d.html", published.Year(), published.Month(), published.Day())
	response, err := web.Get(url)
	if response != nil {
		defer func() { _ = response.Body.Close() }()
	}
	if err != nil {
		return nil, err
	}

	puzzle, err := ParseNYTBeeResponse(response.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to parse nytbee.com response for date %s: %v", published, err)
	}

	puzzle.PublishedDate = published
	return puzzle, nil
}

// ParseNYTBeeResponse converts an HTML page from nytbee.com into a puzzle
// object.
func ParseNYTBeeResponse(in io.Reader) (*Puzzle, error) {
	doc, err := goquery.NewDocumentFromReader(in)
	if err != nil {
		return nil, err
	}

	words := func(selector string) []string {
		var words []string
		doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			words = append(words, strings.ToUpper(strings.TrimSpace(s.Text())))
		})

		return words
	}

	official := words("#main-answer-list ul li")
	unofficial := words("#not_official .answer-list ul li")

	return InferPuzzle(official, unofficial)
}

// InferPuzzle looks at the list of words and determines the puzzle structure
// from them.  In addition it verifies that the provided words have a valid
// structure for the puzzle.
func InferPuzzle(official, unofficial []string) (*Puzzle, error) {
	// Verify we found words
	if len(official) == 0 {
		return nil, errors.New("no official words")
	}
	if len(unofficial) == 0 {
		return nil, errors.New("no unofficial words")
	}

	// Verify the words are all 4 letters or more
	for _, word := range official {
		if len(word) < 4 {
			return nil, fmt.Errorf("official word is too short: %s", word)
		}
	}
	for _, word := range unofficial {
		if len(word) < 4 {
			return nil, fmt.Errorf("unofficial word is too short: %s", word)
		}
	}

	center, letters, err := InferLetters(official)
	if err != nil {
		return nil, fmt.Errorf("error determining letters in puzzle: %+v", err)
	}

	var puzzle Puzzle
	puzzle.CenterLetter = center
	puzzle.Letters = letters
	puzzle.OfficialAnswers = official
	puzzle.UnofficialAnswers = unofficial

	return &puzzle, nil
}

// InferLetters looks at the list of words and determines which letter is the
// center letter (because it's used in every word) and which letters are the
// surrounding letters.  If for some reason the letters can't be inferred then
// an error will be returned.
func InferLetters(words []string) (string, []string, error) {
	// determine the unique letters in the provided word
	unique := func(word string) []string {
		seen := make(map[string]struct{})
		for _, letter := range word {
			seen[string(letter)] = struct{}{}
		}

		var letters []string
		for letter := range seen {
			letters = append(letters, letter)
		}

		return letters
	}

	frequencies := make(map[string]int) // number of words each letter appears in
	for _, word := range words {
		for _, letter := range unique(word) {
			frequencies[letter]++
		}
	}

	var center string
	var letters []string
	for letter, count := range frequencies {
		if count == len(words) {
			// This letter appears in every word, it's the central letter in the
			// puzzle grid.  If we've already determined the central letter of the
			// grid then we have a problem.
			if center != "" {
				return "", nil, errors.New("multiple candidates for center letter")
			}

			center = letter
			continue
		}

		letters = append(letters, letter)
	}

	if center == "" {
		return "", nil, errors.New("unable to determine center letter")
	}

	if len(letters) != 6 {
		return "", nil, fmt.Errorf("unable to determine 6 non-center letters: %v", letters)
	}

	return center, letters, nil
}
