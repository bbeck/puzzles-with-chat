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

	// The total number of points possible in the puzzle (only official answers).
	MaximumScore int `json:"max_score"`

	// The total number of official answers.
	NumOfficialAnswers int `json:"num_official_answers"`

	// The total number of unofficial answers (not including the official ones).
	NumUnofficialAnswers int `json:"num_unofficial_answers"`
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
	puzzle.MaximumScore = p.MaximumScore
	puzzle.NumOfficialAnswers = p.NumOfficialAnswers
	puzzle.NumUnofficialAnswers = p.NumUnofficialAnswers

	return &puzzle
}

// ComputeScore calculates the score for the provided words taken together. No
// checking is done to make sure the words are valid answers, they're all
// assumed to be correct.
func (p *Puzzle) ComputeScore(words []string) int {
	isPangram := func(word string) bool {
		letters := map[string]struct{}{
			p.CenterLetter: {},
		}
		for _, letter := range p.Letters {
			letters[letter] = struct{}{}
		}

		for _, letter := range word {
			delete(letters, string(letter))
		}

		return len(letters) == 0
	}

	var score int
	for _, word := range words {
		if len(word) == 4 {
			score += 1
			continue
		}

		score += len(word)

		// pangrams get a 7 point bonus
		if isPangram(word) {
			score += 7
		}
	}

	return score
}
