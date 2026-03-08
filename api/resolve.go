package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

// Handler — POST /api/resolve
// Marca la incidència com a resolta: posa current_streak a 0.
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

	_, err = db.Exec(`
		UPDATE streak_state
		SET current_streak = 0, updated_at = NOW()
		WHERE id = 1
	`)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "failed to resolve"})
		return
	}

	w.WriteHeader(http.StatusOK)
	writeJSON(w, map[string]interface{}{"resolved": true})
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
