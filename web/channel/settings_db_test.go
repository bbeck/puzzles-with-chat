package channel

import (
	"fmt"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSettings(t *testing.T) {
	s, err := miniredis.Run()
	require.NoError(t, err)

	c, err := redis.Dial("tcp", s.Addr())
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	key := func(s string) string {
		return fmt.Sprintf("settings:%s", s)
	}

	tests := []struct {
		name     string
		channel  string
		setup    func(channel string) error
		expected Settings
	}{
		{
			name:     "no settings",
			channel:  "none",
			expected: Settings{},
		},
		{
			name:    "empty settings",
			channel: "empty",
			setup: func(channel string) error {
				return s.Set(key(channel), `{}`)
			},
			expected: Settings{},
		},
		{
			name:    "correct answers only",
			channel: "correct_answers_only",
			setup: func(channel string) error {
				return s.Set(key(channel), `{"onlyAllowCorrectAnswers":true}`)
			},
			expected: Settings{OnlyAllowCorrectAnswers: true},
		},
		{
			name:    "clue visibility",
			channel: "clue_visibility",
			setup: func(channel string) error {
				return s.Set(key(channel), `{"cluesToShow":"down"}`)
			},
			expected: Settings{CluesToShow: OnlyDownCluesVisible},
		},
		{
			name:    "clue font size",
			channel: "clue_font_size",
			setup: func(channel string) error {
				return s.Set(key(channel), `{"clueFontSize":"large"}`)
			},
			expected: Settings{ClueFontSize: SizeLarge},
		},
		{
			name:    "gets correct settings",
			channel: "correct_key",
			setup: func(channel string) error {
				if err := s.Set(key("incorrect"), `{"clues_to_show":"all"}`); err != nil {
					return err
				}
				return s.Set(key(channel), `{"cluesToShow":"none"}`)
			},
			expected: Settings{CluesToShow: NoCluesVisible},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setup != nil {
				require.NoError(t, test.setup(test.channel))
			}

			actual, err := GetSettings(c, test.channel)
			require.NoError(t, err)
			assert.Equal(t, test.expected.OnlyAllowCorrectAnswers, actual.OnlyAllowCorrectAnswers)
			assert.Equal(t, test.expected.CluesToShow, actual.CluesToShow)
			assert.Equal(t, test.expected.ClueFontSize, actual.ClueFontSize)
		})
	}
}

func TestGetSettings_Error(t *testing.T) {
	s, err := miniredis.Run()
	require.NoError(t, err)

	c, err := redis.Dial("tcp", s.Addr())
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	key := func(s string) string {
		return fmt.Sprintf("settings:%s", s)
	}

	tests := []struct {
		name    string
		channel string
		setup   func(channel string) error
	}{
		{
			name:    "invalid json",
			channel: "invalid_json",
			setup: func(channel string) error {
				return s.Set(key(channel), `{`)
			},
		},
		{
			name:    "incorrect type",
			channel: "incorrect_type",
			setup: func(channel string) error {
				return s.Set(key(channel), `true`)
			},
		},
		{
			name:    "incorrect property type",
			channel: "incorrect_property_type",
			setup: func(channel string) error {
				return s.Set(key(channel), `{"cluesToShow":true}`)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setup != nil {
				require.NoError(t, test.setup(test.channel))
			}

			_, err := GetSettings(c, test.channel)
			require.Error(t, err)
		})
	}
}

func TestSetSettings(t *testing.T) {
	s, err := miniredis.Run()
	require.NoError(t, err)

	c, err := redis.Dial("tcp", s.Addr())
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	key := func(s string) string {
		return fmt.Sprintf("settings:%s", s)
	}

	tests := []struct {
		name     string
		channel  string
		setup    func(channel string) error
		settings Settings
	}{
		{
			name:     "no existing settings",
			channel:  "no_existing",
			settings: Settings{CluesToShow: OnlyAcrossCluesVisible},
		},
		{
			name:    "existing settings",
			channel: "existing",
			setup: func(channel string) error {
				return s.Set(key(channel), `{"clueFontSize":"xlarge"}`)
			},
			settings: Settings{CluesToShow: OnlyDownCluesVisible},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setup != nil {
				require.NoError(t, test.setup(test.channel))
			}

			assert.NoError(t, SetSettings(c, test.channel, test.settings))
			assert.True(t, s.Exists(key(test.channel)))
		})
	}
}

func TestSetSettings_Error(t *testing.T) {
	s, err := miniredis.Run()
	require.NoError(t, err)

	c, err := redis.Dial("tcp", s.Addr())
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	tests := []struct {
		name     string
		channel  string
		settings Settings
	}{
		{
			name:     "invalid settings",
			channel:  "invalid",
			settings: Settings{ClueFontSize: FontSize(17)},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Error(t, SetSettings(c, test.channel, test.settings))
		})
	}
}
