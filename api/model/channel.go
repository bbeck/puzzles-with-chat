package model

// Channel is a representation of a channel and the puzzle that is being solved.
// It can be marshalled to/from JSON.
type Channel struct {
	Name        string `json:"name"`
	Status      Status `json:"status"`
	Description string `json:"description,omitempty"`
}
