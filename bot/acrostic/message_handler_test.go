package acrostic

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
			name:    "answer clue",
			message: "!Q half step",
			expected: Expected{
				"solving":  {"/api/acrostic/channel/answer/Q", `"half step"`},
				"paused":   {},
				"complete": {},
			},
		},
		{
			name:    "answer clue with lowercase clue",
			message: "!q half step",
			expected: Expected{
				"solving":  {"/api/acrostic/channel/answer/q", `"half step"`},
				"paused":   {},
				"complete": {},
			},
		},
		{
			name:    "answer clue long form",
			message: "!answer Q half step",
			expected: Expected{
				"solving":  {"/api/acrostic/channel/answer/Q", `"half step"`},
				"paused":   {},
				"complete": {},
			},
		},
		{
			name:    "answer clue long form, mixed case command",
			message: "!AnSWeR Q half step",
			expected: Expected{
				"solving":  {"/api/acrostic/channel/answer/Q", `"half step"`},
				"paused":   {},
				"complete": {},
			},
		},
		{
			name:    "answer clue long form, lowercase clue",
			message: "!answer q half step",
			expected: Expected{
				"solving":  {"/api/acrostic/channel/answer/q", `"half step"`},
				"paused":   {},
				"complete": {},
			},
		},
		{
			name:    "answer cells",
			message: "!26 vast knowledge",
			expected: Expected{
				"solving":  {"/api/acrostic/channel/answer/26", `"vast knowledge"`},
				"paused":   {},
				"complete": {},
			},
		},
		{
			name:    "answer cells long form",
			message: "!answer 26 vast knowledge",
			expected: Expected{
				"solving":  {"/api/acrostic/channel/answer/26", `"vast knowledge"`},
				"paused":   {},
				"complete": {},
			},
		},
		{
			name:    "answer cells long form, mixed case command",
			message: "!AnSWeR 26 vast knowledge",
			expected: Expected{
				"solving":  {"/api/acrostic/channel/answer/26", `"vast knowledge"`},
				"paused":   {},
				"complete": {},
			},
		},
		{
			name:    "show command",
			message: "!show H",
			expected: Expected{
				"solving":  {"/api/acrostic/channel/show/H", ""},
				"paused":   {"/api/acrostic/channel/show/H", ""},
				"complete": {"/api/acrostic/channel/show/H", ""},
			},
		},
		{
			name:    "show command, lowercase clue",
			message: "!show h",
			expected: Expected{
				"solving":  {"/api/acrostic/channel/show/h", ""},
				"paused":   {"/api/acrostic/channel/show/h", ""},
				"complete": {"/api/acrostic/channel/show/h", ""},
			},
		},
		{
			name:    "show command, mixed case command",
			message: "!ShoW H",
			expected: Expected{
				"solving":  {"/api/acrostic/channel/show/H", ""},
				"paused":   {"/api/acrostic/channel/show/H", ""},
				"complete": {"/api/acrostic/channel/show/H", ""},
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
