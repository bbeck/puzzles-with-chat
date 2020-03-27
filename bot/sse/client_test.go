package sse

import (
	"bytes"
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/iotest"
	"time"
)

func TestOpen(t *testing.T) {
	tests := []struct {
		name       string
		responders []func(context.CancelFunc, http.ResponseWriter)
		verify     func(t *testing.T, connectionCount int, events []Event)
	}{
		{
			name: "reconnects when server closes connection",
			responders: []func(context.CancelFunc, http.ResponseWriter){
				func(cancel context.CancelFunc, w http.ResponseWriter) {
					w.WriteHeader(200)
					write(w, Event{Name: "name", Data: []byte("data1")})
				},
				func(cancel context.CancelFunc, w http.ResponseWriter) {
					w.WriteHeader(200)
					write(w, Event{Name: "name", Data: []byte("data2")})
				},
			},
			verify: func(t *testing.T, count int, events []Event) {
				expected := []Event{
					{Name: "name", Data: []byte("data1")},
					{Name: "name", Data: []byte("data2")},
				}

				assert.Equal(t, 3, count) // Our 2 connections plus the last one.
				assert.ElementsMatch(t, expected, events)
			},
		},
		{
			name: "reconnects when error happens and context is not canceled",
			responders: []func(context.CancelFunc, http.ResponseWriter){
				func(cancel context.CancelFunc, w http.ResponseWriter) {
					w.WriteHeader(404)
				},
				func(cancel context.CancelFunc, w http.ResponseWriter) {
					w.WriteHeader(200)
					write(w, Event{Name: "name", Data: []byte("data")})
				},
			},
			verify: func(t *testing.T, count int, events []Event) {
				expected := []Event{
					{Name: "name", Data: []byte("data")},
				}

				assert.Equal(t, 3, count)
				assert.ElementsMatch(t, expected, events)
			},
		},
		{
			name: "does not reconnect when context is canceled",
			responders: []func(context.CancelFunc, http.ResponseWriter){
				func(cancel context.CancelFunc, w http.ResponseWriter) {
					w.WriteHeader(200)
					cancel()
				},
			},
			verify: func(t *testing.T, count int, events []Event) {
				assert.Equal(t, 1, count)
				assert.Empty(t, events)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var connectionCount int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				connectionCount++
				if connectionCount <= len(test.responders) {
					test.responders[connectionCount-1](cancel, w)
				} else {
					cancel()
					w.WriteHeader(200)
				}
			}))
			defer server.Close()

			oldReconnectDelay := ReconnectDelay
			ReconnectDelay = 1 * time.Millisecond
			defer func() { ReconnectDelay = oldReconnectDelay }()

			time.AfterFunc(100*time.Millisecond, cancel)

			c := Open(ctx, server.URL)
			<-ctx.Done()

			events := drain(c)
			test.verify(t, connectionCount, events)
		})
	}
}

func TestRunOnce(t *testing.T) {
	tests := []struct {
		name   string
		events []Event
	}{
		{
			name: "no events",
		},
		{
			name: "one event",
			events: []Event{
				{Name: "name", Data: []byte("data")},
			},
		},
		{
			name: "multiple events",
			events: []Event{
				{Name: "name1", Data: []byte("data1")},
				{Name: "name2", Data: []byte("data2")},
			},
		},
		{
			name: "just id",
			events: []Event{
				{ID: "id"},
			},
		},
		{
			name: "just name",
			events: []Event{
				{Name: "name"},
			},
		},
		{
			name: "just data",
			events: []Event{
				{Data: []byte("data")},
			},
		},
	}

	client := &http.Client{Timeout: 50 * time.Millisecond}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)

				for _, event := range test.events {
					write(w, event)
				}
			}))
			defer server.Close()

			time.AfterFunc(100*time.Millisecond, func() {
				cancel()
			})

			c := make(chan Event, 10)
			err := RunOnce(ctx, client, server.URL, c)
			assert.NoError(t, err)

			close(c)

			events := drain(c)
			assert.ElementsMatch(t, test.events, events)
		})
	}
}

func TestRunOnce_Error(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		client  *http.Client
		respond func(http.ResponseWriter)
	}{
		{
			name: "error creating request (bad url)",
			url:  ":",
		},
		{
			name:   "error in client.Do (timeout)",
			client: &http.Client{Timeout: 1 * time.Millisecond},
			respond: func(writer http.ResponseWriter) {
				time.Sleep(10 * time.Millisecond)
				writer.WriteHeader(200)
			},
		},
		{
			name: "non-200 response",
			respond: func(writer http.ResponseWriter) {
				writer.WriteHeader(404)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				test.respond(w)
			}))
			defer server.Close()

			var url = server.URL
			if test.url != "" {
				url = test.url
			}

			var client = DefaultSSEClient
			if test.client != nil {
				client = test.client
			}

			c := make(chan Event, 10)
			err := RunOnce(context.Background(), client, url, c)
			assert.Error(t, err)

			close(c)
		})
	}
}

func TestReadEvents(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		expected []Event
	}{
		{
			name: "no events",
		},
		{
			name: "one event",
			lines: []string{
				"event: name",
				"data: data",
				"",
			},
			expected: []Event{
				{Name: "name", Data: []byte("data")},
			},
		},
		{
			name: "multiple events",
			lines: []string{
				"event: name1",
				"data: data1",
				"",
				"event: name2",
				"data: data2",
				"",
			},
			expected: []Event{
				{Name: "name1", Data: []byte("data1")},
				{Name: "name2", Data: []byte("data2")},
			},
		},
		{
			name: "just id",
			lines: []string{
				"id: id",
				"",
			},
			expected: []Event{
				{ID: "id"},
			},
		},
		{
			name: "just name",
			lines: []string{
				"event: name",
				"",
			},
			expected: []Event{
				{Name: "name"},
			},
		},
		{
			name: "just data",
			lines: []string{
				"data: data",
				"",
			},
			expected: []Event{
				{Data: []byte("data")},
			},
		},
		{
			name: "no space after colon",
			lines: []string{
				"id:id",
				"event:name",
				"data:data",
				"",
			},
			expected: []Event{
				{ID: "id", Name: "name", Data: []byte("data")},
			},
		},
		{
			name: "data with embedded newline",
			lines: []string{
				"event: name",
				"data: first",
				"data: second",
				"data: third",
				"",
			},
			expected: []Event{
				{Name: "name", Data: []byte("first\nsecond\nthird")},
			},
		},
		{
			name: "last event",
			lines: []string{
				"event: name",
				"data: data",
				// no extra newline on purpose
			},
			expected: []Event{
				{Name: "name", Data: []byte("data")},
			},
		},
		{
			name: "ignored lines",
			lines: []string{
				"event: name",
				": ignored completely",
				"data: data",
				"",
			},
			expected: []Event{
				{Name: "name", Data: []byte("data")},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var in bytes.Buffer
			for _, line := range test.lines {
				_, err := in.WriteString(line)
				require.NoError(t, err)

				_, err = in.Write([]byte{'\n'})
				require.NoError(t, err)
			}

			c := make(chan Event, 10)
			err := ReadEvents(&in, c)
			assert.NoError(t, err)
			close(c)

			events := drain(c)
			assert.Equal(t, test.expected, events)
		})
	}
}

func TestReadEvents_Error(t *testing.T) {
	in := iotest.TimeoutReader(strings.NewReader("data:"))
	c := make(chan Event, 10)
	err := ReadEvents(in, c)
	assert.Equal(t, iotest.ErrTimeout, err)
	close(c)

	events := drain(c)
	assert.Empty(t, events)
}

func write(w io.Writer, event Event) {
	if event.ID != "" {
		w.Write([]byte(fmt.Sprintf("id:%s\n", event.ID)))
	}
	if event.Name != "" {
		w.Write([]byte(fmt.Sprintf("event:%s\n", event.Name)))
	}
	if event.Data != nil {
		w.Write([]byte(fmt.Sprintf("data:%s\n", event.Data)))
	}
	w.Write([]byte{'\n'})
}

func drain(c <-chan Event) []Event {
	var events []Event
	for {
		event, ok := <-c
		if !ok {
			break
		}

		events = append(events, event)
	}

	return events
}
