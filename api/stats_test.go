package handler

import (
	"testing"
	"time"
)

// madrid returns a time.Time in Europe/Madrid for the given date.
func madrid(year int, month time.Month, day int) time.Time {
	loc, _ := time.LoadLocation("Europe/Madrid")
	return time.Date(year, month, day, 12, 0, 0, 0, loc)
}

func TestComputeStats(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Madrid")
	// Fix "today" to 1 March 2026 for deterministic results.
	now := time.Date(2026, 3, 1, 15, 0, 0, 0, loc)

	tests := []struct {
		name      string
		dates     []time.Time
		wantTotal int
		wantDays  int // days_since_last_incident
		wantNorm  int // normal_days_this_year
		wantLong  int // longest_streak (from computeStats; caller overrides with streak_state)
		wantCurr  int // current_streak
	}{
		{
			name:      "no incidents",
			dates:     nil,
			wantTotal: 0,
			wantDays:  60, // 1 Jan → 1 Mar = 60 days elapsed
			wantNorm:  60,
			wantLong:  0,
			wantCurr:  0,
		},
		{
			name:      "single incident today",
			dates:     []time.Time{madrid(2026, 3, 1)},
			wantTotal: 1,
			wantDays:  0,
			wantNorm:  59,
			wantLong:  1,
			wantCurr:  1,
		},
		{
			name:      "single incident yesterday",
			dates:     []time.Time{madrid(2026, 2, 28)},
			wantTotal: 1,
			wantDays:  1,
			wantNorm:  59,
			wantLong:  1,
			wantCurr:  1, // still within ≤1 day window
		},
		{
			name:      "single incident 5 days ago",
			dates:     []time.Time{madrid(2026, 2, 24)},
			wantTotal: 1,
			wantDays:  5,
			wantNorm:  59,
			wantLong:  1,
			wantCurr:  0, // streak expired
		},
		{
			name: "three consecutive days ending today",
			dates: []time.Time{
				madrid(2026, 2, 27),
				madrid(2026, 2, 28),
				madrid(2026, 3, 1),
			},
			wantTotal: 3,
			wantDays:  0,
			wantNorm:  57,
			wantLong:  3,
			wantCurr:  3,
		},
		{
			name: "two separate streaks — longest wins",
			dates: []time.Time{
				madrid(2026, 1, 5),
				madrid(2026, 1, 6),
				madrid(2026, 1, 7), // streak of 3
				madrid(2026, 2, 10),
				madrid(2026, 2, 11), // streak of 2
			},
			wantTotal: 5,
			wantDays:  18, // Feb 11 → Mar 1: (28-11) + 1 = 18 days
			wantNorm:  55,
			wantLong:  3,
			wantCurr:  0,
		},
		{
			name: "gap in otherwise consecutive days",
			dates: []time.Time{
				madrid(2026, 1, 10),
				madrid(2026, 1, 11),
				madrid(2026, 1, 13), // gap on 12th
				madrid(2026, 1, 14),
			},
			wantTotal: 4,
			wantDays:  46, // 2026-01-14 to 2026-03-01
			wantNorm:  56,
			wantLong:  2,
			wantCurr:  0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := computeStats(tc.dates, now)

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
