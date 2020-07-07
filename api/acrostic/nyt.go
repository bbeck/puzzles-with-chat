package acrostic

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/bbeck/puzzles-with-chat/api/web"
	"html"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// LoadFromNewYorkTimes loads an acrostic puzzle from the New York Times for a
// particular date.
//
// This method uses the xwordinfo.com JSON API to load a New York Times
// acrostic puzzle.  Unfortunately the JSON API for acrostics is not documented.
//
// If the puzzle cannot be loaded or parsed then an error is returned.
func LoadFromNewYorkTimes(date string) (*Puzzle, error) {
	if testPuzzle != nil {
		return testPuzzle, nil
	}

	if testPuzzleLoadError != nil {
		return nil, testPuzzleLoadError
	}

	url := fmt.Sprintf("https://www.xwordinfo.com/JSON/AcData.aspx?date=%s", date)
	response, err := web.Get(url)
	if response != nil {
		defer func() { _ = response.Body.Close() }()
	}
	if err != nil {
		return nil, err
	}

	puzzle, err := ParseXWordInfoPuzzleResponse(response.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to parse xwordinfo.com response for date %s: %v", date, err)
	}

	return puzzle, nil
}

// LoadAvailableNewYorkTimesDates determines all of the historical dates that
// have acrostic puzzles.
//
// This method uses the https://www.xwordinfo.com/SelectAcrostic page and parses
// the HTML on the page to determine the available puzzle dates.
//
// If the dates cannot be determined then an error is returned.
func LoadAvailableNewYorkTimesDates() ([]time.Time, error) {
	if testAvailableDates != nil {
		return testAvailableDates, nil
	}

	if testAvailableDatesLoadError != nil {
		return nil, testAvailableDatesLoadError
	}

	url := "https://www.xwordinfo.com/SelectAcrostic"
	response, err := web.Get(url)
	if response != nil {
		defer func() { _ = response.Body.Close() }()
	}
	if err != nil {
		return nil, err
	}

	dates, err := ParseXWordInfoAvailableDatesResponse(response.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to parse xwordinfo.com response for available dates: %v", err)
	}

	return dates, nil
}

// XWordInfoPuzzle is a representation of the response from the xwordinfo.com
// JSON API when querying for a puzzle.
type XWordInfoPuzzle struct {
	AnswerKey    string   `json:"answerKey"`
	ClueData     []string `json:"clueData"`
	Clues        []string `json:"clues"`
	Cols         int      `json:"cols"`
	Copyright    string   `json:"copyright"`
	Date         string   `json:"date"`
	FullQuote    string   `json:"fullQuote"`
	GridLetters  string   `json:"gridLetters"`
	GridNumbers  []int    `json:"gridNumbers"`
	TitleNumbers []int    `json:"mapTitle"`
	Quote        string   `json:"quote"`
	Rows         int      `json:"rows"`
}

// ParseXWordInfoPuzzleResponse converts a JSON response from xwordinfo.com into a
// puzzle object.
func ParseXWordInfoPuzzleResponse(in io.Reader) (*Puzzle, error) {
	var raw XWordInfoPuzzle
	if err := json.NewDecoder(in).Decode(&raw); err != nil {
		return nil, fmt.Errorf("unable to parse JSON response: %v", err)
	}

	published, err := time.Parse("1/2/2006", raw.Date)
	if err != nil {
		return nil, fmt.Errorf("unable to parse date (%s) from JSON response: %v", raw.Date, err)
	}

	if raw.AnswerKey == "" {
		return nil, fmt.Errorf("empty JSON response")
	}

	// Puzzles can end in multiple black squares because the last row isn't
	// completely filled.  When this happens we'll pad the various strings/arrays
	// so that the below loops don't all need to have special cases within them.
	for len(raw.AnswerKey) < raw.Rows*raw.Cols {
		raw.AnswerKey += " "
	}
	for len(raw.GridLetters) < raw.Rows*raw.Cols {
		raw.GridLetters += " "
	}
	for len(raw.GridNumbers) < raw.Rows*raw.Cols {
		raw.GridNumbers = append(raw.GridNumbers, 0)
	}

	var cells [][]string
	for row := 0; row < raw.Rows; row++ {
		cells = append(cells, make([]string, raw.Cols))
		for col := 0; col < raw.Cols; col++ {
			index := row*raw.Cols + col
			if raw.AnswerKey[index] != ' ' {
				cells[row][col] = string(raw.AnswerKey[index])
			}
		}
	}

	var givens [][]string
	for row := 0; row < raw.Rows; row++ {
		givens = append(givens, make([]string, raw.Cols))
		for col := 0; col < raw.Cols; col++ {
			index := row*raw.Cols + col
			if raw.AnswerKey[index] != ' ' && raw.GridNumbers[index] == 0 {
				givens[row][col] = string(raw.AnswerKey[index])
			}
		}
	}

	var blocks [][]bool
	for row := 0; row < raw.Rows; row++ {
		blocks = append(blocks, make([]bool, raw.Cols))
		for col := 0; col < raw.Cols; col++ {
			index := row*raw.Cols + col
			blocks[row][col] = raw.AnswerKey[index] == ' '
		}
	}

	var numbers [][]int
	for row := 0; row < raw.Rows; row++ {
		numbers = append(numbers, make([]int, raw.Cols))
		for col := 0; col < raw.Cols; col++ {
			index := row*raw.Cols + col
			numbers[row][col] = raw.GridNumbers[index]
		}
	}

	var letters [][]string
	for row := 0; row < raw.Rows; row++ {
		letters = append(letters, make([]string, raw.Cols))
		for col := 0; col < raw.Cols; col++ {
			letter := raw.GridLetters[row*raw.Cols+col]
			if unicode.IsLetter(rune(letter)) {
				letters[row][col] = string(letter)
			}
		}
	}

	clues := make(map[string]string)
	clueNumbers := make(map[string][]int)
	for index := 0; index < len(raw.Clues); index++ {
		letter, err := GetClueLetter(index)
		if err != nil {
			return nil, err
		}

		nums, err := ParseInts(raw.ClueData[index])
		if err != nil {
			return nil, err
		}

		clues[letter] = raw.Clues[index]
		clueNumbers[letter] = nums
	}

	author, title := ParseAuthorAndTitle(raw.Quote)

	// Sometimes the FullQuote field is empty, when that happens use the quote
	// field to determine the quote the puzzle is from.
	quote := raw.FullQuote
	if quote == "" {
		quote = ParseQuote(raw.Quote)
	}

	var puzzle Puzzle
	puzzle.Description = fmt.Sprintf("New York Times puzzle from %s", published.Format("2006-01-02"))
	puzzle.Rows = raw.Rows
	puzzle.Cols = raw.Cols
	puzzle.Publisher = "The New York Times"
	puzzle.PublishedDate = published
	puzzle.Author = author
	puzzle.Title = title
	puzzle.Quote = quote
	puzzle.Cells = cells
	puzzle.Givens = givens
	puzzle.CellBlocks = blocks
	puzzle.CellNumbers = numbers
	puzzle.CellClueLetters = letters
	puzzle.Clues = clues
	puzzle.ClueNumbers = clueNumbers

	return &puzzle, nil
}

// A set of regular expressions to use to extract the author and title of an
// acrostic.  These will be tried in order and the first one to return a match
// is the one that is used.
var AuthorTitleRegexps = []*regexp.Regexp{
	// Most of the time the author and title are easy to extract, the author is
	// the beginning of the string up until a comma and the title is from the
	// comma until a — (not a hyphen) character.
	//
	// Some examples:
	//   KEN DRUSE, THE NEW SHADE GARDEN — Plants are moving...
	//   (MABEL) WAGNALLS, STARS OF THE OPERA — People...
	//   (DORIS) NASH-WORTMAN, TITLE — Quote...
	regexp.MustCompile(`^(?P<author>[^,:]+)[,:] (?P<title>[^—]+) —`),

	// Sometimes however they use a hyphen character to separate the author and
	// title from the rest of the quote.  But more often than this there are
	// authors that have hyphenated last names, so we try the unicode character
	// from above first.
	regexp.MustCompile(`^(?P<author>[^,:]+)[,:] (?P<title>[^-]+) -`),

	// Sometimes there's just a title and no author.
	regexp.MustCompile(`(?P<author>)(?P<title>[^—]+) —`),
}

// ParseAuthorAndTitle extracts the author name and title from the quote field
// of the xwordinfo.com JSON API response.
func ParseAuthorAndTitle(s string) (string, string) {
	// Sometimes the author and title have special characters in them or are too
	// long to fit the clues so they surround some words with parentheses or
	// brackets to indicate that they were left out.
	sanitize := func(s string) string {
		s = html.UnescapeString(s)
		s = strings.ReplaceAll(s, "(", "")
		s = strings.ReplaceAll(s, ")", "")
		s = strings.ReplaceAll(s, "[", "")
		s = strings.ReplaceAll(s, "]", "")
		s = strings.Trim(s, " ")
		return s
	}

	for _, regex := range AuthorTitleRegexps {
		if match := regex.FindStringSubmatch(s); len(match) != 0 {
			author := sanitize(match[1])
			title := sanitize(match[2])
			return author, title
		}
	}

	return "", ""
}

// A set of regular expressions to use to extract the quote of an acrostic.
var QuoteRegexps = []*regexp.Regexp{
	// Most of the time the author and title are separated from the quote by a
	// unicode — (not a hyphen) character and continues to the end of the line.
	//
	// Some examples:
	//   KEN DRUSE, THE NEW SHADE GARDEN — Plants are moving...
	//   (MABEL) WAGNALLS, STARS OF THE OPERA — People...
	//   (DORIS) NASH-WORTMAN, TITLE — Quote...
	regexp.MustCompile(`^(?P<author_and_title>[^—]+) — (?P<quote>.+)$`),
}

func ParseQuote(s string) string {
	// Sometimes the quote has special characters in it that needs to be
	// sanitized.
	sanitize := func(s string) string {
		s = html.UnescapeString(s)
		s = strings.ReplaceAll(s, "[", "")
		s = strings.ReplaceAll(s, "]", "")
		s = strings.Trim(s, " ")
		return s
	}

	for _, regex := range QuoteRegexps {
		if match := regex.FindStringSubmatch(s); len(match) != 0 {
			return sanitize(match[2])
		}
	}

	// If we couldn't locate the quote then return the full string.  It'll
	// probably have the author and title in it, but that's better than nothing.
	return sanitize(s)
}

var ClueLetters = []string{
	"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O",
	"P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
}

func GetClueLetter(index int) (string, error) {
	if index < 0 || index >= len(ClueLetters) {
		return "", fmt.Errorf("invalid clue letter index: %d", index)
	}
	return ClueLetters[index], nil
}

func ParseInts(s string) ([]int, error) {
	var ints []int
	for _, part := range strings.Split(s, ",") {
		n, err := strconv.Atoi(part)
		if err != nil {
			return nil, err
		}

		ints = append(ints, n)
	}

	return ints, nil
}

// ParseXWordInfoAvailableDatesResponse converts an HTML response from the
// select acrostic page on xwordinfo.com into a list of available dates.
func ParseXWordInfoAvailableDatesResponse(in io.Reader) ([]time.Time, error) {
	doc, err := goquery.NewDocumentFromReader(in)
	if err != nil {
		return nil, err
	}

	var dates []time.Time
	doc.Find("a.dtlink").Each(func(i int, s *goquery.Selection) {
		if err != nil {
			return
		}

		href, ok := s.Attr("href")
		if !ok {
			err = fmt.Errorf("unable to determine href for selection: %v", s)
			return
		}

		d := strings.TrimPrefix(href, "/Acrostic?date=")

		var date time.Time
		if date, err = time.Parse("1/2/2006", d); err == nil {
			dates = append(dates, date)
		}
	})

	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	if len(dates) == 0 && err == nil {
		err = errors.New("no dates found")
	}

	return dates, err
}
