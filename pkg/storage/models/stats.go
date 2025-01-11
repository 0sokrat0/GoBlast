package models

import "time"

type Stats struct {
	TotalSent     int64            `json:"total_sent"`
	TotalFailed   int64            `json:"total_failed"`
	ByContentType map[string]int64 `json:"by_content_type"`
	StartTime     time.Time        `json:"-"`
	TimeSpent     float64          `json:"time_spent"`

	ProcessedCount int64            `json:"-"`
	ExpectedCount  int64            `json:"-"`
	ErrorCounts    map[string]int64 `json:"-"`
}
