package spellingbee

import (
	"errors"
	"fmt"
	"github.com/bbeck/twitch-plays-crosswords/api/db"
	"github.com/bbeck/twitch-plays-crosswords/api/model"
	"github.com/gomodule/redigo/redis"
	"sort"
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

// ApplyAnswer applies an answer to the state.  If the answer cannot be applied
// or is incorrect then an error is returned.
func (s *State) ApplyAnswer(answer string, allowUnofficial bool) error {
	allowed := map[string]bool{
		s.Puzzle.CenterLetter: true,
	}
	for _, letter := range s.Puzzle.Letters {
		allowed[letter] = true
	}

	contains := func(words []string, word string) bool {
		index := sort.SearchStrings(words, word)
		return index < len(words) && words[index] == word
	}

	// First, ensure the answer uses the proper letters.
	answer = strings.ToUpper(answer)
	for _, letter := range answer {
		if !allowed[string(letter)] {
			return fmt.Errorf("answer contains letter not in puzzle: %c", letter)
		}
	}

	// Next, make sure the answer wasn't previously given.
	if contains(s.Words, answer) {
		return errors.New("answer already given")
	}

	// Next, ensure the answer is in the list of allowed answers.
	var allAnswers []string
	allAnswers = append(allAnswers, s.Puzzle.OfficialAnswers...)
	if allowUnofficial {
		allAnswers = append(allAnswers, s.Puzzle.UnofficialAnswers...)
	}
	sort.Strings(allAnswers)

	if !contains(allAnswers, answer) {
		return errors.New("answer not in the list of allowed answers")
	}

	// Save the answer to the state and ensure they remain sorted.
	s.Words = append(s.Words, answer)
	sort.Strings(s.Words)

	// Lastly determine if we've found all of the answers and the puzzle is now
	// complete.
	if len(s.Words) == len(allAnswers) {
		s.Status = model.StatusComplete
	}

	return nil
}

// ClearUnofficialAnswers goes through all of the provided answers for a puzzle
// and removes any that are on the unofficial answers list.
func (s *State) ClearUnofficialAnswers() {
	contains := func(words []string, word string) bool {
		index := sort.SearchStrings(words, word)
		return index < len(words) && words[index] == word
	}

	var updatedWords []string
	for _, word := range s.Words {
		if !contains(s.Puzzle.UnofficialAnswers, word) {
			updatedWords = append(updatedWords, word)
		}
	}

	// Shouldn't need to re-sort because they were already in sorted order.
	s.Words = updatedWords

	// TODO: In the future this will have to recompute the score
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

	if testStateLoadError != nil {
		return state, testStateLoadError
	}

	err := db.GetWithTTLRefresh(conn, StateKey(channel), &state, StateTTL)
	return state, err
}

// SetState writes the state for a channel's spelling bee solve to redis.  If
// the state can't be property written then an error will be returned.
func SetState(conn redis.Conn, channel string, state State) error {
	if testStateSaveError != nil {
		return testStateSaveError
	}

	return db.SetWithTTL(conn, StateKey(channel), state, StateTTL)
}

// GetChannelNamesWithState loads the name of all channels that currently have
// a state present in redis.  If there are no channels then an empty slice is
// returned.  This method does not update the expiration times of any state.
func GetChannelNamesWithState(conn redis.Conn) ([]string, error) {
	channels := make([]string, 0)

	if testChannelNamesLoadError != nil {
		return channels, testChannelNamesLoadError
	}

	keys, err := db.ScanKeys(conn, StateKey("*"))
	for _, key := range keys {
		channels = append(channels, strings.Replace(key, StateKey(""), "", 1))
	}

	return channels, err
}
