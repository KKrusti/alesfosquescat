//go:build ignore

// Local development server — ignored by `go build ./...` and vercel-go
// Run with: go run api/dev_server.go   (desde la raíz del proyecto)
// Serves /api/report, /api/stats, /api/resolve on :8787
// Vite proxies /api → :8787 (vite.config.ts)

package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
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

	ip := clientIP(r)
	ipHash := sha256hex(ip)
	today := todayInMadrid()

	db, err := openDB()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "db connection failed"})
		return
	}
	defer db.Close()

	// ── Read current streak state ──────────────────────────────────────
	var incidentStart sql.NullString
	_ = db.QueryRow(
		`SELECT to_char(incident_start, 'YYYY-MM-DD') FROM streak_state WHERE id = 1`,
	).Scan(&incidentStart)

	// If streak already active, nothing to do
	if incidentStart.Valid && incidentStart.String != "" {
		w.WriteHeader(http.StatusOK)
		writeJSON(w, map[string]interface{}{"already_active": true})
		return
	}

	if _, err = db.Exec(
		`INSERT INTO incidents (date) VALUES ($1) ON CONFLICT (date) DO NOTHING`,
		today,
	); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "failed to record incident"})
		return
	}

	// ── Restore or start fresh ─────────────────────────────────────────
	var resolveStart sql.NullString
	_ = db.QueryRow(`
		SELECT to_char(incident_start_saved, 'YYYY-MM-DD')
		  FROM daily_votes
		 WHERE ip_hash = 'community' AND date = $1 AND action = 'resolve'
		 LIMIT 1
	`, today).Scan(&resolveStart)

	activateFrom := today
	if resolveStart.Valid && resolveStart.String != "" {
		activateFrom = resolveStart.String
	}

	if _, err = db.Exec(`
		INSERT INTO streak_state (id, incident_start, updated_at)
		VALUES (1, $1, NOW())
		ON CONFLICT (id) DO UPDATE
		  SET incident_start = $1,
		      updated_at     = NOW()
	`, activateFrom); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "failed to activate streak"})
		return
	}

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
	ipHash := sha256hex(clientIP(r))

	// ── Llegir incident_start i longest_streak actuals ───────────────────
	var incidentStart sql.NullString
	var longestStreak int
	err = db.QueryRow(
		`SELECT to_char(incident_start, 'YYYY-MM-DD'), longest_streak FROM streak_state WHERE id = 1`,
	).Scan(&incidentStart, &longestStreak)
	if err != nil && err != sql.ErrNoRows {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "failed to read streak state"})
		return
	}

	// ── Calcular el streak actual ─────────────────────────────────────────
	currentStreak := 0
	if incidentStart.Valid && incidentStart.String != "" {
		currentStreak = calcStreak(incidentStart.String)
	}

	// ── Inserir fila resolve a daily_votes amb incident_start_saved ───────
	var savedStart interface{}
	if incidentStart.Valid && incidentStart.String != "" {
		savedStart = incidentStart.String
	}
	_, err = db.Exec(`
		INSERT INTO daily_votes (ip_hash, date, action, incident_start_saved)
		VALUES ($1, $2, 'resolve', $3)
		ON CONFLICT (ip_hash, date, action) DO NOTHING
	`, ipHash, today, savedStart)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "failed to record resolve"})
		return
	}

	// ── Actualitzar streak_state ──────────────────────────────────────────
	if currentStreak > 0 {
		newLongest := longestStreak
		if currentStreak > newLongest {
			newLongest = currentStreak
		}
		_, err = db.Exec(`
			UPDATE streak_state
			   SET longest_streak = $1,
			       incident_start = NULL,
			       updated_at     = NOW()
			 WHERE id = 1
		`, newLongest)
	} else {
		_, err = db.Exec(`
			UPDATE streak_state
			   SET incident_start = NULL,
			       updated_at     = NOW()
			 WHERE id = 1
		`)
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "failed to update streak"})
		return
	}

	// ── Eliminar el vot 'report' de l'IP per permetre re-reportar ────────
	_, _ = db.Exec(
		`DELETE FROM daily_votes WHERE ip_hash = $1 AND date = $2 AND action = 'report'`,
		ipHash, today,
	)

	w.WriteHeader(http.StatusOK)
	writeJSON(w, map[string]interface{}{"resolved": true, "date": today})
}

// calcStreak retorna (avui - startDate + 1) en dies. startDate és "YYYY-MM-DD".
func calcStreak(startDate string) int {
	loc, _ := time.LoadLocation("Europe/Madrid")
	if loc == nil {
		loc = time.UTC
	}
	start, err := time.ParseInLocation("2006-01-02", startDate, loc)
	if err != nil {
		return 0
	}
	now := time.Now().In(loc)
	todayMid := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	startMid := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, loc)
	days := int(todayMid.Sub(startMid).Hours()/24) + 1
	if days < 1 {
		return 0
	}
	return days
}

// ── /api/stats ────────────────────────────────────────────────────────────────

type StatsResponse struct {
	TotalThisYear           int    `json:"total_this_year"`
	LongestIncidentStreak   int    `json:"longest_incident_streak"`
	DaysSinceLastIncident   int    `json:"days_since_last_incident"`
	LastIncidentDate        string `json:"last_incident_date"`
	LongestNoIncidentStreak int    `json:"longest_no_incident_streak"`
	CurrentIncidentStreak   int    `json:"current_incident_streak"`
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

	// ── Llegir incident_start i longest_streak ───────────────────────────
	var incidentStartStr sql.NullString
	var longestStored int
	_ = db.QueryRow(
		`SELECT to_char(incident_start, 'YYYY-MM-DD'), longest_streak FROM streak_state WHERE id = 1`,
	).Scan(&incidentStartStr, &longestStored)

	currentStreak := 0
	if incidentStartStr.Valid && incidentStartStr.String != "" {
		currentStreak = calcStreak(incidentStartStr.String)
	}

	longestStreak := longestStored
	if currentStreak > longestStreak {
		longestStreak = currentStreak
	}

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
		dates = append(dates, d)
	}

	s := computeStats(dates, now)
	s.CurrentIncidentStreak = currentStreak
	s.LongestIncidentStreak = longestStreak
	writeJSON(w, s)
}

func computeStats(dates []time.Time, now time.Time) StatsResponse {
	s := StatsResponse{}
	s.TotalThisYear = len(dates)

	loc := now.Location()
	yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, loc)
	todayMidnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	if len(dates) == 0 {
		days := int(todayMidnight.Sub(yearStart).Hours()/24) + 1
		s.DaysSinceLastIncident = days
		s.LongestNoIncidentStreak = days
		return s
	}

	normalDay := func(t time.Time) time.Time {
		d := t.In(loc)
		return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, loc)
	}

	lastRaw := dates[len(dates)-1].In(loc)
	last := time.Date(lastRaw.Year(), lastRaw.Month(), lastRaw.Day(), 0, 0, 0, 0, loc)
	s.LastIncidentDate = last.Format("2006-01-02")
	s.DaysSinceLastIncident = int(todayMidnight.Sub(last).Hours() / 24)

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

	maxGap := int(normalDay(dates[0]).Sub(yearStart).Hours() / 24)
	for i := 1; i < len(dates); i++ {
		prev := normalDay(dates[i-1])
		this := normalDay(dates[i])
		if gap := int(this.Sub(prev).Hours()/24) - 1; gap > maxGap {
			maxGap = gap
		}
	}
	if s.DaysSinceLastIncident > maxGap {
		maxGap = s.DaysSinceLastIncident
	}
	s.LongestNoIncidentStreak = maxGap

	return s
}

// ── helpers ───────────────────────────────────────────────────────────────────

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

func clientIP(r *http.Request) string {
	for _, h := range []string{"CF-Connecting-IP", "X-Forwarded-For", "X-Real-IP"} {
		if val := r.Header.Get(h); val != "" {
			return strings.TrimSpace(strings.SplitN(val, ",", 2)[0])
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func sha256hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h)
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
			os.Setenv(key, val)
		}
	}
}
