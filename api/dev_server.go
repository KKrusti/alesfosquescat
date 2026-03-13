//go:build ignore

// Local development server — ignored by `go build ./...` and vercel-go
// Run with: go run api/dev_server.go   (desde la raíz del proyecto)
// Serves /api/report, /api/stats, /api/resolve, /api/history, /api/interactions on :8787
// Vite proxies /api → :8787 (vite.config.ts)

package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	// Carrega DATABASE_URL des de .env.local (arrel del projecte) si no és al entorn
	if os.Getenv("DATABASE_URL") == "" {
		loadDotEnv("../.env.local")
	}
	if os.Getenv("DATABASE_URL") == "" {
		loadDotEnv(".env.local")
	}

	if os.Getenv("DATABASE_URL") == "" {
		log.Fatal("ERROR: DATABASE_URL no trobat. Comprova que .env.local existeix a l'arrel del projecte.")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/report", reportHandler)
	mux.HandleFunc("/api/stats", statsHandler)
	mux.HandleFunc("/api/resolve", resolveHandler)
	mux.HandleFunc("/api/history", historyHandler)
	mux.HandleFunc("/api/interactions", interactionsHandler)

	addr := ":8787"
	log.Printf("API dev server → http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

// ── /api/report ───────────────────────────────────────────────────────────────

func reportHandler(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w, "POST, OPTIONS")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeJSON(w, map[string]string{"error": "method not allowed"})
		return
	}

	today := todayInMadrid()

	db, err := openDB()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "db connection failed"})
		return
	}
	defer db.Close()

	// Check if an outage is already active (end_date IS NULL)
	var dummy int
	err = db.QueryRow(`SELECT 1 FROM incidents WHERE end_date IS NULL LIMIT 1`).Scan(&dummy)
	if err == nil {
		w.WriteHeader(http.StatusOK)
		writeJSON(w, map[string]interface{}{"already_active": true})
		return
	}
	if err != sql.ErrNoRows {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "failed to check active outage"})
		return
	}

	// Insert new incident (open, end_date = NULL)
	if _, err = db.Exec(
		`INSERT INTO incidents (date) VALUES ($1) ON CONFLICT (date) DO NOTHING`,
		today,
	); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "failed to record incident"})
		return
	}

	_, _ = db.Exec(`INSERT INTO interaction_log (action) VALUES ('report')`)

	w.WriteHeader(http.StatusOK)
	writeJSON(w, map[string]interface{}{"success": true, "date": today})
}

// ── /api/resolve ──────────────────────────────────────────────────────────────

func resolveHandler(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w, "POST, OPTIONS")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
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

	today := todayInMadrid()

	// Find the active (open) incident
	var incidentID int
	var startDate string
	err = db.QueryRow(`
		SELECT id, to_char(date, 'YYYY-MM-DD')
		  FROM incidents
		 WHERE end_date IS NULL
		 LIMIT 1
	`).Scan(&incidentID, &startDate)

	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusOK)
		writeJSON(w, map[string]interface{}{"resolved": false, "already_resolved": true})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "failed to read active incident"})
		return
	}

	if startDate == today {
		// False report — delete it
		if _, err = db.Exec(`DELETE FROM incidents WHERE id = $1`, incidentID); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeJSON(w, map[string]string{"error": "failed to delete incident"})
			return
		}
	} else {
		// Real outage — close it
		if _, err = db.Exec(
			`UPDATE incidents SET end_date = $1 WHERE id = $2`,
			today, incidentID,
		); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeJSON(w, map[string]string{"error": "failed to close incident"})
			return
		}
	}

	_, _ = db.Exec(`INSERT INTO interaction_log (action) VALUES ('resolve')`)

	w.WriteHeader(http.StatusOK)
	writeJSON(w, map[string]interface{}{"resolved": true, "date": today})
}

// ── /api/stats ────────────────────────────────────────────────────────────────

type StatsResponse struct {
	TotalThisYear         int    `json:"total_this_year"`
	LongestIncidentStreak int    `json:"longest_incident_streak"`
	DaysSinceLastIncident int    `json:"days_since_last_incident"`
	LastIncidentDate      string `json:"last_incident_date"`
	NormalDaysThisYear    int    `json:"normal_days_this_year"`
	CurrentIncidentStreak int    `json:"current_incident_streak"`
}

type incidentRow struct {
	StartDate time.Time
	EndDate   *time.Time
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
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

// ── /api/history ──────────────────────────────────────────────────────────────

func historyHandler(w http.ResponseWriter, r *http.Request) {
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

	loc, _ := time.LoadLocation("Europe/Madrid")
	if loc == nil {
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

	type IncidentPeriod struct {
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
		Days      int    `json:"days"`
	}

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
			days = int(todayMid.Sub(startMid).Hours()/24) + 1
			if days < 1 {
				days = 1
			}
		}

		periods = append(periods, IncidentPeriod{StartDate: startStr, EndDate: endStr, Days: days})
	}

	w.WriteHeader(http.StatusOK)
	writeJSON(w, periods)
}

// ── /api/interactions ──────────────────────────────────────────────────────────

func interactionsHandler(w http.ResponseWriter, r *http.Request) {
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

	loc, _ := time.LoadLocation("Europe/Madrid")
	if loc == nil {
		loc = time.UTC
	}

	rows, err := db.Query(
		`SELECT action, created_at FROM interaction_log ORDER BY created_at DESC LIMIT 100`,
	)
	if err != nil {
		log.Printf("[interactions] query error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "query failed"})
		return
	}
	defer rows.Close()

	type Entry struct {
		Action string `json:"action"`
		At     string `json:"at"`
	}
	entries := []Entry{}
	for rows.Next() {
		var action string
		var createdAt time.Time
		if err := rows.Scan(&action, &createdAt); err != nil {
			continue
		}
		entries = append(entries, Entry{
			Action: action,
			At:     createdAt.In(loc).Format("02-01-2006 15:04"),
		})
	}

	w.WriteHeader(http.StatusOK)
	writeJSON(w, entries)
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
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", methods)
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	_ = json.NewEncoder(w).Encode(v)
}

func todayInMadrid() string {
	loc, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		loc = time.UTC
	}
	return time.Now().In(loc).Format("2006-01-02")
}

// loadDotEnv lee variables de entorno desde un archivo .env (formato KEY="VALUE")
func loadDotEnv(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if len(val) >= 2 && ((val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'')) {
			val = val[1 : len(val)-1]
		}
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
}
