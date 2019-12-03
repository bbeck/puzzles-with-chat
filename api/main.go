package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"

	"github.com/bbeck/twitch-plays-crosswords/api/channel"
)

func main() {
	pool := NewRedisPool()
	defer func() { _ = pool.Close() }()

	registry := new(ClientRegistry)

	router := gin.Default()

	// Register handlers for our paths.
	api := router.Group("/api")
	api.PUT("/channel/:channel/settings/:setting", UpdateChannelSetting(pool, registry))
	api.GET("/channel/:channel/events", ChannelEventsHandler(pool, registry))

	// Start the server.
	err := router.Run(":5000")
	if err != nil {
		log.Fatalf("error from main: %+v", err)
	}
}

func NewRedisPool() *redis.Pool {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = ":6379"
	}

	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 300 * time.Second,

		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", host)
		},

		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

// UpdateChannelSetting allows a setting for a channel to be updated.
func UpdateChannelSetting(pool *redis.Pool, registry *ClientRegistry) gin.HandlerFunc {
	return func(c *gin.Context) {
		channelName := c.Param("channel")
		setting := c.Param("setting")

		conn := pool.Get()
		defer func() { _ = conn.Close() }()

		// Load the existing settings for the channel so that we can apply the
		// updates to them.
		settings, err := channel.GetSettings(conn, channelName)
		if err != nil {
			err = fmt.Errorf("unable to read settings for channel %s: %+v", channelName, err)
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		// Apply the update to the settings in memory.
		switch setting {
		case "only_allow_correct_answers":
			var value bool
			if err := c.BindJSON(&value); err != nil {
				err = fmt.Errorf("unable to parse boolean setting value %v: %+v", value, err)
				_ = c.AbortWithError(http.StatusBadRequest, err)
				return
			}
			settings.OnlyAllowCorrectAnswers = value

		case "clues_to_show":
			var value channel.ClueVisibility
			if err := c.BindJSON(&value); err != nil {
				err = fmt.Errorf("unable to parse visibility setting value %s: %+v", value, err)
				_ = c.AbortWithError(http.StatusBadRequest, err)
				return
			}
			settings.CluesToShow = value

		case "clue_font_size":
			var value channel.FontSize
			if err := c.BindJSON(&value); err != nil {
				err = fmt.Errorf("unable to parse font size setting value %s: %+v", value, err)
				_ = c.AbortWithError(http.StatusBadRequest, err)
				return
			}
			settings.ClueFontSize = value

		default:
			err = fmt.Errorf("unrecognized setting name %s", setting)
			_ = c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		// Save the settings back to the database.
		err = channel.SetSettings(conn, channelName, settings)
		if err != nil {
			err = fmt.Errorf("unable to save settings for channel %s: %+v", channelName, err)
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		// Now broadcast the new settings to all of the clients in the channel.
		registry.BroadcastEvent(channelName, Event{
			Kind:    "settings",
			Payload: settings,
		})
	}
}

// ChannelEventsHandler establishes an event stream with a client.  An event
// stream is server side event stream (SSE) with a client's browser that allows
// one way communication from the server to the client.  Clients that call into
// this handler will keep an open connection open to the server waiting to
// receive events as JSON objects.  The server can send events to all clients
// of a channel using the ClientRegistry.BroadcastEvent method.
func ChannelEventsHandler(pool *redis.Pool, registry *ClientRegistry) gin.HandlerFunc {
	return func(c *gin.Context) {
		channelName := c.Param("channel")

		// Construct the stream that all events for this particular client will be
		// placed into.
		stream := make(chan Event, 10)
		defer close(stream)

		// Setup a connection to redis so that we can read settings and the current
		// state of the solve.
		conn := pool.Get()
		defer func() { _ = conn.Close() }()

		// Send application settings if there are any.
		settings, err := channel.GetSettings(conn, channelName)
		if err != nil {
			err = fmt.Errorf("unable to read settings for channel %s: %+v", channelName, err)
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		stream <- Event{
			Kind:    "settings",
			Payload: settings,
		}

		// TODO: Send a periodic ping to keep the connection open?
		// TODO: Send the current state of the puzzle if one is being solved.

		// Register this client with the registry so that it can receive events for
		// the channel.
		id, err := registry.RegisterClient(channelName, stream)
		if err != nil {
			err = fmt.Errorf("unable to register client for channel %s: %+v", channelName, err)
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		defer registry.DeregisterClient(channelName, id)

		// Loop until the client disconnects sending them any events that are
		// queued for them.
		c.Header("Cache-Control", "no-transform")
		c.Stream(func(w io.Writer) bool {
			if msg, ok := <-stream; ok {
				c.SSEvent("message", msg)
				return true
			}

			return false
		})
	}
}
