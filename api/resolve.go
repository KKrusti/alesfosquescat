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

// Handler — POST /api/resolve
// Marca la fi d'un apagón:
//   - Llegeix incident_start actual de streak_state.
//   - Insereix una fila a daily_votes amb action='resolve' i incident_start_saved=incident_start,
//     de manera que si l'usuari torna a reportar el mateix dia, report.go pot restaurar
//     la data original sense cap lògica de backup a streak_state.
//   - Actualitza longest_streak si el streak actual el supera.
//   - Posa incident_start = NULL per indicar que no hi ha apagón actiu.
//
// L'incident queda a la taula incidents (no s'esborra) per preservar l'historial.
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

	db, err := openDB()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "db connection failed"})
		return
	}
	defer db.Close()

	today := todayInMadrid()
	ipHash := sha256hex(clientIP(r))

	// ── Rate limit: max 5 peticions per IP per minut ─────────────────
	if exceeded, err := checkRateLimit(db, ipHash, "resolve", 5); err == nil && exceeded {
		w.WriteHeader(http.StatusTooManyRequests)
		writeJSON(w, map[string]string{"error": "massa peticions, espera un minut"})
		return
	}

	// ── Ja ha resolt avui? Evita actualitzar streak_state dues vegades ───
	var alreadyResolved bool
	if err = db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM daily_votes WHERE ip_hash=$1 AND date=$2 AND action='resolve')`,
		ipHash, today,
	).Scan(&alreadyResolved); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "resolve check failed"})
		return
	}
	if alreadyResolved {
		// L'estat ja és correcte (streak NULL), retornem 200 perquè el frontend
		// no mostri error. El streak_state no es torna a tocar.
		w.WriteHeader(http.StatusOK)
		writeJSON(w, map[string]interface{}{"resolved": true, "date": today})
		return
	}

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
	// ON CONFLICT DO NOTHING: si l'usuari fa doble clic a resoldre, ignorem.
	// incident_start_saved és NULL si el streak ja estava inactiu.
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
		// Streak ja era inactiu (doble clic o estat inconsistent).
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
	writeJSON(w, map[string]interface{}{
		"resolved": true,
		"date":     today,
	})
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
