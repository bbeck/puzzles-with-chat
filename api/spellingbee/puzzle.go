package spellingbee

import "time"

// Puzzle represents a spelling bee puzzle.  The puzzle is comprised of a
// circular grid of 6 letters around a single letter.  The goal is to use the
// letters to make words that all use the center letter and have length 4 or
// greater.  Letters may be reused.
type Puzzle struct {
	// The date that the spelling bee puzzle was published.
	PublishedDate time.Time `json:"published"`

	// The center letter of the spelling bee puzzle.  This letter must be used in
	// every answer.
	CenterLetter string `json:"center"`

	// The non-center letters in the spelling bee puzzle.  Each entry will
	// always be a single letter long and there will always be 6 entries.
	Letters []string `json:"letters"`

	// The list of official answers from The New York Times.
	OfficialAnswers []string `json:"official_answers,omitempty"`

	// The list of unofficial answers from NYTBee.com.
	UnofficialAnswers []string `json:"unofficial_answers,omitempty"`
}

// WithoutAnswers returns a copy of the puzzle that has the answers removed.
// This makes the resulting puzzle suitable to pass to a client that shouldn't
// know the answers to the puzzle.
func (p *Puzzle) WithoutAnswers() *Puzzle {
	var puzzle Puzzle
	puzzle.PublishedDate = p.PublishedDate
	puzzle.CenterLetter = p.CenterLetter
	puzzle.Letters = p.Letters
	puzzle.OfficialAnswers = nil
	puzzle.UnofficialAnswers = nil

	return &puzzle
}
