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

	// The current order of the letters of the puzzle, not including the first
	// letter.
	Letters []string `json:"letters"`

	// The currently discovered words the puzzle.
	Words []string `json:"words"`

	// The current score of the solve.
	Score int `json:"score"`

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

	// Update the score for this answer.
	s.updateScore()

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

	// Since we've modified the words slice we should update the answer.
	s.updateScore()
}

// UpdateScore updates the score for the puzzle based on all of the answers
// that have been provided.
func (s *State) updateScore() {
	var score int
	for _, word := range s.Words {
		if len(word) == 4 {
			score += 1
			continue
		}

		score += len(word)
	}

	s.Score = score
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

// GetAllChannels returns a slice of model.Channel instances for each spelling
// bee that contains state in the database.  If there are no active channels
// then an empty slice is returned.  This method does not update the expiration
// times of any state instance.
func GetAllChannels(conn redis.Conn) ([]model.Channel, error) {
	keys, err := db.ScanKeys(conn, StateKey("*"))
	if err != nil {
		return nil, err
	}

	values, err := db.GetAll(conn, keys, State{})
	if err != nil {
		return nil, err
	}

	channels := make([]model.Channel, 0)
	for key, value := range values {
		name := strings.Replace(key, StateKey(""), "", 1)

		state, ok := value.(State)
		if !ok {
			return nil, fmt.Errorf("unable to convert value to State: %v", value)
		}

		channels = append(channels, model.Channel{
			Name:   name,
			Status: state.Status,
		})
	}

	return channels, nil
}
