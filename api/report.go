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

	// ── Duplicate-vote check ──────────────────────────────────────────
	var alreadyVoted bool
	err = db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM daily_votes WHERE ip_hash=$1 AND date=$2)`,
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
	result, err := db.Exec(
		`INSERT INTO incidents (date) VALUES ($1) ON CONFLICT (date) DO NOTHING`,
		today,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": "failed to record incident"})
		return
	}

	// ── Actualitzar streak counter ───────────────────────────────────────
	// Llegim l'estat actual per decidir com actualitzar-lo.
	rowsInserted, _ := result.RowsAffected()

	var currentStreak, streakBeforeResolve int
	_ = db.QueryRow(
		`SELECT current_streak, streak_before_resolve FROM streak_state WHERE id = 1`,
	).Scan(&currentStreak, &streakBeforeResolve)

	// Casos que requereixen actualització:
	//  A. currentStreak == 0: s'havia marcat com a resolt (per error o no).
	//     Restaurem la racha des d'on estava (streak_before_resolve).
	//     Això cobreix tant el primer report del dia (incident nou, RowsAffected==1)
	//     com el re-report del mateix dia després de resoldre (RowsAffected==0).
	//  B. currentStreak > 0 i incident NOU (RowsAffected==1): incrementem.
	//     Si RowsAffected==0 amb streak actiu, ja estava comptat → no fem res.
	if currentStreak == 0 {
		newStreak := streakBeforeResolve
		if newStreak == 0 {
			newStreak = 1
		}
		if _, err = db.Exec(`
			INSERT INTO streak_state (id, current_streak, longest_streak, streak_before_resolve)
			VALUES (1, $1, $1, 0)
			ON CONFLICT (id) DO UPDATE
			  SET current_streak        = $1,
			      longest_streak        = GREATEST(streak_state.longest_streak, $1),
			      streak_before_resolve = 0,
			      updated_at            = NOW()
		`, newStreak); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeJSON(w, map[string]string{"error": "failed to update streak"})
			return
		}
	} else if rowsInserted == 1 {
		if _, err = db.Exec(`
			UPDATE streak_state
			   SET current_streak = current_streak + 1,
			       longest_streak = GREATEST(longest_streak, current_streak + 1),
			       updated_at     = NOW()
			 WHERE id = 1
		`); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeJSON(w, map[string]string{"error": "failed to update streak"})
			return
		}
	}

	// ── Record this vote ──────────────────────────────────────────────
	if _, err = db.Exec(
		`INSERT INTO daily_votes (ip_hash, date) VALUES ($1, $2)`,
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
