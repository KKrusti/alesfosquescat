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
// Marca la fi d'un apagón: posa current_streak a 0 i desa el valor anterior
// a streak_before_resolve. Elimina el vot de l'IP que resol per permetre
// que pugui tornar a reportar si s'ha equivocat.
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

	// ── Eliminar el vot de l'IP per permetre re-reportar ─────────────────
	_, err = db.Exec(
		`DELETE FROM daily_votes WHERE ip_hash = $1 AND date = $2`,
		ipHash, today,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "failed to clear vote"})
		return
	}

	// ── Guardar current_streak i posar-lo a 0 ────────────────────────────
	// Si current_streak > 0 actualitzem streak_before_resolve amb el valor
	// actual. Si ja és 0 mantenim l'últim streak_before_resolve vàlid per
	// no sobreescriure'l en cas de doble clic.
	_, err = db.Exec(`
		UPDATE streak_state
		   SET streak_before_resolve = CASE WHEN current_streak > 0
		                                    THEN current_streak
		                                    ELSE streak_before_resolve
		                               END,
		       current_streak        = 0,
		       updated_at            = NOW()
		 WHERE id = 1
	`)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "failed to update streak"})
		return
	}

	w.WriteHeader(http.StatusOK)
	writeJSON(w, map[string]interface{}{
		"resolved": true,
		"date":     today,
	})
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
