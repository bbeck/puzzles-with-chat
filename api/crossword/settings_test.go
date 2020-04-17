package crossword

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClueVisibility_String(t *testing.T) {
	tests := []struct {
		name       string
		visibility ClueVisibility
		expected   string
	}{
		{
			name:       "all",
			visibility: AllCluesVisible,
			expected:   "all",
		},
		{
			name:       "none",
			visibility: NoCluesVisible,
			expected:   "none",
		},
		{
			name:       "down",
			visibility: OnlyDownCluesVisible,
			expected:   "down",
		},
		{
			name:       "across",
			visibility: OnlyAcrossCluesVisible,
			expected:   "across",
		},
		{
			name:       "invalid",
			visibility: ClueVisibility(17),
			expected:   "unknown",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := test.visibility.String()
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestClueVisibility_MarshalJSON(t *testing.T) {
	tests := []struct {
		name       string
		visibility ClueVisibility
		expected   []byte
	}{
		{
			name:       "all",
			visibility: AllCluesVisible,
			expected:   []byte(`"all"`),
		},
		{
			name:       "none",
			visibility: NoCluesVisible,
			expected:   []byte(`"none"`),
		},
		{
			name:       "down",
			visibility: OnlyDownCluesVisible,
			expected:   []byte(`"down"`),
		},
		{
			name:       "across",
			visibility: OnlyAcrossCluesVisible,
			expected:   []byte(`"across"`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bs, err := json.Marshal(test.visibility)
			require.NoError(t, err)
			assert.Equal(t, test.expected, bs)
		})
	}
}

func TestClueVisibility_MarshalJSON_Error(t *testing.T) {
	tests := []struct {
		name       string
		visibility ClueVisibility
	}{
		{
			name:       "invalid",
			visibility: ClueVisibility(17),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := json.Marshal(test.visibility)
			require.Error(t, err)
		})
	}
}

func TestClueVisibility_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		bs       []byte
		expected ClueVisibility
	}{
		{
			name:     "all",
			bs:       []byte(`"all"`),
			expected: AllCluesVisible,
		},
		{
			name:     "none",
			bs:       []byte(`"none"`),
			expected: NoCluesVisible,
		},
		{
			name:     "down",
			bs:       []byte(`"down"`),
			expected: OnlyDownCluesVisible,
		},
		{
			name:     "across",
			bs:       []byte(`"across"`),
			expected: OnlyAcrossCluesVisible,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var actual ClueVisibility

			err := json.Unmarshal(test.bs, &actual)
			require.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestClueVisibility_UnmarshalJSON_Error(t *testing.T) {
	tests := []struct {
		name string
		bs   []byte
	}{
		{
			name: "invalid json",
			bs:   []byte(`false`),
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
			var actual ClueVisibility

			err := json.Unmarshal(test.bs, &actual)
			require.Error(t, err)
		})
	}
}
