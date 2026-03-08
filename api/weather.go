package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

const (
	weatherLat  = 41.640199
	weatherLon  = 2.199181
	cacheMaxAge = 12 * time.Hour
	rainMM      = 5.0
	rainProbPct = 60
)

// WeatherResponse is the JSON payload returned by GET /api/weather.
type WeatherResponse struct {
	Alert     bool    `json:"alert"`
	DaysUntil int     `json:"days_until"`
	MM        float64 `json:"mm"`
	Prob      int     `json:"prob"`
}

// Handler — GET /api/weather
// Returns a rain alert for the next 7 days based on Open-Meteo data cached in DB.
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

	// Try to serve from cache if it is still fresh and contains all required fields.
	var cachedRaw string
	var updatedAt time.Time
	cacheErr := db.QueryRow(
		`SELECT data::text, updated_at FROM weather_cache WHERE id = 1`,
	).Scan(&cachedRaw, &updatedAt)
	cacheValid := cacheErr == nil &&
		time.Since(updatedAt) < cacheMaxAge &&
		strings.Contains(cachedRaw, `"prob"`)
	if cacheValid {
		_, _ = w.Write([]byte(cachedRaw))
		return
	}

	// Cache is stale or missing — fetch fresh data from Open-Meteo.
	result, fetchErr := fetchWeather()
	if fetchErr != nil {
		// Fall back to stale cache rather than returning an error.
		if cacheErr == nil {
			_, _ = w.Write([]byte(cachedRaw))
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
		writeJSON(w, map[string]string{"error": "weather service unavailable"})
		return
	}

	data, _ := json.Marshal(result)
	_, _ = db.Exec(`
		INSERT INTO weather_cache (id, data, updated_at)
		VALUES (1, $1::jsonb, NOW())
		ON CONFLICT (id) DO UPDATE
		  SET data       = $1::jsonb,
		      updated_at = NOW()
	`, string(data))

	_, _ = w.Write(data)
}

// fetchWeather calls the Open-Meteo forecast API and returns a WeatherResponse.
// It scans the next 7 days and flags the first day where precipitation_sum > 5 mm
// AND precipitation_probability_max > 60 %.
func fetchWeather() (*WeatherResponse, error) {
	url := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast"+
			"?latitude=%g&longitude=%g"+
			"&daily=precipitation_sum,precipitation_probability_max"+
			"&timezone=Europe%%2FMadrid"+
			"&forecast_days=7",
		weatherLat, weatherLon,
	)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload struct {
		Daily struct {
			Time                 []string  `json:"time"`
			PrecipitationSum     []float64 `json:"precipitation_sum"`
			PrecipitationProbMax []int     `json:"precipitation_probability_max"`
		} `json:"daily"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	result := parseWeatherAlert(
		payload.Daily.Time,
		payload.Daily.PrecipitationSum,
		payload.Daily.PrecipitationProbMax,
	)
	return &result, nil
}

// parseWeatherAlert scans daily forecast slices and returns the first day that
// exceeds both the rain threshold and the probability threshold.
// Pure function — no I/O, safe to unit-test directly.
func parseWeatherAlert(times []string, sums []float64, probs []int) WeatherResponse {
	for i := range times {
		if i >= len(sums) || i >= len(probs) {
			break
		}
		if sums[i] > rainMM && probs[i] > rainProbPct {
			return WeatherResponse{Alert: true, DaysUntil: i, MM: sums[i], Prob: probs[i]}
		}
	}
	return WeatherResponse{}
}

// ── helpers (duplicated per vercel-go per-file isolation requirement) ─────────

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
