package crossword

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"time"
)

// The HTTP client to use when communicating with the api service from the
// crossword integration.
var DefaultCrosswordHTTPClient = &http.Client{
	Timeout: 1 * time.Second,
}

// A regular expression that matches a message that's providing an answer.
// Capture group 1 is the clue and capture group 2 is the answer.
var AnswerRegexp = regexp.MustCompile(
	`^!(?:answer\s+)?([0-9]+[aAdD])\s+(.*)\s*$`,
)

// A regular expression that matches a message that's asking for a clue to be
// made visible.  Capture group 1 is the clue.
var ShowClueRegexp = regexp.MustCompile(
	`^!show\s+(?P<clue>[0-9]+[aAdD])\s*$`,
)

type MessageHandler struct {
	baseURL string
}

func NewMessageHandler(host string) *MessageHandler {
	url := fmt.Sprintf("http://%s/api/crossword", host)
	return &MessageHandler{baseURL: url}
}

// HandleChannelMessage parses a message and if it matches a crossword command
// sends it to the appropriate API endpoint.
func (h *MessageHandler) HandleChannelMessage(channel string, _ string, _ string, message string) {
	if match := AnswerRegexp.FindStringSubmatch(message); len(match) != 0 {
		clue := match[1]
		answer := match[2]

		url := fmt.Sprintf("%s/%s/answer/%s", h.baseURL, channel, clue)
		body, err := PUT(DefaultCrosswordHTTPClient, url, answer)
		defer func() { _ = body.Close() }()
		if err != nil {
			log.Printf("error applying answer, url: %s, answer: %s\n", url, answer)
		}
		return
	}

	if match := ShowClueRegexp.FindStringSubmatch(message); len(match) != 0 {
		clue := match[1]

		url := fmt.Sprintf("%s/%s/show/%s", h.baseURL, channel, clue)
		body, err := GET(DefaultCrosswordHTTPClient, url)
		defer func() { _ = body.Close() }()
		if err != nil {
			log.Printf("error showing clue, url: %s", url)
		}
		return
	}
}

func GET(client *http.Client, url string) (io.ReadCloser, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		err := fmt.Errorf("unable to create HTTP request for url %s: %v", url, err)
		return ioutil.NopCloser(nil), err
	}

	response, err := client.Do(request)
	if err != nil {
		err := fmt.Errorf("unable to perform HTTP GET to url %s: %v", url, err)
		return ioutil.NopCloser(nil), err
	}

	if response.StatusCode != 200 {
		err := fmt.Errorf("received non-200 response for GET to url %s: %v", url, err)
		return response.Body, err
	}

	return response.Body, nil
}

func PUT(client *http.Client, url string, body interface{}) (io.ReadCloser, error) {
	bs, err := json.Marshal(body)
	if err != nil {
		err := fmt.Errorf("unable to marshal body to json: %v", err)
		return ioutil.NopCloser(nil), err
	}

	request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(bs))
	if err != nil {
		err := fmt.Errorf("unable to create HTTP request for url %s: %v", url, err)
		return ioutil.NopCloser(nil), err
	}

	response, err := client.Do(request)
	if err != nil {
		err := fmt.Errorf("unable to perform HTTP PUT to url %s: %v", url, err)
		return ioutil.NopCloser(nil), err
	}

	if response.StatusCode != 200 {
		err := fmt.Errorf("received non-200 response for PUT to url %s: %v", url, err)
		return response.Body, err
	}

	return response.Body, nil
}
