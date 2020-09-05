package spellingbee

import (
	"errors"
	"fmt"
	"github.com/bbeck/puzzles-with-chat/api/db"
	"github.com/bbeck/puzzles-with-chat/api/model"
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

	// The currently discovered words the puzzle mapping to their index.
	Words map[string]int `json:"words"`

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
	answer = strings.ToUpper(answer)

	// First, make sure the answer wasn't previously given.
	if _, found := s.Words[answer]; found {
		return errors.New("answer already given")
	}

	// Next, ensure the answer is in the list of allowed answers.
	var answers []string
	answers = append(answers, s.Puzzle.OfficialAnswers...)
	if allowUnofficial {
		answers = append(answers, s.Puzzle.UnofficialAnswers...)
	}
	sort.Strings(answers)

	index, found := find(answers, answer)
	if !found {
		return errors.New("answer not in the list of allowed answers")
	}

	// Save the answer to the state along with it's index.
	s.Words[answer] = index

	// Update the score for this answer.
	s.Score = s.Puzzle.ComputeScore(keys(s.Words))

	// Lastly determine if we've found all of the answers and the puzzle is now
	// complete.
	if len(s.Words) == len(answers) {
		s.Status = model.StatusComplete
	}

	return nil
}

// RebuildWordMap rebuilds the words map using the set of answers specified by
// the allowUnofficial parameter.  Words that are present that are no longer
// permitted are removed, and indices are adjusted appropriately.
func (s *State) RebuildWordMap(allowUnofficial bool) {
	var answers []string
	answers = append(answers, s.Puzzle.OfficialAnswers...)
	if allowUnofficial {
		answers = append(answers, s.Puzzle.UnofficialAnswers...)
	}
	sort.Strings(answers)

	words := make(map[string]int)
	for word := range s.Words {
		if index, found := find(answers, word); found {
			words[word] = index
		}
	}

	s.Words = words

	// The words may have changed, update the score accordingly.
	s.Score = s.Puzzle.ComputeScore(keys(s.Words))

	// Lastly determine if the puzzle is now solved.
	if len(s.Words) == len(answers) {
		s.Status = model.StatusComplete
	}
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
func GetState(conn db.Connection, channel string) (State, error) {
	var state State

	if testStateLoadError != nil {
		return state, testStateLoadError
	}

	err := db.Get(conn, StateKey(channel), &state)
	return state, err
}

// SetState writes the state for a channel's spelling bee solve to redis.  If
// the state can't be property written then an error will be returned.
func SetState(conn db.Connection, channel string, state State) error {
	if testStateSaveError != nil {
		return testStateSaveError
	}

	return db.SetWithTTL(conn, StateKey(channel), state, StateTTL)
}

// GetAllChannels returns a slice of model.Channel instances for each spelling
// bee that contains state in the database.  If there are no active channels
// then an empty slice is returned.  This method does not update the expiration
// times of any state instance.
func GetAllChannels(conn db.Connection) ([]model.Channel, error) {
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

		var description string
		var publisher string
		var published time.Time
		if state.Puzzle != nil {
			description = state.Puzzle.Description
			publisher = "The New York Times"
			published = state.Puzzle.PublishedDate
		}

		channels = append(channels, model.Channel{
			Name:        name,
			Status:      state.Status,
			Description: description,
			Puzzle: model.PuzzleSource{
				Publisher:     publisher,
				PublishedDate: published,
			},
		})
	}

	sort.Slice(channels, func(i, j int) bool {
		return channels[i].Name < channels[j].Name
	})

	return channels, nil
}

func find(words []string, word string) (int, bool) {
	index := sort.SearchStrings(words, word)
	found := index < len(words) && words[index] == word
	return index, found
}

func keys(words map[string]int) []string {
	var keys []string
	for key := range words {
		keys = append(keys, key)
	}
	return keys
}
