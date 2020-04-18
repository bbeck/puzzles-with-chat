package pubsub

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http/httptest"
	"testing"
	"time"
)

const nl = "\n"

func TestEmitEvent(t *testing.T) {
	tests := []struct {
		name     string
		event    Event
		expected []byte
	}{
		{
			name:     "ping event",
			event:    PingEvent,
			expected: []byte(`event:message` + nl + `data:{"kind":"ping"}` + nl + nl),
		},
		{
			name:     "normal event",
			event:    Event{Kind: "kind", Payload: "payload"},
			expected: []byte(`event:message` + nl + `data:{"kind":"kind","payload":"payload"}` + nl + nl),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			err := EmitEvent(w, test.event)
			require.NoError(t, err)
			assert.Equal(t, test.expected, w.Body.Bytes())
		})
	}
}

func TestEmitEvent_Error(t *testing.T) {
	tests := []struct {
		name  string
		w     io.Writer
		event Event
	}{
		{
			name:  "json.Marshal error",
			event: Event{"kind", make(chan int)}, // channels cannot be converted to JSON
		},
		{
			name: "io.Writer error",
			w:    ErrWriter{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := EmitEvent(test.w, test.event)
			require.Error(t, err)
		})
	}
}

func TestEmitEvents(t *testing.T) {
	w := httptest.NewRecorder()

	latch := NewCountDownLatch(1)
	go func() {
		events := make(chan Event, 10)
		events <- Event{Kind: "event-1"}
		events <- Event{Kind: "event-2"}
		events <- Event{Kind: "event-3"}
		close(events)

		EmitEvents(context.Background(), w, events)
		latch.CountDown()
	}()

	assert.True(t, latch.Wait(100*time.Millisecond))

	expected := []byte(`` +
		`event:message` + nl + `data:{"kind":"event-1"}` + nl + nl +
		`event:message` + nl + `data:{"kind":"event-2"}` + nl + nl +
		`event:message` + nl + `data:{"kind":"event-3"}` + nl + nl)
	assert.Equal(t, expected, w.Body.Bytes())
}

func TestEmitEvents_ContextDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := httptest.NewRecorder()

	latch := NewCountDownLatch(1)
	go func() {
		// Cancel the context immediately, this should terminate the loop.
		cancel()

		EmitEvents(ctx, w, make(chan Event))
		latch.CountDown()
	}()

	assert.True(t, latch.Wait(100*time.Millisecond))
	assert.Empty(t, w.Body.Bytes())
}

func TestEmitEvents_ClosedEventChannel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := httptest.NewRecorder()

	latch := NewCountDownLatch(1)
	go func() {
		// Close the channel immediately, this should terminate the loop.
		events := make(chan Event)
		close(events)

		EmitEvents(ctx, w, events)
		latch.CountDown()
	}()

	assert.True(t, latch.Wait(100*time.Millisecond))
	assert.Empty(t, w.Body.Bytes())
}

func TestEmitEvents_ErrorEmittingEvent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := httptest.NewRecorder()

	latch := NewCountDownLatch(1)
	go func() {
		events := make(chan Event, 1)

		// channels cannot be converted to JSON
		events <- Event{Kind: "kind", Payload: make(chan int)}
		close(events)

		EmitEvents(ctx, w, events)
		latch.CountDown()
	}()

	assert.True(t, latch.Wait(100*time.Millisecond))
	assert.Empty(t, w.Body.Bytes())
}
