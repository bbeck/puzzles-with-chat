package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bbeck/puzzles-with-chat/bot/sse"
)

// ChannelLocator is a SSE event processor that runs in the background to locate
// and keep track of channels that should have their messages processed.  As
// channels are discovered they are passed onto an update callback along with
// the state of their puzzle solve.
type ChannelLocator struct {
	url string
}

// Event is the SSE event that the ChannelLocator receives from the API about
// the active set of channels.
type Event struct {
	Kind    string          `json:"kind"`
	Payload json.RawMessage `json:"payload"`
}

// ChannelsPayload is the payload of an event that is sent out containing the
// current set of located channels organized by puzzle type ID.
type ChannelsPayload map[ID][]struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

func NewChannelLocator(host string) *ChannelLocator {
	url := fmt.Sprintf("http://%s/api/channels", host)
	return &ChannelLocator{url: url}
}

// Update represents an update of the status of a channel's integration.
type Update struct {
	Application ID
	Channel     string
	Status      string
}

type UpdateFunc func(updates []Update)
type FailFunc func(error)

// Run starts an infinite loop connecting to a SSE stream and processing
// received events.  Channels events are passed onto the update function while
// any errors encountered are passed onto the fail function.
func (l *ChannelLocator) Run(ctx context.Context, update UpdateFunc, fail FailFunc) {
	stream := sse.Open(ctx, l.url)

	for {
		select {
		case entry := <-stream:
			var event Event
			if err := json.Unmarshal(entry.Data, &event); err != nil {
				err = fmt.Errorf("unable to parse json '%s': %+v", entry.Data, err)
				fail(err)
				break
			}

			switch event.Kind {
			case "channels":
				var payload ChannelsPayload
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					err = fmt.Errorf("unable to parse payload '%s': %+v", event.Payload, err)
					fail(err)
					break
				}

				ProcessPayload(payload, update)

			case "ping":
				// do nothing

			default:
				err := fmt.Errorf("unrecognized event kind: %s", event.Kind)
				fail(err)
			}

		case <-ctx.Done():
			return
		}
	}
}

func ProcessPayload(payload ChannelsPayload, update UpdateFunc) {
	var updates []Update

	// Flatten the payload into individual updates
	for app, channels := range payload {
		for _, channel := range channels {
			updates = append(updates, Update{
				Application: app,
				Channel:     channel.Name,
				Status:      channel.Status,
			})
		}
	}

	update(updates)
}
