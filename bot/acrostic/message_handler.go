package acrostic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bbeck/puzzles-with-chat/bot/web"
	"log"
	"net/http"
	"regexp"
	"time"
)

// The HTTP client to use when communicating with the api service from the
// acrostic integration.
var DefaultAcrosticHTTPClient = &http.Client{
	Timeout: 1 * time.Second,
}

// A regular expression that matches a message that's providing an answer.
// Capture group 1 is the clue and capture group 2 is the answer.
var AnswerRegexp = regexp.MustCompile(
	`^!(?:answer\s+)?([0-9]+|[A-Za-z])\s+(.*)\s*$`,
)

// A regular expression that matches a message that's asking for a clue to be
// made visible.  Capture group 1 is the clue.
var ShowClueRegexp = regexp.MustCompile(
	`^!show\s+(?P<clue>[A-Za-z])\s*$`,
)

type MessageHandler struct {
	baseURL string
}

func NewMessageHandler(host string) *MessageHandler {
	url := fmt.Sprintf("http://%s/api/acrostic", host)
	return &MessageHandler{baseURL: url}
}

// HandleChannelMessage parses a message and if it matches an acrostic command
// sends it to the appropriate API endpoint.
func (h *MessageHandler) HandleChannelMessage(channel string, _ string, _ string, message string) {
	if match := AnswerRegexp.FindStringSubmatch(message); len(match) != 0 {
		clue := match[1]
		answer := match[2]

		bs, err := json.Marshal(answer)
		if err != nil {
			log.Printf("unable to marshal answer (%s) to json: %v", answer, err)
			return
		}

		url := fmt.Sprintf("%s/%s/answer/%s", h.baseURL, channel, clue)
		response, err := web.PutWithClient(DefaultAcrosticHTTPClient, url, bytes.NewReader(bs))
		defer func() { _ = response.Body.Close() }()
		if err != nil {
			log.Printf("error applying answer, url: %s, answer: %s\n", url, answer)
		}
		return
	}

	if match := ShowClueRegexp.FindStringSubmatch(message); len(match) != 0 {
		clue := match[1]

		url := fmt.Sprintf("%s/%s/show/%s", h.baseURL, channel, clue)
		response, err := web.GetWithClient(DefaultAcrosticHTTPClient, url, nil)
		defer func() { _ = response.Body.Close() }()
		if err != nil {
			log.Printf("error showing clue, url: %s", url)
		}
		return
	}
}
