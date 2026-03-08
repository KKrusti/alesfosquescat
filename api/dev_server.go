//go:build ignore

// Local development server — ignored by Vercel build and `go build ./...`
// Run with: go run api/dev_server.go
// Serves /api/report and /api/stats on :8787; Vite proxies /api → :8787

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
	// Load DATABASE_URL from .env.local if not already set
	if os.Getenv("DATABASE_URL") == "" {
		loadEnvLocal()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/report", reportHandler)
	mux.HandleFunc("/api/stats", statsHandler)

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

	var alreadyVoted bool
	err = db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM daily_votes WHERE ip_hash=$1 AND date=$2)`,
		ipHash, today,
	).Scan(&alreadyVoted)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "vote check failed"})
		return
	}
	if alreadyVoted {
		w.WriteHeader(http.StatusConflict)
		writeJSON(w, map[string]interface{}{
			"already_voted": true,
			"message":       "Ja ho sabem, tio 🕯️",
		})
		return
	}

	result, err := db.Exec(
		`INSERT INTO incidents (date) VALUES ($1) ON CONFLICT (date) DO NOTHING`,
		today,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "failed to record incident"})
		return
	}

	if n, _ := result.RowsAffected(); n == 1 {
		if _, err = db.Exec(`
			INSERT INTO streak_state (id, current_streak, longest_streak)
			VALUES (1, 1, 1)
			ON CONFLICT (id) DO UPDATE
			  SET current_streak = streak_state.current_streak + 1,
			      longest_streak = GREATEST(streak_state.longest_streak, streak_state.current_streak + 1),
			      updated_at     = NOW()
		`); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeJSON(w, map[string]string{"error": "failed to update streak"})
			return
		}
	}

	if _, err = db.Exec(
		`INSERT INTO daily_votes (ip_hash, date) VALUES ($1, $2)`,
		ipHash, today,
	); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "failed to record vote"})
		return
	}

	w.WriteHeader(http.StatusOK)
	writeJSON(w, map[string]interface{}{"success": true, "date": today})
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

	var currentStreak, longestStreak int
	_ = db.QueryRow(
		`SELECT current_streak, longest_streak FROM streak_state WHERE id = 1`,
	).Scan(&currentStreak, &longestStreak)

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

	yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	todayMidnight := now.Truncate(24 * time.Hour)

	if len(dates) == 0 {
		days := int(todayMidnight.Sub(yearStart).Hours()/24) + 1
		s.DaysSinceLastIncident = days
		s.LongestNoIncidentStreak = days
		return s
	}

	last := dates[len(dates)-1].Truncate(24 * time.Hour)
	s.LastIncidentDate = last.Format("2006-01-02")
	s.DaysSinceLastIncident = int(todayMidnight.Sub(last).Hours() / 24)

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

	maxGap := int(dates[0].Truncate(24*time.Hour).Sub(yearStart).Hours() / 24)
	for i := 1; i < len(dates); i++ {
		prev := dates[i-1].Truncate(24 * time.Hour)
		this := dates[i].Truncate(24 * time.Hour)
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

// loadEnvLocal reads DATABASE_URL from ../.env.local (project root)
func loadEnvLocal() {
	data, err := os.ReadFile("../.env.local")
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
		val := strings.Trim(strings.TrimSpace(parts[1]), `"`)
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}
