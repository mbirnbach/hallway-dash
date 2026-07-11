// Package config loads all runtime configuration from environment variables.
package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port string

	StaticDir      string
	BackgroundsDir string

	Latitude      float64
	Longitude     float64
	LocationLabel string

	CalendarICSURL string
	WasteICSURL    string

	AGSCode string

	JokeAPIURL string

	HAURL            string
	HAToken          string
	HATempEntity     string
	HAHumidityEntity string

	ClockFormat string
	AccentColor string
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envFloatOr(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}

// loadDotEnv reads a ".env" file in the working directory, if present,
// and applies KEY=VALUE pairs via os.Setenv for any key not already set
// in the real environment. Docker Compose reads .env natively for
// variable substitution, but a plain `go run .` during local
// development does not — this makes both paths behave the same way.
// Missing file is not an error; it's expected in production where env
// vars are injected directly (Docker, Unraid, etc.).
func loadDotEnv() {
	f, err := os.Open(".env")
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if key == "" {
			continue
		}
		if _, alreadySet := os.LookupEnv(key); !alreadySet {
			os.Setenv(key, value)
		}
	}
}

// Load reads configuration from the environment (loading a local .env
// file first, if present), applying the same defaults used by the
// design reference (Berlin coordinates, 24h clock, amber accent) so the
// dashboard still renders something sensible when a variable is unset.
func Load() Config {
	loadDotEnv()

	return Config{
		Port: envOr("PORT", "8080"),

		StaticDir:      envOr("STATIC_DIR", "static"),
		BackgroundsDir: envOr("BACKGROUNDS_DIR", "backgrounds"),

		Latitude:      envFloatOr("LATITUDE", 52.52),
		Longitude:     envFloatOr("LONGITUDE", 13.405),
		LocationLabel: envOr("LOCATION_LABEL", "Zuhause"),

		CalendarICSURL: os.Getenv("CALENDAR_ICS_URL"),
		WasteICSURL:    os.Getenv("WASTE_ICS_URL"),

		AGSCode: os.Getenv("AGS_CODE"),

		JokeAPIURL: envOr("JOKE_API_URL", "https://nopunintended.xyz"),

		HAURL:            os.Getenv("HA_URL"),
		HAToken:          os.Getenv("HA_TOKEN"),
		HATempEntity:     os.Getenv("HA_TEMP_ENTITY"),
		HAHumidityEntity: os.Getenv("HA_HUMIDITY_ENTITY"),

		ClockFormat: envOr("CLOCK_FORMAT", "24h"),
		AccentColor: envOr("ACCENT_COLOR", "#F5A524"),
	}
}

// ShowIndoor reports whether enough Home Assistant config is present to
// attempt fetching the indoor sensor block at all.
func (c Config) ShowIndoor() bool {
	return c.HAURL != "" && c.HAToken != "" && c.HATempEntity != ""
}
