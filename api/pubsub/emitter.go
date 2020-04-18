package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

var PingEvent = Event{Kind: "ping"}

// EmitEvents will loop and send events to the provided HTTP response.  The
// events will be formatted according to the W3C working draft for Server-Sent
// Events found at: https://www.w3.org/TR/2009/WD-eventsource-20090421.  This
// method is useful for implementing the server side component of a SSE
// endpoint.
//
// EmitEvents will block until either the events channel is closed, the
// provided context is done, or an error occurs while emitting an event.
//
// If no events are available on the events channel for 30 seconds then a ping
// event will be synthesized and emitted automatically in order to keep the
// connection with the client alive.
func EmitEvents(ctx context.Context, w http.ResponseWriter, events <-chan Event) {
	w.Header().Set("Cache-Control", "no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	for {
		select {
		case <-ctx.Done():
			return

		case msg, ok := <-events:
			if !ok {
				return
			}

			if err := EmitEvent(w, msg); err != nil {
				return
			}

		case <-time.After(30 * time.Second):
			if err := EmitEvent(w, PingEvent); err != nil {
				return
			}
		}
	}
}

// EmitEvent marshals an event to JSON and sends it as a SSE message to the
// provided io.Writer.  If the provided io.Writer implements the http.Flusher
// interface than the writer will be flushed after the write occurs.
func EmitEvent(w io.Writer, event Event) error {
	bs, err := json.Marshal(event)
	if err != nil {
		log.Printf("unable to marshal event '%+v' to json: %+v\n", event, err)
		return err
	}

	if _, err := fmt.Fprintf(w, "event:message\ndata:%s\n\n", bs); err != nil {
		log.Printf("error while writing message to http.ResponseWriter: %+v", err)
		return err
	}

	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	return nil
}
