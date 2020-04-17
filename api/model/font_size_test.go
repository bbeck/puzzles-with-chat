package model

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFontSize_String(t *testing.T) {
	tests := []struct {
		name     string
		size     FontSize
		expected string
	}{
		{
			name:     "normal",
			size:     FontSizeNormal,
			expected: "normal",
		},
		{
			name:     "large",
			size:     FontSizeLarge,
			expected: "large",
		},
		{
			name:     "xlarge",
			size:     FontSizeXLarge,
			expected: "xlarge",
		},
		{
			name:     "invalid",
			size:     FontSize(17),
			expected: "unknown",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := test.size.String()
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestFontSize_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		size     FontSize
		expected []byte
	}{
		{
			name:     "normal",
			size:     FontSizeNormal,
			expected: []byte(`"normal"`),
		},
		{
			name:     "large",
			size:     FontSizeLarge,
			expected: []byte(`"large"`),
		},
		{
			name:     "xlarge",
			size:     FontSizeXLarge,
			expected: []byte(`"xlarge"`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bs, err := json.Marshal(test.size)
			require.NoError(t, err)
			assert.Equal(t, test.expected, bs)
		})
	}
}

func TestFontSize_MarshalJSON_Error(t *testing.T) {
	tests := []struct {
		name string
		size FontSize
	}{
		{
			name: "invalid",
			size: FontSize(17),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := json.Marshal(test.size)
			require.Error(t, err)
		})
	}
}

func TestFontSize_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		bs       []byte
		expected FontSize
	}{
		{
			name:     "normal",
			bs:       []byte(`"normal"`),
			expected: FontSizeNormal,
		},
		{
			name:     "large",
			bs:       []byte(`"large"`),
			expected: FontSizeLarge,
		},
		{
			name:     "xlarge",
			bs:       []byte(`"xlarge"`),
			expected: FontSizeXLarge,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var actual FontSize

			err := json.Unmarshal(test.bs, &actual)
			require.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestFontSize_UnmarshalJSON_Error(t *testing.T) {
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
			var actual FontSize

			err := json.Unmarshal(test.bs, &actual)
			require.Error(t, err)
		})
	}
}
