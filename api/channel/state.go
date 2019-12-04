package channel

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/bbeck/twitch-plays-crosswords/api/crossword"
)

// State represents the state of an active channel that is attempting to solve
// a crossword.
type State struct {
	// The status of the channel's crossword solve.
	Status Status `json:"status"`

	// The crossword puzzle that's being solved.  May not always be present, for
	// example when the state is being serialized to be sent to the browser.
	Puzzle *crossword.Puzzle `json:"puzzle,omitempty"`

	// The currently filled in cells of the crossword.
	Cells [][]string `json:"cells,omitempty"`

	// Whether or not an across clue with a given clue number has had an answer
	// filled in.
	AcrossCluesFilled map[int]bool `json:"across_clues_filled,omitempty"`

	// Whether or not a down clue with a given clue number has had an answer
	// filled in.
	DownCluesFilled map[int]bool `json:"down_clues_filled,omitempty"`

	// The time that we last started or resumed solving the puzzle.  If the
	// channel has not yet started solving the puzzle or is in a non-playing state
	// this will be nil.
	LastStartTime *time.Time `json:"last_start_time,omitempty"`

	// The total time spent on solving the puzzle up to the last start time.
	TotalSolveDuration Duration `json:"total_solve_duration"`
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
