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
      // stats just won't show — non-critical
    } finally {
      setStatsLoading(false)
    }
  }, [])

  useEffect(() => { fetchStats() }, [fetchStats])

  const totalThisYear = stats?.total_this_year ?? null

  return (
    <div className="min-h-screen bg-[#050510] relative overflow-x-hidden">

      {/* ── Star field ── */}
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

      {/* ── Sky gradient overlay ── */}
      <div
        className="fixed inset-0 pointer-events-none"
        style={{
          background:
            'radial-gradient(ellipse 70% 55% at 50% 0%, rgba(15,15,60,0.45) 0%, transparent 100%)',
        }}
      />

      {/* ── Page content ── */}
      <main className="relative z-10 flex flex-col items-center px-4 pt-10 pb-16 min-h-screen">

        {/* Header */}
        <header className="text-center mb-8 animate-fade-in">
          <h1
            className="text-5xl sm:text-6xl md:text-7xl font-black tracking-[0.06em] text-signal-500 uppercase"
            style={{
              fontFamily: 'Anton, sans-serif',
              textShadow: '0 0 40px rgba(251,191,36,0.35), 0 0 80px rgba(251,191,36,0.12)',
            }}
          >
            ALESFOSQUESCAT
          </h1>
          <p className="mt-2 text-white/30 font-mono text-xs uppercase tracking-[0.35em]">
            el comptador de dies sense enllumenat public de Santa Eulàlia de Ronçana
          </p>
        </header>

        {/* Big counter */}
        <div className="text-center mb-6 animate-slide-up" style={{ animationDelay: '100ms', animationFillMode: 'both' }}>
          {statsLoading ? (
            <div className="h-24 w-40 bg-white/5 rounded animate-pulse mx-auto" />
          ) : (
            <div className="animate-count-in">
              <div
                className="text-8xl sm:text-9xl font-black leading-none text-signal-500"
                style={{
                  fontFamily: 'Anton, sans-serif',
                  textShadow: '0 0 50px rgba(252,211,77,0.5)',
                }}
              >
                {totalThisYear ?? '—'}
              </div>
              <p className="mt-2 text-white/45 font-mono text-sm uppercase tracking-[0.3em]">
                dies sense llum aquest any
              </p>
            </div>
          )}
        </div>

        {/* Bat Signal */}
        <div
          className="my-2 animate-fade-in"
          style={{ animationDelay: '200ms', animationFillMode: 'both' }}
        >
          <BatSignal onSuccess={fetchStats} />
        </div>

        {/* Divider */}
        <div className="w-full max-w-2xl my-6 border-t border-white/8" />

        {/* Stats section */}
        <section
          className="w-full max-w-2xl animate-slide-up"
          style={{ animationDelay: '350ms', animationFillMode: 'both' }}
        >
          <h2 className="text-center text-xs font-mono text-white/25 uppercase tracking-[0.45em] mb-4">
            estadístiques
          </h2>
          <Stats stats={stats} loading={statsLoading} />
        </section>

        {/* Footer */}
        <footer className="mt-auto pt-10 text-center">
          <p className="text-white/15 font-mono text-xs tracking-widest">
            alesfosquescat — {new Date().getFullYear()}
          </p>
        </footer>
      </main>
    </div>
  )
}
