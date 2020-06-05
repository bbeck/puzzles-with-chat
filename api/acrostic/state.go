package acrostic

import (
	"fmt"
	"github.com/bbeck/puzzles-with-chat/api/model"
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
