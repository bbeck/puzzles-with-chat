package main

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestChannelManager_Diff(t *testing.T) {
	tests := []struct {
		name            string
		initial         []string
		poll            []string
		expectedAdded   []string
		expectedRemoved []string
	}{
		{
			name: "nil channels",
			poll: nil,
		},
		{
			name: "empty channels",
			poll: []string{},
		},
		{
			name:          "one added channel",
			poll:          []string{"a"},
			expectedAdded: []string{"a"},
		},
		{
			name:          "multiple added channels",
			poll:          []string{"a", "b"},
			expectedAdded: []string{"a", "b"},
		},
		{
			name:            "one removed channel",
			initial:         []string{"a"},
			poll:            []string{},
			expectedRemoved: []string{"a"},
		},
		{
			name:            "multiple removed channels",
			initial:         []string{"a", "b"},
			poll:            []string{},
			expectedRemoved: []string{"a", "b"},
		},
		{
			name:            "one added and removed channel",
			initial:         []string{"a", "b"},
			poll:            []string{"b", "c"},
			expectedAdded:   []string{"c"},
			expectedRemoved: []string{"a"},
		},
		{
			name:            "multiple added and removed channel",
			initial:         []string{"a", "b"},
			poll:            []string{"c", "d"},
			expectedAdded:   []string{"c", "d"},
			expectedRemoved: []string{"a", "b"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manager := NewChannelManager(nil)
			for _, channel := range test.initial {
				manager.channels[channel] = struct{}{}
			}

			added, removed := manager.Diff(test.poll)
			assert.ElementsMatch(t, test.expectedAdded, added)
			assert.ElementsMatch(t, test.expectedRemoved, removed)
		})
	}
}

func TestChannelManager_Run(t *testing.T) {
	tests := []struct {
		name            string
		initial         []string
		poll            func() ([]string, error)
		expectedAdded   []string
		expectedRemoved []string
		expectedError   error
	}{
		{
			name:          "one added channel",
			poll:          func() ([]string, error) { return []string{"a"}, nil },
			expectedAdded: []string{"a"},
		},
		{
			name:          "multiple added channels",
			poll:          func() ([]string, error) { return []string{"a", "b"}, nil },
			expectedAdded: []string{"a", "b"},
		},
		{
			name:            "one removed channel",
			initial:         []string{"a"},
			poll:            func() ([]string, error) { return []string{}, nil },
			expectedRemoved: []string{"a"},
		},
		{
			name:            "multiple removed channels",
			initial:         []string{"a", "b"},
			poll:            func() ([]string, error) { return []string{}, nil },
			expectedRemoved: []string{"a", "b"},
		},
		{
			name:            "one added and removed channel",
			initial:         []string{"a", "b"},
			poll:            func() ([]string, error) { return []string{"b", "c"}, nil },
			expectedAdded:   []string{"c"},
			expectedRemoved: []string{"a"},
		},
		{
			name:            "multiple added and removed channel",
			initial:         []string{"a", "b"},
			poll:            func() ([]string, error) { return []string{"c", "d"}, nil },
			expectedAdded:   []string{"c", "d"},
			expectedRemoved: []string{"a", "b"},
		},
		{
			name:          "poll returns error",
			poll:          func() ([]string, error) { return nil, errors.New("forced error") },
			expectedError: errors.New("forced error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var manager *ChannelManager

			poll := func() ([]string, error) {
				// Ensure that we only run a single time
				manager.Stop()
				return test.poll()
			}

			// Setup a count down latch with the number of expected calls to handlers.
			var count int
			count += len(test.expectedAdded)
			count += len(test.expectedRemoved)
			if test.expectedError != nil {
				count++
			}

			latch := NewCountDownLatch(count)

			var added []string
			onAdd := func(channel string) {
				added = append(added, channel)
				latch.CountDown()
			}

			var removed []string
			onRemove := func(channel string) {
				removed = append(removed, channel)
				latch.CountDown()
			}

			var err error
			onError := func(e error) {
				err = e
				latch.CountDown()
			}

			// Setup the channel manager with our handlers.
			manager = NewChannelManager(poll)
			manager.OnAddChannel = onAdd
			manager.OnRemoveChannel = onRemove
			manager.OnError = onError
			for _, channel := range test.initial {
				manager.channels[channel] = struct{}{}
			}
			manager.Run(time.Millisecond)

			success := latch.Wait(50 * time.Millisecond)
			assert.True(t, success)
			assert.ElementsMatch(t, test.expectedAdded, added)
			assert.ElementsMatch(t, test.expectedRemoved, removed)
			assert.Equal(t, test.expectedError, err)
		})
	}
}
