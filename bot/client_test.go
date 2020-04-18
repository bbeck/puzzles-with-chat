package main

import (
	"bufio"
	"fmt"
	"github.com/gempir/go-twitch-irc/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net"
	"os"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name   string
		setup  func()
		verify func(t *testing.T, client Client, err error)
	}{
		{
			name: "no ENV variable results in local client",
			verify: func(t *testing.T, client Client, err error) {
				assert.NoError(t, err)

				_, ok := client.(*LocalClient)
				assert.True(t, ok)
			},
		},
		{
			name: "ENV=local results in local client",
			setup: func() {
				os.Setenv("ENV", "local")
			},
			verify: func(t *testing.T, client Client, err error) {
				assert.NoError(t, err)

				_, ok := client.(*LocalClient)
				assert.True(t, ok)
			},
		},
		{
			name: "no TWITCH_USERNAME variable results in error",
			setup: func() {
				os.Setenv("ENV", "production")
				os.Setenv("TWITCH_OAUTH_TOKEN", "token")
			},
			verify: func(t *testing.T, client Client, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "no TWITCH_OAUTH_TOKEN variable results in error",
			setup: func() {
				os.Setenv("ENV", "production")
				os.Setenv("TWITCH_USERNAME", "token")
			},
			verify: func(t *testing.T, client Client, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "ENV=production results in twitch.Client",
			setup: func() {
				os.Setenv("ENV", "production")
				os.Setenv("TWITCH_USERNAME", "token")
				os.Setenv("TWITCH_OAUTH_TOKEN", "token")
			},
			verify: func(t *testing.T, client Client, err error) {
				assert.NoError(t, err)

				_, ok := client.(*twitch.Client)
				assert.True(t, ok)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// These tests manipulate the environment, so we save a copy before each
			// test to ensure that it doesn't get permanently changed by the test case.
			SaveEnvironmentVars(t)

			if test.setup != nil {
				test.setup()
			}

			client, err := NewClient(nil)
			test.verify(t, client, err)
		})
	}

	os.Setenv("ENV", "local")

	client, err := NewClient(nil)
	assert.NoError(t, err)

	_, ok := client.(*LocalClient)
	assert.True(t, ok)
}

func TestLocalClient_Connect(t *testing.T) {
	client := &LocalClient{port: GetFreePort(t)}

	listening := NewCountDownLatch(1)
	closed := NewCountDownLatch(1)

	go func() {
		listening.CountDown()
		defer closed.CountDown()

		client.Connect()
	}()

	// Wait for the goroutine to start listening before we attempt to connect.
	assert.True(t, listening.Wait(100*time.Millisecond))

	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", client.port))
	require.NotNil(t, conn)
	assert.NoError(t, err)

	// Close the connection (should cause an EOF on the client).
	conn.Close()

	// Wait for the client to process the disconnect error.
	assert.True(t, closed.Wait(100*time.Millisecond))
}

func TestLocalClient_REPL(t *testing.T) {
	tests := []struct {
		name                string
		inputs              []string
		expectedNumMessages int
		verify              func(t *testing.T, messages []SeenMessage)
	}{
		{
			name:                "input passed onto integration",
			inputs:              []string{"hi there"},
			expectedNumMessages: 1,
			verify: func(t *testing.T, messages []SeenMessage) {
				assert.Equal(t, "hi there", messages[0].message)
			},
		},
		{
			name:   "/channel doesn't send to integration",
			inputs: []string{"/channel foo"},
		},
		{
			name: "/channel changes channel name",
			inputs: []string{
				"/channel foo",
				"test",
			},
			expectedNumMessages: 1,
			verify: func(t *testing.T, messages []SeenMessage) {
				assert.Equal(t, "foo", messages[0].channel)
			},
		},
		{
			name:   "/user doesn't send to integration",
			inputs: []string{"/user foo"},
		},
		{
			name: "/user changes user name",
			inputs: []string{
				"/user foo",
				"test",
			},
			expectedNumMessages: 1,
			verify: func(t *testing.T, messages []SeenMessage) {
				assert.Equal(t, "foo", messages[0].username)
			},
		},
		{
			name: "/user changes user id",
			inputs: []string{
				"/user foo",
				"test",
				"/user bar",
				"test",
			},
			expectedNumMessages: 2,
			verify: func(t *testing.T, messages []SeenMessage) {
				assert.NotEqual(t, messages[0].userid, messages[1].userid)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := NewTestMessageHandler(test.expectedNumMessages)

			client := &LocalClient{
				port:     GetFreePort(t),
				handlers: []MessageHandler{handler},
			}

			listening := NewCountDownLatch(1)
			closed := NewCountDownLatch(1)

			go func() {
				listening.CountDown()
				defer closed.CountDown()

				client.Connect() // We don't care about any errors returned when disconnecting
			}()

			// Wait for the goroutine to start listening before we attempt to connect.
			assert.True(t, listening.Wait(100*time.Millisecond))

			conn, err := net.Dial("tcp", fmt.Sprintf(":%d", client.port))
			require.NotNil(t, conn)
			assert.NoError(t, err)

			// Send our inputs.
			writer := bufio.NewWriter(conn)
			for _, input := range test.inputs {
				writer.WriteString(input)
				writer.WriteByte('\n')
			}
			writer.Flush()

			// Wait for the integration to receive the proper number of messages
			// before disconnecting.
			assert.True(t, handler.latch.Wait(100*time.Millisecond))

			// Now that we're done, disconnect.
			conn.Close()

			// Wait for the client to process the disconnect error.
			assert.True(t, closed.Wait(100*time.Millisecond))

			// Ensure that we received the correct number of messages (we may have
			// received more).
			assert.Equal(t, test.expectedNumMessages, len(handler.seen))

			// Now verify that the integration was called with the correct messages.
			if test.verify != nil {
				test.verify(t, handler.seen)
			}
		})
	}
}

func GetFreePort(t *testing.T) int {
	addr, err := net.ResolveTCPAddr("tcp", ":0")
	require.NoError(t, err)

	listener, err := net.ListenTCP("tcp", addr)
	require.NoError(t, err)
	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port
}

type SeenMessage struct {
	channel  string
	userid   string
	username string
	message  string
}

type TestMessageHandler struct {
	latch *CountDownLatch
	seen  []SeenMessage
}

func NewTestMessageHandler(expected int) *TestMessageHandler {
	latch := NewCountDownLatch(expected)
	return &TestMessageHandler{latch: latch}
}

func (i *TestMessageHandler) GetActiveChannelNames() ([]string, error) {
	return nil, nil
}

func (i *TestMessageHandler) HandleChannelMessage(channel, userid, username, message string) {
	i.seen = append(i.seen, SeenMessage{
		channel:  channel,
		userid:   userid,
		username: username,
		message:  message,
	})
	i.latch.CountDown()
}
