package crossword

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

var StateTTL = 4 * time.Hour

// StateKey returns the key that should be used in redis to store a particular
// crossword solve's state.
func StateKey(name string) string {
	return fmt.Sprintf("%s:crossword:state", name)
}

// GetState loads the state for a crossword solve from redis.  If the state
// can't be loaded then an error will be returned.  If there is no state, then
// the zero value will be returned.  After a state is read, its expiration time
// is automatically updated.
func GetState(c redis.Conn, name string) (*State, error) {
	var state State

	// We could create a transaction to read the key and update its expiration
	// time atomically, but that's not really necessary.  The worst that will
	// happen is that we race to update the expiration time and in either case
	// last one wins which is what we want to happen.
	bs, err := redis.Bytes(c.Do("GET", StateKey(name)))
	if err == redis.ErrNil {
		// There wasn't a state in redis, this is expected prior to a puzzle being
		// selected.  When this happens we'll use the zero value of the state and
		// don't need to update the expiration time.
		return &state, nil
	}
	if err != nil {
		return nil, err
	}

	_, err = c.Do("EXPIRE", StateKey(name), StateTTL.Seconds())
	if err != nil {
		return nil, err
	}

	return &state, json.Unmarshal(bs, &state)
}

// SetState writes the state for a channel's crossword solve to redis.  If the
// settings can't be property written then an error will be returned.
func SetState(c redis.Conn, name string, state *State) error {
	if state == nil {
		return errors.New("cannot save nil state")
	}

	bs, err := json.Marshal(state)
	if err != nil {
		return err
	}

	_, err = c.Do("SETEX", StateKey(name), StateTTL.Seconds(), bs)
	return err
}
