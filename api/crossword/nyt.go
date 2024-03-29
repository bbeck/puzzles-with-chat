package crossword

import (
	"encoding/json"
	"fmt"
	"github.com/bbeck/puzzles-with-chat/api/web"
	"html"
	"io"
	"strconv"
	"strings"
	"time"
)

var XWordInfoHeaders = map[string]string{
	"Referer": "https://www.xwordinfo.com/JSON",
}

// LoadFromNewYorkTimes loads a crossword puzzle from the New York Times for a
// particular date.
//
// This method uses the xwordinfo.com JSON API to load a New York Times
// crossword puzzle.  While organized slightly differently from the XPF API the
// returned data is mostly the same.  Documentation for the JSON API can be
// found here: https://www.xwordinfo.com/JSON/
//
// If the puzzle cannot be loaded or parsed then an error is returned.
func LoadFromNewYorkTimes(date string) (*Puzzle, error) {
	if testPuzzle != nil {
		return testPuzzle, nil
	}

	if testPuzzleLoadError != nil {
		return nil, testPuzzleLoadError
	}

	url := fmt.Sprintf("https://www.xwordinfo.com/JSON/Data.ashx?date=%s", date)
	response, err := web.GetWithHeaders(url, XWordInfoHeaders)
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
	Title     string `json:"title"`
	Author    string `json:"author"`
	Publisher string `json:"publisher"`
	Date      string `json:"date"`
	Notepad   string `json:"notepad"`
	JNotes    string `json:"jnotes"`
	Size      struct {
		Rows int `json:"rows"`
		Cols int `json:"cols"`
	} `json:"size"`
	Grid         []string `json:"grid"`
	GridNums     []int    `json:"gridnums"`
	Circles      []int    `json:"circles"`
	ShadeCircles bool   `json:"shadecircles"`
	Clues        struct {
		Across []string `json:"across"`
		Down   []string `json:"down"`
	} `json:"clues"`
	Answers struct {
		Across []string `json:"across"`
		Down   []string `json:"down"`
	} `json:"answers"`
}

// ParseXWordInfoResponse converts a JSON response from xwordinfo.com into a
// puzzle object.
func ParseXWordInfoResponse(in io.Reader) (*Puzzle, error) {
	var raw XWordInfoPuzzle
	if err := json.NewDecoder(in).Decode(&raw); err != nil {
		return nil, fmt.Errorf("unable to parse JSON response: %v", err)
	}

	// If xwordinfo.com doesn't have a puzzle it still returns a valid JSON object
	// but most of the fields are empty or missing.  Since the main component of a
	// puzzle is the grid, we'll use it as a marker of an empty, but valid
	// response.
	if len(raw.Grid) == 0 {
		return nil, fmt.Errorf("empty JSON response")
	}

	published, err := time.Parse("1/2/2006", raw.Date)
	if err != nil {
		return nil, fmt.Errorf("unable to parse date (%s) from JSON response: %v", raw.Date, err)
	}

	var cells [][]string
	for row := 0; row < raw.Size.Rows; row++ {
		cells = append(cells, make([]string, raw.Size.Cols))
		for col := 0; col < raw.Size.Cols; col++ {
			index := row*raw.Size.Cols + col
			if raw.Grid[index] != "." {
				cells[row][col] = raw.Grid[index]
			}
		}
	}

	var blocks [][]bool
	for row := 0; row < raw.Size.Rows; row++ {
		blocks = append(blocks, make([]bool, raw.Size.Cols))
		for col := 0; col < raw.Size.Cols; col++ {
			index := row*raw.Size.Cols + col
			blocks[row][col] = raw.Grid[index] == "."
		}
	}

	var numbers [][]int
	for row := 0; row < raw.Size.Rows; row++ {
		numbers = append(numbers, make([]int, raw.Size.Cols))
		for col := 0; col < raw.Size.Cols; col++ {
			index := row*raw.Size.Cols + col
			numbers[row][col] = raw.GridNums[index]
		}
	}

	var circles [][]bool
	var shades [][]bool
	for row := 0; row < raw.Size.Rows; row++ {
		circles = append(circles, make([]bool, raw.Size.Cols))
		shades = append(shades, make([]bool, raw.Size.Cols))

		for col := 0; col < raw.Size.Cols; col++ {
			index := row*raw.Size.Cols + col

			// Not every puzzle has circles, so make sure we check first if the index
			// is present before trying to look it up.  We don't do this check outside
			// of the loop because we want to take advantage of the initialization of
			// the circles array that the loop does for us.
			if len(raw.Circles) > 0 {
				// Whether or not the cell should be circled or shaded is indicated by
				// the ShadeCircles property of the response.
				if raw.ShadeCircles {
					shades[row][col] = raw.Circles[index] == 1
				} else {
					circles[row][col] = raw.Circles[index] == 1
				}
			}
		}
	}

	across := make(map[int]string)
	for _, c := range raw.Clues.Across {
		num, clue, err := ParseXWordInfoClue(c)
		if err != nil {
			return nil, fmt.Errorf("unable to parse clue text %s: %v", c, err)
		}

		across[num] = clue
	}

	down := make(map[int]string)
	for _, c := range raw.Clues.Down {
		num, clue, err := ParseXWordInfoClue(c)
		if err != nil {
			return nil, fmt.Errorf("unable to parse clue text %s: %v", c, err)
		}

		down[num] = clue
	}

	var puzzle Puzzle
	puzzle.Description = fmt.Sprintf("New York Times puzzle from %s", published.Format("2006-01-02"))
	puzzle.Rows = raw.Size.Rows
	puzzle.Cols = raw.Size.Cols
	puzzle.Title = raw.Title
	puzzle.Publisher = raw.Publisher
	puzzle.PublishedDate = published
	puzzle.Author = raw.Author
	puzzle.Cells = cells
	puzzle.CellBlocks = blocks
	puzzle.CellClueNumbers = numbers
	puzzle.CellCircles = circles
	puzzle.CellShades = shades
	puzzle.CluesAcross = across
	puzzle.CluesDown = down

	if raw.Notepad != "" && raw.JNotes != "" {
		puzzle.Notes = raw.Notepad + "<br/>" + raw.JNotes
	} else if raw.Notepad != "" {
		puzzle.Notes = raw.Notepad
	} else {
		puzzle.Notes = raw.JNotes
	}

	return &puzzle, nil
}

// ParseXWordInfoClue parses the text of a clue from the New York Times into
// its clue number and clue text.
func ParseXWordInfoClue(s string) (int, string, error) {
	// Clues look like the following:
	//   1. 4.0 is a great one
	//   13. &quot;Look out!&quot;
	//   67. ___ raving mad
	//
	// Because of this we need to make sure we split only after the first
	// decimal point as there may be other decimal points in the clue.  Also
	// we should unescape HTML characters that may be present as well.
	parts := strings.SplitN(s, ".", 2)
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("incorrect number of parts when parsing clue text: %s", s)
	}

	num, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", fmt.Errorf("unable to parse clue number %s: %v", parts[0], err)
	}

	clue := html.UnescapeString(parts[1])
	clue = strings.Trim(clue, " \t\n")
	return num, clue, nil
}

var NYTFirstPuzzleDate = time.Date(1942, time.February, 15, 0, 0, 0, 0, time.UTC)

// Prior to 1950-09-11 puzzles were only on Sundays.
var NYTSwitchToDailyDate = time.Date(1950, time.September, 11, 0, 0, 0, 0, time.UTC)

// LoadAvailableNYTDates calculates the set of available dates for crossword
// puzzles from The New York Times.
func LoadAvailableNYTDates() []time.Time {
	now := time.Now().UTC()

	var dates []time.Time
	for date := NYTFirstPuzzleDate; date.Before(now) || date.Equal(now); date = date.AddDate(0, 0, 1) {
		if date.Before(NYTSwitchToDailyDate) && date.Weekday() != time.Sunday {
			continue
		}

		dates = append(dates, date)
	}

	return dates
}
