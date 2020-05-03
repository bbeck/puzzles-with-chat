package main

import (
	"context"
	"fmt"
	"github.com/bbeck/twitch-plays-crosswords/bot/crossword"
	"github.com/bbeck/twitch-plays-crosswords/bot/spellingbee"
	"io"
	"log"
	"os"
	"time"
)

type ID string

// A MessageHandler represents an implementation of a bot that processes chat
// messages from a client in order to play a game in a channel.
type MessageHandler interface {
	HandleChannelMessage(channel, userID, username, message string)
}

func main() {
	host, ok := os.LookupEnv("API_HOST")
	if !ok {
		log.Fatal("missing API_HOST environment variable")
	}

	handlers := map[ID]MessageHandler{
		"crossword":   crossword.NewMessageHandler(host),
		"spellingbee": spellingbee.NewMessageHandler(host),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// The message router gets notified when the locator discovers channels and
	// routes messages from a client to the appropriate message handler(s).
	router := NewMessageRouter(handlers)

	// Create a new client that sends messages to each handler.
	client, err := NewClient(router)
	if err != nil {
		log.Fatalf("unable to create client: %v", err)
	}

	// The channel monitor that will be used to keep track of which channels the
	// client should be monitoring and router should be sending messages to.
	monitor := ChannelMonitor{
		OnAddChannel: func(channel string) {
			client.Join(channel)
			log.Printf("joined channel %s", channel)
		},
		OnRemoveChannel: func(channel string) {
			client.Depart(channel)
			log.Printf("parted channel %s", channel)
		},
		OnUpdateChannels: router.UpdateChannels,
	}

	// Start the channel locator to discover which channels the client should
	// join.
	locator := NewChannelLocator(host)
	onError := func(err error) {
		log.Printf("error while locating channels: %+v", err)
	}
	go locator.Run(ctx, monitor.Update, onError)

	for {
		err := client.Connect()
		if err != nil && err != io.EOF {
			err = fmt.Errorf("received error from client connect: %w", err)
			log.Println(err)
		}
		time.Sleep(1 * time.Second)
	}
}
