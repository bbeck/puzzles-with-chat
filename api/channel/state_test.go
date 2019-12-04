package channel

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			require.Error(t, err)
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
			require.Error(t, err)
		})
	}
}

func TestDuration_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		duration Duration
		expected []byte
	}{
		{
			name:     "empty",
			duration: Duration{},
			expected: []byte(`"0s"`),
		},
		{
			name:     "1 second",
			duration: Duration{time.Second},
			expected: []byte(`"1s"`),
		},
		{
			name:     "1 minute",
			duration: Duration{time.Minute},
			expected: []byte(`"1m0s"`),
		},
		{
			name:     "1 hour",
			duration: Duration{time.Hour},
			expected: []byte(`"1h0m0s"`),
		},
		{
			name:     "24 hours",
			duration: Duration{24 * time.Hour},
			expected: []byte(`"24h0m0s"`),
		},
		{
			name:     "2 hours 12 minutes 9 seconds",
			duration: Duration{2*time.Hour + 12*time.Minute + 9*time.Second},
			expected: []byte(`"2h12m9s"`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := json.Marshal(test.duration)
			require.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestDuration_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		bs       []byte
		expected Duration
	}{
		{
			name:     "empty",
			bs:       []byte(`"0s"`),
			expected: Duration{},
		},
		{
			name:     "1 second",
			bs:       []byte(`"1s"`),
			expected: Duration{time.Second},
		},
		{
			name:     "1 minute",
			bs:       []byte(`"1m0s"`),
			expected: Duration{time.Minute},
		},
		{
			name:     "1 hour",
			bs:       []byte(`"1h0m0s"`),
			expected: Duration{time.Hour},
		},
		{
			name:     "24 hours",
			bs:       []byte(`"24h0m0s"`),
			expected: Duration{24 * time.Hour},
		},
		{
			name:     "2 hours 12 minutes 9 seconds",
			bs:       []byte(`"2h12m9s"`),
			expected: Duration{2*time.Hour + 12*time.Minute + 9*time.Second},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var actual Duration

			err := json.Unmarshal(test.bs, &actual)
			require.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestDuration_UnmarshalJSON_Error(t *testing.T) {
	tests := []struct {
		name string
		bs   []byte
	}{
		{
			name: "invalid type",
			bs:   []byte(`true`),
		},
		{
			name: "empty value",
			bs:   []byte(`""`),
		},
		{
			name: "incorrect value",
			bs:   []byte(`"1x2y"`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var actual Duration

			err := json.Unmarshal(test.bs, &actual)
			require.Error(t, err)
		})
	}
}
