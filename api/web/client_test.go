package web

//
// NOTE: This file is hard linked in both api/web/client_test.go and
// bot/web/client_test.go changes made in one file will automatically be
// reflected in the other.
//
// A hardlink was used to allow different docker containers to see the file
// properly.
//

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer server.Close()

	_, err := Get(server.URL)
	assert.NoError(t, err)
}

func TestGet_Error(t *testing.T) {
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

			client := test.client
			if client == nil {
				client = DefaultHTTPClient
			}

			_, err := GetWithClient(client, url, nil)
			assert.Error(t, err)
		})
	}
}

func TestGetWithHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("foo") != "bar" {
			w.WriteHeader(400)
			return
		}

		w.WriteHeader(200)
	}))
	defer server.Close()

	_, err := GetWithHeaders(server.URL, map[string]string{"foo": "bar"})
	assert.NoError(t, err)
}

func TestPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer server.Close()

	_, err := Post(server.URL, strings.NewReader(""))
	assert.NoError(t, err)
}

func TestPost_Error(t *testing.T) {
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

			client := test.client
			if client == nil {
				client = DefaultHTTPClient
			}

			response, err := PostWithClient(client, url, strings.NewReader(""))
			if response != nil {
				defer response.Body.Close()
			}
			require.Error(t, err)
		})
	}
}

func TestPut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer server.Close()

	_, err := Put(server.URL, strings.NewReader(""))
	assert.NoError(t, err)
}

func TestPut_Error(t *testing.T) {
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

			client := test.client
			if client == nil {
				client = DefaultHTTPClient
			}

			response, err := PutWithClient(client, url, strings.NewReader(""))
			if response != nil {
				defer response.Body.Close()
			}
			require.Error(t, err)
		})
	}
}
