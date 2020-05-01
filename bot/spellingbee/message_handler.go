package spellingbee

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bbeck/twitch-plays-crosswords/bot/web"
	"log"
	"net/http"
	"regexp"
	"time"
)

// The HTTP client to use when communicating with the api service from the
// spelling bee integration.
var DefaultSpellingBeeHTTPClient = &http.Client{
	Timeout: 1 * time.Second,
}

// A regular expression that matches a message that's providing an answer.
// Capture group 1 is the answer.
var AnswerRegexp = regexp.MustCompile(
	`^!(?:answer\s+)?([^\s]+)\s*$`,
)

// A regular expression that matches a message that's asking for the letters
// to be shuffled into a new order.  There are no capture groups.
var ShuffleRegexp = regexp.MustCompile(
	`^!shuffle\s*$`,
)

type MessageHandler struct {
	baseURL string
}

func NewMessageHandler(host string) *MessageHandler {
	url := fmt.Sprintf("http://%s/api/spellingbee", host)
	return &MessageHandler{baseURL: url}
}

// HandleChannelMessage parses a message and if it matches a spelling bee
// command sends it to the appropriate API endpoint.
func (h *MessageHandler) HandleChannelMessage(channel string, _ string, _ string, message string) {
	// We need to check for !shuffle first since the regexp patterns are overlapping.
	if match := ShuffleRegexp.FindStringSubmatch(message); len(match) != 0 {
		url := fmt.Sprintf("%s/%s/shuffle", h.baseURL, channel)
		response, err := web.GetWithClient(DefaultSpellingBeeHTTPClient, url, nil)
		defer func() { _ = response.Body.Close() }()
		if err != nil {
			log.Printf("error shuffling letters, url: %s", url)
		}
		return
	}

	if match := AnswerRegexp.FindStringSubmatch(message); len(match) != 0 {
		answer := match[1]

		bs, err := json.Marshal(answer)
		if err != nil {
			log.Printf("unable to marshal answer (%s) to json: %v", answer, err)
			return
		}

		url := fmt.Sprintf("%s/%s/answer", h.baseURL, channel)
		response, err := web.PostWithClient(DefaultSpellingBeeHTTPClient, url, bytes.NewReader(bs))
		defer func() { _ = response.Body.Close() }()
		if err != nil {
			log.Printf("error applying answer, url: %s, answer: %s\n", url, answer)
		}
		return
	}
}
