package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// IncidentPeriod represents a contiguous block of outage days.
type IncidentPeriod struct {
	StartDate string `json:"start_date"` // YYYY-MM-DD
	Days      int    `json:"days"`
}

// Handler — GET /api/history
// Returns all incident periods for the current year, grouped from consecutive
// dates in the incidents table, sorted most-recent first.
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

	rows, err := db.Query(
		`SELECT date FROM incidents WHERE EXTRACT(YEAR FROM date) = $1 ORDER BY date ASC`,
		now.Year(),
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "query failed"})
		return
	}
	defer rows.Close()

	var dates []time.Time
	for rows.Next() {
		var d time.Time
		if err := rows.Scan(&d); err != nil {
			continue
		}
		dates = append(dates, d.In(loc))
	}

	periods := groupIntoPeriods(dates, loc)

	// Reverse to return most-recent first
	for i, j := 0, len(periods)-1; i < j; i, j = i+1, j-1 {
		periods[i], periods[j] = periods[j], periods[i]
	}

	w.WriteHeader(http.StatusOK)
	writeJSON(w, periods)
}

// groupIntoPeriods groups consecutive dates into IncidentPeriod entries.
func groupIntoPeriods(dates []time.Time, loc *time.Location) []IncidentPeriod {
	if len(dates) == 0 {
		return []IncidentPeriod{}
	}

	day := func(t time.Time) time.Time {
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
	}

	var periods []IncidentPeriod
	start := day(dates[0])
	count := 1

	for i := 1; i < len(dates); i++ {
		prev := day(dates[i-1])
		curr := day(dates[i])
		if int(curr.Sub(prev).Hours()/24) == 1 {
			count++
		} else {
			periods = append(periods, IncidentPeriod{StartDate: start.Format("2006-01-02"), Days: count})
			start = curr
			count = 1
		}
	}
	periods = append(periods, IncidentPeriod{StartDate: start.Format("2006-01-02"), Days: count})

	return periods
}

// ── helpers ───────────────────────────────────────────────────────────────────

func openDB() (*sql.DB, error) {
	return sql.Open("postgres", os.Getenv("DATABASE_URL"))
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
