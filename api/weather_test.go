package handler

import (
	"testing"
)

func TestParseWeatherAlert(t *testing.T) {
	times := []string{"2026-03-08", "2026-03-09", "2026-03-10", "2026-03-11"}

	tests := []struct {
		name      string
		sums      []float64
		probs     []int
		wantAlert bool
		wantDay   int
		wantMM    float64
		wantProb  int
	}{
		{
			name:      "no rain at all",
			sums:      []float64{0, 0, 0, 0},
			probs:     []int{10, 5, 0, 20},
			wantAlert: false,
		},
		{
			name:      "high mm but low probability",
			sums:      []float64{0, 12.0, 0, 0},
			probs:     []int{0, 50, 0, 0},
			wantAlert: false,
		},
		{
			name:      "high probability but low mm",
			sums:      []float64{0, 2.0, 0, 0},
			probs:     []int{0, 90, 0, 0},
			wantAlert: false,
		},
		{
			name:      "exactly at threshold — not triggered (> not >=)",
			sums:      []float64{5.0, 0, 0, 0},
			probs:     []int{60, 0, 0, 0},
			wantAlert: false,
		},
		{
			name:      "alert today (index 0)",
			sums:      []float64{18.5, 0, 0, 0},
			probs:     []int{75, 0, 0, 0},
			wantAlert: true,
			wantDay:   0,
			wantMM:    18.5,
			wantProb:  75,
		},
		{
			name:      "alert in 3 days (index 3)",
			sums:      []float64{0, 0, 0, 9.2},
			probs:     []int{0, 0, 0, 80},
			wantAlert: true,
			wantDay:   3,
			wantMM:    9.2,
			wantProb:  80,
		},
		{
			name:      "picks first matching day, not last",
			sums:      []float64{0, 6.0, 0, 20.0},
			probs:     []int{0, 70, 0, 95},
			wantAlert: true,
			wantDay:   1,
			wantMM:    6.0,
			wantProb:  70,
		},
		{
			name:      "empty slices",
			sums:      []float64{},
			probs:     []int{},
			wantAlert: false,
		},
		{
			name:      "mismatched slice lengths — shorter probs wins",
			sums:      []float64{10.0, 10.0},
			probs:     []int{90},
			wantAlert: true,
			wantDay:   0,
			wantMM:    10.0,
			wantProb:  90,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseWeatherAlert(times, tc.sums, tc.probs)
			if got.Alert != tc.wantAlert {
				t.Errorf("Alert: got %v, want %v", got.Alert, tc.wantAlert)
			}
			if !tc.wantAlert {
				return
			}
			if got.DaysUntil != tc.wantDay {
				t.Errorf("DaysUntil: got %d, want %d", got.DaysUntil, tc.wantDay)
			}
			if got.MM != tc.wantMM {
				t.Errorf("MM: got %v, want %v", got.MM, tc.wantMM)
			}
			if got.Prob != tc.wantProb {
				t.Errorf("Prob: got %d, want %d", got.Prob, tc.wantProb)
			}
		})
	}
}
