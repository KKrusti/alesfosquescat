package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// InteractionEntry is one logged action returned by GET /api/interactions.
type InteractionEntry struct {
	Action string `json:"action"` // "report" | "resolve"
	At     string `json:"at"`     // "DD-MM-YYYY HH:mm" in Europe/Madrid
}

// Handler — GET /api/interactions
// Returns the last 100 report/resolve actions, most-recent first.
// Timestamps are formatted in Europe/Madrid so the frontend never has to
// deal with timezone conversion.
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

	rows, err := db.Query(
		`SELECT action, created_at FROM interaction_log ORDER BY created_at DESC LIMIT 100`,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "query failed"})
		return
	}
	defer rows.Close()

	entries := []InteractionEntry{}
	for rows.Next() {
		var action string
		var createdAt time.Time
		if err := rows.Scan(&action, &createdAt); err != nil {
			continue
		}
		entries = append(entries, InteractionEntry{
			Action: action,
			At:     createdAt.In(loc).Format("02-01-2006 15:04"),
		})
	}

	w.WriteHeader(http.StatusOK)
	writeJSON(w, entries)
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
