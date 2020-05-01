package web

//
// NOTE: This file is hard linked in both api/web/client.go and
// bot/web/client.go changes made in one file will automatically be reflected in
// the other.
//
// A hardlink was used to allow different docker containers to see the file
// properly.
//

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
	request, err := http.NewRequest(http.MethodGet, url, nil)
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
		return response, fmt.Errorf("received %d response for GET from url %s", response.StatusCode, url)
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
	request, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("unable to create http request for url %s: %v", url, err)
	}

	response, err := client.Do(request)
	if err != nil {
		return response, fmt.Errorf("unable to POST to url %s: %v", url, err)
	}

	if response.StatusCode != 200 {
		return response, fmt.Errorf("received %d response for POST to url %s", response.StatusCode, url)
	}

	return response, nil
}

// Put performs a HTTP PUT to a URL using the supplied body and the default
// HTTP client.
func Put(url string, body io.Reader) (*http.Response, error) {
	return PutWithClient(DefaultHTTPClient, url, body)
}

// PutWithClient performs a HTTP PUT to a URL using the supplied body and HTTP
// client.
func PutWithClient(client *http.Client, url string, body io.Reader) (*http.Response, error) {
	request, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		return nil, fmt.Errorf("unable to create http request for url %s: %v", url, err)
	}

	response, err := client.Do(request)
	if err != nil {
		return response, fmt.Errorf("unable to PUT to url %s: %v", url, err)
	}

	if response.StatusCode != 200 {
		return response, fmt.Errorf("received %d response for PUT to url %s", response.StatusCode, url)
	}

	return response, nil
}
