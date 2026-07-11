// Package weather proxies the Open-Meteo forecast API and reshapes its
// response into the flat structure the frontend consumes, so the browser
// never talks to Open-Meteo directly and the JS doesn't need to know
// about Open-Meteo's field-of-arrays layout.
package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Current struct {
	Temperature float64 `json:"temperature"`
	FeelsLike   float64 `json:"feelsLike"`
	WeatherCode int     `json:"weatherCode"`
}

type Day struct {
	Date        string  `json:"date"`
	WeatherCode int     `json:"weatherCode"`
	TempMax     float64 `json:"tempMax"`
	TempMin     float64 `json:"tempMin"`
	Pop         int     `json:"pop"`
	Sunrise     string  `json:"sunrise"`
	Sunset      string  `json:"sunset"`
}

type Data struct {
	Current *Current `json:"current"`
	Daily   []Day    `json:"daily"`
}

// openMeteoResponse mirrors only the fields we use from
// https://open-meteo.com/en/docs
type openMeteoResponse struct {
	Current struct {
		Temperature float64 `json:"temperature_2m"`
		FeelsLike   float64 `json:"apparent_temperature"`
		WeatherCode int     `json:"weather_code"`
	} `json:"current"`
	Daily struct {
		Time                 []string  `json:"time"`
		WeatherCode          []int     `json:"weather_code"`
		TempMax              []float64 `json:"temperature_2m_max"`
		TempMin              []float64 `json:"temperature_2m_min"`
		PrecipitationProbMax []int     `json:"precipitation_probability_max"`
		Sunrise              []string  `json:"sunrise"`
		Sunset               []string  `json:"sunset"`
	} `json:"daily"`
}

type Fetcher struct {
	Client    *http.Client
	Latitude  float64
	Longitude float64
}

func (f Fetcher) Fetch(ctx context.Context) (Data, error) {
	url := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%g&longitude=%g&current=temperature_2m,apparent_temperature,weather_code&daily=weather_code,temperature_2m_max,temperature_2m_min,precipitation_probability_max,sunrise,sunset&timezone=auto&forecast_days=6",
		f.Latitude, f.Longitude,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Data{}, err
	}
	resp, err := f.Client.Do(req)
	if err != nil {
		return Data{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Data{}, fmt.Errorf("open-meteo returned status %d", resp.StatusCode)
	}
	var raw openMeteoResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return Data{}, err
	}

	data := Data{
		Current: &Current{
			Temperature: raw.Current.Temperature,
			FeelsLike:   raw.Current.FeelsLike,
			WeatherCode: raw.Current.WeatherCode,
		},
	}
	n := len(raw.Daily.Time)
	for i := 0; i < n; i++ {
		day := Day{Date: raw.Daily.Time[i]}
		if i < len(raw.Daily.WeatherCode) {
			day.WeatherCode = raw.Daily.WeatherCode[i]
		}
		if i < len(raw.Daily.TempMax) {
			day.TempMax = raw.Daily.TempMax[i]
		}
		if i < len(raw.Daily.TempMin) {
			day.TempMin = raw.Daily.TempMin[i]
		}
		if i < len(raw.Daily.PrecipitationProbMax) {
			day.Pop = raw.Daily.PrecipitationProbMax[i]
		}
		if i == 0 {
			if len(raw.Daily.Sunrise) > 0 {
				day.Sunrise = formatClock(raw.Daily.Sunrise[0])
			}
			if len(raw.Daily.Sunset) > 0 {
				day.Sunset = formatClock(raw.Daily.Sunset[0])
			}
		}
		data.Daily = append(data.Daily, day)
	}
	return data, nil
}

// formatClock turns Open-Meteo's "2026-07-11T05:32" into "05:32".
func formatClock(iso string) string {
	t, err := time.Parse("2006-01-02T15:04", iso)
	if err != nil {
		return ""
	}
	return t.Format("15:04")
}
