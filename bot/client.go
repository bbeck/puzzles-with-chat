package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/gempir/go-twitch-irc/v2"
	"io"
	"net"
	"os"
	"regexp"
	"strings"
)

// A Client represents a source of chat messages from one or more channels.
type Client interface {
	// Connect to the service and emit chat messages as they happen.  This method
	// will block and only returns when the client is disconnected.
	Connect() error

	// Join one or more channels and begin processing messages from them.
	Join(channels ...string)

	// Depart from a channel and stop processing messages from it.
	Depart(channel string)
}

// An Integration represents an implementation of a bot to play a game in a
// client channel.
type Integration interface {
	GetActiveChannelNames() ([]string, error)
	HandleChannelMessage(channel, userid, username, message string)
}

// NewClient constructs a new client instance that's wired to the provided
// integrations and will send all channel messages to each integration.
func NewClient(integrations []Integration) (Client, error) {
	env, ok := os.LookupEnv("ENV")
	if !ok {
		env = "local"
	}

	if env == "local" {
		// When running locally, spawn a client that reads chat messages from a
		// local network socket and doesn't actually connect to Twitch.
		return &LocalClient{integrations: integrations}, nil
	}

	// In a non-local environment we return an actual client that's hooked up to
	// Twitch.  In order to do this we'll need to make sure we have the proper
	// Twitch API credentials.
	username, ok := os.LookupEnv("TWITCH_USERNAME")
	if !ok {
		return nil, errors.New("missing TWITCH_USERNAME environment variable")
	}

	token, ok := os.LookupEnv("TWITCH_OAUTH_TOKEN")
	if !ok {
		return nil, errors.New("missing TWITCH_OAUTH_TOKEN environment variable")
	}

	client := twitch.NewClient(username, token)

	// Wire up the Twitch client to send every message in a channel to each of the
	// integrations.
	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		channel := message.Channel
		uid := message.User.ID
		user := message.User.DisplayName

		for _, integration := range integrations {
			integration.HandleChannelMessage(channel, uid, user, message.Message)
		}
	})

	return client, nil
}

// LocalClient listens on a local network socket and returns messages based on
// the commands it receives.
type LocalClient struct {
	port         int
	integrations []Integration
}

func (c *LocalClient) Join(...string) {}
func (c *LocalClient) Depart(string)  {}

// Connect implements a small REPL on a network socket that allows a user to
// use the connection as a means for providing input into the bot.  Only one
// connection at a time is supported.
func (c *LocalClient) Connect() error {
	address := ":5000"
	if c.port != 0 {
		address = fmt.Sprintf(":%d", c.port)
	}

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer func() { _ = listener.Close() }()

	conn, err := listener.Accept()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	return c.REPL(conn)
}

// A regular expression that matches a command to the LocalClient to set the
// channel that a message should be sent to.
var SetChannelRegexp = regexp.MustCompile(`^/channel ([^\s]+)$`)

// A regular expression that matches a command to the LocalClient to set the
// user that a message should be sent from.
var SetUserRegexp = regexp.MustCompile(`^/user ([^\s]+)$`)

// REPL runs a read, eval, print loop for a connected client.
func (c *LocalClient) REPL(conn net.Conn) error {
	channel := "default"
	user := "anonymous"
	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)

	dedent := func(s string) string {
		formatted := ""
		for _, line := range strings.Split(s, "\n") {
			formatted = fmt.Sprintf("%s\n%s", formatted, strings.TrimSpace(line))
		}

		return strings.TrimSpace(formatted)
	}

	write := func(s string) error {
		_, err := writer.WriteString(s)
		if err != nil {
			return err
		}

		return writer.Flush()
	}

	id := func(s string) string {
		hash := md5.New()
		_, _ = io.WriteString(hash, s)
		return hex.EncodeToString(hash.Sum(nil))
	}

	// Print the banner to let the user know they've connected.
	banner := dedent(`
    ============================================================================
		Connected to local bot interface.  Please enter chat commands to provide
		answers to the puzzle.  Special chat commands are available as follows:
		
		/channel <name> - Sets the channel that answers are submitted to.
		/user <name>    - Sets the username that answers as submitted as.
    ============================================================================
	`)
	if err := write(fmt.Sprintf("\n%s\n\n", banner)); err != nil {
		return err
	}

	// Loop until the client disconnects processing commands.
	for {
		// Prompt
		if err := write(fmt.Sprintf("[%s@%s] ", user, channel)); err != nil {
			return err
		}

		input, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		input = strings.TrimSpace(input)

		if match := SetChannelRegexp.FindStringSubmatch(input); len(match) != 0 {
			channel = match[1]
			continue
		}

		if match := SetUserRegexp.FindStringSubmatch(input); len(match) != 0 {
			user = match[1]
			continue
		}

		for _, integration := range c.integrations {
			integration.HandleChannelMessage(channel, id(user), user, input)
		}
	}
}
