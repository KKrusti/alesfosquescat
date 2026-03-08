package handler

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// Handler — POST /api/report
// Records an outage for today.
//
// Guard: if a streak is already active (incident_start IS NOT NULL), returns
// already_active — the frontend uses the live stats to show this before calling.
//
// If no active streak:
//   - If the community resolved today, restores incident_start_saved so the
//     counter returns to the pre-resolve value (false-resolve correction).
//   - Otherwise starts a fresh streak from today.
func Handler(w http.ResponseWriter, r *http.Request) {
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

	// ── Rate limit: max 10 requests per IP per minute ─────────────────
	if exceeded, err := checkRateLimit(db, ipHash, "report", 10); err == nil && exceeded {
		w.WriteHeader(http.StatusTooManyRequests)
		writeJSON(w, map[string]string{"error": "massa peticions, espera un minut"})
		return
	}

	// ── Read current streak state ──────────────────────────────────────
	var incidentStart sql.NullString
	_ = db.QueryRow(
		`SELECT to_char(incident_start, 'YYYY-MM-DD') FROM streak_state WHERE id = 1`,
	).Scan(&incidentStart)

	// If a streak is already active, nothing to do — the frontend shows this
	// state before even calling the API, but handle it here as a safety net.
	if incidentStart.Valid && incidentStart.String != "" {
		w.WriteHeader(http.StatusOK)
		writeJSON(w, map[string]interface{}{"already_active": true})
		return
	}

	// ── Upsert today's incident ────────────────────────────────────────
	if _, err = db.Exec(
		`INSERT INTO incidents (date) VALUES ($1) ON CONFLICT (date) DO NOTHING`,
		today,
	); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "failed to record incident"})
		return
	}

	// ── Determine streak start: restore if resolved today, else start fresh ──
	var resolveStart sql.NullString
	_ = db.QueryRow(`
		SELECT to_char(incident_start_saved, 'YYYY-MM-DD')
		  FROM daily_votes
		 WHERE ip_hash = 'community' AND date = $1 AND action = 'resolve'
		 LIMIT 1
	`, today).Scan(&resolveStart)

	restored := resolveStart.Valid && resolveStart.String != ""
	activateFrom := today
	if restored {
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

	// Best-effort log — do not block the response if this fails
	_, _ = db.Exec(`INSERT INTO interaction_log (action) VALUES ('report')`)

	w.WriteHeader(http.StatusOK)
	writeJSON(w, map[string]interface{}{
		"success":  true,
		"restored": restored,
		"date":     today,
	})
}

// ── helpers ──────────────────────────────────────────────────────────────────

func openDB() (*sql.DB, error) {
	return sql.Open("postgres", os.Getenv("DATABASE_URL"))
}

// setCORSHeaders restricts cross-origin access to the configured ALLOWED_ORIGIN.
// If the env var is not set, no cross-origin access is granted (same-origin only).
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

// clientIP returns the real client IP. On Vercel, the platform appends the
// real client IP as the last entry in X-Forwarded-For, so we read the
// rightmost entry to prevent spoofing via client-controlled prepended IPs.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[len(parts)-1])
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

// checkRateLimit returns true if the IP has exceeded maxPerMinute requests
// for the given endpoint in the current 1-minute window.
// Old windows are pruned on each call (best-effort).
func checkRateLimit(db *sql.DB, ipHash, endpoint string, maxPerMinute int) (bool, error) {
	loc, _ := time.LoadLocation("Europe/Madrid")
	if loc == nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	window := now.Format("2006-01-02T15:04")
	cutoff := now.Add(-10 * time.Minute).Format("2006-01-02T15:04")

	_, _ = db.Exec(`DELETE FROM rate_limits WHERE window < $1`, cutoff)

	var count int
	err := db.QueryRow(`
		INSERT INTO rate_limits (ip_hash, endpoint, window, count)
		VALUES ($1, $2, $3, 1)
		ON CONFLICT (ip_hash, endpoint, window)
		DO UPDATE SET count = rate_limits.count + 1
		RETURNING count
	`, ipHash, endpoint, window).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > maxPerMinute, nil
}
