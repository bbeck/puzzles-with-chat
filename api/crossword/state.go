package crossword

import (
	"fmt"
	"github.com/bbeck/puzzles-with-chat/api/db"
	"github.com/bbeck/puzzles-with-chat/api/model"
	"sort"
	"strconv"
	"strings"
	"time"
)

// State represents the state of an active channel that is attempting to solve
// a crossword.
type State struct {
	// The status of the channel's crossword solve.
	Status model.Status `json:"status"`

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
	TotalSolveDuration model.Duration `json:"total_solve_duration"`
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
		s.Status = model.StatusComplete
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

// StateKey returns the key that should be used in redis to store a particular
// crossword solve's state.
func StateKey(name string) string {
	return fmt.Sprintf("%s:crossword:state", name)
}

// StateTTL determines how long a particular crossword's solve state should
// remain in redis in the absence of any activity.
var StateTTL = 4 * time.Hour

// GetState loads the state for a crossword solve from redis.  If the state
// can't be loaded then an error will be returned.  If there is no state, then
// the zero value will be returned.  After a state is read, its expiration time
// is automatically updated.
func GetState(conn db.Connection, channel string) (State, error) {
	var state State

	if testStateLoadError != nil {
		return state, testStateLoadError
	}

	err := db.Get(conn, StateKey(channel), &state)
	return state, err
}

// SetState writes the state for a channel's crossword solve to redis.  If the
// state can't be property written then an error will be returned.
func SetState(conn db.Connection, channel string, state State) error {
	if testStateSaveError != nil {
		return testStateSaveError
	}

	return db.SetWithTTL(conn, StateKey(channel), state, StateTTL)
}

// GetAllChannels returns a slice of model.Channel instances for each crossword
// that contains state in the database.  If there are no active channels then an
// empty slice is returned.  This method does not update the expiration times
// of any state instance.
func GetAllChannels(conn db.Connection) ([]model.Channel, error) {
	keys, err := db.ScanKeys(conn, StateKey("*"))
	if err != nil {
		return nil, err
	}

	values, err := db.GetAll(conn, keys, State{})
	if err != nil {
		return nil, err
	}

	channels := make([]model.Channel, 0)
	for key, value := range values {
		name := strings.Replace(key, StateKey(""), "", 1)

		state, ok := value.(State)
		if !ok {
			return nil, fmt.Errorf("unable to convert value to State: %v", value)
		}

		var description string
		if state.Puzzle != nil {
			description = state.Puzzle.Description
		}

		channels = append(channels, model.Channel{
			Name:        name,
			Status:      state.Status,
			Description: description,
		})
	}

	sort.Slice(channels, func(i, j int) bool {
		return channels[i].Name < channels[j].Name
	})

	return channels, nil
}
