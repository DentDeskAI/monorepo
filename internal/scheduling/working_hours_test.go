package scheduling

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseWorkingHours_ArrayObjectAndNullFormats(t *testing.T) {
	raw := json.RawMessage(`{
		"mon": ["09:00", "18:00"],
		"tue": {"open":"10:00","close":"19:00"},
		"sun": null
	}`)

	got := parseWorkingHours(raw)

	require.Contains(t, got, "mon")
	require.Contains(t, got, "tue")
	require.Contains(t, got, "sun")
	assert.Equal(t, &dayHours{Open: "09:00", Close: "18:00"}, got["mon"])
	assert.Equal(t, &dayHours{Open: "10:00", Close: "19:00"}, got["tue"])
	assert.Nil(t, got["sun"])
}

func TestParseWorkingHours_InvalidJSONReturnsEmptyMap(t *testing.T) {
	assert.Empty(t, parseWorkingHours(json.RawMessage(`{invalid`)))
	assert.Empty(t, parseWorkingHours(nil))
}

func TestWeekdayKey(t *testing.T) {
	loc := time.FixedZone("UTC+6", 6*60*60)
	tests := map[string]time.Time{
		"mon": time.Date(2026, 5, 4, 12, 0, 0, 0, loc),
		"tue": time.Date(2026, 5, 5, 12, 0, 0, 0, loc),
		"wed": time.Date(2026, 5, 6, 12, 0, 0, 0, loc),
		"thu": time.Date(2026, 5, 7, 12, 0, 0, 0, loc),
		"fri": time.Date(2026, 5, 8, 12, 0, 0, 0, loc),
		"sat": time.Date(2026, 5, 9, 12, 0, 0, 0, loc),
		"sun": time.Date(2026, 5, 10, 12, 0, 0, 0, loc),
	}

	for want, input := range tests {
		assert.Equal(t, want, weekdayKey(input))
	}
}

func TestDayAtAndFloorToDay(t *testing.T) {
	loc := time.FixedZone("UTC+6", 6*60*60)
	base := time.Date(2026, 5, 4, 15, 45, 12, 123, loc)

	at, err := dayAt(base, "09:30")
	require.NoError(t, err)
	assert.Equal(t, time.Date(2026, 5, 4, 9, 30, 0, 0, loc), at)

	floor := floorToDay(base)
	assert.Equal(t, time.Date(2026, 5, 4, 0, 0, 0, 0, loc), floor)
}

func TestDayAt_InvalidFormat(t *testing.T) {
	_, err := dayAt(time.Now(), "930")
	assert.Error(t, err)

	_, err = dayAt(time.Now(), "aa:bb")
	assert.Error(t, err)
}
