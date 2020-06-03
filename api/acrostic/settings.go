package acrostic

import "github.com/bbeck/puzzles-with-chat/api/model"

// Settings represents the optional behaviors that can be enabled or disabled
// by a streamer for their channel's acrostic solves.
type Settings struct {
	// When enabled only correct answers will be filled into the puzzle grid.
	OnlyAllowCorrectAnswers bool `json:"only_allow_correct_answers"`

	// What font size should the clues be rendered with.
	ClueFontSize model.FontSize `json:"clue_font_size"`
}
