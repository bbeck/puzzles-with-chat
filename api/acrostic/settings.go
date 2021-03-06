package acrostic

import (
	"fmt"
	"github.com/bbeck/puzzles-with-chat/api/db"
	"github.com/bbeck/puzzles-with-chat/api/model"
	"github.com/gomodule/redigo/redis"
)

// Settings represents the optional behaviors that can be enabled or disabled
// by a streamer for their channel's acrostic solves.
type Settings struct {
	// When enabled only correct answers will be filled into the puzzle grid.
	OnlyAllowCorrectAnswers bool `json:"only_allow_correct_answers"`

	// What font size should the clues be rendered with.
	ClueFontSize model.FontSize `json:"clue_font_size"`
}

// SettingsKey returns the key that should be used in redis to store a
// particular channel's acrostic settings.
func SettingsKey(name string) string {
	return fmt.Sprintf("%s:acrostic:settings", name)
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
