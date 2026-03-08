package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// Handler — POST /api/resolve
// Marca la fi d'un apagón (acció global, no lligada a cap IP):
//   - Llegeix incident_start actual de streak_state.
//   - Insereix una fila a daily_votes amb ip_hash='community' i action='resolve'
//     guardant incident_start_saved, perquè si algun veí torna a reportar el
//     mateix dia, report.go pugui restaurar la data original.
//   - Actualitza longest_streak si el streak actual el supera.
//   - Posa incident_start = NULL per indicar que no hi ha apagón actiu.
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
	// Usem ip_hash='community' perquè resolve és una acció global sense IP.
	// ON CONFLICT DO NOTHING: evita duplicats si es clica dues vegades avui.
	// incident_start_saved permet a report.go restaurar la data si algun veí
	// torna a reportar el mateix dia després d'una resolució.
	var savedStart interface{}
	if incidentStart.Valid && incidentStart.String != "" {
		savedStart = incidentStart.String
	}
	_, err = db.Exec(`
		INSERT INTO daily_votes (ip_hash, date, action, incident_start_saved)
		VALUES ('community', $1, 'resolve', $2)
		ON CONFLICT (ip_hash, date, action) DO NOTHING
	`, today, savedStart)
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

	// Best-effort log — do not block the response if this fails
	_, _ = db.Exec(`INSERT INTO interaction_log (action) VALUES ('resolve')`)

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
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, err
	}
	// Serverless: one connection per invocation, release immediately after use
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

func todayInMadrid() string {
	loc, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		loc = time.UTC
	}
	return time.Now().In(loc).Format("2006-01-02")
}
