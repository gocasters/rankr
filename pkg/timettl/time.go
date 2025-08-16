package timettl

import (
	"fmt"
	"time"
)

func GetYear() string {
	now := time.Now()
	return fmt.Sprintf("%d", now.Year())
}

func GetMonth() string {
	now := time.Now()
	return fmt.Sprintf("%d-%02d", now.Year(), now.Month())
}

func GetWeek() string {
	now := time.Now()
	year, week := now.ISOWeek()
	return fmt.Sprintf("%d-W%02d", year, week)
}
