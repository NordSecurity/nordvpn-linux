package cli

import (
	"fmt"
	"io"
	"regexp"
)

var timespanRegexp = regexp.MustCompile(`-?\d+\s*[A-Za-z]*\s*`)

func parseTimespan(ts string) (int, int, int, int, error) {
	var years int = 0
	var months int = 0
	var days int = 0
	var seconds int = 0
	matches := timespanRegexp.FindAllString(ts, -1)

	if matches == nil {
		return 0, 0, 0, 0, fmt.Errorf("Time span parsing error: '%s'", ts)
	}

	for _, m := range matches {
		var num int
		var unit string
		n, err := fmt.Sscanf(m, "%d%s", &num, &unit)

		if err == io.EOF && n == 1 && unit == "" {
			// pass
		} else if err != nil {
			return 0, 0, 0, 0, fmt.Errorf("Time span argument '%s' parsing error: %s", m, err)
		}

		// Using time span syntax as defined by systemd
		// https://www.freedesktop.org/software/systemd/man/latest/systemd.time.html
		switch unit {
		case "", "seconds", "second", "sec", "s":
			seconds += num
		case "minutes", "minute", "min", "m":
			seconds += num * 60
		case "hours", "hour", "hr", "h":
			seconds += num * 3600
		case "days", "day", "d":
			days += num
		case "weeks", "week", "w":
			days += num * 7
		case "months", "month", "M":
			months += num
		case "years", "year", "y":
			years += num
		default:
			return 0, 0, 0, 0, fmt.Errorf("Time span unit parsing error: '%s'", unit)
		}
	}

	return years, months, days, seconds, nil
}
