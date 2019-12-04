package channel

import (
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetState(t *testing.T) {
	s, err := miniredis.Run()
	require.NoError(t, err)

	c, err := redis.Dial("tcp", s.Addr())
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	tests := []struct {
		name     string
		channel  string
		setup    func(channel string) error
		expected State
	}{
		{
			name:     "no state",
			channel:  "none",
			expected: State{},
		},
		{
			name:    "empty state",
			channel: "empty",
			setup: func(channel string) error {
				return s.Set(StateKey(channel), `{}`)
			},
			expected: State{},
		},
		{
			name:    "solving status",
			channel: "solving",
			setup: func(channel string) error {
				return s.Set(StateKey(channel), `{"status":"solving"}`)
			},
			expected: State{Status: StatusSolving},
		},
		{
			name:    "gets correct state",
			channel: "correct_key",
			setup: func(channel string) error {
				if err := s.Set(StateKey("incorrect"), `{"status":"created"}`); err != nil {
					return err
				}
				return s.Set(StateKey(channel), `{"status":"solving"}`)
			},
			expected: State{Status: StatusSolving},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setup != nil {
				require.NoError(t, test.setup(test.channel))
			}

			actual, err := GetState(c, test.channel)
			require.NoError(t, err)
			assert.Equal(t, test.expected.Status, actual.Status)
		})
	}
}

func TestGetState_UpdatesTTL(t *testing.T) {
	s, err := miniredis.Run()
	require.NoError(t, err)

	c, err := redis.Dial("tcp", s.Addr())
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	tests := []struct {
		name     string
		channel  string
		setup    func(channel string) error
		expected time.Duration
	}{
		{
			name:    "missing state",
			channel: "missing",
		},
		{
			name:    "existing_state",
			channel: "existing_state",
			setup: func(channel string) error {
				return s.Set(StateKey(channel), `{"status":"solving"}`)
			},
			expected: StateTTL,
		},
		{
			name:    "existing_ttl",
			channel: "existing_ttl",
			setup: func(channel string) error {
				if err := s.Set(StateKey(channel), `{"status":"solving"}`); err != nil {
					return err
				}

				s.SetTTL(StateKey(channel), 2*time.Hour)
				return nil
			},
			expected: StateTTL,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setup != nil {
				require.NoError(t, test.setup(test.channel))
			}

			_, err := GetState(c, test.channel)
			require.NoError(t, err)
			assert.Equal(t, test.expected, s.TTL(StateKey(test.channel)))
		})
	}
}

func TestGetState_Error(t *testing.T) {
	s, err := miniredis.Run()
	require.NoError(t, err)

	c, err := redis.Dial("tcp", s.Addr())
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	tests := []struct {
		name    string
		channel string
		setup   func(channel string) error
	}{
		{
			name:    "invalid json",
			channel: "invalid_json",
			setup: func(channel string) error {
				return s.Set(StateKey(channel), `{`)
			},
		},
		{
			name:    "incorrect redis type",
			channel: "incorrect_redis_type",
			setup: func(channel string) error {
				s.HSet(StateKey(channel), "field", "value")
				return nil
			},
		},
		{
			name:    "incorrect json type",
			channel: "incorrect_json_type",
			setup: func(channel string) error {
				return s.Set(StateKey(channel), `true`)
			},
		},
		{
			name:    "incorrect property type",
			channel: "incorrect_property_type",
			setup: func(channel string) error {
				return s.Set(StateKey(channel), `{"status":true}`)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setup != nil {
				require.NoError(t, test.setup(test.channel))
			}

			_, err := GetState(c, test.channel)
			require.Error(t, err)
		})
	}
}

func TestSetState(t *testing.T) {
	s, err := miniredis.Run()
	require.NoError(t, err)

	c, err := redis.Dial("tcp", s.Addr())
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	tests := []struct {
		name    string
		channel string
		setup   func(channel string) error
		state   *State
	}{
		{
			name:    "no existing state",
			channel: "no_existing",
			state:   &State{Status: StatusComplete},
		},
		{
			name:    "existing state",
			channel: "existing",
			setup: func(channel string) error {
				return s.Set(StateKey(channel), `{"status": "solving"}`)
			},
			state: &State{Status: StatusComplete},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setup != nil {
				require.NoError(t, test.setup(test.channel))
			}

			assert.NoError(t, SetState(c, test.channel, test.state))
			assert.True(t, s.Exists(StateKey(test.channel)))
			assert.Equal(t, StateTTL, s.TTL(StateKey(test.channel)))
		})
	}

	fmt.Println(s.Dump())
}

func TestSetState_Error(t *testing.T) {
	s, err := miniredis.Run()
	require.NoError(t, err)

	c, err := redis.Dial("tcp", s.Addr())
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	tests := []struct {
		name    string
		channel string
		state   *State
	}{
		{
			name:    "nil state",
			channel: "nil",
		},
		{
			name:    "invalid state status",
			channel: "invalid_status",
			state:   &State{Status: Status(17)},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Error(t, SetState(c, test.channel, test.state))
		})
	}
}
