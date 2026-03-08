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
	TotalThisYear           int    `json:"total_this_year"`
	LongestIncidentStreak   int    `json:"longest_incident_streak"`
	DaysSinceLastIncident   int    `json:"days_since_last_incident"`
	LastIncidentDate        string `json:"last_incident_date"`
	LongestNoIncidentStreak int    `json:"longest_no_incident_streak"`
	CurrentIncidentStreak   int    `json:"current_incident_streak"`
}

// Handler — GET /api/stats
// Retorna les estadístiques de l'any en curs calculades des de la taula incidents.
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

	// ── Read streak counters ──────────────────────────────────────────
	var currentStreak, longestStreak int
	_ = db.QueryRow(
		`SELECT current_streak, longest_streak FROM streak_state WHERE id = 1`,
	).Scan(&currentStreak, &longestStreak)

	// ── Read incident dates for date-based stats ──────────────────────
	rows, err := db.Query(
		`SELECT date FROM incidents
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

	var dates []time.Time
	for rows.Next() {
		var d time.Time
		if err := rows.Scan(&d); err != nil {
			continue
		}
		dates = append(dates, d)
	}

	s := computeStats(dates, now)
	s.CurrentIncidentStreak = currentStreak
	s.LongestIncidentStreak = longestStreak
	writeJSON(w, s)
}

// computeStats calculates all derived statistics from the sorted incident dates.
func computeStats(dates []time.Time, now time.Time) StatsResponse {
	s := StatsResponse{}
	s.TotalThisYear = len(dates)

	yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	todayMidnight := now.Truncate(24 * time.Hour)

	if len(dates) == 0 {
		daysSinceYearStart := int(todayMidnight.Sub(yearStart).Hours()/24) + 1
		s.DaysSinceLastIncident = daysSinceYearStart
		s.LongestNoIncidentStreak = daysSinceYearStart
		return s
	}

	// ── Last incident ─────────────────────────────────────────────────
	last := dates[len(dates)-1].Truncate(24 * time.Hour)
	s.LastIncidentDate = last.Format("2006-01-02")
	s.DaysSinceLastIncident = int(todayMidnight.Sub(last).Hours() / 24)

	// ── Longest incident streak (consecutive days with incidents) ─────
	maxStreak, cur := 1, 1
	for i := 1; i < len(dates); i++ {
		prev := dates[i-1].Truncate(24 * time.Hour)
		this := dates[i].Truncate(24 * time.Hour)
		if int(this.Sub(prev).Hours()/24) == 1 {
			cur++
			if cur > maxStreak {
				maxStreak = cur
			}
		} else {
			cur = 1
		}
	}
	s.LongestIncidentStreak = maxStreak

	// ── Current incident streak (only if last incident ≤ 1 day ago) ──
	if s.DaysSinceLastIncident <= 1 {
		cur = 1
		for i := len(dates) - 1; i > 0; i-- {
			prev := dates[i-1].Truncate(24 * time.Hour)
			this := dates[i].Truncate(24 * time.Hour)
			if int(this.Sub(prev).Hours()/24) == 1 {
				cur++
			} else {
				break
			}
		}
		s.CurrentIncidentStreak = cur
	}

	// ── Longest gap without incidents (max consecutive clean days) ────
	// Gap before first incident
	maxGap := int(dates[0].Truncate(24*time.Hour).Sub(yearStart).Hours() / 24)

	// Gaps between consecutive incidents
	for i := 1; i < len(dates); i++ {
		prev := dates[i-1].Truncate(24 * time.Hour)
		this := dates[i].Truncate(24 * time.Hour)
		gap := int(this.Sub(prev).Hours()/24) - 1
		if gap > maxGap {
			maxGap = gap
		}
	}

	// Gap from last incident to today
	if s.DaysSinceLastIncident > maxGap {
		maxGap = s.DaysSinceLastIncident
	}
	s.LongestNoIncidentStreak = maxGap

	return s
}

// ── helpers (used only by this handler) ──────────────────────────────────────

func openDB() (*sql.DB, error) {
	return sql.Open("postgres", os.Getenv("DATABASE_URL"))
}

func setCORSHeaders(w http.ResponseWriter, methods string) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", methods)
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	_ = json.NewEncoder(w).Encode(v)
}
