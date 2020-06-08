package acrostic

import (
	"fmt"
	"github.com/bbeck/puzzles-with-chat/api/db"
	"github.com/bbeck/puzzles-with-chat/api/model"
	"sort"
	"strings"
	"time"
)

// State represents the state of an active channel that is attempting to solve
// an acrostic.
type State struct {
	// The status of the channel's acrostic solve.
	Status model.Status `json:"status"`

	// The acrostic puzzle that's being solved.  May not always be present, for
	// example when the state is being serialized to be sent to the browser.
	Puzzle *Puzzle `json:"puzzle,omitempty"`

	// The currently filled in cells of the acrostic.
	Cells [][]string `json:"cells"`

	// Whether or not a clue with a given clue letter has had an answer filled in.
	CluesFilled map[string]bool `json:"clues_filled"`

	// The time that we last started or resumed solving the puzzle.  If the
	// channel has not yet started solving the puzzle or is in a non-playing state
	// this will be nil.
	LastStartTime *time.Time `json:"last_start_time,omitempty"`

	// The total time spent on solving the puzzle up to the last start time.
	TotalSolveDuration model.Duration `json:"total_solve_duration"`
}

// ApplyClueAnswer applies an answer for a clue to the state.  If the clue
// cannot be identified or the answer doesn't fit property (too short or too
// long) then an error will be returned.  If the onlyCorrect parameter is true
// then only correct values will be permitted and an error is returned if any
// part of the answer is incorrect or would remove a correct cell.
func (s *State) ApplyClueAnswer(clue string, answer string, onlyCorrect bool) error {
	clue = strings.ToUpper(clue)
	nums, ok := s.Puzzle.ClueNumbers[clue]
	if !ok {
		return fmt.Errorf("invalid clue identifier: %s", clue)
	}

	// Ignore spaces within the answer and ensure the answer is all uppercase.
	answer = strings.ReplaceAll(answer, " ", "")
	answer = strings.ToUpper(answer)

	// Ensure that we have a proper length answer
	if len(nums) != len(answer) {
		return fmt.Errorf("unable to apply answer %s to clue %s, incompatible sizes", answer, clue)
	}

	// Cache the coordinates of each cell of the answer.
	xs := make(map[int]int)
	ys := make(map[int]int)
	for _, num := range nums {
		x, y, err := s.Puzzle.GetCellCoordinates(num)
		if err != nil {
			return err
		}

		xs[num] = x
		ys[num] = y
	}

	// Check to see if the answer is correct when required.
	if onlyCorrect {
		for i, num := range nums {
			x := xs[num]
			y := ys[num]

			existing := s.Cells[y][x]
			expected := s.Puzzle.Cells[y][x]
			desired := string(answer[i])

			// We can't change a correct value to an incorrect or empty one.
			if existing != "" && desired != existing {
				return fmt.Errorf("unable to apply answer %s to clue %s, changes correct value", answer, clue)
			}

			// We can't write an incorrect value into a cell
			if desired != "." && desired != expected {
				return fmt.Errorf("unable to apply answer %s to clue %s, incorrect", answer, clue)
			}
		}
	}

	// Apply the answer to the cells.
	filled := true
	for i, num := range nums {
		x := xs[num]
		y := ys[num]

		if answer[i] != '.' {
			s.Cells[y][x] = string(answer[i])
		} else {
			s.Cells[y][x] = ""
			filled = false
		}
	}

	// Update whether or not the clue was filled.
	s.CluesFilled[clue] = filled

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

	return nil
}

// ApplyCellAnswer applies an answer to the cells to the state.  If the starting
// cell number specified outside the bounds of the puzzle or the answer doesn't
// fit within the bounds of the puzzle then an error will be returned.  If the
// onlyCorrect parameter is true then only correct values will be permitted and
// an error is returned if any part of the answer is incorrect or would remove a
// correct cell.
func (s *State) ApplyCellAnswer(start int, answer string, onlyCorrect bool) error {
	if start <= 0 {
		return fmt.Errorf("invalid starting index: %d", start)
	}

	// Ignore spaces within the answer and ensure the answer is all uppercase.
	answer = strings.ReplaceAll(answer, " ", "")
	answer = strings.ToUpper(answer)

	// Ensure that we have a non-empty answer.
	if len(answer) == 0 {
		return fmt.Errorf("empty answer")
	}

	// Compute the coordinates of each letter of the answer.  We do this ahead of
	// time and not in a loop while setting cell values because we may discover
	// an error in identifying a cell's coordinate.
	xs := make([]int, len(answer))
	ys := make([]int, len(answer))
	for i := 0; i < len(answer); i++ {
		x, y, err := s.Puzzle.GetCellCoordinates(start + i)
		if err != nil {
			return err
		}

		xs[i] = x
		ys[i] = y
	}

	// Check to see if the answer is correct when required.
	if onlyCorrect {
		for i := 0; i < len(answer); i++ {
			x := xs[i]
			y := ys[i]

			existing := s.Cells[y][x]
			expected := s.Puzzle.Cells[y][x]
			desired := string(answer[i])

			// We can't change a correct value to an incorrect or empty one.
			if existing != "" && desired != existing {
				return fmt.Errorf("unable to apply answer %s starting at index %d, changes correct value", answer, start)
			}

			// We can't write an incorrect value into a cell
			if desired != "." && desired != expected {
				return fmt.Errorf("unable to apply answer %s starting at index %d, incorrect", answer, start)
			}
		}
	}

	// Apply the answer.
	for i := 0; i < len(answer); i++ {
		if answer[i] != '.' {
			s.Cells[ys[i]][xs[i]] = string(answer[i])
		} else {
			s.Cells[ys[i]][xs[i]] = ""
		}
	}

	// Update which clues have been filled.
	if err := s.UpdateFilledClues(); err != nil {
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

	return nil
}

// UpdateFilledClues looks at each clue in the puzzle and determines if a
// complete answer has been provided for the clue, if so then the corresponding
// entry in CluesFilled will be set to true.  This method doesn't check that the
// provided answer is correct, just that one is present.
func (s *State) UpdateFilledClues() error {
	for clue, nums := range s.Puzzle.ClueNumbers {
		complete := true
		for _, num := range nums {
			x, y, err := s.Puzzle.GetCellCoordinates(num)
			if err != nil {
				return err
			}

			if s.Cells[y][x] == "" {
				complete = false
				break
			}
		}

		s.CluesFilled[clue] = complete
	}

	return nil
}

// ClearIncorrectCells will look at each filled in cell of the acrostic and
// clear it if it is filled in with an incorrect answer.  The CluesFilled
// field will also be updated to indicate any clues that are now unanswered due
// to cleared cells.
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

// GetAllChannels returns a slice of model.Channel instances for each acrostic
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

// StateKey returns the key that should be used in redis to store a particular
// crossword solve's state.
func StateKey(name string) string {
	return fmt.Sprintf("%s:acrostic:state", name)
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
