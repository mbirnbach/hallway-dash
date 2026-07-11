// Command hallway-dash serves the wall-display dashboard: a static
// frontend plus JSON endpoints that aggregate weather, calendars,
// civil-protection alerts, a joke, and Home Assistant sensor data. All
// third-party calls happen here so the browser never talks to those
// services directly.
package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/marvinbirnbach/hallway-dash/internal/alerts"
	"github.com/marvinbirnbach/hallway-dash/internal/calendar"
	"github.com/marvinbirnbach/hallway-dash/internal/config"
	"github.com/marvinbirnbach/hallway-dash/internal/homeassistant"
	"github.com/marvinbirnbach/hallway-dash/internal/joke"
	"github.com/marvinbirnbach/hallway-dash/internal/pollcache"
	"github.com/marvinbirnbach/hallway-dash/internal/weather"
)

// publicConfig is what the frontend needs to know about server config;
// secrets (HA token, raw ICS URLs) never leave the backend.
type publicConfig struct {
	LocationLabel string `json:"locationLabel"`
	ClockFormat   string `json:"clockFormat"`
	AccentColor   string `json:"accentColor"`
	ShowIndoor    bool   `json:"showIndoor"`
}

func main() {
	cfg := config.Load()
	httpClient := &http.Client{Timeout: 15 * time.Second}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	weatherCache := pollcache.New("weather", 15*time.Minute, weather.Fetcher{
		Client:    httpClient,
		Latitude:  cfg.Latitude,
		Longitude: cfg.Longitude,
	}.Fetch)

	jokeCache := pollcache.New("joke", 60*time.Minute, joke.Fetcher{
		Client:     httpClient,
		PrimaryURL: cfg.JokeAPIURL,
	}.Fetch)

	alertsCache := pollcache.New("alerts", 5*time.Minute, alerts.Fetcher{
		Client:  httpClient,
		AGSCode: cfg.AGSCode,
	}.Fetch)

	calendarCache := pollcache.New("calendar", 15*time.Minute, calendar.Fetcher{
		Client: httpClient,
		Sources: []calendar.Source{
			{URL: cfg.CalendarICSURL, IsWaste: false},
			{URL: cfg.WasteICSURL, IsWaste: true},
		},
	}.Fetch)

	indoorCache := pollcache.New("indoor", 60*time.Second, homeassistant.Fetcher{
		Client:         httpClient,
		BaseURL:        cfg.HAURL,
		Token:          cfg.HAToken,
		TempEntity:     cfg.HATempEntity,
		HumidityEntity: cfg.HAHumidityEntity,
	}.Fetch)

	weatherCache.Start(ctx)
	jokeCache.Start(ctx)
	alertsCache.Start(ctx)
	calendarCache.Start(ctx)
	indoorCache.Start(ctx)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/config", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, publicConfig{
			LocationLabel: cfg.LocationLabel,
			ClockFormat:   cfg.ClockFormat,
			AccentColor:   cfg.AccentColor,
			ShowIndoor:    cfg.ShowIndoor(),
		})
	})
	mux.HandleFunc("GET /api/weather", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, weatherCache.Get())
	})
	mux.HandleFunc("GET /api/joke", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, jokeCache.Get())
	})
	mux.HandleFunc("GET /api/alerts", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]*alerts.Alert{"alert": alertsCache.Get()})
	})
	mux.HandleFunc("GET /api/calendar", func(w http.ResponseWriter, r *http.Request) {
		events := calendarCache.Get()
		if events == nil {
			events = []calendar.Event{}
		}
		writeJSON(w, map[string][]calendar.Event{"events": events})
	})
	mux.HandleFunc("GET /api/indoor", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, indoorCache.Get())
	})

	mux.Handle("/backgrounds/", http.StripPrefix("/backgrounds/", http.FileServer(http.Dir(cfg.BackgroundsDir))))
	mux.Handle("/", http.FileServer(http.Dir(cfg.StaticDir)))

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("hallway-dash listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(v)
}
