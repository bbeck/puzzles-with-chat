package model

import "time"

// Channel is a representation of a channel and the puzzle that is being solved.
// It can be marshalled to/from JSON.
type Channel struct {
	Name        string       `json:"name"`
	Status      Status       `json:"status"`
	Description string       `json:"description,omitempty"`
	Puzzle      PuzzleSource `json:"puzzle"`
}

// PuzzleSource is a representation of the source of a puzzle that's being
// solved.  It can be marshalled to/from JSON.
type PuzzleSource struct {
	Publisher     string    `json:"publisher"`
	PublishedDate time.Time `json:"published"`
}
