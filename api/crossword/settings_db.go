package crossword

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gomodule/redigo/redis"
)

// SettingsKey returns the key that should be used in redis to store a
// particular channel's settings.
func SettingsKey(name string) string {
	return fmt.Sprintf("%s:crossword:settings", name)
}

// GetSettings will load settings for the provided channel name.  If the
// settings can't be properly loaded then an error will be returned.
func GetSettings(c redis.Conn, name string) (*Settings, error) {
	bs, err := redis.Bytes(c.Do("GET", SettingsKey(name)))
	if err == redis.ErrNil {
		// There weren't any settings in redis, this is expected when a channel is
		// new and hasn't solved a puzzle yet.  When this happens we'll return the
		// zero value which are the default settings.
		return &Settings{}, nil
	}
	if err != nil {
		return nil, err
	}

	var settings Settings
	return &settings, json.Unmarshal(bs, &settings)
}

// SetSettings will write settings for the provided channel name.  If the
// settings can't be properly written then an error will be returned.
func SetSettings(c redis.Conn, name string, settings *Settings) error {
	if settings == nil {
		return errors.New("cannot save nil settings")
	}

	bs, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	_, err = c.Do("SET", SettingsKey(name), bs)
	return err
}
