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
			if letter != ' ' {
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

	author, title, err := ParseAuthorAndTitle(raw.Quote)
	if err != nil {
		return nil, err
	}

	var puzzle Puzzle
	puzzle.Description = fmt.Sprintf("New York Times puzzle from %s", published.Format("2006-01-02"))
	puzzle.Rows = raw.Rows
	puzzle.Cols = raw.Cols
	puzzle.Publisher = "The New York Times"
	puzzle.PublishedDate = published
	puzzle.Author = author
	puzzle.Title = title
	puzzle.Quote = raw.FullQuote
	puzzle.Cells = cells
	puzzle.CellBlocks = blocks
	puzzle.CellNumbers = numbers
	puzzle.CellClueLetters = letters
	puzzle.Clues = clues
	puzzle.ClueNumbers = clueNumbers

	return &puzzle, nil
}

// A regular expression that matches the author and title of and acrostic.  This
// is meant to be executed on the quote field for an xwordinfo.com JSON API
// response.
var AuthorTitleRegexp = regexp.MustCompile(
	`^(?P<author>[^,]+), (?P<title>[^-–—]+) [-–—]`,
)

// ParseAuthorAndTitle extracts the author name and title from the quote field
// of the xwordinfo.com JSON API response.  If the author cannot be determined
// then an error is returned.
func ParseAuthorAndTitle(s string) (string, string, error) {
	if match := AuthorTitleRegexp.FindStringSubmatch(s); len(match) != 0 {
		author := match[1]
		title := match[2]

		// Sometimes the author has special characters in it that are escaped or
		// it's too long relative to the number of clues and they leave off the
		// first name.  When this happens the first name is surrounded by
		// parentheses that need to be removed.
		author = html.UnescapeString(author)
		author = strings.ReplaceAll(author, "(", "")
		author = strings.ReplaceAll(author, ")", "")

		// Sometimes the title has special characters in it that are escaped.
		title = html.UnescapeString(title)

		return author, title, nil
	}

	return "", "", fmt.Errorf("unable to determine author and title from quote: %s", s)
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
