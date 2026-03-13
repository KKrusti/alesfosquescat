package handler

import (
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"regexp"

	_ "github.com/lib/pq"
)

// validDate matches YYYY-MM-DD dates only, preventing SQL injection via the
// date parameter. The DB layer uses parameterised queries as a second layer.
var validDate = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// Handler — GET /api/admin  →  list incidents
//
//	DELETE /api/admin?date=YYYY-MM-DD  →  delete one incident
//
// Both methods require the query param `t` to match the ADMIN_TOKEN env var.
// Comparison is done with subtle.ConstantTimeCompare to prevent timing attacks.
func Handler(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w, "GET, DELETE, OPTIONS")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// ── Token validation ──────────────────────────────────────────────
	if !validToken(r.URL.Query().Get("t")) {
		w.WriteHeader(http.StatusUnauthorized)
		writeJSON(w, map[string]string{"error": "unauthorized"})
		return
	}

	db, err := openDB()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "db connection failed"})
		return
	}
	defer db.Close()

	switch r.Method {
	case http.MethodGet:
		handleList(w, db)
	case http.MethodDelete:
		handleDelete(w, r, db)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeJSON(w, map[string]string{"error": "method not allowed"})
	}
}

// handleList returns all incidents ordered by date descending.
func handleList(w http.ResponseWriter, db *sql.DB) {
	rows, err := db.Query(
		`SELECT to_char(date, 'YYYY-MM-DD') FROM incidents ORDER BY date DESC`,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "query failed"})
		return
	}
	defer rows.Close()

	dates := []string{}
	for rows.Next() {
		var d string
		if err := rows.Scan(&d); err != nil {
			continue
		}
		dates = append(dates, d)
	}
	writeJSON(w, map[string]interface{}{"incidents": dates})
}

// handleDelete removes one incident by date and, if it matches the active
// incident_start in streak_state, clears the streak too.
func handleDelete(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	date := r.URL.Query().Get("date")
	if !validDate.MatchString(date) {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, map[string]string{"error": "invalid date format"})
		return
	}

	res, err := db.Exec(`DELETE FROM incidents WHERE date = $1`, date)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "delete failed"})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		w.WriteHeader(http.StatusNotFound)
		writeJSON(w, map[string]string{"error": "incident not found"})
		return
	}

	// If the deleted date matches the active streak start, clear it.
	_, _ = db.Exec(`
		UPDATE streak_state
		   SET incident_start = NULL,
		       updated_at     = NOW()
		 WHERE id = 1
		   AND to_char(incident_start, 'YYYY-MM-DD') = $1
	`, date)

	_, _ = db.Exec(`INSERT INTO interaction_log (action) VALUES ('admin_delete')`)

	writeJSON(w, map[string]interface{}{"deleted": true, "date": date})
}

// validToken compares the provided token against ADMIN_TOKEN using constant-time
// comparison. Returns false if the env var is empty (fail-closed).
func validToken(provided string) bool {
	secret := os.Getenv("ADMIN_TOKEN")
	if secret == "" || provided == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(provided), []byte(secret)) == 1
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
