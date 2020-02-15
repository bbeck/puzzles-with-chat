package crossword

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// The HTTP client to use when communicating with the xwordinfo site.
var XWordInfoHTTPClient = &http.Client{
	Timeout: 5 * time.Second,
}

// LoadFromNewYorkTimes loads a crossword puzzle from the New York Times for a
// particular json.
//
// This method uses the xwordinfo.com JSON API to load a New York Times
// crossword puzzle.  While organized slightly differently from the XPF API the
// returned data is mostly the same.  Documentation for the JSON API can be
// found here: https://www.xwordinfo.com/JSON/
//
// If the puzzle cannot be loaded or parsed then an error is returned.
func LoadFromNewYorkTimes(date string) (*Puzzle, error) {
	if testCachedPuzzle != nil {
		return testCachedPuzzle, nil
	}

	if testCachedError != nil {
		return nil, testCachedError
	}

	url := fmt.Sprintf("https://www.xwordinfo.com/JSON/Data.aspx?date=%s", date)
	response, err := FetchXWordInfo(XWordInfoHTTPClient, url)
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

func FetchXWordInfo(client *http.Client, url string) (*http.Response, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create http request at %s: %v", url, err)
	}
	request.Header.Add("Referer", "https://www.xwordinfo.com/JSON")

	response, err := client.Do(request)
	if err != nil {
		return response, fmt.Errorf("unable to get crossword at %s: %v", url, err)
	}

	if response.StatusCode != 200 {
		return response, fmt.Errorf("received non-200 response when getting crossword at %s: %v", url, err)
	}

	return response, nil
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
	Grid     []string `json:"grid"`
	GridNums []int    `json:"gridnums"`
	Circles  []int    `json:"circles"`
	Clues    struct {
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
	for row := 0; row < raw.Size.Rows; row++ {
		circles = append(circles, make([]bool, raw.Size.Cols))
		for col := 0; col < raw.Size.Cols; col++ {
			index := row*raw.Size.Cols + col

			// Not every puzzle has circles, so make sure we check first if the index
			// is present before trying to look it up.  We don't do this check outside
			// of the loop because we want to take advantage of the initialization of
			// the circles array that the loop does for us.
			if len(raw.Circles) > 0 {
				circles[row][col] = raw.Circles[index] == 1
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
