package main

import (
	"fmt"
	"github.com/bbeck/twitch-plays-crosswords/bot/crossword"
	"log"
	"time"
)

func main() {
	var integrations []Integration

	crosswords, err := crossword.NewIntegration()
	if err != nil {
		log.Fatalf("unable to create crossword integration: %v", err)
	}
	integrations = append(integrations, crosswords)

	client, err := NewClient(integrations)
	if err != nil {
		log.Fatalf("unable to create client: %v", err)
	}

	manager := NewChannelManager(func() ([]string, error) {
		var channels []string
		for _, integration := range integrations {
			cs, err := integration.GetActiveChannelNames()
			if err != nil {
				return nil, err
			}

			channels = append(channels, cs...)
		}

		return channels, nil
	})
	manager.OnAddChannel = func(channel string) {
		client.Join(channel)
		log.Printf("Joined channel %s\n", channel)
	}
	manager.OnRemoveChannel = func(channel string) {
		client.Depart(channel)
		log.Printf("Parted channel %s\n", channel)
	}
	manager.Run(10 * time.Second)

	for {
		log.Println("Connecting...")
		err := client.Connect()
		if err != nil {
			err = fmt.Errorf("received error from connect: %w", err)
			log.Println(err)
		}
		time.Sleep(1 * time.Second)
	}
}
