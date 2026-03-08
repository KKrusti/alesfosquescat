-- alesfosquescat — Neon Postgres schema
-- Run once on your Neon project before deploying

-- tabla incidents: un registro por día con apagón confirmado
CREATE TABLE IF NOT EXISTS incidents (
  id         SERIAL PRIMARY KEY,
  date       DATE UNIQUE NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- tabla daily_votes: registra reports i resolucions per IP i dia.
--   action: 'report' (apagón confirmat) o 'resolve' (apagón resolt).
--   incident_start_saved: només s'omple en action='resolve'; guarda el valor
--     d'incident_start en el moment de resoldre, perquè si l'usuari torna a
--     reportar el mateix dia es pugui restaurar la data original sense tocar
--     streak_state.
--   PK (ip_hash, date, action): permet un report i un resolve per IP i dia.
CREATE TABLE IF NOT EXISTS daily_votes (
  ip_hash              TEXT NOT NULL,
  date                 DATE NOT NULL,
  action               TEXT NOT NULL DEFAULT 'report',
  incident_start_saved DATE,
  created_at           TIMESTAMPTZ DEFAULT NOW(),
  PRIMARY KEY (ip_hash, date, action)
);

-- tabla streak_state: una sola fila (id=1).
--   incident_start: data d'inici del apagón actiu (NULL = inactiu).
--     current_streak es calcula en temps real com (avui - incident_start + 1).
--   longest_streak: màxim historial, actualitzat en resoldre.
CREATE TABLE IF NOT EXISTS streak_state (
  id             INT PRIMARY KEY DEFAULT 1,
  longest_streak INT NOT NULL DEFAULT 0,
  incident_start DATE,
  updated_at     TIMESTAMPTZ DEFAULT NOW(),
  CONSTRAINT single_row CHECK (id = 1)
);

-- Migracions idempotents
-- daily_votes: nova PK (ip_hash, date, action) i nova columna incident_start_saved
ALTER TABLE daily_votes ADD COLUMN IF NOT EXISTS action               TEXT NOT NULL DEFAULT 'report';
ALTER TABLE daily_votes ADD COLUMN IF NOT EXISTS incident_start_saved DATE;
-- streak_state: columnes legacy (ja no s'usen, es mantenen per compatibilitat)
ALTER TABLE streak_state ADD COLUMN IF NOT EXISTS incident_start      DATE;
ALTER TABLE streak_state ADD COLUMN IF NOT EXISTS longest_streak      INT NOT NULL DEFAULT 0;

-- Seed the single row (noop if already exists)
INSERT INTO streak_state (id) VALUES (1) ON CONFLICT DO NOTHING;

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

-- Índexs per rendiment
CREATE INDEX IF NOT EXISTS idx_incidents_date      ON incidents(date);
CREATE INDEX IF NOT EXISTS idx_daily_votes_date    ON daily_votes(date);
CREATE INDEX IF NOT EXISTS idx_daily_votes_hash    ON daily_votes(ip_hash);
CREATE INDEX IF NOT EXISTS idx_rate_limits_window       ON rate_limits(window);
CREATE INDEX IF NOT EXISTS idx_interaction_log_created  ON interaction_log(created_at DESC);
