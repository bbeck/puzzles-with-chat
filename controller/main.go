package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/bbeck/puzzles-with-chat/controller/sse"
	"github.com/bbeck/puzzles-with-chat/controller/web"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var channels = map[string]bool{
	"agenderwitchery":  true,
	"aidanwould":       true,
	"bbeck":            true,
	"mistaeksweremade": true,
}

func main() {
	host, ok := os.LookupEnv("API_HOST")
	if !ok {
		log.Fatal("missing API_HOST environment variable")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		signal.Stop(signals)
		cancel()
	}()

	events := sse.Open(ctx, fmt.Sprintf("http://%s/api/channels", host))
	actions := make(chan SwitchPuzzle, 10)

	log.Printf("controlling channels: %v\n", channels)

	for {
		select {
		case e := <-events:
			if err := HandleEvent(e, actions); err != nil {
				log.Printf("received error %v while processing event %v\n", err, e)
			}
		case a := <-actions:
			body, err := json.Marshal(map[string]string{
				"new_york_times_date": a.Date.Format("2006-01-02"),
			})
			if err != nil {
				log.Printf("unable to marshal body for action %v: %v\n", a, err)
				break
			}

			time.AfterFunc(20*time.Second, func() {
				log.Printf("executing action: %+v\n", a)
				_, err := web.Put(fmt.Sprintf("http://%s/api/crossword/%s", host, a.Channel), bytes.NewReader(body))
				if err != nil {
					log.Printf("received error when changing puzzle: %+v\n", err)
					return
				}
				_, err = web.Put(fmt.Sprintf("http://%s/api/crossword/%s/status", host, a.Channel), nil)
				if err != nil {
					log.Printf("received error when starting solve: %+v\n", err)
				}
			})
		case <-ctx.Done():
			return
		}
	}
}

func HandleEvent(e sse.Event, actions chan<- SwitchPuzzle) error {
	var event Event
	if err := json.Unmarshal(e.Data, &event); err != nil {
		err = fmt.Errorf("unable to parse json '%s': %+v", e.Data, err)
		return err
	}

	switch event.Kind {
	case "channels":
		var payload Payload
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			err = fmt.Errorf("unable to parse payload '%s': %+v", event.Payload, err)
			return err
		}
		return HandlePayload(payload, actions)

	case "ping":
		return nil

	default:
		err := fmt.Errorf("unrecognized event kind: %s", event.Kind)
		return err
	}
}

func HandlePayload(payload Payload, actions chan<- SwitchPuzzle) error {
	for _, channel := range payload["crossword"] {
		if channel.Puzzle.Publisher != "The New York Times" {
			continue
		}

		if !channels[channel.Name] {
			continue
		}

		if channel.Status != "complete" {
			continue
		}

		actions <- SwitchPuzzle{
			Channel:   channel.Name,
			Publisher: channel.Puzzle.Publisher,
			Date:      channel.Puzzle.PublishedDate.AddDate(0, 0, -1),
		}

		return nil
	}

	return nil
}

// Event is the SSE event that is received from the API about the active set of
// channels.
type Event struct {
	Kind    string          `json:"kind"`
	Payload json.RawMessage `json:"payload"`
}

// Payload is the payload of an event that is sent out containing the current
// set of located channels organized by puzzle type ID.
type Payload map[string][]struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Puzzle struct {
		Publisher     string    `json:"publisher"`
		PublishedDate time.Time `json:"published"`
	} `json:"puzzle"`
}

// SwitchPuzzle represents the puzzle we want to switch a channel to.
type SwitchPuzzle struct {
	Channel   string
	Publisher string
	Date      time.Time
}
