package acrostic

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bbeck/puzzles-with-chat/api/model"
	"github.com/bbeck/puzzles-with-chat/api/pubsub"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"path"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

var Channel = ChannelClient{name: "channel"}

func TestRoute_UpdatePuzzle_NewYorkTimes(t *testing.T) {
	// This acts as a small integration test updating the date of the New York
	// Times acrostic we're working on and ensuring the proper values are written
	// to the database.
	router, pool, registry := NewTestRouter(t)
	events := NewEventSubscription(t, registry, Channel.name)

	// Force a specific puzzle to be loaded so we don't make a network call.
	ForcePuzzleToBeLoaded(t, "xwordinfo-nyt-20200524.json")

	response := Channel.PUT("/", `{"new_york_times_date": "2020-05-24"}`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.Equal(t, model.StatusSelected, state.Status)
		assert.NotNil(t, state.Puzzle)
		assert.Equal(t, 0, len(state.CluesFilled))
		assert.Nil(t, state.LastStartTime)
		assert.Equal(t, 0., state.TotalSolveDuration.Seconds())
	})
}

func TestRoute_UpdatePuzzle_NewYorkTimes_WithGivens(t *testing.T) {
	// This acts as a small integration test updating the date of the New York
	// Times acrostic we're working on and ensuring the proper values are written
	// to the database.
	router, pool, registry := NewTestRouter(t)
	events := NewEventSubscription(t, registry, Channel.name)

	// Force a specific puzzle to be loaded so we don't make a network call.
	ForcePuzzleToBeLoaded(t, "xwordinfo-nyt-20011007-given-cell.json")

	response := Channel.PUT("/", `{"new_york_times_date": "2001-10-07"}`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.Equal(t, "-", state.Cells[5][18])
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
			ForcePuzzleToBeLoaded(t, "xwordinfo-nyt-20200524.json")

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
				ForcePuzzleToBeLoaded(t, "xwordinfo-nyt-20200524.json")
			}

			ForceErrorDuringStateSave(t, test.forceStateSaveError)

			response := Channel.PUT("/", test.json, router)
			assert.Equal(t, test.expected, response.Code)
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
	state := NewState(t, "xwordinfo-nyt-20200524.json")
	require.NoError(t, SetState(conn, Channel.name, state))

	// Request showing a clue.
	response := Channel.GET("/show/A", router)
	require.Equal(t, http.StatusOK, response.Code)
	VerifyShowClue(t, events, func(clue string) {
		assert.Equal(t, "A", clue)
	})

	// Request showing a lowercase clue.
	response = Channel.GET("/show/b", router)
	require.Equal(t, http.StatusOK, response.Code)
	VerifyShowClue(t, events, func(clue string) {
		assert.Equal(t, "B", clue)
	})

	// Request showing a malformed clue.
	response = Channel.GET("/show/1", router)
	require.Equal(t, http.StatusBadRequest, response.Code)

	// Request showing a properly formed, but non-existent clue.  This doesn't
	// fail because it doesn't mutate the state of the puzzle in any way.
	response = Channel.GET("/show/X", router)
	require.Equal(t, http.StatusOK, response.Code)
	VerifyShowClue(t, events, func(clue string) {
		assert.Equal(t, "X", clue)
	})
}

func TestRoute_ToggleStatus(t *testing.T) {
	// This acts as a small integration test toggling the status of an acrostic
	// being solved.
	router, pool, registry := NewTestRouter(t)
	conn := NewRedisConnection(t, pool)
	events := NewEventSubscription(t, registry, Channel.name)

	// Set a state that has a puzzle selected but not yet started.
	state := NewState(t, "xwordinfo-nyt-20200524.json")
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

			state := NewState(t, "xwordinfo-nyt-20200524.json")
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
	// This acts as a small integration test of applying answers to an acrostic
	// being solved.
	router, pool, registry := NewTestRouter(t)
	conn := NewRedisConnection(t, pool)
	events := NewEventSubscription(t, registry, Channel.name)

	state := NewState(t, "xwordinfo-nyt-20200524.json")
	state.Status = model.StatusSolving
	require.NoError(t, SetState(conn, Channel.name, state))

	// Apply a correct clue answer.
	response := Channel.PUT("/answer/A", `"WHALES"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.True(t, state.CluesFilled["A"])
		assert.Equal(t, "W", state.Cells[1][10])
		assert.Equal(t, "H", state.Cells[5][9])
		assert.Equal(t, "A", state.Cells[2][4])
		assert.Equal(t, "L", state.Cells[7][14])
		assert.Equal(t, "E", state.Cells[0][18])
		assert.Equal(t, "S", state.Cells[2][24])
	})

	// Apply a correct cells answer.
	response = Channel.PUT("/answer/1", `"PEOPLE"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.Equal(t, "P", state.Cells[0][0])
		assert.Equal(t, "E", state.Cells[0][1])
		assert.Equal(t, "O", state.Cells[0][2])
		assert.Equal(t, "P", state.Cells[0][3])
		assert.Equal(t, "L", state.Cells[0][4])
		assert.Equal(t, "E", state.Cells[0][5])
	})

	// Apply an incorrect clue answer.
	response = Channel.PUT("/answer/B", `"METALLICA"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.True(t, state.CluesFilled["B"])
		assert.Equal(t, "M", state.Cells[6][15])
		assert.Equal(t, "E", state.Cells[2][18])
		assert.Equal(t, "T", state.Cells[7][23])
		assert.Equal(t, "A", state.Cells[1][9])
		assert.Equal(t, "L", state.Cells[4][25])
		assert.Equal(t, "L", state.Cells[0][12])
		assert.Equal(t, "I", state.Cells[7][7])
		assert.Equal(t, "C", state.Cells[5][23])
		assert.Equal(t, "A", state.Cells[1][26])
	})

	// Apply an incorrect cells answer.
	response = Channel.PUT("/answer/7", `"SCREAM"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.Equal(t, "S", state.Cells[0][7])
		assert.Equal(t, "C", state.Cells[0][8])
		assert.Equal(t, "R", state.Cells[0][9])
		assert.Equal(t, "E", state.Cells[0][10])
		assert.Equal(t, "A", state.Cells[0][11])
		assert.Equal(t, "M", state.Cells[0][12])
	})

	// Pause the solve.
	response = Channel.PUT("/status", ``, router)
	require.Equal(t, http.StatusOK, response.Code)

	// Try to apply an answer.
	response = Channel.PUT("/answer/C", `"GYPSY"`, router)
	assert.Equal(t, http.StatusConflict, response.Code)
}

func TestRoute_UpdateAnswer_OnlyAllowCorrectAnswers(t *testing.T) {
	// This acts as a small integration test toggling the status of an acrostic
	// being solved.
	router, pool, registry := NewTestRouter(t)
	conn := NewRedisConnection(t, pool)
	events := NewEventSubscription(t, registry, Channel.name)

	// Change the settings to only allow correct answers.
	settings := Settings{OnlyAllowCorrectAnswers: true}
	require.NoError(t, SetSettings(conn, Channel.name, settings))

	state := NewState(t, "xwordinfo-nyt-20200524.json")
	state.Status = model.StatusSolving
	require.NoError(t, SetState(conn, Channel.name, state))

	// Apply a correct clue answer.
	response := Channel.PUT("/answer/A", `"WHALES"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.True(t, state.CluesFilled["A"])
		assert.Equal(t, "W", state.Cells[1][10])
		assert.Equal(t, "H", state.Cells[5][9])
		assert.Equal(t, "A", state.Cells[2][4])
		assert.Equal(t, "L", state.Cells[7][14])
		assert.Equal(t, "E", state.Cells[0][18])
		assert.Equal(t, "S", state.Cells[2][24])
	})

	// Apply a correct cells answer.
	response = Channel.PUT("/answer/1", `"PEOPLE"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		assert.Equal(t, "P", state.Cells[0][0])
		assert.Equal(t, "E", state.Cells[0][1])
		assert.Equal(t, "O", state.Cells[0][2])
		assert.Equal(t, "P", state.Cells[0][3])
		assert.Equal(t, "L", state.Cells[0][4])
		assert.Equal(t, "E", state.Cells[0][5])
	})

	// Apply an incorrect clue answer.
	response = Channel.PUT("/answer/B", `"METALLICA"`, router)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	// Apply an incorrect cells answer.
	response = Channel.PUT("/answer/7", `"SCREAM"`, router)
	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestRoute_UpdateAnswer_SolvedPuzzleStopsTimer(t *testing.T) {
	// This acts as a small integration test ensuring that the timer stops
	// counting once the acrostic has been solved.
	router, pool, registry := NewTestRouter(t)
	conn := NewRedisConnection(t, pool)
	events := NewEventSubscription(t, registry, Channel.name)

	// Setup a state that has the entire puzzle solved except for the last answer.
	state := NewState(t, "xwordinfo-nyt-20200524.json")
	state.Status = model.StatusSolving
	state.ApplyClueAnswer("A", "WHALES", false)
	state.ApplyClueAnswer("B", "AEROSMITH", false)
	state.ApplyClueAnswer("C", "GYPSY", false)
	state.ApplyClueAnswer("D", "NASHVILLE", false)
	state.ApplyClueAnswer("E", "ALLEMANDE", false)
	state.ApplyClueAnswer("F", "LORGNETTE", false)
	state.ApplyClueAnswer("G", "LEITMOTIF", false)
	state.ApplyClueAnswer("H", "SHARPED", false)
	state.ApplyClueAnswer("I", "SEATTLE", false)
	state.ApplyClueAnswer("J", "TEHRAN", false)
	state.ApplyClueAnswer("K", "ACCORDION", false)
	state.ApplyClueAnswer("L", "REPEAT", false)
	state.ApplyClueAnswer("M", "SYMPHONY", false)
	state.ApplyClueAnswer("N", "OMAHA", false)
	state.ApplyClueAnswer("O", "FLAWLESS", false)
	state.ApplyClueAnswer("P", "THAILAND", false)
	state.ApplyClueAnswer("Q", "HALFSTEP", false)
	state.ApplyClueAnswer("R", "ENTRACTE", false)
	state.ApplyClueAnswer("S", "OCTAVES", false)
	state.ApplyClueAnswer("T", "PROKOFIEV", false)
	state.ApplyClueAnswer("U", "EARDRUM", false)
	state.ApplyClueAnswer("V", "RHAPSODIC", false)
	require.NoError(t, SetState(conn, Channel.name, state))

	// Apply the last answer, but wait a bit first to ensure that a non-zero
	// amount of time has passed in the solve.
	time.Sleep(10 * time.Millisecond)

	response := Channel.PUT("/answer/W", `"ASSASSINS"`, router)
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

			state := NewState(t, "xwordinfo-nyt-20200524.json")
			state.Status = model.StatusSolving
			require.NoError(t, SetState(conn, Channel.name, state))

			response := Channel.PUT("/answer/A", test.json, router)
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

			state := NewState(t, "xwordinfo-nyt-20200524.json")
			state.Status = model.StatusSolving
			require.NoError(t, SetState(conn, Channel.name, state))

			ForceErrorDuringSettingsLoad(t, test.loadSettingsError)
			ForceErrorDuringStateLoad(t, test.loadStateError)
			ForceErrorDuringStateSave(t, test.saveStateError)

			response := Channel.PUT("/answer/A", `"WHALES"`, router)
			assert.NotEqual(t, http.StatusOK, response.Code)
		})
	}
}

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

	response = Channel.PUT("/setting/clue_font_size", `"xlarge"`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifySettings(t, pool, events, func(s Settings) {
		assert.Equal(t, model.FontSizeXLarge, s.ClueFontSize)
	})
}

func TestRoute_UpdateSetting_ClearsIncorrectCells(t *testing.T) {
	// This acts as a small integration test toggling the OnlyAllowCorrectAnswers
	// setting and ensuring that it clears any incorrect answer cells.
	router, pool, registry := NewTestRouter(t)
	conn := NewRedisConnection(t, pool)
	events := NewEventSubscription(t, registry, Channel.name)

	// Set a state that has an incorrect answer filled in for 1a.
	state := NewState(t, "xwordinfo-nyt-20200524.json")
	state.Status = model.StatusSolving
	state.Cells[0][0] = "P"
	state.Cells[0][1] = "U"
	state.Cells[0][2] = "R"
	state.Cells[0][3] = "P"
	state.Cells[0][4] = "L"
	state.Cells[0][5] = "E"
	require.NoError(t, SetState(conn, Channel.name, state))

	// Set the OnlyAllowCorrectAnswers setting to true
	response := Channel.PUT("/setting/only_allow_correct_answers", `true`, router)
	assert.Equal(t, http.StatusOK, response.Code)
	VerifyState(t, pool, events, func(state State) {
		state.Cells[0][0] = "P"
		state.Cells[0][1] = ""
		state.Cells[0][2] = ""
		state.Cells[0][3] = "P"
		state.Cells[0][4] = "L"
		state.Cells[0][5] = "E"
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
			name:    "clue_font_size",
			setting: "clue_font_size",
			json:    `{`,
		},
		{
			name:    "invalid setting name",
			setting: "foo_bar_baz",
			json:    `false`,
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

			state := NewState(t, "xwordinfo-nyt-20200524.json")
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
	state := NewState(t, "xwordinfo-nyt-20200524.json")
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
	response = Channel.PUT("/answer/A", `"WHALES"`, router)
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

	// Now complete the puzzle, this should cause a complete event to be sent.
	solution := `"PEOPLE SELDOM APPRECIATE THE VAST KNOWLEDGE WHICH ORCHESTRA ` +
		`PLAYERS POSSESS MOST OF THEM PLAY SEVERAL INSTRUMENTS AND THEY ALL HOLD ` +
		`AS A CREED THAT A FALSE NOTE IS A SIN AND A VARIATION IN RHYTHM IS A ` +
		`FALL FROM GRACE"`
	response = Channel.PUT("/answer/1", solution, router)
	assert.Equal(t, http.StatusOK, response.Code)

	events = flush()
	assert.Equal(t, 2, len(events)) // last state update event and complete event
	assert.Equal(t, "complete", events[1].Kind)

	complete := map[string]interface{}{
		"author": "MABEL WAGNALLS",
		"title":  "STARS OF THE OPERA",
		"text":   `<p>People seldom appreciate the vast knowledge of music and the remarkable ability in sight-reading which these orchestra players possess. Not one of them but has worked at his art from childhood; most of them play several different instruments; and they all hold as a creed that a false note is a sin, and a variation in rhythm is a fall from grace.</p>`,
	}
	assert.Equal(t, complete, events[1].Payload)

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

			state := NewState(t, "xwordinfo-nyt-20200524.json")
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

func TestRoute_GetAvailableDates(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*testing.T)
		source   string
		expected []string
	}{
		{
			name: "new york times",
			setup: func(t *testing.T) {
				ForceAvailableDatesToBeLoaded(t, "xwordinfo-select-acrostic-20211225.html")
			},
			source: "new_york_times",
			expected: []string{
				"1999-09-12",
				"2000-01-02",
				"2001-01-14",
				"2002-01-13",
				"2003-01-12",
				"2004-01-11",
				"2005-01-09",
				"2006-01-08",
				"2007-01-07",
				"2008-01-06",
				"2009-01-04",
				"2010-01-03",
				"2011-01-02",
				"2012-01-01",
				"2013-01-13",
				"2014-01-12",
				"2015-01-11",
				"2016-01-10",
				"2017-01-08",
				"2018-01-07",
				"2019-01-06",
				"2020-01-05",
				"2020-06-07",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setup != nil {
				test.setup(t)
			}

			router, _, _ := NewTestRouter(t)
			response := GET("/acrostic/dates", router)
			assert.Equal(t, http.StatusOK, response.Code)

			var dates map[string][]string
			require.NoError(t, render.DecodeJSON(response.Result().Body, &dates))

			for _, expected := range test.expected {
				index := sort.SearchStrings(dates[test.source], expected)
				require.True(t, index != len(dates))
				assert.Equal(t, expected, dates[test.source][index])
			}
		})
	}
}

func TestRoute_GetAvailableDates_Error(t *testing.T) {
	tests := []struct {
		name                string
		forceDatesLoadError error
	}{
		{
			name:                "error loading available dates",
			forceDatesLoadError: errors.New("forced error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router, _, _ := NewTestRouter(t)

			ForceErrorDuringAvailableDatesLoad(t, test.forceDatesLoadError)

			response := GET("/acrostic/dates", router)
			assert.Equal(t, http.StatusInternalServerError, response.Code)
		})
	}
}

// VerifyState performs common verifications for state objects in both event
// and database forms and then calls a custom verify function for test specific
// verifications.
func VerifyState(t *testing.T, pool *redis.Pool, events <-chan pubsub.Event, fn func(s State)) {
	t.Helper()

	// First check that we've received a single state event that has the correct
	// values
	found := Events(events, "state")
	require.Equal(t, 1, len(found), "incorrect number of events found")

	state := found[0].Payload.(State)
	assert.Nil(t, state.Puzzle.Cells) // Events should never have the solution
	fn(state)

	// Next check that the database has a valid state object
	conn := NewRedisConnection(t, pool)
	state, err := GetState(conn, Channel.name)
	require.NoError(t, err)
	assert.NotNil(t, state.Puzzle.Cells) // Database should always have the solution
	fn(state)
}

// VerifySettings performs test specific verifications on the settings objects
// in both event and database forms.
func VerifySettings(t *testing.T, pool *redis.Pool, events <-chan pubsub.Event, fn func(s Settings)) {
	t.Helper()

	// First check that we've received a single settings event with the correct
	// values
	found := Events(events, "settings")
	require.Equal(t, 1, len(found), "incorrect number of events found")

	settings := found[0].Payload.(Settings)
	fn(settings)

	// Next check that the database has a valid settings object
	conn := NewRedisConnection(t, pool)
	settings, err := GetSettings(conn, Channel.name)
	require.NoError(t, err)
	fn(settings)
}

// VerifyShowClue performs common verifications for show clue events.
func VerifyShowClue(t *testing.T, events <-chan pubsub.Event, fn func(clue string)) {
	t.Helper()

	found := Events(events, "show_clue")
	require.Equal(t, 1, len(found))
	fn(found[0].Payload.(string))
}

// Events extracts events of a particular kind from a channel.  It consumes all
// events in the channel that are available at the time of the call.
func Events(events <-chan pubsub.Event, kind string) []pubsub.Event {
	var found []pubsub.Event

	for done := false; !done; {
		select {
		case event := <-events:
			if event.Kind != kind {
				continue
			}

			found = append(found, event)

		default:
			done = true
		}
	}

	return found
}

// GET performs a HTTP GET request to the provided router.
func GET(url string, router chi.Router) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, url, nil)
	router.ServeHTTP(recorder, request)
	return recorder
}

// ChannelClient is a client that makes requests against the URL of a particular
// user's channel.
type ChannelClient struct {
	name string
}

func (c ChannelClient) GET(url string, router chi.Router) *httptest.ResponseRecorder {
	url = path.Join("/acrostic", c.name, url)
	return GET(url, router)
}

func (c ChannelClient) PUT(url, body string, router chi.Router) *httptest.ResponseRecorder {
	url = path.Join("/acrostic", c.name, url)
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
	url = path.Join("/acrostic", c.name, url)
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
