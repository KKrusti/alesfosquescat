package handler

import (
	"testing"
	"time"
)

// TestTodayInMadrid verifies that todayInMadrid returns a valid YYYY-MM-DD string.
func TestTodayInMadrid(t *testing.T) {
	got := todayInMadrid()
	if len(got) != 10 {
		t.Errorf("todayInMadrid() = %q, expected length 10", got)
	}
	if got[4] != '-' || got[7] != '-' {
		t.Errorf("todayInMadrid() = %q, expected YYYY-MM-DD format", got)
	}
	// Verify the date is parseable in Europe/Madrid timezone
	loc, _ := time.LoadLocation("Europe/Madrid")
	if _, err := time.ParseInLocation("2006-01-02", got, loc); err != nil {
		t.Errorf("todayInMadrid() = %q, parse error: %v", got, err)
	}
}
