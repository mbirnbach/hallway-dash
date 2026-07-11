// Package joke fetches a random joke from the configured primary source
// and falls back to JokeAPI.dev if the primary is unavailable or returns
// something unparsable.
package joke

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const fallbackURL = "https://v2.jokeapi.dev/joke/Any?safe-mode&format=json&lang=de"

type Data struct {
	Text string `json:"text"`
}

type Fetcher struct {
	Client     *http.Client
	PrimaryURL string
}

func (f Fetcher) Fetch(ctx context.Context) (Data, error) {
	if f.PrimaryURL != "" {
		if text, err := fetchPlainText(ctx, f.Client, f.PrimaryURL); err == nil && text != "" {
			return Data{Text: text}, nil
		}
	}
	text, err := fetchJokeAPI(ctx, f.Client)
	if err != nil {
		return Data{}, err
	}
	return Data{Text: text}, nil
}

// fetchPlainText hits nopunintended.xyz, which returns the joke as a raw
// text body on every call (no JSON wrapper, no dedicated path).
func fetchPlainText(ctx context.Context, client *http.Client, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("joke primary source returned status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8192))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(body)), nil
}

type jokeAPIResponse struct {
	Error    bool   `json:"error"`
	Type     string `json:"type"`
	Joke     string `json:"joke"`
	Setup    string `json:"setup"`
	Delivery string `json:"delivery"`
}

func fetchJokeAPI(ctx context.Context, client *http.Client) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fallbackURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("jokeapi.dev returned status %d", resp.StatusCode)
	}
	var data jokeAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	if data.Error {
		return "", fmt.Errorf("jokeapi.dev returned an error response")
	}
	if data.Type == "twopart" {
		return strings.TrimSpace(data.Setup + " " + data.Delivery), nil
	}
	return strings.TrimSpace(data.Joke), nil
}
