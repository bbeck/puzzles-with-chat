package spellingbee

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestMessageHandler_HandleChannelMessage(t *testing.T) {
	// expected outcomes indexed by channel status
	type Expected map[string]struct {
		path, body string
	}

	tests := []struct {
		name     string
		message  string // the message the channel received
		expected Expected
	}{
		{
			name:    "not a command",
			message: "hello there",
			expected: Expected{
				"solving":  {},
				"paused":   {},
				"complete": {},
			},
		},
		{
			name:    "answer command",
			message: "!railroad",
			expected: Expected{
				"solving":  {"/api/spellingbee/channel/answer", `"railroad"`},
				"paused":   {},
				"complete": {},
			},
		},
		{
			name:    "answer command long form",
			message: "!answer railroad",
			expected: Expected{
				"solving":  {"/api/spellingbee/channel/answer", `"railroad"`},
				"paused":   {},
				"complete": {},
			},
		},
		{
			name:    "answer command long form, mixed case command",
			message: "!AnSWeR railroad",
			expected: Expected{
				"solving":  {"/api/spellingbee/channel/answer", `"railroad"`},
				"paused":   {},
				"complete": {},
			},
		},
		{
			name:    "shuffle command",
			message: "!shuffle",
			expected: Expected{
				"solving":  {"/api/spellingbee/channel/shuffle", ""},
				"paused":   {},
				"complete": {},
			},
		},
		{
			name:    "shuffle command, mixed case command",
			message: "!sHuFfLe",
			expected: Expected{
				"solving":  {"/api/spellingbee/channel/shuffle", ""},
				"paused":   {},
				"complete": {},
			},
		},
	}

	for _, test := range tests {
		for status, expected := range test.expected {
			t.Run(fmt.Sprintf("%s (%s status)", test.name, status), func(t *testing.T) {
				var path, body string
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					defer r.Body.Close()
					w.WriteHeader(200)

					bs, err := ioutil.ReadAll(r.Body)
					require.NoError(t, err)

					path = r.URL.Path
					body = string(bs)
				}))
				defer server.Close()

				parsed, err := url.Parse(server.URL)
				require.NoError(t, err)

				handler := NewMessageHandler(parsed.Host)
				handler.HandleChannelMessage("channel", status, test.message)

				assert.Equal(t, expected.path, path)
				assert.Equal(t, expected.body, body)
			})
		}
	}
}
