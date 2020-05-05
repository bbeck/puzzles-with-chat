package spellingbee

import (
	"fmt"
	"github.com/bbeck/puzzles-with-chat/api/db"
	"github.com/bbeck/puzzles-with-chat/api/model"
	"github.com/gomodule/redigo/redis"
)

// Settings represents the optional behaviors that can be enabled or disabled
// by a streamer for their channel's spelling bee solves.
type Settings struct {
	// When enabled unofficial answers will be allowed.
	AllowUnofficialAnswers bool `json:"allow_unofficial_answers"`

	// What font size words should be rendered with.
	FontSize model.FontSize `json:"font_size"`
}

// SettingsKey returns the key that should be used in redis to store a
// particular channel's settings.
func SettingsKey(name string) string {
	return fmt.Sprintf("%s:spellingbee:settings", name)
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
