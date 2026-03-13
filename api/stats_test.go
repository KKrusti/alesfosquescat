package handler

import (
	"testing"
	"time"
)

// madrid returns a time.Time in Europe/Madrid for the given date at noon.
func madrid(year int, month time.Month, day int) time.Time {
	loc, _ := time.LoadLocation("Europe/Madrid")
	return time.Date(year, month, day, 12, 0, 0, 0, loc)
}

// madridPtr returns a pointer to a time.Time in Europe/Madrid.
func madridPtr(year int, month time.Month, day int) *time.Time {
	t := madrid(year, month, day)
	return &t
}

func TestComputeStats(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Madrid")
	// Fix "today" to 1 March 2026 for deterministic results.
	now := time.Date(2026, 3, 1, 15, 0, 0, 0, loc)

	tests := []struct {
		name      string
		incidents []incidentRow
		wantTotal int
		wantDays  int // days_since_last_incident
		wantNorm  int // normal_days_this_year
		wantLong  int // longest_incident_streak
		wantCurr  int // current_incident_streak
	}{
		{
			name:      "no incidents",
			incidents: nil,
			wantTotal: 0,
			wantDays:  60, // 1 Jan → 1 Mar = 60 days elapsed
			wantNorm:  60,
			wantLong:  0,
			wantCurr:  0,
		},
		{
			name: "single active incident started today",
			incidents: []incidentRow{
				{StartDate: madrid(2026, 3, 1), EndDate: nil},
			},
			wantTotal: 1,
			wantDays:  0,
			wantNorm:  59,
			wantLong:  1,
			wantCurr:  1,
		},
		{
			name: "single resolved incident yesterday",
			incidents: []incidentRow{
				{StartDate: madrid(2026, 2, 28), EndDate: madridPtr(2026, 2, 28)},
			},
			wantTotal: 1,
			wantDays:  1, // today - feb28 = 1
			wantNorm:  59,
			wantLong:  1,
			wantCurr:  0,
		},
		{
			name: "single resolved incident 5 days ago",
			incidents: []incidentRow{
				{StartDate: madrid(2026, 2, 24), EndDate: madridPtr(2026, 2, 24)},
			},
			wantTotal: 1,
			wantDays:  5,
			wantNorm:  59,
			wantLong:  1,
			wantCurr:  0,
		},
		{
			name: "three-day active outage starting feb27",
			incidents: []incidentRow{
				{StartDate: madrid(2026, 2, 27), EndDate: nil},
			},
			wantTotal: 3, // feb27, feb28, mar01
			wantDays:  0,
			wantNorm:  57,
			wantLong:  3,
			wantCurr:  3,
		},
		{
			name: "two resolved outages — longer one wins",
			incidents: []incidentRow{
				{StartDate: madrid(2026, 1, 5), EndDate: madridPtr(2026, 1, 7)},  // 3 days
				{StartDate: madrid(2026, 2, 10), EndDate: madridPtr(2026, 2, 11)}, // 2 days
			},
			wantTotal: 5,
			wantDays:  18, // mar1 - feb11 = 18
			wantNorm:  55,
			wantLong:  3,
			wantCurr:  0,
		},
		{
			name: "two non-consecutive single-day resolved outages",
			incidents: []incidentRow{
				{StartDate: madrid(2026, 1, 10), EndDate: madridPtr(2026, 1, 11)}, // 2 days
				{StartDate: madrid(2026, 1, 13), EndDate: madridPtr(2026, 1, 14)}, // 2 days
			},
			wantTotal: 4,
			wantDays:  46, // mar1 - jan14 = 46
			wantNorm:  56,
			wantLong:  2,
			wantCurr:  0,
		},
		{
			name: "active long outage alongside prior resolved one",
			incidents: []incidentRow{
				{StartDate: madrid(2026, 1, 5), EndDate: madridPtr(2026, 1, 6)}, // 2 days, closed
				{StartDate: madrid(2026, 2, 27), EndDate: nil},                  // 3 days, active
			},
			wantTotal: 5,
			wantDays:  0,
			wantNorm:  55,
			wantLong:  3,
			wantCurr:  3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := computeStats(tc.incidents, now)

			if got.TotalThisYear != tc.wantTotal {
				t.Errorf("TotalThisYear: got %d, want %d", got.TotalThisYear, tc.wantTotal)
			}
			if got.DaysSinceLastIncident != tc.wantDays {
				t.Errorf("DaysSinceLastIncident: got %d, want %d", got.DaysSinceLastIncident, tc.wantDays)
			}
			if got.NormalDaysThisYear != tc.wantNorm {
				t.Errorf("NormalDaysThisYear: got %d, want %d", got.NormalDaysThisYear, tc.wantNorm)
			}
			if got.LongestIncidentStreak != tc.wantLong {
				t.Errorf("LongestIncidentStreak: got %d, want %d", got.LongestIncidentStreak, tc.wantLong)
			}
			if got.CurrentIncidentStreak != tc.wantCurr {
				t.Errorf("CurrentIncidentStreak: got %d, want %d", got.CurrentIncidentStreak, tc.wantCurr)
			}
		})
	}
}
