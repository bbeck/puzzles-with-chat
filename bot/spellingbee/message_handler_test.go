package spellingbee

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestMessageHandler_HandleChannelMessage(t *testing.T) {
	tests := []struct {
		name         string
		message      string
		expectedPath string
		expectedBody string
	}{
		{
			name:    "not a command",
			message: "hello there",
		},
		{
			name:         "answer command",
			message:      "!railroad",
			expectedPath: "/api/spellingbee/channel/answer",
			expectedBody: `"railroad"`,
		},
		{
			name:         "shuffle command",
			message:      "!shuffle",
			expectedPath: "/api/spellingbee/channel/shuffle",
		},
	}

	for _, test := range tests {
		var path, body string
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				w.WriteHeader(200)

				path = r.URL.Path

				bs, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				body = string(bs)
			}))
			defer server.Close()

			parsed, err := url.Parse(server.URL)
			require.NoError(t, err)

			handler := NewMessageHandler(parsed.Host)
			handler.HandleChannelMessage("channel", "uid", "user", test.message)

			assert.Equal(t, test.expectedPath, path)
			assert.Equal(t, test.expectedBody, body)
		})
	}
}
