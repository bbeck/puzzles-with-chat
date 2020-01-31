package crossword

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// State represents the state of an active channel that is attempting to solve
// a crossword.
type State struct {
	// The status of the channel's crossword solve.
	Status Status `json:"status"`

	// The crossword puzzle that's being solved.  May not always be present, for
	// example when the state is being serialized to be sent to the browser.
	Puzzle *Puzzle `json:"puzzle,omitempty"`

	// The currently filled in cells of the crossword.
	Cells [][]string `json:"cells"`

	// Whether or not an across clue with a given clue number has had an answer
	// filled in.
	AcrossCluesFilled map[int]bool `json:"across_clues_filled"`

	// Whether or not a down clue with a given clue number has had an answer
	// filled in.
	DownCluesFilled map[int]bool `json:"down_clues_filled"`

	// The time that we last started or resumed solving the puzzle.  If the
	// channel has not yet started solving the puzzle or is in a non-playing state
	// this will be nil.
	LastStartTime *time.Time `json:"last_start_time,omitempty"`

	// The total time spent on solving the puzzle up to the last start time.
	TotalSolveDuration Duration `json:"total_solve_duration"`
}

// ApplyAnswer applies an answer for a clue to the state.  If the clue cannot
// be identified or the answer doesn't fit property (too short or too long) then
// an error will be returned.  If the onlyCorrect parameter is true then only
// correct cells will be permitted and an error is returned if any part of the
// answer is incorrect or would remove a correct cell.
func (s *State) ApplyAnswer(clue string, answer string, onlyCorrect bool) error {
	num, direction, err := ParseClue(clue)
	if err != nil {
		return err
	}

	cells, err := ParseAnswer(answer)
	if err != nil {
		return err
	}

	minX, minY, maxX, maxY, err := s.Puzzle.GetAnswerCoordinates(num, direction)
	if err != nil {
		return err
	}

	// Check to see if our cell values are compatible with the size of the answer.
	if len(cells) != (maxX-minX)+(maxY-minY)+1 {
		return fmt.Errorf("unable to apply answer %s to %s, incompatible sizes", answer, clue)
	}

	// Determine the way to iterate through the grid.
	var dx, dy int
	if direction == "a" {
		dx = 1
	} else {
		dy = 1
	}

	// Check to see if the answer is correct when required.
	if onlyCorrect {
		for x, y := minX, minY; x <= maxX && y <= maxY; x, y = x+dx, y+dy {
			existing := s.Cells[y][x]
			expected := s.Puzzle.Cells[y][x]
			desired := cells[y-minY+x-minX]

			// We can't change a correct value to an incorrect or empty one.
			if existing != "" && desired != existing {
				return fmt.Errorf("unable to apply answer %s to %s, changes correct value", answer, clue)
			}

			// We can't write an incorrect value into a cell.
			if desired != "" && desired != expected {
				return fmt.Errorf("unable to apply answer %s to %s, incorrect", answer, clue)
			}
		}
	}

	// Write the cells of our answer.
	for x, y := minX, minY; x <= maxX && y <= maxY; x, y = x+dx, y+dy {
		s.Cells[y][x] = cells[y-minY+x-minX]
	}

	// Now that we've filled in an answer we may have completed one or more clues.
	// Do a quick scan of all of the clues to make sure AcrossCluesFilled and
	// DownCluesFilled are up to date.
	err = s.UpdateFilledClues()
	if err != nil {
		return err
	}

	// Also determine if the puzzle is finished with all correct answers and
	// update the Status if so.
	complete := true
	for y := 0; y < s.Puzzle.Rows; y++ {
		for x := 0; x < s.Puzzle.Cols; x++ {
			if s.Cells[y][x] != s.Puzzle.Cells[y][x] {
				complete = false
			}
		}
	}
	if complete {
		s.Status = StatusComplete
	}

	// TODO: This method should probably also return information about whether or
	// not the answer was correct, and if so how many clues where completed as a
	// result of applying this answer.
	return nil
}

// ClearIncorrectCells will look at each filled in cell of the crossword and
// clear it if it is filled in with an incorrect answer.  The AcrossCluesFilled
// and DownCluesFilled fields will also be updated to indicate any clues that
// are now unanswered due to cleared cells.
func (s *State) ClearIncorrectCells() error {
	for y := 0; y < s.Puzzle.Rows; y++ {
		for x := 0; x < s.Puzzle.Cols; x++ {
			if s.Cells[y][x] != "" && s.Cells[y][x] != s.Puzzle.Cells[y][x] {
				s.Cells[y][x] = ""
			}
		}
	}

	// Now that we may have modified one or more cells we need to determine which
	// clues are answered and which aren't.
	return s.UpdateFilledClues()
}

// UpdateFilledClues looks at each clue in the puzzle and determines if a
// complete answer has been provided for the clue, if so then the corresponding
// entry in AcrossCluesFilled or DownCluesFilled will be set to true.  This
// method doesn't check that the provided answer is correct, just that one is
// present.
func (s *State) UpdateFilledClues() error {
	for num := range s.Puzzle.CluesAcross {
		minX, y, maxX, _, err := s.Puzzle.GetAnswerCoordinates(num, "a")
		if err != nil {
			return fmt.Errorf("somehow got invalid clue id %d from CluesAcross", num)
		}

		complete := true
		for x := minX; x <= maxX; x++ {
			if s.Cells[y][x] == "" {
				complete = false
				break
			}
		}

		s.AcrossCluesFilled[num] = complete
	}

	for num := range s.Puzzle.CluesDown {
		x, minY, _, maxY, err := s.Puzzle.GetAnswerCoordinates(num, "d")
		if err != nil {
			return fmt.Errorf("somehow got invalid clue id %d from CluesDown", num)
		}

		complete := true
		for y := minY; y <= maxY; y++ {
			if s.Cells[y][x] == "" {
				complete = false
				break
			}
		}

		s.DownCluesFilled[num] = complete
	}

	return nil
}

// ParseClue parses the identifier of a clue into its number and direction.
// If the clue cannot be parsed for some reason then an error will be returned.
func ParseClue(clue string) (int, string, error) {
	clue = strings.ToLower(strings.TrimSpace(clue))
	if len(clue) <= 1 {
		return 0, "", fmt.Errorf("unable to parse clue: %s", clue)
	}

	dir := clue[len(clue)-1:]
	if dir != "a" && dir != "d" {
		return 0, "", fmt.Errorf("unable to parse clue: %s", clue)
	}

	num, err := strconv.Atoi(clue[:len(clue)-1])
	if err != nil {
		return 0, "", fmt.Errorf("unable to parse clue: %s", clue)
	}

	return num, dir, nil
}

// ParseAnswer parses an answer string into a list of cell values.  The answer
// string is parsed in such a way to look for cell values that contain multiple
// characters (aka a rebus).  It does this by looking for parenthesized groups
// of letters.  For example the string "(red)velvet" would be interpreted as
// ["red", "v", "e", "l", "v", "e", "t"] and fit as the answer for a 7 cell
// clue.
//
// Additionally if an answer contains a "." character anywhere that particular
// cell will be left empty.  This allows strings like "....s" to be entered to
// indicate that the answer is known to be plural, but the other letters aren't
// known yet.  Within a rebus cell "." characters are kept as-is.
//
// Whitespace within answers is removed and ignored.  This makes it more natural
// to specify answers like "red velvet cake".
func ParseAnswer(answer string) ([]string, error) {
	var cells []string
	var inside bool

	for _, c := range strings.ToUpper(answer) {
		switch {
		case c == ' ':
			continue

		case c == '(':
			if inside {
				return nil, fmt.Errorf("unable to parse answer, nested groups: %s", answer)
			}
			inside = true
			cells = append(cells, "")

		case c == ')':
			if !inside {
				return nil, fmt.Errorf("unable to parse answer, unbalanced parens: %s", answer)
			}
			inside = false

		case inside:
			if len(cells) != 0 {
				cells[len(cells)-1] = cells[len(cells)-1] + string(c)
			}

		default:
			if c != '.' {
				cells = append(cells, string(c))
			} else {
				cells = append(cells, "")
			}
		}
	}

	if inside {
		return nil, fmt.Errorf("unable to parse answer, unbalanced parens: %s", answer)
	}

	if len(cells) == 0 {
		return nil, fmt.Errorf("unable to parse answer, empty cells: %s", answer)
	}

	return cells, nil
}

// The status of a channel's crossword solve.
type Status int

const (
	// The channel has been created but a puzzle hasn't yet been selected.
	StatusCreated Status = iota

	// The channel has a puzzle selected but is paused in its attempt to solve it.
	StatusPaused

	// The channel is actively trying to solve the puzzle.
	StatusSolving

	// The puzzle that was being solved is complete.
	StatusComplete
)

func (s Status) String() string {
	switch s {
	case StatusComplete:
		return "complete"
	case StatusCreated:
		return "created"
	case StatusPaused:
		return "paused"
	case StatusSolving:
		return "solving"
	default:
		return "unknown"
	}
}

func (s Status) MarshalJSON() ([]byte, error) {
	var ok bool
	switch s {
	case StatusComplete:
		ok = true
	case StatusCreated:
		ok = true
	case StatusPaused:
		ok = true
	case StatusSolving:
		ok = true
	}

	if !ok {
		return nil, fmt.Errorf("unable to marshal invalid channel state: %v", s)
	}

	return json.Marshal(s.String())
}

func (s *Status) UnmarshalJSON(bs []byte) error {
	var str string
	if err := json.Unmarshal(bs, &str); err != nil {
		return err
	}

	switch str {
	case "complete":
		*s = StatusComplete
	case "created":
		*s = StatusCreated
	case "paused":
		*s = StatusPaused
	case "solving":
		*s = StatusSolving
	default:
		return fmt.Errorf("unable to unmarshal invalid channel state: %s", str)
	}

	return nil
}

// Duration is an aliasing of time.Duration that supports marshalling to/from
// JSON.
type Duration struct {
	time.Duration
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(bs []byte) error {
	var s string
	if err := json.Unmarshal(bs, &s); err != nil {
		return err
	}

	td, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	*d = Duration{td}
	return nil
}
