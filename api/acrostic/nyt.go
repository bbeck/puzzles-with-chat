package acrostic

import (
	"encoding/json"
	"fmt"
	"github.com/bbeck/puzzles-with-chat/api/web"
	"io"
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

	puzzle, err := ParseXWordInfoResponse(response.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to parse xwordinfo.com response for date %s: %v", date, err)
	}

	return puzzle, nil
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

// ParseXWordInfoResponse converts a JSON response from xwordinfo.com into a
// puzzle object.
func ParseXWordInfoResponse(in io.Reader) (*Puzzle, error) {
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

	var puzzle Puzzle
	puzzle.Description = fmt.Sprintf("New York Times puzzle from %s", published.Format("2006-01-02"))
	puzzle.Rows = raw.Rows
	puzzle.Cols = raw.Cols
	puzzle.Publisher = "The New York Times"
	puzzle.PublishedDate = published
	puzzle.Cells = cells
	puzzle.CellBlocks = blocks
	puzzle.CellNumbers = numbers
	puzzle.CellClueLetters = letters
	puzzle.Clues = clues
	puzzle.ClueNumbers = clueNumbers

	return &puzzle, nil
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
