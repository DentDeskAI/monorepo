package scheduler

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type dayHours struct {
	Open  string // "09:00"
	Close string // "19:00"
}

// parseWorkingHours handles the array format used in migrations:
//
//	{"mon":["09:00","19:00"],"sun":null}
//
// Returns a map from lowercase 3-letter weekday key to dayHours (nil = closed).
func parseWorkingHours(raw json.RawMessage) map[string]*dayHours {
	result := map[string]*dayHours{}
	if raw == nil {
		return result
	}

	// Unmarshal as map of raw values so we can handle null and arrays.
	var raw2 map[string]json.RawMessage
	if err := json.Unmarshal(raw, &raw2); err != nil {
		return result
	}

	for day, v := range raw2 {
		if string(v) == "null" {
			result[day] = nil
			continue
		}
		// Try array form: ["09:00","19:00"]
		var arr []string
		if err := json.Unmarshal(v, &arr); err == nil && len(arr) == 2 {
			result[day] = &dayHours{Open: arr[0], Close: arr[1]}
			continue
		}
		// Try object form: {"open":"09:00","close":"19:00"}
		var obj struct {
			Open  string `json:"open"`
			Close string `json:"close"`
		}
		if err := json.Unmarshal(v, &obj); err == nil && obj.Open != "" {
			result[day] = &dayHours{Open: obj.Open, Close: obj.Close}
		}
	}
	return result
}

// weekdayKey returns the lowercase 3-letter key used in working_hours JSON.
func weekdayKey(t time.Time) string {
	switch t.Weekday() {
	case time.Monday:
		return "mon"
	case time.Tuesday:
		return "tue"
	case time.Wednesday:
		return "wed"
	case time.Thursday:
		return "thu"
	case time.Friday:
		return "fri"
	case time.Saturday:
		return "sat"
	default:
		return "sun"
	}
}

// dayAt returns a time.Time set to the given day at hhmm ("09:00").
func dayAt(base time.Time, hhmm string) (time.Time, error) {
	parts := strings.SplitN(hhmm, ":", 2)
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("invalid time %q", hhmm)
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return time.Time{}, fmt.Errorf("invalid time %q", hhmm)
	}
	return time.Date(base.Year(), base.Month(), base.Day(), h, m, 0, 0, base.Location()), nil
}

// floorToDay returns t truncated to midnight.
func floorToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
