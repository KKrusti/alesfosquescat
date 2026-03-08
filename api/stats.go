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

	// ── Read streak state ─────────────────────────────────────────────
	// incident_start: data d'inici de l'apagón actiu (NULL si no n'hi ha).
	// longest_streak: màxim historial guardat (actualitzat en resoldre).
	// El current_streak es calcula en temps real com (avui - incident_start + 1).
	var incidentStart sql.NullTime
	var longestStored int
	_ = db.QueryRow(
		`SELECT incident_start, longest_streak FROM streak_state WHERE id = 1`,
	).Scan(&incidentStart, &longestStored)

	// Calcular current streak dinàmicament
	currentStreak := 0
	if incidentStart.Valid {
		todayMid := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		startIn := incidentStart.Time.In(loc)
		startMid := time.Date(startIn.Year(), startIn.Month(), startIn.Day(), 0, 0, 0, 0, loc)
		days := int(todayMid.Sub(startMid).Hours()/24) + 1
		if days > 0 {
			currentStreak = days
		}
	}

	// longest és el màxim entre l'historial guardat i el streak actiu actual
	longestStreak := longestStored
	if currentStreak > longestStreak {
		longestStreak = currentStreak
	}

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

	loc := now.Location()
	yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, loc)
	// Medianoche d'avui en timezone Madrid (Truncate opera en UTC, no en local)
	todayMidnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	// Dies transcorreguts des de l'1 de gener fins avui (inclusiu)
	daysElapsed := int(todayMidnight.Sub(yearStart).Hours()/24) + 1

	if len(dates) == 0 {
		s.DaysSinceLastIncident = daysElapsed
		s.NormalDaysThisYear = daysElapsed
		return s
	}

	// ── Last incident ─────────────────────────────────────────────────
	// Normalitzem la data de l'incident a medianoche Madrid
	lastRaw := dates[len(dates)-1].In(loc)
	last := time.Date(lastRaw.Year(), lastRaw.Month(), lastRaw.Day(), 0, 0, 0, 0, loc)
	s.LastIncidentDate = last.Format("2006-01-02")
	s.DaysSinceLastIncident = int(todayMidnight.Sub(last).Hours() / 24)

	// ── Longest incident streak (consecutive days with incidents) ─────
	// normalDay normalitza un time.Time a medianoche en timezone Madrid
	normalDay := func(t time.Time) time.Time {
		d := t.In(loc)
		return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, loc)
	}

	maxStreak, cur := 1, 1
	for i := 1; i < len(dates); i++ {
		prev := normalDay(dates[i-1])
		this := normalDay(dates[i])
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
			prev := normalDay(dates[i-1])
			this := normalDay(dates[i])
			if int(this.Sub(prev).Hours()/24) == 1 {
				cur++
			} else {
				break
			}
		}
		s.CurrentIncidentStreak = cur
	}

	// ── Dies amb normalitat aquest any ───────────────────────────────
	// Dies transcorreguts des de l'1 de gener fins avui, menys les nits amb incident.
	normalDays := daysElapsed - len(dates)
	if normalDays < 0 {
		normalDays = 0
	}
	s.NormalDaysThisYear = normalDays

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
