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
// Registra un apagón per avui. Màxim 1 cop per dia per IP (hash SHA-256).
//
// Si el streak estava inactiu (incident_start IS NULL):
//   - Comprova si hi ha una fila action='resolve' per aquesta IP avui:
//     si n'hi ha, restaura incident_start_saved (l'usuari havia resolt per error).
//     si no n'hi ha, activa amb incident_start = avui.
//
// Si el streak ja estava actiu (incident_start NOT NULL):
//   - No toca streak_state; el streak creix automàticament cada dia.
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

	// ── Rate limit: max 10 peticions per IP per minut ────────────────
	if exceeded, err := checkRateLimit(db, ipHash, "report", 10); err == nil && exceeded {
		w.WriteHeader(http.StatusTooManyRequests)
		writeJSON(w, map[string]string{"error": "massa peticions, espera un minut"})
		return
	}

	// ── Duplicate-vote check (action='report') ────────────────────────
	var alreadyVoted bool
	err = db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM daily_votes WHERE ip_hash=$1 AND date=$2 AND action='report')`,
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

	// ── Upsert incident for today ─────────────────────────────────────
	if _, err = db.Exec(
		`INSERT INTO incidents (date) VALUES ($1) ON CONFLICT (date) DO NOTHING`,
		today,
	); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "failed to record incident"})
		return
	}

	// ── Activar streak si estava inactiu ─────────────────────────────
	var incidentStart sql.NullString
	_ = db.QueryRow(
		`SELECT to_char(incident_start, 'YYYY-MM-DD') FROM streak_state WHERE id = 1`,
	).Scan(&incidentStart)

	if !incidentStart.Valid || incidentStart.String == "" {
		// Streak inactiu → determinar la data d'inici.
		// Si hi ha un resolve de l'IP d'avui, restaurem incident_start_saved.
		// Això cobreix el cas "resolt per error → torna a reportar el mateix dia".
		var resolveStart sql.NullString
		_ = db.QueryRow(`
			SELECT to_char(incident_start_saved, 'YYYY-MM-DD')
			  FROM daily_votes
			 WHERE ip_hash = $1 AND date = $2 AND action = 'resolve'
			 LIMIT 1
		`, ipHash, today).Scan(&resolveStart)

		var activateFrom string
		if resolveStart.Valid && resolveStart.String != "" {
			activateFrom = resolveStart.String
		} else {
			activateFrom = today
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
	}
	// Si incident_start ja tenia valor, el streak ja estava actiu → no fer res.

	// ── Record this vote ──────────────────────────────────────────────
	if _, err = db.Exec(
		`INSERT INTO daily_votes (ip_hash, date, action) VALUES ($1, $2, 'report')`,
		ipHash, today,
	); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "failed to record vote"})
		return
	}

	w.WriteHeader(http.StatusOK)
	writeJSON(w, map[string]interface{}{
		"success": true,
		"date":    today,
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
