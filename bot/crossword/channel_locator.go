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
	url := fmt.Sprintf("http://%s/api/crossword/channels", host)
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

			switch message.Kind {
			case "channels":
				update(message.Channels)

			case "ping":
				break

			default:
				err := fmt.Errorf("received unrecognized message: %+v\n", message)
				fail(err)
			}

		case <-ctx.Done():
			return
		}
	}
}
