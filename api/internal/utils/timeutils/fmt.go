package timeutils

import (
	"fmt"
	"time"
)

func FmtDuration(d time.Duration) string {
	d = d.Round(time.Second)
	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if days > 0 {
		return fmt.Sprintf("%d day%s, %02d:%02d:%02d", days, pluralize(days), h, m, s)
	}

	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}

	if m > 0 {
		if s == 0 {
			return fmt.Sprintf("%d minute%s", m, pluralize(m))
		}

		return fmt.Sprintf("%d minute%s, %d second%s", m, pluralize(m), s, pluralize(s))
	}

	return fmt.Sprintf("%d second%s", s, pluralize(s))
}

func RoundToSingleUnit(d time.Duration) string {
	d = d.Round(time.Second)

	days := d / (24 * time.Hour)
	hours := (d % (24 * time.Hour)) / time.Hour
	minutes := (d % time.Hour) / time.Minute
	seconds := (d % time.Minute) / time.Second

	if days > 0 {
		if hours >= 12 {
			return fmt.Sprintf("%d days", days+1)
		}
		return fmt.Sprintf("%d day%s", days, pluralize(days))
	}

	if hours > 0 {
		if minutes >= 30 {
			return fmt.Sprintf("%d hours", hours+1)
		}
		return fmt.Sprintf("%d hour%s", hours, pluralize(hours))
	}

	if minutes > 0 {
		if seconds >= 30 {
			return fmt.Sprintf("%d minutes", minutes+1)
		}
		return fmt.Sprintf("%d minute%s", minutes, pluralize(minutes))
	}

	return fmt.Sprintf("%d second%s", seconds, pluralize(seconds))
}

func pluralize(n time.Duration) string {
	if n == 1 {
		return ""
	}
	return "s"
}
