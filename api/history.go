package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// IncidentPeriod represents a single outage period from the incidents table.
type IncidentPeriod struct {
	StartDate string `json:"start_date"` // YYYY-MM-DD
	EndDate   string `json:"end_date"`   // YYYY-MM-DD, empty string if still active
	Days      int    `json:"days"`       // duration in days; ongoing outages count up to today
}

// Handler — GET /api/history
// Returns all incident periods for the current year, most-recent first.
// Each row in the incidents table is one period (start_date, end_date).
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
	todayMid := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	rows, err := db.Query(
		`SELECT to_char(date, 'YYYY-MM-DD'), end_date
		   FROM incidents
		  WHERE EXTRACT(YEAR FROM date) = $1
		  ORDER BY date DESC`,
		now.Year(),
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "query failed"})
		return
	}
	defer rows.Close()

	periods := []IncidentPeriod{}
	for rows.Next() {
		var startStr string
		var endDate sql.NullTime
		if err := rows.Scan(&startStr, &endDate); err != nil {
			continue
		}

		start, err := time.ParseInLocation("2006-01-02", startStr, loc)
		if err != nil {
			continue
		}
		startMid := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, loc)

		var endStr string
		var days int
		if endDate.Valid {
			endMid := time.Date(endDate.Time.Year(), endDate.Time.Month(), endDate.Time.Day(), 0, 0, 0, 0, loc)
			days = int(endMid.Sub(startMid).Hours()/24) + 1
			endStr = endMid.Format("2006-01-02")
		} else {
			// Active outage: count days up to today
			days = int(todayMid.Sub(startMid).Hours()/24) + 1
			if days < 1 {
				days = 1
			}
		}

		periods = append(periods, IncidentPeriod{
			StartDate: startStr,
			EndDate:   endStr,
			Days:      days,
		})
	}

	w.WriteHeader(http.StatusOK)
	writeJSON(w, periods)
}

// ── helpers ───────────────────────────────────────────────────────────────────

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
