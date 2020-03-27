package crossword

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bbeck/twitch-plays-crosswords/bot/sse"
)

type ChannelLocator struct {
	url string
}

type ChannelsMessage struct {
	Kind     string   `json:"kind"`
	Channels []string `json:"payload"`
}

func NewChannelLocator(host string) *ChannelLocator {
	url := fmt.Sprintf("http://%s/api/crossword/events", host)
	return &ChannelLocator{url: url}
}

func (l *ChannelLocator) Run(ctx context.Context, update func([]string), fail func(error)) {
	events := sse.Open(ctx, l.url)
	for {
		select {
		case event := <-events:
			var message ChannelsMessage
			if err := json.Unmarshal(event.Data, &message); err != nil {
				err = fmt.Errorf("unable to parse json '%s': %+v", event.Data, err)
				fail(err)
				continue
			}

			if message.Kind != "channels" {
				err := fmt.Errorf("received non-channels message: %+v\n", message)
				fail(err)
				continue
			}

			update(message.Channels)

		case <-ctx.Done():
			return
		}
	}
}
