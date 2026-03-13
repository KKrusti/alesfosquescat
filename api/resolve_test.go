package handler

import (
	"testing"
	"time"
)

func TestCalcStreak(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Madrid")

	// Fix "now" to a deterministic time so sub-second execution doesn't affect results.
	// We monkey-patch via the function signature: calcStreak uses time.Now() internally,
	// so we test edge cases relative to the actual day boundary.
	// Instead, we derive expected values from today and known offsets.
	nowMadrid := time.Now().In(loc)
	todayStr := nowMadrid.Format("2006-01-02")
	yesterdayStr := nowMadrid.AddDate(0, 0, -1).Format("2006-01-02")
	threeDaysAgoStr := nowMadrid.AddDate(0, 0, -3).Format("2006-01-02")

	tests := []struct {
		name      string
		startDate string
		wantMin   int // minimum acceptable value (at least this many days)
		wantMax   int // maximum acceptable value
	}{
		{
			name:      "started today — streak is 1",
			startDate: todayStr,
			wantMin:   1,
			wantMax:   1,
		},
		{
			name:      "started yesterday — streak is 2",
			startDate: yesterdayStr,
			wantMin:   2,
			wantMax:   2,
		},
		{
			name:      "started 3 days ago — streak is 4",
			startDate: threeDaysAgoStr,
			wantMin:   4,
			wantMax:   4,
		},
		{
			name:      "invalid date — returns 0",
			startDate: "not-a-date",
			wantMin:   0,
			wantMax:   0,
		},
		{
			name:      "future date — returns 0",
			startDate: nowMadrid.AddDate(0, 0, 1).Format("2006-01-02"),
			wantMin:   0,
			wantMax:   0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := calcStreak(tc.startDate)
			if got < tc.wantMin || got > tc.wantMax {
				t.Errorf("calcStreak(%q) = %d, want between %d and %d",
					tc.startDate, got, tc.wantMin, tc.wantMax)
			}
		})
	}
}
