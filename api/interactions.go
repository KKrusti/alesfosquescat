package handler

import (
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

// InteractionEntry is one logged action returned by GET /api/interactions.
type InteractionEntry struct {
	ID     int    `json:"id"`
	Action string `json:"action"` // "report" | "resolve" | "admin_delete" ...
	At     string `json:"at"`     // "DD-MM-YYYY HH:mm" in Europe/Madrid
}

// Handler — GET /api/interactions
//
//	Returns the last 100 actions, most-recent first. Public endpoint.
//
// DELETE /api/interactions?t=TOKEN&id=ID
//
//	Deletes one entry from interaction_log. Requires admin token.
func Handler(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w, "GET, DELETE, OPTIONS")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
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
		handleGetInteractions(w, db)
	case http.MethodDelete:
		handleDeleteInteraction(w, r, db)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeJSON(w, map[string]string{"error": "method not allowed"})
	}
}

func handleGetInteractions(w http.ResponseWriter, db *sql.DB) {
	loc, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		loc = time.UTC
	}

	rows, err := db.Query(
		`SELECT id, action, created_at FROM interaction_log ORDER BY created_at DESC LIMIT 100`,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "query failed"})
		return
	}
	defer rows.Close()

	entries := []InteractionEntry{}
	for rows.Next() {
		var id int
		var action string
		var createdAt time.Time
		if err := rows.Scan(&id, &action, &createdAt); err != nil {
			continue
		}
		entries = append(entries, InteractionEntry{
			ID:     id,
			Action: action,
			At:     createdAt.In(loc).Format("02-01-2006 15:04"),
		})
	}

	w.WriteHeader(http.StatusOK)
	writeJSON(w, entries)
}

func handleDeleteInteraction(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Token validation — same mechanism as admin.go
	secret := os.Getenv("ADMIN_TOKEN")
	provided := r.URL.Query().Get("t")
	if secret == "" || provided == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(secret)) != 1 {
		w.WriteHeader(http.StatusUnauthorized)
		writeJSON(w, map[string]string{"error": "unauthorized"})
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, map[string]string{"error": "invalid id"})
		return
	}

	res, err := db.Exec(`DELETE FROM interaction_log WHERE id = $1`, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "delete failed"})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		w.WriteHeader(http.StatusNotFound)
		writeJSON(w, map[string]string{"error": "entry not found"})
		return
	}

	writeJSON(w, map[string]bool{"deleted": true})
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
