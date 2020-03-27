package crossword

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
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
			message:      "!1a q and a",
			expectedPath: "/api/crossword/channel/answer/1a",
			expectedBody: `"q and a"`,
		},
		{
			name:         "show command",
			message:      "!show 1a",
			expectedPath: "/api/crossword/channel/show/1a",
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

func TestGET_Error(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		client  *http.Client
		respond func(http.ResponseWriter)
	}{
		{
			name:    "error creating request (bad url)",
			baseURL: ":",
		},
		{
			name:   "error in client.Do (timeout)",
			client: &http.Client{Timeout: 1 * time.Millisecond},
			respond: func(writer http.ResponseWriter) {
				time.Sleep(10 * time.Millisecond)
				writer.WriteHeader(200)
			},
		},
		{
			name: "non-200 response",
			respond: func(w http.ResponseWriter) {
				w.WriteHeader(404)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				test.respond(w)
			}))
			defer server.Close()

			var base = test.baseURL
			if base == "" {
				base = server.URL
			}

			var client = test.client
			if client == nil {
				client = &http.Client{Timeout: 100 * time.Millisecond}
			}

			body, err := GET(client, base)
			defer body.Close()

			assert.Error(t, err)
		})
	}
}

func TestPUT_Error(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		body    interface{}
		client  *http.Client
		respond func(http.ResponseWriter)
	}{
		{
			name:    "error creating request (bad url)",
			baseURL: ":",
		},
		{
			name:   "error in client.Do (timeout)",
			client: &http.Client{Timeout: 1 * time.Millisecond},
			respond: func(writer http.ResponseWriter) {
				time.Sleep(10 * time.Millisecond)
				writer.WriteHeader(200)
			},
		},
		{
			name: "non-200 response",
			respond: func(w http.ResponseWriter) {
				w.WriteHeader(404)
			},
		},
		{
			name: "cannot encode body to json",
			body: make(chan struct{}),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				test.respond(w)
			}))
			defer server.Close()

			var base = test.baseURL
			if base == "" {
				base = server.URL
			}

			var client = test.client
			if client == nil {
				client = &http.Client{Timeout: 100 * time.Millisecond}
			}

			body, err := PUT(client, base, test.body)
			defer body.Close()

			assert.Error(t, err)
		})
	}
}
