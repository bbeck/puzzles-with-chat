package model

// Channel is a representation of a channel and the status of it in solving a
// puzzle.  It can be marshalled to/from JSON.
type Channel struct {
	Name   string `json:"name"`
	Status Status `json:"status"`
}
