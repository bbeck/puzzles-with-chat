package sse

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Event struct {
	Name string
	ID   string
	Data []byte
}

// DefaultSSEClient is the default client to use when connecting to a SSE
// endpoint.  It's configured with no timeout so that it can stay connected
// indefinitely receiving events.
var DefaultSSEClient = &http.Client{Timeout: 0}

// The amount of time the client should wait between reconnecting to the SSE
// endpoint.
var ReconnectDelay = 1 * time.Second

func Open(ctx context.Context, url string) <-chan Event {
	return OpenWithClient(ctx, DefaultSSEClient, url)
}

func OpenWithClient(ctx context.Context, client *http.Client, url string) <-chan Event {
	events := make(chan Event, 10)
	go func() {
		for {
			err := RunOnce(ctx, client, url, events)

			// If the context was canceled then we're done and should exit.
			if errors.Is(err, context.Canceled) {
				close(events)
				break
			}

			// Otherwise, log the error and go again.
			if err != nil {
				log.Printf("received error from RunOnce: %+v", err)
				<-time.After(ReconnectDelay)
			}
		}
	}()

	return events
}

// RunOnce will use the provided HTTP client to connect to the specified URL and
// process the resulting response as a Server-Sent event stream.  Currently only
// an HTTP status code of 200 is allowed.
func RunOnce(ctx context.Context, client *http.Client, url string, events chan<- Event) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != 200 {
		return fmt.Errorf("received %d response for url: %s", response.StatusCode, url)
	}

	return ReadEvents(response.Body, events)
}

// ReadEvents parses and interprets an event stream according to the W3C working
// draft for Server-Sent Events found at:
// https://www.w3.org/TR/2009/WD-eventsource-20090421.
// If a non-EOF error occurs while reading data then it is returned.
func ReadEvents(in io.Reader, events chan<- Event) error {
	// Buffer the body so that we can read line by line.
	reader := bufio.NewReader(in)

	// Keep processing messages until we get an EOF or error from the reader.
	var event Event
	for {
		line, err := reader.ReadBytes('\n')

		switch {
		case err == io.EOF:
			// Make sure to dispatch the last event if we've parsed any fields for it.
			if event.ID != "" || event.Name != "" || event.Data != nil {
				events <- event
			}

			return nil
		case err != nil:
			return err
		}

		switch {
		case bytes.HasPrefix(line, []byte{':'}):
			break

		case bytes.HasPrefix(line, []byte("id:")):
			event.ID = string(trim(line[3:]))

		case bytes.HasPrefix(line, []byte("event:")):
			event.Name = string(trim(line[6:]))

		case bytes.HasPrefix(line, []byte("data:")):
			if event.Data == nil {
				event.Data = trim(line[5:])
				break
			}

			event.Data = append(event.Data, '\n')
			event.Data = append(event.Data, trim(line[5:])...)

		case bytes.Equal(line, []byte("\n")) || bytes.Equal(line, []byte("\r")) || bytes.Equal(line, []byte("\r\n")):
			events <- event
			event = Event{}
		}
	}
}

// Trim a string according to the W3C working draft for Server-Sent Events.  A
// single leading space should be removed from a field value as well as a single
// trailing newline character.
func trim(bs []byte) []byte {
	if bs[0] == ' ' {
		bs = bs[1:]
	}

	L := len(bs)
	if bs[L-1] == '\n' {
		bs = bs[:L-1]
	}

	return bs
}
