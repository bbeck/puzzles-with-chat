package model

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStatus_String(t *testing.T) {
	tests := []struct {
		name     string
		state    Status
		expected string
	}{
		{
			name:     "created",
			state:    StatusCreated,
			expected: "created",
		},
		{
			name:     "selected",
			state:    StatusSelected,
			expected: "selected",
		},
		{
			name:     "paused",
			state:    StatusPaused,
			expected: "paused",
		},
		{
			name:     "solving",
			state:    StatusSolving,
			expected: "solving",
		},
		{
			name:     "complete",
			state:    StatusComplete,
			expected: "complete",
		},
		{
			name:     "invalid",
			state:    Status(17),
			expected: "unknown",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := test.state.String()
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestStatus_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		state    Status
		expected []byte
	}{
		{
			name:     "created",
			state:    StatusCreated,
			expected: []byte(`"created"`),
		},
		{
			name:     "selected",
			state:    StatusSelected,
			expected: []byte(`"selected"`),
		},
		{
			name:     "paused",
			state:    StatusPaused,
			expected: []byte(`"paused"`),
		},
		{
			name:     "solving",
			state:    StatusSolving,
			expected: []byte(`"solving"`),
		},
		{
			name:     "complete",
			state:    StatusComplete,
			expected: []byte(`"complete"`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bs, err := json.Marshal(test.state)
			require.NoError(t, err)
			assert.Equal(t, test.expected, bs)
		})
	}
}

func TestStatus_MarshalJSON_Error(t *testing.T) {
	tests := []struct {
		name  string
		state Status
	}{
		{
			name:  "invalid",
			state: Status(17),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := json.Marshal(test.state)
			assert.Error(t, err)
		})
	}
}

func TestStatus_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		bs       []byte
		expected Status
	}{
		{
			name:     "created",
			bs:       []byte(`"created"`),
			expected: StatusCreated,
		},
		{
			name:     "selected",
			bs:       []byte(`"selected"`),
			expected: StatusSelected,
		},
		{
			name:     "paused",
			bs:       []byte(`"paused"`),
			expected: StatusPaused,
		},
		{
			name:     "solving",
			bs:       []byte(`"solving"`),
			expected: StatusSolving,
		},
		{
			name:     "complete",
			bs:       []byte(`"complete"`),
			expected: StatusComplete,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var actual Status

			err := json.Unmarshal(test.bs, &actual)
			require.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestStatus_UnmarshalJSON_Error(t *testing.T) {
	tests := []struct {
		name string
		bs   []byte
	}{
		{
			name: "wrong type",
			bs:   []byte(`true`),
		},
		{
			name: "empty value",
			bs:   []byte(`""`),
		},
		{
			name: "invalid value",
			bs:   []byte(`"asdf"`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var actual Status

			err := json.Unmarshal(test.bs, &actual)
			assert.Error(t, err)
		})
	}
}
