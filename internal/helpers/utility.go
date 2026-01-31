package helpers

import (
	"fmt"
	"time"
)

func GenerateDateRange(start, end string) ([]string, error) {
	layout := "20060102"

	startTime, err := time.Parse(layout, start)
	if err != nil {
		return nil, err
	}

	endTime := startTime
	if end != "" {
		endTime, err = time.Parse(layout, end)
		if err != nil {
			return nil, err
		}
	}

	if startTime.After(endTime) {
		return nil, fmt.Errorf("start_date > end_date")
	}

	var dates []string
	for d := startTime; !d.After(endTime); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d.Format(layout))
	}
	return dates, nil
}

func FormatBigNumber(n float64) string {
	if n >= 1e12 {
		return fmt.Sprintf("%.2f T", n/1e12)
	}
	if n >= 1e9 {
		return fmt.Sprintf("%.2f M", n/1e9)
	}
	if n >= 1e6 {
		return fmt.Sprintf("%.2f Jt", n/1e6)
	}
	return fmt.Sprintf("%.0f", n)
}
