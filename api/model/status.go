package model

import (
	"encoding/json"
	"fmt"
)

// Status is an enumeration representing the supported statuses a channel can
// be in with respect to solving a puzzle.  It can be marshalled to/from JSON as
// well as implements the fmt.Stringer interface for human readability.
type Status int

const (
	// The channel exists, but a puzzle hasn't yet been selected.
	StatusCreated Status = iota

	// A puzzle has been selected, but not yet started.
	StatusSelected

	// The channel has a puzzle selected but is paused in its attempt to solve it.
	StatusPaused

	// The channel is actively trying to solve the puzzle.
	StatusSolving

	// The puzzle that was being solved is complete.
	StatusComplete
)

func (s Status) String() string {
	switch s {
	case StatusCreated:
		return "created"
	case StatusSelected:
		return "selected"
	case StatusPaused:
		return "paused"
	case StatusSolving:
		return "solving"
	case StatusComplete:
		return "complete"
	default:
		return "unknown"
	}
}

func (s Status) MarshalJSON() ([]byte, error) {
	switch s {
	case StatusCreated:
	case StatusSelected:
	case StatusPaused:
	case StatusSolving:
	case StatusComplete:
	default:
		return nil, fmt.Errorf("unrecognized status: %v", s)
	}

	return json.Marshal(s.String())
}

func (s *Status) UnmarshalJSON(bs []byte) error {
	var str string
	if err := json.Unmarshal(bs, &str); err != nil {
		return err
	}

	switch str {
	case "created":
		*s = StatusCreated
	case "selected":
		*s = StatusSelected
	case "paused":
		*s = StatusPaused
	case "solving":
		*s = StatusSolving
	case "complete":
		*s = StatusComplete
	default:
		return fmt.Errorf("unrecognized status string: %s", str)
	}

	return nil
}
