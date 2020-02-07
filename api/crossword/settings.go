package crossword

import (
	"encoding/json"
	"fmt"
)

// Settings represents the optional behaviors that can be enabled or disabled
// by a streamer for their channel's crossword solves.
type Settings struct {
	// When enabled only correct answers will be filled into the puzzle grid.
	OnlyAllowCorrectAnswers bool `json:"only_allow_correct_answers"`

	// Which clues should be shown to users.  Can be all of the clues, none of the
	// clues, only across clues or only down clues.
	CluesToShow ClueVisibility `json:"clues_to_show"`

	// What font size should the clues be rendered with.
	ClueFontSize FontSize `json:"clue_font_size"`

	// Whether or not notes field should shown.
	ShowNotes bool `json:"show_notes"`
}

// ClueVisibility is an enumeration representing which clues should be shown.
type ClueVisibility int

const (
	AllCluesVisible ClueVisibility = iota
	NoCluesVisible
	OnlyDownCluesVisible
	OnlyAcrossCluesVisible
)

func (v ClueVisibility) String() string {
	switch v {
	case AllCluesVisible:
		return "all"
	case NoCluesVisible:
		return "none"
	case OnlyDownCluesVisible:
		return "down"
	case OnlyAcrossCluesVisible:
		return "across"
	default:
		return "unknown"
	}
}

func (v ClueVisibility) MarshalJSON() ([]byte, error) {
	var ok bool
	switch v {
	case AllCluesVisible:
		ok = true
	case NoCluesVisible:
		ok = true
	case OnlyDownCluesVisible:
		ok = true
	case OnlyAcrossCluesVisible:
		ok = true
	}

	if !ok {
		return nil, fmt.Errorf("unable to marshal invalid clue visibility: %v", v)
	}

	return json.Marshal(v.String())
}

func (v *ClueVisibility) UnmarshalJSON(bs []byte) error {
	var str string
	if err := json.Unmarshal(bs, &str); err != nil {
		return err
	}

	switch str {
	case "all":
		*v = AllCluesVisible
	case "none":
		*v = NoCluesVisible
	case "down":
		*v = OnlyDownCluesVisible
	case "across":
		*v = OnlyAcrossCluesVisible
	default:
		return fmt.Errorf("unable to unmarshal invalid clue visibility: %s", str)
	}

	return nil
}

// FontSize is an enumeration representing the supported sizes of the clue font.
type FontSize int

const (
	SizeNormal FontSize = iota
	SizeLarge
	SizeXLarge
)

func (s FontSize) String() string {
	switch s {
	case SizeNormal:
		return "normal"
	case SizeLarge:
		return "large"
	case SizeXLarge:
		return "xlarge"
	default:
		return "unknown"
	}
}

func (s FontSize) MarshalJSON() ([]byte, error) {
	var ok bool
	switch s {
	case SizeNormal:
		ok = true
	case SizeLarge:
		ok = true
	case SizeXLarge:
		ok = true
	}

	if !ok {
		return nil, fmt.Errorf("unable to marshal invalid clue size: %v", s)
	}

	return json.Marshal(s.String())
}

func (s *FontSize) UnmarshalJSON(bs []byte) error {
	var str string
	if err := json.Unmarshal(bs, &str); err != nil {
		return err
	}

	switch str {
	case "normal":
		*s = SizeNormal
	case "large":
		*s = SizeLarge
	case "xlarge":
		*s = SizeXLarge
	default:
		return fmt.Errorf("unable to unmarshal invalid clue size: %s", str)
	}

	return nil
}
