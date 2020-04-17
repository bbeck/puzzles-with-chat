package spellingbee

import (
	"fmt"
	"github.com/bbeck/twitch-plays-crosswords/api/db"
	"github.com/bbeck/twitch-plays-crosswords/api/model"
	"github.com/gomodule/redigo/redis"
	"strings"
	"time"
)

// State represents the state of an active channel that is attempting to solve
// a spelling bee puzzle.
type State struct {
	// The status of the channel's solve.
	Status model.Status `json:"status"`

	// The spelling bee puzzle that's being solved.  May not always be present,
	// for example when the state is being serialized to be sent to the browser.
	Puzzle *Puzzle `json:"puzzle,omitempty"`

	// The currently discovered words the puzzle.
	Words []string `json:"words"`

	// The time that we last started or resumed solving the puzzle.  If the
	// channel has not yet started solving the puzzle or is in a non-playing state
	// this will be nil.
	LastStartTime *time.Time `json:"last_start_time,omitempty"`

	// The total time spent on solving the puzzle up to the last start time.
	TotalSolveDuration model.Duration `json:"total_solve_duration"`
}

// StateKey returns the key that should be used in redis to store a particular
// spelling bee solve's state.
func StateKey(name string) string {
	return fmt.Sprintf("%s:spellingbee:state", name)
}

// StateTTL determines how long a particular crossword's solve state should
// remain in redis in the absence of any activity.
var StateTTL = 4 * time.Hour

// GetState loads the state for a spelling bee solve from redis.  If the state
// can't be loaded then an error will be returned.  If there is no state, then
// the zero value will be returned.  After a state is read, its expiration time
// is automatically updated.
func GetState(conn redis.Conn, channel string) (State, error) {
	var state State
	err := db.GetWithTTLRefresh(conn, StateKey(channel), &state, StateTTL)
	return state, err
}

// SetState writes the state for a channel's spelling bee solve to redis.  If
// the state can't be property written then an error will be returned.
func SetState(conn redis.Conn, channel string, state State) error {
	return db.SetWithTTL(conn, StateKey(channel), state, StateTTL)
}

// GetChannelNamesWithState loads the name of all channels that currently have
// a state present in redis.  If there are no channels then an empty slice is
// returned.  This method does not update the expiration times of any state.
func GetChannelNamesWithState(conn redis.Conn) ([]string, error) {
	keys, err := db.ScanKeys(conn, StateKey("*"))

	channels := make([]string, 0)
	for _, key := range keys {
		channels = append(channels, strings.Replace(key, StateKey(""), "", 1))
	}

	return channels, err
}
