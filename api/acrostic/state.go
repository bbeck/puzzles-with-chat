package acrostic

import (
	"github.com/bbeck/puzzles-with-chat/api/model"
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
