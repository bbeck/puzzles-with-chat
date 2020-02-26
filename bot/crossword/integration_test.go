package crossword

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestIntegration_GetActiveChannelNames(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)

		names, err := json.Marshal([]string{"channel1", "channel2"})
		require.NoError(t, err)

		w.Write(names)
	}))
	defer server.Close()

	integration := &Integration{baseURL: server.URL}
	names, err := integration.GetActiveChannelNames()
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"channel1", "channel2"}, names)
}

func TestIntegration_GetActiveChannelNames_Error(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		client  *http.Client
		respond func(http.ResponseWriter)
	}{
		{
			name: "error creating request (bad url)",
			url:  ":",
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
			respond: func(writer http.ResponseWriter) {
				writer.WriteHeader(404)
			},
		},
		{
			name: "invalid json response",
			respond: func(writer http.ResponseWriter) {
				writer.WriteHeader(200)
				writer.Write([]byte("not json"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				test.respond(w)
			}))
			defer server.Close()

			url := test.url
			if url == "" {
				url = server.URL
			}

			if test.client != nil {
				old := DefaultCrosswordHTTPClient
				defer func() { DefaultCrosswordHTTPClient = old }()

				DefaultCrosswordHTTPClient = test.client
			}

			integration := &Integration{baseURL: url}
			_, err := integration.GetActiveChannelNames()
			assert.Error(t, err)
		})
	}
}

func TestIntegration_HandleChannelMessage(t *testing.T) {
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
			expectedPath: "/channel/answer/1a",
			expectedBody: `"q and a"`,
		},
		{
			name:         "show command",
			message:      "!show 1a",
			expectedPath: "/channel/show/1a",
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

			integration := &Integration{baseURL: server.URL}
			integration.HandleChannelMessage("channel", "id", "name", test.message)

			assert.Equal(t, test.expectedPath, path)
			assert.Equal(t, test.expectedBody, body)
		})
	}
}

func TestIntegration_GET_Error(t *testing.T) {
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

			var url = test.baseURL
			if url == "" {
				url = server.URL
			}

			if test.client != nil {
				old := DefaultCrosswordHTTPClient
				defer func() { DefaultCrosswordHTTPClient = old }()

				DefaultCrosswordHTTPClient = test.client
			}

			integration := &Integration{baseURL: url}
			body, err := integration.GET("")
			defer body.Close()

			assert.Error(t, err)
		})
	}
}

func TestIntegration_PUT_Error(t *testing.T) {
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

			var url = test.baseURL
			if url == "" {
				url = server.URL
			}

			if test.client != nil {
				old := DefaultCrosswordHTTPClient
				defer func() { DefaultCrosswordHTTPClient = old }()

				DefaultCrosswordHTTPClient = test.client
			}

			integration := &Integration{baseURL: url}
			body, err := integration.PUT("", test.body)
			defer body.Close()

			assert.Error(t, err)
		})
	}
}
