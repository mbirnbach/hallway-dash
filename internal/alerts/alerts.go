// Package alerts fetches active NINA/BBK civil-protection warnings for a
// configured region (by AGS code) and picks the single highest-severity
// alert to display, mirroring the design reference's client-side logic.
package alerts

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

type Alert struct {
	Headline string `json:"headline"`
	Severity string `json:"severity"`
}

var severityRank = map[string]int{
	"Extreme":  3,
	"Severe":   3,
	"Moderate": 2,
	"Minor":    1,
}

type ninaItem struct {
	I18nTitle struct {
		De string `json:"de"`
	} `json:"i18nTitle"`
	Payload struct {
		Data struct {
			Headline string `json:"headline"`
			Severity string `json:"severity"`
		} `json:"data"`
	} `json:"payload"`
}

type Fetcher struct {
	Client  *http.Client
	AGSCode string
}

// Fetch returns nil (no error) when there's no active alert or the
// feature isn't configured — the wall display simply hides the banner.
func (f Fetcher) Fetch(ctx context.Context) (*Alert, error) {
	ags := strings.TrimSpace(f.AGSCode)
	if ags == "" {
		return nil, nil
	}
	id := ags
	if len(id) < 12 {
		id = (id + "000000000000")[:12]
	} else {
		id = id[:12]
	}

	url := fmt.Sprintf("https://warnung.bund.de/api31/dashboard/%s.json", id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := f.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nina dashboard returned status %d", resp.StatusCode)
	}

	var items []ninaItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}

	alerts := make([]Alert, 0, len(items))
	for _, item := range items {
		headline := item.I18nTitle.De
		if headline == "" {
			headline = item.Payload.Data.Headline
		}
		if headline == "" {
			headline = "Warnung"
		}
		severity := item.Payload.Data.Severity
		if severity == "" {
			severity = "Minor"
		}
		alerts = append(alerts, Alert{Headline: headline, Severity: severity})
	}

	sort.SliceStable(alerts, func(i, j int) bool {
		return severityRank[alerts[i].Severity] > severityRank[alerts[j].Severity]
	})
	return &alerts[0], nil
}
