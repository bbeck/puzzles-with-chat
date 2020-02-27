package web

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// The default HTTP client to use when fetching a URL.
var DefaultHTTPClient = &http.Client{
	Timeout: 3 * time.Second,
}

// Get performs a HTTP GET of a URL using the default HTTP client.
func Get(url string) (*http.Response, error) {
	return GetWithHeaders(url, nil)
}

// GetWithHeaders performs a HTTP GET of a URL using the default HTTP client
// and passing in the supplied headers.
func GetWithHeaders(url string, headers map[string]string) (*http.Response, error) {
	return GetWithClient(DefaultHTTPClient, url, headers)
}

// GetWithClient performs a HTTP GET of a URL with custom headers using the
// supplied HTTP client.
func GetWithClient(client *http.Client, url string, headers map[string]string) (*http.Response, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create http request for url %s: %v", url, err)
	}

	for key, value := range headers {
		request.Header.Add(key, value)
	}

	response, err := client.Do(request)
	if err != nil {
		return response, fmt.Errorf("unable to GET from url %s: %v", url, err)
	}

	if response.StatusCode != 200 {
		return response, fmt.Errorf("received non-200 response for GET from url %s: %v", url, err)
	}

	return response, nil
}

// Post performs a HTTP POST to a URL using the supplied body and the default
// HTTP client.
func Post(url string, body io.Reader) (*http.Response, error) {
	return PostWithClient(DefaultHTTPClient, url, body)
}

// PostWithClient performs a HTTP POST to a URL using the supplied body and HTTP
// client.
func PostWithClient(client *http.Client, url string, body io.Reader) (*http.Response, error) {
	request, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("unable to create http request for url %s: %v", url, err)
	}

	response, err := client.Do(request)
	if err != nil {
		return response, fmt.Errorf("unable to POST to url %s: %v", url, err)
	}

	if response.StatusCode != 200 {
		return response, fmt.Errorf("received non-200 response for POST to url %s: %v", url, err)
	}

	return response, nil
}
