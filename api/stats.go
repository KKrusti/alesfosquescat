package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// StatsResponse is the JSON payload returned by GET /api/stats.
type StatsResponse struct {
	TotalThisYear         int    `json:"total_this_year"`
	LongestIncidentStreak int    `json:"longest_incident_streak"`
	DaysSinceLastIncident int    `json:"days_since_last_incident"`
	LastIncidentDate      string `json:"last_incident_date"`
	NormalDaysThisYear    int    `json:"normal_days_this_year"`
	CurrentIncidentStreak int    `json:"current_incident_streak"`
}

// incidentRow holds one row from the incidents table.
// EndDate is nil when the outage is still active (end_date IS NULL).
type incidentRow struct {
	StartDate time.Time
	EndDate   *time.Time
}

// Handler — GET /api/stats
// Returns statistics for the current year computed from the incidents table.
func Handler(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w, "GET, OPTIONS")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeJSON(w, map[string]string{"error": "method not allowed"})
		return
	}

	db, err := openDB()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "db connection failed"})
		return
	}
	defer db.Close()

	loc, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)

	// ── Read all incidents for the current year ───────────────────────
	rows, err := db.Query(
		`SELECT date, end_date FROM incidents
		  WHERE EXTRACT(YEAR FROM date) = $1
		  ORDER BY date ASC`,
		now.Year(),
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "query failed"})
		return
	}
	defer rows.Close()

	var incidents []incidentRow
	for rows.Next() {
		var start time.Time
		var end sql.NullTime
		if err := rows.Scan(&start, &end); err != nil {
			continue
		}
		row := incidentRow{StartDate: start}
		if end.Valid {
			t := end.Time
			row.EndDate = &t
		}
		incidents = append(incidents, row)
	}

	writeJSON(w, computeStats(incidents, now))
}

// computeStats calculates all derived statistics from the incident rows.
// Each row represents one outage period (start_date, optional end_date).
// If end_date is nil, the outage is still active.
func computeStats(incidents []incidentRow, now time.Time) StatsResponse {
	loc := now.Location()
	todayMid := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, loc)
	daysElapsed := int(todayMid.Sub(yearStart).Hours()/24) + 1

	if len(incidents) == 0 {
		return StatsResponse{
			DaysSinceLastIncident: daysElapsed,
			NormalDaysThisYear:    daysElapsed,
		}
	}

	day := func(t time.Time) time.Time {
		d := t.In(loc)
		return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, loc)
	}

	totalDays := 0
	longest := 0
	currentStreak := 0
	var lastEndDay time.Time
	hasLastEnd := false

	for _, inc := range incidents {
		start := day(inc.StartDate)
		var dur int

		if inc.EndDate != nil {
			end := day(*inc.EndDate)
			dur = int(end.Sub(start).Hours()/24) + 1
			if !hasLastEnd || end.After(lastEndDay) {
				lastEndDay = end
				hasLastEnd = true
			}
		} else {
			// Active outage: duration counts from start to today (inclusive)
			dur = int(todayMid.Sub(start).Hours()/24) + 1
			if dur < 1 {
				dur = 1
			}
			currentStreak = dur
		}

		totalDays += dur
		if dur > longest {
			longest = dur
		}
	}

	var daysSince int
	var lastDateStr string

	if currentStreak > 0 {
		// Active outage: find its start date for display
		for _, inc := range incidents {
			if inc.EndDate == nil {
				lastDateStr = day(inc.StartDate).Format("2006-01-02")
				break
			}
		}
		daysSince = 0
	} else if hasLastEnd {
		daysSince = int(todayMid.Sub(lastEndDay).Hours() / 24)
		lastDateStr = lastEndDay.Format("2006-01-02")
	} else {
		daysSince = daysElapsed
	}

	normalDays := daysElapsed - totalDays
	if normalDays < 0 {
		normalDays = 0
	}

	return StatsResponse{
		TotalThisYear:         totalDays,
		LongestIncidentStreak: longest,
		DaysSinceLastIncident: daysSince,
		LastIncidentDate:      lastDateStr,
		NormalDaysThisYear:    normalDays,
		CurrentIncidentStreak: currentStreak,
	}
}

// ── helpers ──────────────────────────────────────────────────────────────────

func openDB() (*sql.DB, error) {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(0)
	return db, nil
}

func setCORSHeaders(w http.ResponseWriter, methods string) {
	if origin := os.Getenv("ALLOWED_ORIGIN"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
	}
	w.Header().Set("Access-Control-Allow-Methods", methods)
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	_ = json.NewEncoder(w).Encode(v)
}
