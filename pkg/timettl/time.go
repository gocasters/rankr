package timettl

import (
	"fmt"
	"time"
)

// GetYear returns current year as string (e.g., "2025")
func GetYear() string {
	now := time.Now()
	return fmt.Sprintf("%d", now.Year())
}

// GetMonth returns current year-month as string (e.g., "2025-11")
func GetMonth() string {
	now := time.Now()
	return fmt.Sprintf("%d-%02d", now.Year(), now.Month())
}

// GetWeek returns current ISO week as string (e.g., "2025-W44")
func GetWeek() string {
	now := time.Now()
	year, week := now.ISOWeek()
	return fmt.Sprintf("%d-W%02d", year, week)
}

// GetDay returns current date as string (e.g., "2025-11-02")
func GetDay() string {
	now := time.Now()
	return fmt.Sprintf("%d-%02d-%02d", now.Year(), now.Month(), now.Day())
}

// CalculateEndOfPeriod returns the expiration time for a given timeframe
// This ensures all keys for the same period expire at the same time
func CalculateEndOfPeriod(timeframe string) (time.Time, error) {
	now := time.Now().UTC()

	switch timeframe {
	case "daily":
		// Expire at end of current day (23:59:59 UTC)
		return time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.UTC), nil

	case "weekly":
		// Expire at end of current ISO week (Sunday 23:59:59 UTC)
		// ISO week starts Monday, ends Sunday
		weekday := int(now.Weekday())
		if weekday == 0 { // Sunday
			weekday = 7
		}
		daysUntilSunday := 7 - weekday
		endOfWeek := now.AddDate(0, 0, daysUntilSunday)
		return time.Date(endOfWeek.Year(), endOfWeek.Month(), endOfWeek.Day(), 23, 59, 59, 0, time.UTC), nil

	case "monthly":
		// Expire at end of current month (last second)
		year := now.Year()
		month := now.Month()

		// First day of next month
		firstOfNextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC)

		// Last second of current month
		return firstOfNextMonth.Add(-time.Second), nil

	case "yearly":
		// Expire at end of current year (December 31, 23:59:59 UTC)
		return time.Date(now.Year(), 12, 31, 23, 59, 59, 0, time.UTC), nil

	case "all_time":
		// No expiration for all_time keys
		return time.Time{}, nil

	default:
		return time.Time{}, fmt.Errorf("unknown timeframe: %s", timeframe)
	}
}

// GetExpirationDuration returns the duration until end of period
// Useful for debugging or calculating time remaining
func GetExpirationDuration(timeframe string) (time.Duration, error) {
	endTime, err := CalculateEndOfPeriod(timeframe)
	if err != nil {
		return 0, err
	}

	if endTime.IsZero() {
		// No expiration for all_time
		return 0, nil
	}

	now := time.Now().UTC()
	duration := endTime.Sub(now)

	// Ensure non-negative duration
	if duration < 0 {
		return 0, fmt.Errorf("calculated expiration time is in the past")
	}

	return duration, nil
}

// IsWithinPeriod checks if a given time is within the same period as now
// Useful for validating if a key belongs to current period
func IsWithinPeriod(t time.Time, timeframe string) bool {
	now := time.Now().UTC()
	t = t.UTC()

	switch timeframe {
	case "daily":
		return t.Year() == now.Year() && t.YearDay() == now.YearDay()

	case "weekly":
		tYear, tWeek := t.ISOWeek()
		nowYear, nowWeek := now.ISOWeek()
		return tYear == nowYear && tWeek == nowWeek

	case "monthly":
		return t.Year() == now.Year() && t.Month() == now.Month()

	case "yearly":
		return t.Year() == now.Year()

	case "all_time":
		return true

	default:
		return false
	}
}

// GetPeriodKey returns the period string for current time
// Useful for generating leaderboard keys
func GetPeriodKey(timeframe string) (string, error) {
	switch timeframe {
	case "daily":
		return GetDay(), nil
	case "weekly":
		return GetWeek(), nil
	case "monthly":
		return GetMonth(), nil
	case "yearly":
		return GetYear(), nil
	case "all_time":
		return "", nil
	default:
		return "", fmt.Errorf("unknown timeframe: %s", timeframe)
	}
}
