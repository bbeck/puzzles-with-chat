package channel

import (
	"encoding/json"
	"fmt"

	"github.com/gomodule/redigo/redis"
)

// GetSettings will load settings for the provided channel name.  If the
// settings can't be properly loaded then an error will be returned.
func GetSettings(c redis.Conn, name string) (Settings, error) {
	var settings Settings

	key := fmt.Sprintf("settings:%s", name)
	bs, err := redis.Bytes(c.Do("GET", key))
	if err == nil {
		err = json.Unmarshal(bs, &settings)
	} else if err == redis.ErrNil {
		// There weren't any settings in redis, this is expected when a channel is
		// new and hasn't solved a puzzle yet.  When this happens we'll use the
		// default values for the settings.
		err = nil
	}

	return settings, err
}

// SetSettings will write settings for the provided channel name.  If the
// settings can't be properly written then an error will be returned.
func SetSettings(c redis.Conn, name string, settings Settings) error {
	bs, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("settings:%s", name)
	_, err = c.Do("SET", key, bs)
	return err
}
