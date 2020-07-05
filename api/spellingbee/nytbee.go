package spellingbee

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/bbeck/puzzles-with-chat/api/web"
	"io"
	"sort"
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

	puzzle.Description = fmt.Sprintf("New York Times puzzle from %s", published.Format("2006-01-02"))
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

	// Over time the format of the HTML has changed slightly putting the answers
	// in a slightly different place.  Also when the site was first live it didn't
	// always include unofficial answers.  These selectors are all of the known
	// official answer locations along with a boolean that indicates whether or
	// not this particular file format includes unofficial answers.
	selectors := map[string]bool{
		"#main-answer-list ul li":           true,
		"#top-container #answer-list ul li": false,
		"#top-container .answer-list ul li": false,
		"#answer-list ul.column-list li":    false,
	}

	var official []string
	var unofficialAnswersRequired bool
	for selector, unofficialRequired := range selectors {
		official = words(selector)
		if official != nil {
			unofficialAnswersRequired = unofficialRequired
			break
		}
	}

	unofficial := words("#not_official .answer-list ul li")

	puzzle, err := InferPuzzle(official, unofficial, unofficialAnswersRequired)
	if err != nil {
		return nil, err
	}

	// Add calculated values to the created puzzle object.
	puzzle.MaximumOfficialScore = puzzle.ComputeScore(puzzle.OfficialAnswers)
	puzzle.MaximumUnofficialScore = puzzle.MaximumOfficialScore + puzzle.ComputeScore(puzzle.UnofficialAnswers)
	puzzle.NumOfficialAnswers = len(puzzle.OfficialAnswers)
	puzzle.NumUnofficialAnswers = len(puzzle.UnofficialAnswers)

	return puzzle, nil
}

// InferPuzzle looks at the list of words and determines the puzzle structure
// from them.  In addition it verifies that the provided words have a valid
// structure for the puzzle.
func InferPuzzle(official, unofficial []string, unofficialRequired bool) (*Puzzle, error) {
	// Verify we found words
	if len(official) == 0 {
		return nil, errors.New("no official words")
	}
	if unofficialRequired && len(unofficial) == 0 {
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

	var words []string
	words = append(words, official...)
	words = append(words, unofficial...)

	center, letters, err := InferLetters(words)
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
			// This letter appears in every word so it must be the central letter in
			// the puzzle grid.  There are cases where multiple letters can be
			// considered the central letter (like when there is a single vowel in
			// the grid), but in those cases the puzzle really isn't any different
			// with either candidate as the central letter.  So we'll just choose the
			// one that appears first alphabetically.
			if center == "" {
				center = letter
				continue
			}

			if letter < center {
				letters = append(letters, center)
				center = letter
				continue
			}
		}

		letters = append(letters, letter)
	}

	if center == "" {
		return "", nil, errors.New("unable to determine center letter")
	}

	if len(letters) != 6 {
		return "", nil, fmt.Errorf("unable to determine 6 non-center letters: %v", letters)
	}

	// Make sure the letters are always in the same order.
	sort.Strings(letters)

	return center, letters, nil
}

var NYTBeeFirstPuzzleDate = time.Date(2018, time.July, 29, 0, 0, 0, 0, time.UTC)

// LoadAvailableNYTBeeDates calculates the set of available dates for spelling
// bee puzzles on the nytbee.com site.  Since the spelling bee puzzle is
// available daily this is just a computation and doesn't require any network
// calls.
func LoadAvailableNYTBeeDates() []time.Time {
	now := time.Now().UTC()

	var dates []time.Time
	for date := NYTBeeFirstPuzzleDate; date.Before(now) || date.Equal(now); date = date.AddDate(0, 0, 1) {
		dates = append(dates, date)
	}

	return dates
}
