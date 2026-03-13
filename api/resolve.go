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
// Marks the end of the current outage.
//
//   - Finds the open incident (end_date IS NULL).
//   - If started today: DELETE it (false report).
//   - Otherwise: set end_date = today (real outage resolved).
//
// If no active outage is found, returns already_resolved.
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

	// ── Find the active (open) incident ───────────────────────────────
	var incidentID int
	var startDate string
	err = db.QueryRow(`
		SELECT id, to_char(date, 'YYYY-MM-DD')
		  FROM incidents
		 WHERE end_date IS NULL
		 LIMIT 1
	`).Scan(&incidentID, &startDate)

	if err == sql.ErrNoRows {
		// No active outage — nothing to resolve
		w.WriteHeader(http.StatusOK)
		writeJSON(w, map[string]interface{}{"resolved": false, "already_resolved": true})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "failed to read active incident"})
		return
	}

	// ── Close or delete the incident ──────────────────────────────────
	if startDate == today {
		// Outage started today: this is a false report — remove it entirely
		if _, err = db.Exec(`DELETE FROM incidents WHERE id = $1`, incidentID); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeJSON(w, map[string]string{"error": "failed to delete incident"})
			return
		}
	} else {
		// Real multi-day outage: set end_date to mark it as resolved
		if _, err = db.Exec(
			`UPDATE incidents SET end_date = $1 WHERE id = $2`,
			today, incidentID,
		); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeJSON(w, map[string]string{"error": "failed to close incident"})
			return
		}
	}

	// Best-effort log — do not block the response if this fails
	_, _ = db.Exec(`INSERT INTO interaction_log (action) VALUES ('resolve')`)

	w.WriteHeader(http.StatusOK)
	writeJSON(w, map[string]interface{}{
		"resolved": true,
		"date":     today,
	})
}

// ── helpers (duplicated per vercel-go isolation requirement) ──────────────────

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

func todayInMadrid() string {
	loc, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		loc = time.UTC
	}
	return time.Now().In(loc).Format("2006-01-02")
}
