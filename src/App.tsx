import { useState, useEffect, useCallback } from 'react'
import { BatSignal } from './components/BatSignal'
import { Stats } from './components/Stats'
import type { StatsResponse } from './types'

// ── Deterministic star field (seeded PRNG — stable across renders) ──────────
function mulberry32(seed: number) {
  return () => {
    seed = (seed + 0x6d2b79f5) | 0
    let t = Math.imul(seed ^ (seed >>> 15), 1 | seed)
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296
  }
}

interface Star { x: number; y: number; r: number; delay: number; opacity: number }

const rng = mulberry32(0xdeadbeef)
const STARS: Star[] = Array.from({ length: 55 }, () => ({
  x: rng() * 100,
  y: rng() * 100,
  r: 0.15 + rng() * 0.45,
  delay: rng() * 5,
  opacity: 0.3 + rng() * 0.7,
}))

export default function App() {
  const [stats, setStats] = useState<StatsResponse | null>(null)
  const [statsLoading, setStatsLoading] = useState(true)

  const fetchStats = useCallback(async () => {
    try {
      const res = await fetch('/api/stats')
      if (res.ok) setStats(await res.json())
    } catch {
      // non-critical — stats just won't show
    } finally {
      setStatsLoading(false)
    }
  }, [])

  useEffect(() => { fetchStats() }, [fetchStats])

  const totalThisYear = stats?.total_this_year ?? null

  return (
    <div className="bg-[#050510] relative overflow-x-hidden">

      {/* ── Star field (fixed, full viewport) ── */}
      <svg
        className="fixed inset-0 w-full h-full pointer-events-none"
        xmlns="http://www.w3.org/2000/svg"
        preserveAspectRatio="xMidYMid slice"
        viewBox="0 0 100 100"
        aria-hidden="true"
      >
        {STARS.map((s, i) => (
          <circle
            key={i}
            cx={s.x}
            cy={s.y}
            r={s.r}
            fill="white"
            opacity={s.opacity}
            style={{
              animation: `twinkle ${2.5 + s.delay}s ease-in-out infinite`,
              animationDelay: `${s.delay}s`,
            }}
          />
        ))}
      </svg>

      {/* ── Sky gradient ── */}
      <div
        className="fixed inset-0 pointer-events-none"
        style={{
          background:
            'radial-gradient(ellipse 80% 50% at 50% 0%, rgba(15,15,60,0.5) 0%, transparent 100%)',
        }}
      />

      {/* ── Page content ── */}
      <main className="relative z-10 flex flex-col items-center px-5 pt-safe pb-safe">

        {/* ── HERO: title + counter + signal ── */}
        <section className="w-full flex flex-col items-center gap-3 pt-8 pb-6">

          {/* Title */}
          <h1
            className="text-3xl sm:text-5xl md:text-6xl font-black tracking-[0.05em] text-signal-500 uppercase text-center leading-none animate-fade-in"
            style={{
              fontFamily: 'Anton, sans-serif',
              textShadow: '0 0 30px rgba(251,191,36,0.35)',
            }}
          >
            ALES FOSQUES CAT
          </h1>
          <p className="text-white/30 font-mono text-[10px] sm:text-xs uppercase tracking-[0.35em] animate-fade-in">
            el comptador d&apos;apagons del poble
          </p>

          {/* Main counter */}
          <div
            className="text-center mt-1 animate-slide-up"
            style={{ animationDelay: '80ms', animationFillMode: 'both' }}
          >
            {statsLoading ? (
              <div className="h-20 w-32 bg-white/5 rounded-lg animate-pulse mx-auto" />
            ) : (
              <div className="animate-count-in">
                <div
                  className="text-7xl sm:text-8xl font-black leading-none text-signal-500"
                  style={{
                    fontFamily: 'Anton, sans-serif',
                    textShadow: '0 0 40px rgba(252,211,77,0.45)',
                  }}
                >
                  {totalThisYear ?? '—'}
                </div>
                <p className="mt-1.5 text-white/40 font-mono text-xs uppercase tracking-[0.25em]">
                  dies sense llum aquest any
                </p>
              </div>
            )}
          </div>

          {/* Bat Signal — hero element */}
          <div
            className="animate-fade-in"
            style={{ animationDelay: '160ms', animationFillMode: 'both' }}
          >
            <BatSignal onSuccess={fetchStats} />
          </div>
        </section>

        {/* ── STATS ── */}
        <section
          className="w-full max-w-lg animate-slide-up border-t border-white/8 pt-6 pb-8"
          style={{ animationDelay: '300ms', animationFillMode: 'both' }}
        >
          <h2 className="text-center text-[10px] font-mono text-white/25 uppercase tracking-[0.45em] mb-4">
            estadístiques
          </h2>
          <Stats stats={stats} loading={statsLoading} />
        </section>

        <footer className="pb-4 text-center">
          <p className="text-white/15 font-mono text-[10px] tracking-widest">
            ales fosques cat — {new Date().getFullYear()}
          </p>
        </footer>
      </main>
    </div>
  )
}
