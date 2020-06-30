package main

import (
	"context"
	"fmt"
	"github.com/bbeck/puzzles-with-chat/bot/acrostic"
	"github.com/bbeck/puzzles-with-chat/bot/crossword"
	"github.com/bbeck/puzzles-with-chat/bot/spellingbee"
	"io"
	"log"
	"os"
	"time"
)

type ID string

// A MessageHandler represents an implementation of a bot that processes chat
// messages from a client in order to play a game in a channel.
type MessageHandler interface {
	HandleChannelMessage(channel, status, message string)
}

func main() {
	host, ok := os.LookupEnv("API_HOST")
	if !ok {
		log.Fatal("missing API_HOST environment variable")
	}

	handlers := map[ID]MessageHandler{
		"acrostic":    acrostic.NewMessageHandler(host),
		"crossword":   crossword.NewMessageHandler(host),
		"spellingbee": spellingbee.NewMessageHandler(host),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// The message router gets notified whenever an integration is discovered or
	// changes status.  Then then uses this information to route messages received
	// from channels to the appropriate message handler(s).
	router := NewMessageRouter(handlers)

	// Create a new client that sends messages to the router.
	client, err := NewClient(router)
	if err != nil {
		log.Fatalf("unable to create client: %v", err)
	}

	// The channel monitor that will be used to keep track of which channels the
	// client should be monitoring and router should be sending messages to.
	monitor := ChannelMonitor{
		OnChannelAdded: func(channel string) {
			client.Join(channel)
			log.Printf("joined channel %s", channel)
		},
		OnChannelRemoved: func(channel string) {
			client.Depart(channel)
			log.Printf("parted channel %s", channel)
		},
		OnIntegrationAdded:   router.AddIntegration,
		OnIntegrationRemoved: router.RemoveIntegration,
		OnIntegrationUpdated: func(app ID, channel string, oldStatus, newStatus string) {
			router.UpdateIntegrationStatus(app, channel, newStatus)
		},
	}

	// Start the channel locator to receive any channel updates.
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
