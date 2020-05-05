package crossword

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bbeck/puzzles-with-chat/api/model"
	"github.com/bbeck/puzzles-with-chat/api/pubsub"
	"github.com/go-chi/chi"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"sync"
	"testing"
	"time"
)

var Channel = ChannelClient{name: "channel"}

func TestRoute_UpdateSetting(t *testing.T) {
	// This acts as a small integration test updating each setting in turn and
	// making sure the proper value is written to the database and that clients
	// receive events notifying them of the setting change.
	router, pool, registry := NewTestRouter(t)
	events := NewEventSubscription(t, registry, Channel.name)

	// Update each setting, one at a time.
	response := Channel.PUT("/setting/only_allow_correct_answers", `true`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifySettings(t, pool, events, func(s Settings) {
		assert.True(t, s.OnlyAllowCorrectAnswers)
	})

	response = Channel.PUT("/setting/clues_to_show", `"none"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifySettings(t, pool, events, func(s Settings) {
		assert.Equal(t, NoCluesVisible, s.CluesToShow)
	})

	response = Channel.PUT("/setting/clue_font_size", `"xlarge"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifySettings(t, pool, events, func(s Settings) {
		assert.Equal(t, model.FontSizeXLarge, s.ClueFontSize)
	})

	response = Channel.PUT("/setting/show_notes", `true`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifySettings(t, pool, events, func(s Settings) {
		assert.True(t, s.ShowNotes)
	})
}

func TestRoute_UpdateSetting_ClearsIncorrectCells(t *testing.T) {
	// This acts as a small integration test toggling the OnlyAllowCorrectAnswers
	// setting and ensuring that it clears any incorrect answer cells.
	router, pool, registry := NewTestRouter(t)
	conn := NewRedisConnection(t, pool)
	events := NewEventSubscription(t, registry, Channel.name)

	// Set a state that has an incorrect answer filled in for 1a.
	state := NewState(t, "xwordinfo-nyt-20181231.json")
	state.Status = model.StatusSolving
	state.Cells[0][0] = "Q"
	state.Cells[0][1] = "N"
	state.Cells[0][2] = "O"
	state.Cells[0][3] = "R"
	state.Cells[0][4] = "A"
	require.NoError(t, SetState(conn, Channel.name, state))

	// Set the OnlyAllowCorrectAnswers setting to true
	response := Channel.PUT("/setting/only_allow_correct_answers", `true`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.False(t, state.AcrossCluesFilled[1])
		assert.Equal(t, "Q", state.Cells[0][0])
		assert.Equal(t, "", state.Cells[0][1])
		assert.Equal(t, "", state.Cells[0][2])
		assert.Equal(t, "", state.Cells[0][3])
		assert.Equal(t, "A", state.Cells[0][4])
	})
}

func TestRoute_UpdateSetting_JSONError(t *testing.T) {
	tests := []struct {
		name    string
		setting string
		json    string
	}{
		{
			name:    "only_allow_correct_answers",
			setting: "only_allow_correct_answers",
			json:    `{`,
		},
		{
			name:    "clues_to_show",
			setting: "clues_to_show",
			json:    `{`,
		},
		{
			name:    "clue_font_size",
			setting: "clue_font_size",
			json:    `{`,
		},
		{
			name:    "invalid setting name",
			setting: "foo_bar_baz",
			json:    `false`,
		},
		{
			name:    "show_notes",
			setting: "show_notes",
			json:    `{`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router, _, _ := NewTestRouter(t)

			response := Channel.PUT(fmt.Sprintf("/setting/%s", test.setting), test.json, router)
			assert.NotEqual(t, http.StatusOK, response.Code)
		})
	}
}

func TestRoute_UpdateSettings_LoadSaveError(t *testing.T) {
	tests := []struct {
		name                   string
		forceSettingsLoadError error
		forceSettingsSaveError error
		forceStateLoadError    error
		forceStateSaveError    error
	}{
		{
			name:                   "error loading settings",
			forceSettingsLoadError: errors.New("forced error"),
		},
		{
			name:                   "error saving settings",
			forceSettingsSaveError: errors.New("forced error"),
		},
		{
			name:                "error loading state",
			forceStateLoadError: errors.New("forced error"),
		},
		{
			name:                "error saving state",
			forceStateSaveError: errors.New("forced error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router, pool, _ := NewTestRouter(t)
			conn := NewRedisConnection(t, pool)

			state := NewState(t, "xwordinfo-nyt-20181231.json")
			state.Status = model.StatusSolving
			require.NoError(t, SetState(conn, Channel.name, state))

			ForceErrorDuringSettingsLoad(t, test.forceSettingsLoadError)
			ForceErrorDuringSettingsSave(t, test.forceSettingsSaveError)
			ForceErrorDuringStateLoad(t, test.forceStateLoadError)
			ForceErrorDuringStateSave(t, test.forceStateSaveError)

			response := Channel.PUT("/setting/only_allow_correct_answers", `true`, router)
			assert.NotEqual(t, http.StatusOK, response.Code)
		})
	}
}

func TestRoute_UpdatePuzzle_NewYorkTimes(t *testing.T) {
	// This acts as a small integration test updating the date of the New York
	// Times crossword we're working on and ensuring the proper values are written
	// to the database.
	router, pool, registry := NewTestRouter(t)
	events := NewEventSubscription(t, registry, Channel.name)

	// Force a specific puzzle to be loaded so we don't make a network call.
	ForcePuzzleToBeLoaded(t, "xwordinfo-nyt-20181231.json")

	response := Channel.PUT("/", `{"new_york_times_date": "2018-12-31"}`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.Equal(t, model.StatusSelected, state.Status)
		assert.NotNil(t, state.Puzzle)
		assert.Equal(t, 0, len(state.AcrossCluesFilled))
		assert.Equal(t, 0, len(state.DownCluesFilled))
		assert.Nil(t, state.LastStartTime)
		assert.Equal(t, 0., state.TotalSolveDuration.Seconds())
	})
}

func TestRoute_UpdatePuzzle_WallStreetJournal(t *testing.T) {
	// This acts as a small integration test updating the date of the Wall Street
	// Journal crossword we're working on and ensuring the proper values are
	// written to the database.
	router, pool, registry := NewTestRouter(t)
	events := NewEventSubscription(t, registry, Channel.name)

	// Force a specific puzzle to be loaded so we don't make a network call.
	ForcePuzzleToBeLoaded(t, "puzzle-wsj-20190102.json")

	response := Channel.PUT("/", `{"wall_street_journal_date": "2019-01-02"}`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.Equal(t, model.StatusSelected, state.Status)
		assert.NotNil(t, state.Puzzle)
		assert.Equal(t, 0, len(state.AcrossCluesFilled))
		assert.Equal(t, 0, len(state.DownCluesFilled))
		assert.Nil(t, state.LastStartTime)
		assert.Equal(t, 0., state.TotalSolveDuration.Seconds())
	})
}

func TestRoute_UpdatePuzzle_PuzFile(t *testing.T) {
	// This acts as a small integration test uploading a .puz file of the
	// crossword we're working on and ensuring the proper values are written to
	// the database.
	router, pool, registry := NewTestRouter(t)
	events := NewEventSubscription(t, registry, Channel.name)

	// Force a specific puzzle to be loaded so we don't make a network call.
	ForcePuzzleToBeLoaded(t, "puzzle-wp-20051206.json")

	response := Channel.PUT("/", `{"puz_file_bytes": "unused"}`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.Equal(t, model.StatusSelected, state.Status)
		assert.NotNil(t, state.Puzzle)
		assert.Equal(t, 0, len(state.AcrossCluesFilled))
		assert.Equal(t, 0, len(state.DownCluesFilled))
		assert.Nil(t, state.LastStartTime)
		assert.Equal(t, 0., state.TotalSolveDuration.Seconds())
	})
}

func TestRoute_UpdatePuzzle_PuzURL(t *testing.T) {
	// This acts as a small integration test retrieving a .puz file from a URL of
	// the crossword we're working on and ensuring the proper values are written
	// to the database.
	router, pool, registry := NewTestRouter(t)
	events := NewEventSubscription(t, registry, Channel.name)

	// Force a specific puzzle to be loaded so we don't make a network call.
	ForcePuzzleToBeLoaded(t, "puzzle-wp-20051206.json")

	response := Channel.PUT("/", `{"puz_file_url": "unused"}`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.Equal(t, model.StatusSelected, state.Status)
		assert.NotNil(t, state.Puzzle)
		assert.Equal(t, 0, len(state.AcrossCluesFilled))
		assert.Equal(t, 0, len(state.DownCluesFilled))
		assert.Nil(t, state.LastStartTime)
		assert.Equal(t, 0., state.TotalSolveDuration.Seconds())
	})
}

func TestRoute_UpdatePuzzle_JSONError(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected int
	}{
		{
			name:     "bad json",
			json:     `{"new_york_times_date": }`,
			expected: http.StatusBadRequest,
		},
		{
			name:     "invalid json",
			json:     `{}`,
			expected: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router, _, _ := NewTestRouter(t)
			ForcePuzzleToBeLoaded(t, "xwordinfo-nyt-20181231.json")

			response := Channel.PUT("/", test.json, router)
			assert.Equal(t, test.expected, response.Code)
		})
	}
}

func TestRoute_UpdatePuzzle_LoadSaveError(t *testing.T) {
	tests := []struct {
		name                 string
		json                 string
		forcePuzzleLoadError error
		forceStateSaveError  error
		expected             int
	}{
		{
			name:                 "nyt error loading puzzle",
			json:                 `{"new_york_times_date": "unused"}`,
			forcePuzzleLoadError: errors.New("forced error"),
			expected:             http.StatusInternalServerError,
		},
		{
			name:                 "wsj error loading puzzle",
			json:                 `{"wall_street_journal_date": "unused"}`,
			forcePuzzleLoadError: errors.New("forced error"),
			expected:             http.StatusInternalServerError,
		},
		{
			name:                 "puz bytes error loading puzzle",
			json:                 `{"puz_file_bytes": "unused"}`,
			forcePuzzleLoadError: errors.New("forced error"),
			expected:             http.StatusInternalServerError,
		},
		{
			name:                 "puz url error loading puzzle",
			json:                 `{"puz_file_url": "unused"}`,
			forcePuzzleLoadError: errors.New("forced error"),
			expected:             http.StatusInternalServerError,
		},
		{
			name:                "error saving state",
			json:                `{"new_york_times_date": "unused"}`,
			forceStateSaveError: errors.New("forced error"),
			expected:            http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router, _, _ := NewTestRouter(t)

			if test.forcePuzzleLoadError != nil {
				ForceErrorDuringPuzzleLoad(t, test.forcePuzzleLoadError)
			} else {
				ForcePuzzleToBeLoaded(t, "xwordinfo-nyt-20181231.json")
			}

			ForceErrorDuringStateSave(t, test.forceStateSaveError)

			response := Channel.PUT("/", test.json, router)
			assert.Equal(t, test.expected, response.Code)
		})
	}
}

func TestRoute_ToggleStatus(t *testing.T) {
	// This acts as a small integration test toggling the status of a crossword
	// being solved.
	router, pool, registry := NewTestRouter(t)
	conn := NewRedisConnection(t, pool)
	events := NewEventSubscription(t, registry, Channel.name)

	// Set a state that has a puzzle selected but not yet started.
	state := NewState(t, "xwordinfo-nyt-20181231.json")
	require.NoError(t, SetState(conn, Channel.name, state))

	// Toggle the status, it should transition to solving.
	response := Channel.PUT("/status", ``, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.Equal(t, model.StatusSolving, state.Status)
		assert.NotNil(t, state.LastStartTime)
	})

	// Toggle the status again, it should transition to paused. Make sure we
	// sleep for at least a nanosecond first so that the solve was unpaused for
	// a non-zero duration.
	time.Sleep(1 * time.Nanosecond)
	response = Channel.PUT("/status", ``, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.Equal(t, model.StatusPaused, state.Status)
		assert.Nil(t, state.LastStartTime)
		assert.True(t, state.TotalSolveDuration.Seconds() > 0.)
	})

	// Toggle the status again, it should transition back to solving.
	response = Channel.PUT("/status", ``, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.Equal(t, model.StatusSolving, state.Status)
		assert.NotNil(t, state.LastStartTime)
		assert.True(t, state.TotalSolveDuration.Seconds() > 0.)
	})

	// Force the puzzle to be complete.
	state.Status = model.StatusComplete
	require.NoError(t, SetState(conn, Channel.name, state))

	// Try to toggle the status one more time.  Now that the puzzle is complete
	// it should return an HTTP error.
	response = Channel.PUT("/status", ``, router)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	state, err := GetState(conn, Channel.name)
	require.NoError(t, err)
	assert.Equal(t, model.StatusComplete, state.Status)
}

func TestRoute_ToggleStatus_Error(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  model.Status
		loadStateError error
		saveStateError error
	}{
		{
			name:          "status created",
			initialStatus: model.StatusCreated,
		},
		{
			name:           "error loading state",
			initialStatus:  model.StatusSelected,
			loadStateError: errors.New("forced error"),
		},
		{
			name:           "error saving state",
			initialStatus:  model.StatusSelected,
			saveStateError: errors.New("forced error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router, pool, _ := NewTestRouter(t)
			conn := NewRedisConnection(t, pool)

			state := NewState(t, "xwordinfo-nyt-20181231.json")
			state.Status = test.initialStatus
			require.NoError(t, SetState(conn, Channel.name, state))

			if test.loadStateError != nil {
				ForceErrorDuringStateLoad(t, test.loadStateError)
			}

			if test.saveStateError != nil {
				ForceErrorDuringStateSave(t, test.saveStateError)
			}

			response := Channel.PUT("/status", "", router)
			assert.NotEqual(t, http.StatusOK, response.Code)
		})
	}
}

func TestRoute_UpdateAnswer_AllowIncorrectAnswers(t *testing.T) {
	// This acts as a small integration test of applying answers to a crossword
	// being solved.
	router, pool, registry := NewTestRouter(t)
	conn := NewRedisConnection(t, pool)
	events := NewEventSubscription(t, registry, Channel.name)

	state := NewState(t, "xwordinfo-nyt-20181231.json")
	state.Status = model.StatusSolving
	require.NoError(t, SetState(conn, Channel.name, state))

	// Apply a correct across answer.
	response := Channel.PUT("/answer/1a", `"QANDA"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.True(t, state.AcrossCluesFilled[1])
		assert.Equal(t, "Q", state.Cells[0][0])
		assert.Equal(t, "A", state.Cells[0][1])
		assert.Equal(t, "N", state.Cells[0][2])
		assert.Equal(t, "D", state.Cells[0][3])
		assert.Equal(t, "A", state.Cells[0][4])
	})

	// Apply a correct down answer.
	response = Channel.PUT("/answer/1d", `"QTIP"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.True(t, state.DownCluesFilled[1])
		assert.Equal(t, "Q", state.Cells[0][0])
		assert.Equal(t, "T", state.Cells[1][0])
		assert.Equal(t, "I", state.Cells[2][0])
		assert.Equal(t, "P", state.Cells[3][0])
	})

	// Apply an incorrect across answer.
	response = Channel.PUT("/answer/6a", `"FLOOR"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.True(t, state.AcrossCluesFilled[6])
		assert.Equal(t, "F", state.Cells[0][6])
		assert.Equal(t, "L", state.Cells[0][7])
		assert.Equal(t, "O", state.Cells[0][8])
		assert.Equal(t, "O", state.Cells[0][9])
		assert.Equal(t, "R", state.Cells[0][10])
	})

	// Apply an incorrect down answer.
	response = Channel.PUT("/answer/11d", `"HEYA"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.True(t, state.DownCluesFilled[11])
		assert.Equal(t, "H", state.Cells[0][12])
		assert.Equal(t, "E", state.Cells[1][12])
		assert.Equal(t, "Y", state.Cells[2][12])
		assert.Equal(t, "A", state.Cells[3][12])
	})

	// Pause the solve.
	response = Channel.PUT("/status", ``, router)
	require.Equal(t, http.StatusOK, response.Code)

	// Try to apply an answer.
	response = Channel.PUT("/answer/6a", `"ATTIC"`, router)
	assert.Equal(t, http.StatusConflict, response.Code)
}

func TestRoute_UpdateAnswer_OnlyAllowCorrectAnswers(t *testing.T) {
	// This acts as a small integration test toggling the status of a crossword
	// being solved.
	router, pool, registry := NewTestRouter(t)
	conn := NewRedisConnection(t, pool)
	events := NewEventSubscription(t, registry, Channel.name)

	// Change the settings to only allow correct answers.
	settings := Settings{OnlyAllowCorrectAnswers: true}
	require.NoError(t, SetSettings(conn, Channel.name, settings))

	state := NewState(t, "xwordinfo-nyt-20181231.json")
	state.Status = model.StatusSolving
	require.NoError(t, SetState(conn, Channel.name, state))

	// Apply a correct across answer.
	response := Channel.PUT("/answer/1a", `"QANDA"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.True(t, state.AcrossCluesFilled[1])
		assert.Equal(t, "Q", state.Cells[0][0])
		assert.Equal(t, "A", state.Cells[0][1])
		assert.Equal(t, "N", state.Cells[0][2])
		assert.Equal(t, "D", state.Cells[0][3])
		assert.Equal(t, "A", state.Cells[0][4])
	})

	// Apply a correct down answer.
	response = Channel.PUT("/answer/1d", `"QTIP"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.True(t, state.DownCluesFilled[1])
		assert.Equal(t, "Q", state.Cells[0][0])
		assert.Equal(t, "T", state.Cells[1][0])
		assert.Equal(t, "I", state.Cells[2][0])
		assert.Equal(t, "P", state.Cells[3][0])
	})

	// Apply an incorrect across answer.
	response = Channel.PUT("/answer/6a", `"FLOOR"`, router)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	// Apply an incorrect down answer.
	response = Channel.PUT("/answer/11d", `"HEYA"`, router)
	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestRoute_UpdateAnswer_SolvedPuzzleStopsTimer(t *testing.T) {
	// This acts as a small integration test ensuring that the timer stops
	// counting once the crossword has been solved.
	router, pool, registry := NewTestRouter(t)
	conn := NewRedisConnection(t, pool)
	events := NewEventSubscription(t, registry, Channel.name)

	// Setup a state that has the entire puzzle solved except for the last answer.
	state := NewState(t, "xwordinfo-nyt-20181231.json")
	state.Status = model.StatusSolving
	state.ApplyAnswer("1a", "Q AND A", false)
	state.ApplyAnswer("6a", "ATTIC", false)
	state.ApplyAnswer("11a", "HON", false)
	state.ApplyAnswer("14a", "THIRD", false)
	state.ApplyAnswer("15a", "LAID ASIDE", false)
	state.ApplyAnswer("17a", "IM TOO OLD FOR THIS", false)
	state.ApplyAnswer("19a", "PERU", false)
	state.ApplyAnswer("20a", "LEAF", false)
	state.ApplyAnswer("21a", "PEONS", false)
	state.ApplyAnswer("22a", "DOG TAG", false)
	state.ApplyAnswer("24a", "LOL", false)
	state.ApplyAnswer("25a", "HAVE NO OOMPH", false)
	state.ApplyAnswer("30a", "MATTE", false)
	state.ApplyAnswer("33a", "IMPLORED", false)
	state.ApplyAnswer("35a", "ERR", false)
	state.ApplyAnswer("36a", "RANGE", false)
	state.ApplyAnswer("38a", "EMO", false)
	state.ApplyAnswer("39a", "WAIT HERE", false)
	state.ApplyAnswer("42a", "EGYPT", false)
	state.ApplyAnswer("44a", "BOO OFF STAGE", false)
	state.ApplyAnswer("47a", "ERS", false)
	state.ApplyAnswer("48a", "EUGENE", false)
	state.ApplyAnswer("51a", "SHARI", false)
	state.ApplyAnswer("54a", "SINN", false)
	state.ApplyAnswer("56a", "WING", false)
	state.ApplyAnswer("58a", "ITS A ZOO OUT THERE", false)
	state.ApplyAnswer("61a", "STEGOSAUR", false)
	state.ApplyAnswer("62a", "HIT ON", false)
	state.ApplyAnswer("63a", "IPA", false)
	state.ApplyAnswer("64a", "NURSE", false)
	require.NoError(t, SetState(conn, Channel.name, state))

	// Apply the last answer, but wait a bit first to ensure that a non-zero
	// amount of time has passed in the solve.
	time.Sleep(10 * time.Millisecond)

	response := Channel.PUT("/answer/65a", `"OZONE"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		require.Equal(t, model.StatusComplete, state.Status)
		assert.Nil(t, state.LastStartTime)
		assert.True(t, state.TotalSolveDuration.Seconds() > 0)
	})
}

func TestRoute_UpdateAnswer_Error(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected int
	}{
		{
			name:     "empty answer",
			json:     `""`,
			expected: http.StatusBadRequest,
		},
		{
			name:     "too long answer",
			json:     `"` + RandomString(1023) + `"`,
			expected: http.StatusRequestEntityTooLarge,
		},
		{
			name:     "malformed json",
			json:     `"`,
			expected: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router, pool, _ := NewTestRouter(t)
			conn := NewRedisConnection(t, pool)

			state := NewState(t, "xwordinfo-nyt-20181231.json")
			state.Status = model.StatusSolving
			require.NoError(t, SetState(conn, Channel.name, state))

			response := Channel.PUT("/answer/1a", test.json, router)
			assert.Equal(t, test.expected, response.Code)
		})
	}
}

func TestRoute_UpdateAnswer_LoadSaveError(t *testing.T) {
	tests := []struct {
		name              string
		loadSettingsError error
		loadStateError    error
		saveStateError    error
	}{
		{
			name:              "error loading settings",
			loadSettingsError: errors.New("forced error"),
		},
		{
			name:           "error loading state",
			loadStateError: errors.New("forced error"),
		},
		{
			name:           "error saving state",
			saveStateError: errors.New("forced error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router, pool, _ := NewTestRouter(t)
			conn := NewRedisConnection(t, pool)

			state := NewState(t, "xwordinfo-nyt-20181231.json")
			state.Status = model.StatusSolving
			require.NoError(t, SetState(conn, Channel.name, state))

			ForceErrorDuringSettingsLoad(t, test.loadSettingsError)
			ForceErrorDuringStateLoad(t, test.loadStateError)
			ForceErrorDuringStateSave(t, test.saveStateError)

			response := Channel.PUT("/answer/1a", `"QANDA"`, router)
			assert.NotEqual(t, http.StatusOK, response.Code)
		})
	}
}

func TestRoute_ShowClue(t *testing.T) {
	// This acts as a small integration test requesting clues to be shown and
	// making sure events are properly emitted.
	router, pool, registry := NewTestRouter(t)
	conn := NewRedisConnection(t, pool)
	events := NewEventSubscription(t, registry, Channel.name)

	// Force a specific puzzle to be loaded so we don't make a network call.
	state := NewState(t, "xwordinfo-nyt-20181231.json")
	require.NoError(t, SetState(conn, Channel.name, state))

	// Request showing an across clue.
	response := Channel.GET("/show/1a", router)
	require.Equal(t, http.StatusOK, response.Code)
	VerifyShowClue(t, events, func(clue string) {
		assert.Equal(t, "1a", clue)
	})

	// Request showing a down clue.
	response = Channel.GET("/show/16d", router)
	require.Equal(t, http.StatusOK, response.Code)
	VerifyShowClue(t, events, func(clue string) {
		assert.Equal(t, "16d", clue)
	})

	// Request showing a malformed clue.
	response = Channel.GET("/show/1x", router)
	require.Equal(t, http.StatusBadRequest, response.Code)

	// Request showing a properly formed, but non-existent clue.  This doesn't
	// fail because it doesn't mutate the state of the puzzle in any way.
	response = Channel.GET("/show/999a", router)
	require.Equal(t, http.StatusOK, response.Code)
	VerifyShowClue(t, events, func(clue string) {
		assert.Equal(t, "999a", clue)
	})
}

func TestRoute_GetEvents(t *testing.T) {
	// This acts as a small integration test ensuring that the event stream
	// receives the events put into a registry.
	router, pool, _ := NewTestRouter(t)
	conn := NewRedisConnection(t, pool)

	// Connect to the stream when there's no puzzle selected, we should receive
	// just the channel's settings.
	_, stop := Channel.SSE("/events", router)
	events := stop()
	assert.Equal(t, 1, len(events))
	assert.Equal(t, "settings", events[0].Kind)

	// Select a puzzle.
	state := NewState(t, "xwordinfo-nyt-20181231.json")
	require.NoError(t, SetState(conn, Channel.name, state))

	// Now reconnect to the stream and we should receive both the settings and the
	// channel's current state.
	flush, stop := Channel.SSE("/events", router)
	events = flush()
	assert.Equal(t, 2, len(events))
	assert.Equal(t, "settings", events[0].Kind)
	assert.Equal(t, "state", events[1].Kind)

	// Toggle the status to solving, this should cause the state to be sent again.
	response := Channel.PUT("/status", ``, router)
	require.Equal(t, http.StatusOK, response.Code)

	events = flush()
	assert.Equal(t, 1, len(events))
	assert.Equal(t, "state", events[0].Kind)

	// Now specify an answer, this should cause the state to be sent again.
	response = Channel.PUT("/answer/1a", `"QANDA"`, router)
	assert.Equal(t, http.StatusOK, response.Code)

	events = flush()
	assert.Equal(t, 1, len(events))
	assert.Equal(t, "state", events[0].Kind)

	// Now change a setting, this should cause the settings to be sent again.
	response = Channel.PUT("/setting/clue_font_size", `"xlarge"`, router)
	assert.Equal(t, http.StatusOK, response.Code)

	events = flush()
	assert.Equal(t, 1, len(events))
	assert.Equal(t, "settings", events[0].Kind)

	// Disconnect, there shouldn't be any events anymore.
	events = stop()
	assert.Equal(t, 0, len(events))
}

func TestRoute_GetEvents_LoadSaveError(t *testing.T) {
	tests := []struct {
		name                   string
		forceSettingsLoadError error
		forceStateLoadError    error
	}{
		{
			name:                   "error loading settings",
			forceSettingsLoadError: errors.New("forced error"),
		},
		{
			name:                "error loading state",
			forceStateLoadError: errors.New("forced error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router, pool, _ := NewTestRouter(t)
			conn := NewRedisConnection(t, pool)

			state := NewState(t, "xwordinfo-nyt-20181231.json")
			require.NoError(t, SetState(conn, Channel.name, state))

			ForceErrorDuringSettingsLoad(t, test.forceSettingsLoadError)
			ForceErrorDuringStateLoad(t, test.forceStateLoadError)

			// This won't start a background goroutine to send events because the
			// request will fail before reaching that part of the code.
			response := Channel.GET("/events", router)
			assert.NotEqual(t, http.StatusOK, response.Code)
		})
	}
}

// VerifySettings performs test specific verifications on the settings objects
// in both event and database forms.
func VerifySettings(t *testing.T, pool *redis.Pool, events <-chan pubsub.Event, fn func(s Settings)) {
	t.Helper()

	// First check that we've received an event with the correct value
	select {
	case event := <-events:
		// Ignore any non-settings events.
		if event.Kind != "settings" {
			return
		}
		fn(event.Payload.(Settings))

	default:
		assert.Fail(t, "no settings event available")
	}

	// Next check that the database has a valid settings object
	conn := NewRedisConnection(t, pool)
	settings, err := GetSettings(conn, Channel.name)
	require.NoError(t, err)
	fn(settings)
}

// VerifyState performs common verifications for state objects in both event
// and database forms and then calls a custom verify function for test specific
// verifications.
func VerifyState(t *testing.T, pool *redis.Pool, events <-chan pubsub.Event, fn func(s State)) {
	t.Helper()

	// First check that we've received an event with the correct value
	select {
	case event := <-events:
		// Ignore any non-state events.
		if event.Kind != "state" {
			return
		}

		state := event.Payload.(State)
		assert.Nil(t, state.Puzzle.Cells) // Events should never have the solution
		fn(state)

	default:
		assert.Fail(t, "no state event available")
	}

	// Next check that the database has a valid settings object
	conn := NewRedisConnection(t, pool)
	state, err := GetState(conn, Channel.name)
	require.NoError(t, err)
	assert.NotNil(t, state.Puzzle.Cells) // Database should always have the solution
	fn(state)
}

// VerifyShowClue performs common verifications for show clue events.
func VerifyShowClue(t *testing.T, events <-chan pubsub.Event, fn func(clue string)) {
	t.Helper()

	// First check that we've received an event with the correct value
	select {
	case event := <-events:
		if event.Kind != "show_clue" {
			return
		}
		fn(event.Payload.(string))

	default:
		assert.Fail(t, "no show_clue event available")
	}
}

// ChannelClient is a client that makes requests against the URL of a particular
// user's channel.
type ChannelClient struct {
	name string
}

func (c ChannelClient) GET(url string, router chi.Router) *httptest.ResponseRecorder {
	url = path.Join("/crossword", c.name, url)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, url, nil)
	router.ServeHTTP(recorder, request)
	return recorder
}

func (c ChannelClient) PUT(url, body string, router chi.Router) *httptest.ResponseRecorder {
	url = path.Join("/crossword", c.name, url)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, url, strings.NewReader(body))
	router.ServeHTTP(recorder, request)
	return recorder
}

// SSE performs a streaming request to the provided router.  Because the router
// won't immediately return, this request is done in a background goroutine.
// When the main thread wishes to read events that have been received thus far
// the flush method can be called and it will return any queued up events.  When
// the main thread wishes to close the connection to the router the stop method
// can be called and it will return any unread events.
func (c ChannelClient) SSE(url string, router chi.Router) (flush func() []pubsub.Event, stop func() []pubsub.Event) {
	url = path.Join("/crossword", c.name, url)
	recorder := CreateTestResponseRecorder()

	flush = func() []pubsub.Event {
		// Give the router a chance to write everything it needs to.
		time.Sleep(10 * time.Millisecond)

		reader, err := recorder.Body()
		if err != nil {
			return nil
		}

		var events []pubsub.Event
		for {
			bs, err := reader.ReadBytes('\n')
			if err != nil {
				break
			}

			if !bytes.HasPrefix(bs, []byte("data:")) {
				continue
			}

			var event pubsub.Event
			json.Unmarshal(bs[5:], &event)
			events = append(events, event)
		}

		return events
	}

	stop = func() []pubsub.Event {
		// Give the router a chance to write everything it needs to.
		time.Sleep(10 * time.Millisecond)

		recorder.Close()
		return flush()
	}

	request := httptest.NewRequest(http.MethodGet, url, nil)
	go router.ServeHTTP(recorder, request)

	return flush, stop
}

func RandomString(n int) string {
	var alphabet = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")

	bs := make([]rune, n)
	for i := range bs {
		bs[i] = alphabet[rand.Intn(len(alphabet))]
	}

	return string(bs)
}

// Create a http.ResponseWriter that synchronizes whenever reads or writes
// happen so that there are no races in a multiple goroutine environment.
// Additionally implement the http.CloseNotifier interface so that requests can
// be stopped by tests.
type TestResponseRecorder struct {
	sync.Mutex
	headers http.Header
	body    *bytes.Buffer
	close   chan bool
}

func CreateTestResponseRecorder() *TestResponseRecorder {
	return &TestResponseRecorder{
		headers: make(http.Header),
		body:    new(bytes.Buffer),
		close:   make(chan bool, 1),
	}
}

func (r *TestResponseRecorder) Header() http.Header {
	r.Lock()
	defer r.Unlock()

	return r.headers
}

func (r *TestResponseRecorder) Write(bs []byte) (int, error) {
	r.Lock()
	defer r.Unlock()

	return r.body.Write(bs)
}

func (r *TestResponseRecorder) Body() (*bufio.Reader, error) {
	r.Lock()
	defer r.Unlock()

	bs, err := ioutil.ReadAll(r.body)
	if err != nil {
		return nil, err
	}
	r.body.Reset()
	return bufio.NewReader(bytes.NewReader(bs)), nil
}

func (r *TestResponseRecorder) CloseNotify() <-chan bool {
	r.Lock()
	defer r.Unlock()

	return r.close
}

func (r *TestResponseRecorder) Close() {
	r.Lock()
	defer r.Unlock()

	r.close <- true
}

func (r *TestResponseRecorder) WriteHeader(int) {
	// Not used
}
