package main

import (
	"context"
	"fmt"
	"github.com/bbeck/twitch-plays-crosswords/bot/crossword"
	"io"
	"log"
	"os"
	"time"
)

type Integration struct {
	ID             string
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
	HandleChannelMessage(channel, userid, username, message string)
}

func main() {
	host, ok := os.LookupEnv("API_HOST")
	if !ok {
		log.Fatal("missing API_HOST environment variable")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var integrations = []Integration{
		{
			ID:             "crossword",
			ChannelLocator: crossword.NewChannelLocator(host),
			MessageHandler: crossword.NewMessageHandler(host),
		},
	}

	var handlers []MessageHandler
	for _, integration := range integrations {
		handlers = append(handlers, integration.MessageHandler)
	}

	// Create a new client that sends messages to each handler.
	client, err := NewClient(handlers)
	if err != nil {
		log.Fatalf("unable to create client: %v", err)
	}

	var manager ChannelManager
	manager.OnAddChannel = func(channel string) {
		client.Join(channel)
		log.Printf("joined channel %s", channel)
	}
	manager.OnRemoveChannel = func(channel string) {
		client.Depart(channel)
		log.Printf("parted channel %s", channel)
	}

	// Start the channel locators to discover which channels should be joined.
	for _, integration := range integrations {
		onError := func(err error) {
			log.Printf("error while locating channels in integration %s: %+v", integration.ID, err)
		}

		go integration.ChannelLocator.Run(ctx, manager.Update, onError)
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
