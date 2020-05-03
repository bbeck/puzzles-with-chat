package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bbeck/twitch-plays-crosswords/bot/sse"
)

// ChannelLocator is a SSE event processor that runs in the background to locate
// channels that should be joined for active puzzles.  As channels are
// discovered they are passed onto an update callback.
type ChannelLocator struct {
	url string
}

// ChannelsEvent is the SSE event that the ChannelLocator unmarshalls data into.
// It contains all of the fields representing which channels are solving a
// puzzle.
type ChannelsEvent struct {
	Kind    string `json:"kind"`
	Payload struct {
		Crosswords []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"crossword"`
		SpellingBees []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"spellingbee"`
	} `json:"payload"`
}

func NewChannelLocator(host string) *ChannelLocator {
	url := fmt.Sprintf("http://%s/api/channels", host)
	return &ChannelLocator{url: url}
}

type UpdateChannelsFunc func(map[ID][]string)
type FailFunc func(error)

func (l *ChannelLocator) Run(ctx context.Context, update UpdateChannelsFunc, fail FailFunc) {
	stream := sse.Open(ctx, l.url)

	for {
		select {
		case event := <-stream:
			channels, err := ProcessEvent(event.Data)
			if err != nil {
				fail(err)
				break
			}

			if channels != nil {
				update(channels)
			}

		case <-ctx.Done():
			return
		}
	}
}

func ProcessEvent(bs []byte) (map[ID][]string, error) {
	var event ChannelsEvent
	if err := json.Unmarshal(bs, &event); err != nil {
		err = fmt.Errorf("unable to parse json '%s': %+v", bs, err)
		return nil, err
	}

	switch event.Kind {
	case "channels":
		channels := make(map[ID][]string)
		for _, channel := range event.Payload.Crosswords {
			channels["crossword"] = append(channels["crossword"], channel.Name)
		}
		for _, channel := range event.Payload.SpellingBees {
			channels["spellingbee"] = append(channels["spellingbee"], channel.Name)
		}

		return channels, nil

	case "ping":
		return nil, nil

	default:
		return nil, fmt.Errorf("received unrecognized message: %s\n", bs)
	}
}
