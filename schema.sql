-- alesfosquescat — Neon Postgres schema
-- Run once on your Neon project before deploying

-- tabla incidents: un registro por apagón con fecha de inicio y fin.
--   date:     fecha de inicio del apagón (start_date).
--   end_date: fecha en que se resolvió (NULL = apagón activo).
--   Cuando end_date = NULL hay un apagón en curso.
--   Si start_date == end_date fue un falso reporte resuelto el mismo día.
CREATE TABLE IF NOT EXISTS incidents (
  id         SERIAL PRIMARY KEY,
  date       DATE UNIQUE NOT NULL,
  end_date   DATE,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- tabla rate_limits: límit de peticions per IP i minut per endpoint.
--   window: "YYYY-MM-DDTHH:MM" (finestra d'1 minut en Europe/Madrid).
--   Les files antigues s'esborren automàticament des dels handlers Go.
CREATE TABLE IF NOT EXISTS rate_limits (
  ip_hash  TEXT NOT NULL,
  endpoint TEXT NOT NULL,
  window   TEXT NOT NULL,
  count    INT  NOT NULL DEFAULT 1,
  PRIMARY KEY (ip_hash, endpoint, window)
);

-- tabla interaction_log: registre de cada acció de reportar/resoldre.
--   Permet mostrar un historial d'activitat amb hora del servidor (Europe/Madrid).
--   No s'associa a cap IP; és un log públic i anònim.
CREATE TABLE IF NOT EXISTS interaction_log (
  id         SERIAL PRIMARY KEY,
  action     TEXT        NOT NULL, -- 'report' | 'resolve'
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- tabla weather_cache: una sola fila (id=1) amb el resultat d'Open-Meteo.
--   data: JSON { alert, days_until, mm } actualitzat cada 12h pel cron de Vercel.
CREATE TABLE IF NOT EXISTS weather_cache (
  id         INT PRIMARY KEY DEFAULT 1,
  data       JSONB NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT single_row_weather CHECK (id = 1)
);

-- ── Migrations ────────────────────────────────────────────────────────────────

-- Add end_date to incidents (idempotent)
ALTER TABLE incidents ADD COLUMN IF NOT EXISTS end_date DATE;

-- Migrate historical data: mark all existing closed incidents with end_date = date
-- (each old row was a 1-day outage start; without real end dates we use start = end).
-- Leave active outage (matching streak_state.incident_start) with end_date = NULL.
-- If streak_state doesn't exist yet, all rows are treated as closed.
UPDATE incidents
   SET end_date = date
 WHERE end_date IS NULL
   AND date NOT IN (
       SELECT incident_start
         FROM streak_state
        WHERE incident_start IS NOT NULL
   );

-- Merge legacy 1-day rows into multi-day periods (gaps-and-islands).
-- Rows where end_date = date are legacy: one row per day instead of one per period.
-- This groups consecutive dates into a single row (keeps the lowest id, deletes the rest).
-- Idempotent: rows already merged (end_date != date) are untouched.
DO $$
BEGIN
  -- Step 1: update the first row of each island with the correct end_date.
  WITH numbered AS (
    SELECT id, date,
           date - (ROW_NUMBER() OVER (ORDER BY date) || ' days')::INTERVAL AS grp
      FROM incidents
     WHERE end_date = date
  ),
  islands AS (
    SELECT MIN(id) AS keep_id, MAX(date) AS new_end
      FROM numbered
     GROUP BY grp
     HAVING COUNT(*) > 1
  )
  UPDATE incidents i
     SET end_date = s.new_end
    FROM islands s
   WHERE i.id = s.keep_id;

  -- Step 2: delete the redundant rows (all but the first of each island).
  WITH numbered AS (
    SELECT id, date,
           date - (ROW_NUMBER() OVER (ORDER BY date) || ' days')::INTERVAL AS grp
      FROM incidents
     WHERE end_date = date
  ),
  islands AS (
    SELECT MIN(id) AS keep_id, array_agg(id ORDER BY date) AS all_ids
      FROM numbered
     GROUP BY grp
     HAVING COUNT(*) > 1
  )
  DELETE FROM incidents
   WHERE id IN (
     SELECT unnest(all_ids[2:]) FROM islands
   );
END $$;

-- ── Indexes ───────────────────────────────────────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_incidents_date           ON incidents(date);
CREATE INDEX IF NOT EXISTS idx_incidents_end_date       ON incidents(end_date);
CREATE INDEX IF NOT EXISTS idx_rate_limits_window       ON rate_limits(window);
CREATE INDEX IF NOT EXISTS idx_interaction_log_created  ON interaction_log(created_at DESC);
