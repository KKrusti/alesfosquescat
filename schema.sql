-- alesfosquescat — Neon Postgres schema
-- Run once on your Neon project before deploying

-- tabla incidents: un registro por día con apagón confirmado
CREATE TABLE IF NOT EXISTS incidents (
  id         SERIAL PRIMARY KEY,
  date       DATE UNIQUE NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- tabla daily_votes: evitar duplicados por usuario/día
CREATE TABLE IF NOT EXISTS daily_votes (
  ip_hash    TEXT NOT NULL,
  date       DATE NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  PRIMARY KEY (ip_hash, date)
);

-- tabla streak_state: comptador persistent de ratxes (una sola fila, id=1)
CREATE TABLE IF NOT EXISTS streak_state (
  id                    INT PRIMARY KEY DEFAULT 1,
  current_streak        INT NOT NULL DEFAULT 0,
  longest_streak        INT NOT NULL DEFAULT 0,
  streak_before_resolve INT NOT NULL DEFAULT 0,
  updated_at            TIMESTAMPTZ DEFAULT NOW(),
  CONSTRAINT single_row CHECK (id = 1)
);

-- Migració: afegir columna streak_before_resolve si no existeix (idempotent)
ALTER TABLE streak_state ADD COLUMN IF NOT EXISTS streak_before_resolve INT NOT NULL DEFAULT 0;

-- Seed the single row (noop if already exists)
INSERT INTO streak_state (id) VALUES (1) ON CONFLICT DO NOTHING;

-- Índexs per rendiment
CREATE INDEX IF NOT EXISTS idx_incidents_date    ON incidents(date);
CREATE INDEX IF NOT EXISTS idx_daily_votes_date  ON daily_votes(date);
CREATE INDEX IF NOT EXISTS idx_daily_votes_hash  ON daily_votes(ip_hash);
