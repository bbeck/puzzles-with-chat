package crossword

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"regexp"
	"time"
)

// The HTTP client to use when communicating with api service from the crossword
// integration.
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

type Integration struct {
	baseURL string
}

func NewIntegration() (*Integration, error) {
	host, ok := os.LookupEnv("API_HOST")
	if !ok {
		return nil, errors.New("missing API_HOST environment variable")
	}

	return &Integration{
		baseURL: fmt.Sprintf("http://%s/api/crossword", host),
	}, nil
}

// GetActiveChannelNames calls the API service to see which channels are
// currently solving a crossword.
func (c *Integration) GetActiveChannelNames() ([]string, error) {
	request, err := http.NewRequest(http.MethodGet, c.baseURL, nil)
	if err != nil {
		err := fmt.Errorf("unable to create http request: %v", err)
		return nil, err
	}

	response, err := DefaultCrosswordHTTPClient.Do(request)
	if response != nil {
		defer func() { _ = response.Body.Close() }()
	}
	if err != nil {
		err := fmt.Errorf("unable to get active crosswords from api: %v", err)
		return nil, err
	}

	if response.StatusCode != 200 {
		err := fmt.Errorf("received non-200 response when getting active crosswords: %v", err)
		return nil, err
	}

	var names []string
	err = json.NewDecoder(response.Body).Decode(&names)
	if err != nil {
		err := fmt.Errorf("unable to parse active crossword JSON response: %v", err)
		return nil, err
	}

	return names, nil
}

// HandleChannelMessage parses a message and if it matches a crossword command
// sends it to the appropriate API endpoint.
func (c *Integration) HandleChannelMessage(channel string, uid string, user string, message string) {
	if match := AnswerRegexp.FindStringSubmatch(message); len(match) != 0 {
		clue := match[1]
		answer := match[2]

		body, _ := c.PUT(path.Join(channel, "/answer", clue), answer)
		defer func() { _ = body.Close() }()
		return
	}

	if match := ShowClueRegexp.FindStringSubmatch(message); len(match) != 0 {
		clue := match[1]

		body, _ := c.GET(path.Join(channel, "/show", clue))
		defer func() { _ = body.Close() }()
		return
	}
}

func (c *Integration) GET(path string) (io.ReadCloser, error) {
	url := fmt.Sprintf("%s/%s", c.baseURL, path)

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		err := fmt.Errorf("unable to create HTTP request for url %s: %v", url, err)
		return ioutil.NopCloser(nil), err
	}

	response, err := DefaultCrosswordHTTPClient.Do(request)
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

func (c *Integration) PUT(path string, body interface{}) (io.ReadCloser, error) {
	url := fmt.Sprintf("%s/%s", c.baseURL, path)

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

	response, err := DefaultCrosswordHTTPClient.Do(request)
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
