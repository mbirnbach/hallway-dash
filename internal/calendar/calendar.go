// Package calendar fetches and parses the two ICS feeds (personal +
// waste/bin collection), expanding RRULE recurrences within the window
// needed to fill the frontend's rolling 4-week grid.
package calendar

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/teambition/rrule-go"
)

type Event struct {
	Date    string `json:"date"` // YYYY-MM-DD
	Title   string `json:"title"`
	IsWaste bool   `json:"isWaste"`
}

type Source struct {
	URL     string
	IsWaste bool
}

type Fetcher struct {
	Client  *http.Client
	Sources []Source
}

// Fetch returns every occurrence (recurring events expanded) across all
// configured feeds that falls within [-35, +35] days of now, which
// always comfortably covers the rolling 4-week grid the frontend
// computes independently (current week's Monday through 27 days later).
func (f Fetcher) Fetch(ctx context.Context) ([]Event, error) {
	now := time.Now()
	rangeStart := now.AddDate(0, 0, -35)
	rangeEnd := now.AddDate(0, 0, 35)

	var all []Event
	for _, src := range f.Sources {
		if src.URL == "" {
			continue
		}
		events, err := fetchOne(ctx, f.Client, src.URL, src.IsWaste, rangeStart, rangeEnd)
		if err != nil {
			// One bad feed shouldn't blank out the other; log and continue.
			continue
		}
		all = append(all, events...)
	}
	return all, nil
}

func fetchOne(ctx context.Context, client *http.Client, url string, isWaste bool, rangeStart, rangeEnd time.Time) ([]Event, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ics feed returned status %d", resp.StatusCode)
	}

	cal, err := ics.ParseCalendar(resp.Body)
	if err != nil {
		return nil, err
	}

	var events []Event
	for _, ev := range cal.Events() {
		summaryProp := ev.GetProperty(ics.ComponentPropertySummary)
		if summaryProp == nil || strings.TrimSpace(summaryProp.Value) == "" {
			continue
		}
		title := ics.FromText(summaryProp.Value)

		start, allDay, ok := eventStart(ev)
		if !ok {
			continue
		}

		exceptions := exceptionDates(ev)

		var occurrences []time.Time
		if rruleProp := ev.GetProperty(ics.ComponentPropertyRrule); rruleProp != nil && rruleProp.Value != "" {
			occurrences = expandRRule(start, rruleProp.Value, rangeStart, rangeEnd)
		} else {
			occurrences = []time.Time{start}
		}

		for _, occ := range occurrences {
			if occ.Before(rangeStart) || occ.After(rangeEnd) {
				continue
			}
			key := dateKey(occ, allDay)
			if exceptions[key] {
				continue
			}
			events = append(events, Event{Date: key, Title: title, IsWaste: isWaste})
		}
	}
	return events, nil
}

// eventStart returns the DTSTART of an event along with whether it's an
// all-day (VALUE=DATE) event, since golang-ical exposes those via two
// different accessors.
func eventStart(ev *ics.VEvent) (time.Time, bool, bool) {
	dtstart := ev.GetProperty(ics.ComponentPropertyDtStart)
	if dtstart == nil {
		return time.Time{}, false, false
	}
	allDay := false
	for _, v := range dtstart.ICalParameters["VALUE"] {
		if v == "DATE" {
			allDay = true
		}
	}
	if allDay {
		t, err := ev.GetAllDayStartAt()
		if err != nil {
			return time.Time{}, false, false
		}
		return t, true, true
	}
	t, err := ev.GetStartAt()
	if err != nil {
		return time.Time{}, false, false
	}
	return t, false, true
}

func exceptionDates(ev *ics.VEvent) map[string]bool {
	out := map[string]bool{}
	exdates, err := ev.GetExDates()
	if err != nil {
		return out
	}
	for _, d := range exdates {
		out[dateKey(d, true)] = true
	}
	return out
}

// expandRRule uses rrule-go to expand the recurrence rule (as found
// verbatim in the RRULE property) between rangeStart and rangeEnd.
func expandRRule(start time.Time, rruleValue string, rangeStart, rangeEnd time.Time) []time.Time {
	option, err := rrule.StrToROptionInLocation(rruleValue, start.Location())
	if err != nil {
		return []time.Time{start}
	}
	option.Dtstart = start
	rule, err := rrule.NewRRule(*option)
	if err != nil {
		return []time.Time{start}
	}
	return rule.Between(rangeStart, rangeEnd, true)
}

func dateKey(t time.Time, allDay bool) string {
	if !allDay {
		t = t.Local()
	}
	return t.Format("2006-01-02")
}
