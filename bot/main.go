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

type Integration struct {
	ID             ID
	ChannelLocator ChannelLocator
	MessageHandler MessageHandler
}

// ChannelLocator represents an implementation that runs in the background to
// locate channels that a chat client should join in order to play a game.  As
// channels are discovered they are passed onto a channel manager.  If an error
// occurs while locating channels then the fail function is called.
type ChannelLocator interface {
	Run(ctx context.Context, update func([]string), fail func(error))
}

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

	var integrations = []Integration{
		{
			ID:             "crossword",
			ChannelLocator: crossword.NewChannelLocator(host),
			MessageHandler: crossword.NewMessageHandler(host),
		},
		{
			ID:             "spellingbee",
			ChannelLocator: spellingbee.NewChannelLocator(host),
			MessageHandler: spellingbee.NewMessageHandler(host),
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// The message router gets notified when the locator discovers channels and
	// routes messages from a client to the appropriate message handler(s).
	router := NewMessageRouter(integrations)

	// Create a new client that sends messages to each handler.
	client, err := NewClient(router)
	if err != nil {
		log.Fatalf("unable to create client: %v", err)
	}

	// The channel manager that will be used to keep track of which channels the
	// client should be monitoring.
	manager := ChannelManager{
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

	// Start each channel locator to discover which channels the client should
	// join.
	for _, integration := range integrations {
		id := integration.ID

		onUpdate := func(channels []string) {
			// Let the manager know the most recent set of channels for this
			// integration.
			manager.Update(id, channels)
		}

		onError := func(err error) {
			log.Printf("error while locating channels in integration %s: %+v", integration.ID, err)
		}

		go integration.ChannelLocator.Run(ctx, onUpdate, onError)
	}

	for {
		err := client.Connect()
		if err != nil && err != io.EOF {
			err = fmt.Errorf("received error from client connect: %w", err)
			log.Println(err)
		}
		time.Sleep(1 * time.Second)
	}
}
