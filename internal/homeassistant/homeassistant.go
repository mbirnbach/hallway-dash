// Package homeassistant fetches indoor temperature/humidity sensor
// states from a Home Assistant instance. All requests happen
// server-side so the long-lived access token never reaches the browser.
package homeassistant

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Data struct {
	Temperature *float64 `json:"temperature"`
	Humidity    *float64 `json:"humidity"`
}

type Fetcher struct {
	Client         *http.Client
	BaseURL        string
	Token          string
	TempEntity     string
	HumidityEntity string
}

type stateResponse struct {
	State string `json:"state"`
}

// Fetch returns a zero-value Data{} (both fields nil, no error) when
// Home Assistant isn't configured, so the frontend just shows "—".
func (f Fetcher) Fetch(ctx context.Context) (Data, error) {
	baseURL := strings.TrimRight(f.BaseURL, "/")
	if baseURL == "" || f.Token == "" || f.TempEntity == "" {
		return Data{}, nil
	}

	temp, err := f.fetchState(ctx, baseURL, f.TempEntity)
	if err != nil {
		return Data{}, err
	}

	var hum *float64
	if f.HumidityEntity != "" {
		hum, err = f.fetchState(ctx, baseURL, f.HumidityEntity)
		if err != nil {
			return Data{}, err
		}
	}

	return Data{Temperature: temp, Humidity: hum}, nil
}

func (f Fetcher) fetchState(ctx context.Context, baseURL, entity string) (*float64, error) {
	url := fmt.Sprintf("%s/api/states/%s", baseURL, entity)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+f.Token)
	req.Header.Set("Accept", "application/json")

	resp, err := f.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("home assistant returned status %d for %s", resp.StatusCode, entity)
	}

	var state stateResponse
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return nil, err
	}
	value, err := strconv.ParseFloat(state.State, 64)
	if err != nil {
		return nil, nil // e.g. "unavailable" — treat as no reading, not an error
	}
	return &value, nil
}
