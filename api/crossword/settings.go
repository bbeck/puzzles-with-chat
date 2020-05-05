package crossword

import (
	"encoding/json"
	"fmt"
	"github.com/bbeck/puzzles-with-chat/api/db"
	"github.com/bbeck/puzzles-with-chat/api/model"
	"github.com/gomodule/redigo/redis"
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
	ClueFontSize model.FontSize `json:"clue_font_size"`

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

// SettingsKey returns the key that should be used in redis to store a
// particular channel's settings.
func SettingsKey(name string) string {
	return fmt.Sprintf("%s:crossword:settings", name)
}

// GetSettings will load settings for the provided channel name.  If the
// settings can't be properly loaded then an error will be returned.
func GetSettings(conn redis.Conn, channel string) (Settings, error) {
	var settings Settings

	if testSettingsLoadError != nil {
		return settings, testSettingsLoadError
	}

	err := db.Get(conn, SettingsKey(channel), &settings)
	return settings, err
}

// SetSettings will write settings for the provided channel name.  If the
// settings can't be properly written then an error will be returned.
func SetSettings(conn redis.Conn, channel string, settings Settings) error {
	if testSettingsSaveError != nil {
		return testSettingsSaveError
	}

	return db.Set(conn, SettingsKey(channel), settings)
}
