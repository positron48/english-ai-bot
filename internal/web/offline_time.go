package web

import (
	"strings"
	"time"
)

func parseOfflineTimestamp(values ...string) *time.Time {
	for _, raw := range values {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		layouts := []string{
			time.RFC3339Nano,
			time.RFC3339,
			"2006-01-02T15:04:05.000Z",
			"2006-01-02T15:04:05Z",
		}
		for _, layout := range layouts {
			if t, err := time.Parse(layout, raw); err == nil {
				utc := t.UTC()
				return &utc
			}
		}
	}
	return nil
}

func offlineTimestampOrNow(values ...string) time.Time {
	if t := parseOfflineTimestamp(values...); t != nil {
		return *t
	}
	return time.Now().UTC()
}
