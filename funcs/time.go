package funcs

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	gotime "time"

	"github.com/flanksource/commons/duration"
	"github.com/flanksource/commons/properties"
	"github.com/flanksource/commons/utils"
	"github.com/timberio/go-datemath"

	"github.com/flanksource/gomplate/v3/conv"
	"github.com/flanksource/gomplate/v3/time"
)

// TimeNS -
//
// Deprecated: don't use
func TimeNS() *TimeFuncs {
	return &TimeFuncs{
		ANSIC:       gotime.ANSIC,
		UnixDate:    gotime.UnixDate,
		RubyDate:    gotime.RubyDate,
		RFC822:      gotime.RFC822,
		RFC822Z:     gotime.RFC822Z,
		RFC850:      gotime.RFC850,
		RFC1123:     gotime.RFC1123,
		RFC1123Z:    gotime.RFC1123Z,
		RFC3339:     gotime.RFC3339,
		RFC3339Nano: gotime.RFC3339Nano,
		Kitchen:     gotime.Kitchen,
		Stamp:       gotime.Stamp,
		StampMilli:  gotime.StampMilli,
		StampMicro:  gotime.StampMicro,
		StampNano:   gotime.StampNano,
	}
}

// AddTimeFuncs -
//
// Deprecated: use CreateTimeFuncs instead
func AddTimeFuncs(f map[string]interface{}) {
	for k, v := range CreateTimeFuncs(context.Background()) {
		f[k] = v
	}
}

// CreateTimeFuncs -
func CreateTimeFuncs(ctx context.Context) map[string]interface{} {
	ns := &TimeFuncs{
		ctx:         ctx,
		ANSIC:       gotime.ANSIC,
		UnixDate:    gotime.UnixDate,
		RubyDate:    gotime.RubyDate,
		RFC822:      gotime.RFC822,
		RFC822Z:     gotime.RFC822Z,
		RFC850:      gotime.RFC850,
		RFC1123:     gotime.RFC1123,
		RFC1123Z:    gotime.RFC1123Z,
		RFC3339:     gotime.RFC3339,
		RFC3339Nano: gotime.RFC3339Nano,
		Kitchen:     gotime.Kitchen,
		Stamp:       gotime.Stamp,
		StampMilli:  gotime.StampMilli,
		StampMicro:  gotime.StampMicro,
		StampNano:   gotime.StampNano,
	}

	return map[string]interface{}{
		"time":              func() interface{} { return ns },
		"in_business_hours": ns.InBusinessHour,
		"parseDateTime":     ParseDateTime,
	}
}

// TimeFuncs -
type TimeFuncs struct {
	ctx         context.Context
	ANSIC       string
	UnixDate    string
	RubyDate    string
	RFC822      string
	RFC822Z     string
	RFC850      string
	RFC1123     string
	RFC1123Z    string
	RFC3339     string
	RFC3339Nano string
	Kitchen     string
	Stamp       string
	StampMilli  string
	StampMicro  string
	StampNano   string
}

// ZoneName - return the local system's time zone's name
func (TimeFuncs) ZoneName() string {
	return time.ZoneName()
}

// ZoneOffset - return the local system's time zone's name
func (TimeFuncs) ZoneOffset() int {
	return time.ZoneOffset()
}

// Parse -
func (TimeFuncs) Parse(layout string, value interface{}) (gotime.Time, error) {
	return gotime.Parse(layout, conv.ToString(value))
}

// ParseLocal -
func (f TimeFuncs) ParseLocal(layout string, value interface{}) (gotime.Time, error) {
	tz := os.Getenv("TZ")
	if tz == "" {
		tz = "Local"
	}
	return f.ParseInLocation(layout, tz, value)
}

// ParseInLocation -
func (TimeFuncs) ParseInLocation(layout, location string, value interface{}) (gotime.Time, error) {
	loc, err := gotime.LoadLocation(location)
	if err != nil {
		return gotime.Time{}, err
	}
	return gotime.ParseInLocation(layout, conv.ToString(value), loc)
}

// Now -
func (TimeFuncs) Now() gotime.Time {
	return gotime.Now()
}

// Unix - convert UNIX time (in seconds since the UNIX epoch) into a time.Time for further processing
// Takes a string or number (int or float)
func (TimeFuncs) Unix(in interface{}) (gotime.Time, error) {
	sec, nsec, err := parseNum(in)
	if err != nil {
		return gotime.Time{}, err
	}
	return gotime.Unix(sec, nsec), nil
}

// Nanosecond -
func (TimeFuncs) Nanosecond(n interface{}) gotime.Duration {
	return gotime.Nanosecond * gotime.Duration(conv.ToInt64(n))
}

// Microsecond -
func (TimeFuncs) Microsecond(n interface{}) gotime.Duration {
	return gotime.Microsecond * gotime.Duration(conv.ToInt64(n))
}

// Millisecond -
func (TimeFuncs) Millisecond(n interface{}) gotime.Duration {
	return gotime.Millisecond * gotime.Duration(conv.ToInt64(n))
}

// Second -
func (TimeFuncs) Second(n interface{}) gotime.Duration {
	return gotime.Second * gotime.Duration(conv.ToInt64(n))
}

// Minute -
func (TimeFuncs) Minute(n interface{}) gotime.Duration {
	return gotime.Minute * gotime.Duration(conv.ToInt64(n))
}

// Hour -
func (TimeFuncs) Hour(n interface{}) gotime.Duration {
	return gotime.Hour * gotime.Duration(conv.ToInt64(n))
}

// ParseDuration -
func (TimeFuncs) ParseDuration(n interface{}) (gotime.Duration, error) {
	// Using commons duration package to support more time units
	d, err := duration.ParseDuration(conv.ToString(n))
	return gotime.Duration(d), err
}

// Since -
func (TimeFuncs) Since(n gotime.Time) gotime.Duration {
	return gotime.Since(n)
}

// Until -
func (TimeFuncs) Until(n gotime.Time) gotime.Duration {
	return gotime.Until(n)
}

// InTimeRange reports whether the time of day of t falls within [start, end] (both inclusive).
// start and end are "HH:MM" or "HH:MM:SS" strings.
//
// Examples:
//
//	InTimeRange(t, "09:00", "17:00")    → true for 9:00:00–17:00:59
//	InTimeRange(t, "09:30", "17:30")    → true for 9:30:00–17:30:59
//	InTimeRange(t, "09:00:00", "17:00:30") → true for 9:00:00–17:00:30
func (TimeFuncs) InTimeRange(t any, start, end string) (bool, error) {
	ts, err := toTime(t)
	if err != nil {
		return false, err
	}
	startSecs, err := parseTimeOfDay(start)
	if err != nil {
		return false, err
	}
	endSecs, err := parseTimeOfDay(end)
	if err != nil {
		return false, err
	}
	tSecs := ts.Hour()*3600 + ts.Minute()*60 + ts.Second()
	return tSecs >= startSecs && tSecs <= endSecs, nil
}

// parseTimeOfDay parses a "HH:MM" or "HH:MM:SS" string and returns
// the total number of seconds since midnight.
func parseTimeOfDay(s string) (int, error) {
	for _, layout := range []string{"15:04:05", "15:04"} {
		if t, err := gotime.Parse(layout, s); err == nil {
			return t.Hour()*3600 + t.Minute()*60 + t.Second(), nil
		}
	}
	return 0, fmt.Errorf("cannot parse %q as a time of day (expected HH:MM or HH:MM:SS)", s)
}

// toTime converts a timestamp value to a time.Time.
// Supported input types: time.Time, or a string in RFC3339Nano / RFC3339 /
// "2006-01-02T15:04:05" / "2006-01-02 15:04:05" format.
func toTime(v any) (gotime.Time, error) {
	switch t := v.(type) {
	case gotime.Time:
		return t, nil
	case string:
		layouts := []string{
			gotime.RFC3339Nano,
			gotime.RFC3339,
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
		}
		for _, layout := range layouts {
			if ts, err := gotime.Parse(layout, t); err == nil {
				return ts, nil
			}
		}
		return gotime.Time{}, fmt.Errorf("cannot parse %q as a timestamp", t)
	}
	return gotime.Time{}, fmt.Errorf("cannot convert %T to a timestamp", v)
}

// InBusinessHour returns nil when no business hours are configured.
func (TimeFuncs) InBusinessHour(value string) (any, error) {
	in, err := inBusinessHour(value)
	if err != nil {
		return nil, err
	}
	if in == nil {
		return nil, nil
	}
	return *in, nil
}

func inBusinessHour(value string) (*bool, error) {
	intervals, err := properties.BusinessHours()
	if err != nil {
		return nil, err
	}
	if len(intervals) == 0 {
		return nil, nil
	}

	t := utils.ParseTime(value)
	if t == nil {
		return nil, fmt.Errorf("failed to parse time %q", value)
	}

	in := intervals.ContainsTime(*t)
	return &in, nil
}

// convert a number input to a pair of int64s, representing the integer portion and the decimal remainder
// this can handle a string as well as any integer or float type
// precision is at the "nano" level (i.e. 1e+9)
func parseNum(in interface{}) (integral int64, fractional int64, err error) {
	if s, ok := in.(string); ok {
		ss := strings.Split(s, ".")
		if len(ss) > 2 {
			return 0, 0, fmt.Errorf("can not parse '%s' as a number - too many decimal points", s)
		}
		if len(ss) == 1 {
			integral, err := strconv.ParseInt(s, 0, 64)
			return integral, 0, err
		}
		integral, err := strconv.ParseInt(ss[0], 0, 64)
		if err != nil {
			return integral, 0, err
		}
		fractional, err = strconv.ParseInt(padRight(ss[1], "0", 9), 0, 64)
		return integral, fractional, err
	}
	if s, ok := in.(fmt.Stringer); ok {
		return parseNum(s.String())
	}
	if i, ok := in.(int); ok {
		return int64(i), 0, nil
	}
	if u, ok := in.(uint64); ok {
		return int64(u), 0, nil
	}
	if f, ok := in.(float64); ok {
		return 0, 0, fmt.Errorf("can not parse floating point number (%f) - use a string instead", f)
	}
	if in == nil {
		return 0, 0, nil
	}
	return 0, 0, nil
}

// pads a number with zeroes
func padRight(in, pad string, length int) string {
	for {
		in += pad
		if len(in) > length {
			return in[0:length]
		}
	}
}

// ParseDateTime handles various datetime formats including datemath expressions
func ParseDateTime(timeStr string) *gotime.Time {
	if timeStr == "" {
		return nil
	}

	// Handle datemath expressions using the go-datemath library
	if strings.HasPrefix(timeStr, "now") || strings.Contains(timeStr, "/") {
		parsedTime, err := datemath.ParseAndEvaluate(timeStr, datemath.WithNow(gotime.Now()))
		if err != nil {
			return nil
		}
		return &parsedTime
	}

	// Handle Unix timestamps (seconds)
	if val, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
		if val > 1000000000 && val < 10000000000 { // Valid Unix timestamp range
			t := gotime.Unix(val, 0)
			return &t
		}
	}

	// Handle Unix timestamps (milliseconds)
	if val, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
		if val > 1000000000000 && val < 10000000000000 { // Valid Unix timestamp in ms
			t := gotime.Unix(val/1000, (val%1000)*1000000)
			return &t
		}
	}

	// Try parsing as RFC3339Nano (with milliseconds)
	if t, err := gotime.Parse(gotime.RFC3339Nano, timeStr); err == nil {
		return &t
	}

	// Try parsing as RFC3339
	if t, err := gotime.Parse(gotime.RFC3339, timeStr); err == nil {
		return &t
	}

	// Try parsing as RFC3339 without timezone but with milliseconds
	if t, err := gotime.Parse("2006-01-02T15:04:05.999999999", timeStr); err == nil {
		return &t
	}

	// Try parsing as RFC3339 without timezone
	if t, err := gotime.Parse("2006-01-02T15:04:05", timeStr); err == nil {
		return &t
	}

	// Try parsing as ISO date
	if t, err := gotime.Parse("2006-01-02", timeStr); err == nil {
		return &t
	}

	// Try parsing as date with time and milliseconds
	if t, err := gotime.Parse("2006-01-02 15:04:05.999999999", timeStr); err == nil {
		return &t
	}

	// Try parsing as date with time (common log format)
	if t, err := gotime.Parse("2006-01-02 15:04:05", timeStr); err == nil {
		return &t
	}

	// Try parsing as date with time and timezone
	if t, err := gotime.Parse("2006-01-02 15:04:05 MST", timeStr); err == nil {
		return &t
	}

	return nil
}
