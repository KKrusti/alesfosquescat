import { useState, useEffect, useCallback } from 'react'
import { BatSignal } from './components/BatSignal'
import { Stats } from './components/Stats'
import type { StatsResponse } from './types'

export default function App() {
  const [stats, setStats] = useState<StatsResponse | null>(null)
  const [statsLoading, setStatsLoading] = useState(true)

  const fetchStats = useCallback(async () => {
    try {
      const res = await fetch('/api/stats')
      if (res.ok) setStats(await res.json())
    } catch { /* non-critical */ }
    finally { setStatsLoading(false) }
  }, [])

  useEffect(() => { fetchStats() }, [fetchStats])

  const total = stats?.total_this_year ?? null

  return (
    <div className="min-h-screen bg-[#0d0d1c]">

      {/* ── Header ─────────────────────────────────────────────────── */}
      <header className="sticky top-0 z-20 bg-[#0d0d1c]/95 backdrop-blur-sm border-b border-white/8 pt-safe">
        <div className="max-w-lg mx-auto px-4 py-3 flex items-center gap-3">

          {/* Amber accent stripe — the visual signature of the layout */}
          <div className="w-[3px] self-stretch bg-amber-400 rounded-full shrink-0" />

          <div className="flex-1 min-w-0">
            <p className="text-amber-400 font-bold text-[15px] leading-tight tracking-tight">
              alesfosquestcat
            </p>
            <p className="text-white/35 text-[11px] leading-tight mt-0.5">
              Santa Eulàlia de Ronçana
            </p>
          </div>

          {/* Live indicator */}
          <div className="flex items-center gap-1.5 px-2.5 py-1 rounded border border-amber-500/25 bg-amber-500/8 shrink-0">
            <span className="w-1.5 h-1.5 rounded-full bg-amber-400 animate-pulse" />
            <span className="text-[10px] text-amber-400/80 uppercase tracking-widest font-medium">
              en directe
            </span>
          </div>
        </div>
      </header>

      {/* ── Content ─────────────────────────────────────────────────── */}
      <main className="max-w-lg mx-auto px-4 pb-safe">

        {/* ── Estat actual ── */}
        <section className="pt-5 pb-4">
          <p className="section-label">Estat actual</p>

          <div className="status-card">
            {/* Left colored bar */}
            <div className="w-[3px] self-stretch bg-amber-400 rounded-full shrink-0 my-0.5" />

            <div className="flex-1 min-w-0">
              {statsLoading ? (
                <div className="space-y-2">
                  <div className="h-14 w-24 bg-white/6 rounded animate-pulse" />
                  <div className="h-3 w-36 bg-white/4 rounded animate-pulse" />
                </div>
              ) : (
                <>
                  <div
                    className="text-6xl sm:text-7xl font-black leading-none text-amber-400"
                    style={{ fontFamily: 'Anton, sans-serif' }}
                  >
                    {total ?? '—'}
                  </div>
                  <p className="text-white/40 text-xs mt-1.5 tracking-wide">
                    dies sense llum · {new Date().getFullYear()}
                  </p>
                </>
              )}
            </div>

            {/* Status badge */}
            <div className="shrink-0 self-start mt-0.5">
              <div className="flex items-center gap-1.5 px-2 py-1 rounded border border-amber-500/20 bg-amber-500/8">
                <span className="w-1.5 h-1.5 rounded-full bg-amber-400 animate-pulse shrink-0" />
                <span className="text-[10px] text-amber-400/70 uppercase tracking-wider whitespace-nowrap">
                  Apagons
                </span>
              </div>
            </div>
          </div>
        </section>

        <div className="section-divider" />

        {/* ── Reportar ── */}
        <section className="py-5">
          <p className="section-label">Reportar incidència</p>
          <div className="flex flex-col items-center">
            <BatSignal onSuccess={fetchStats} />
          </div>
        </section>

        <div className="section-divider" />

        {/* ── Estadístiques ── */}
        <section className="py-5">
          <p className="section-label">Dades anuals</p>
          <Stats stats={stats} loading={statsLoading} />
        </section>

        {/* Footer */}
        <footer className="py-5 border-t border-white/6 text-center">
          <p className="text-white/15 text-[10px] tracking-widest">
            alesfosquestcat · Santa Eulàlia de Ronçana · {new Date().getFullYear()}
          </p>
        </footer>
      </main>
    </div>
  )
}
