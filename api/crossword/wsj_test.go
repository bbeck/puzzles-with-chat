package crossword

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sort"
	"testing"
	"time"
)

func TestLoadAvailableWSJDates(t *testing.T) {
	tests := []struct {
		name     string
		expected time.Time
	}{
		{
			name:     "2013-01-04",
			expected: time.Date(2013, time.January, 4, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "2014-01-03",
			expected: time.Date(2014, time.January, 3, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "2015-01-02",
			expected: time.Date(2015, time.January, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "2016-01-02",
			expected: time.Date(2016, time.January, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "2017-01-03",
			expected: time.Date(2017, time.January, 3, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "2018-01-02",
			expected: time.Date(2018, time.January, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "2019-01-02",
			expected: time.Date(2019, time.January, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "2020-01-02",
			expected: time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC),
		},
	}

	dates := LoadAvailableWSJDates()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.True(t, sort.SliceIsSorted(dates, func(i, j int) bool {
				return dates[i].Before(dates[j])
			}))

			index := sort.Search(len(dates), func(i int) bool {
				return dates[i].Equal(test.expected) || dates[i].After(test.expected)
			})
			assert.Equal(t, test.expected, dates[index])
		})
	}
}
