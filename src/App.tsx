import { useState, useEffect, useCallback } from 'react'
import { BatSignal } from './components/BatSignal'
import { Stats } from './components/Stats'
import type { StatsResponse } from './types'

// Schedules a one-shot callback at the next local midnight
function onNextMidnight(cb: () => void): () => void {
  const now = new Date()
  const midnight = new Date(now)
  midnight.setDate(midnight.getDate() + 1)
  midnight.setHours(0, 0, 0, 0)
  const tid = setTimeout(cb, midnight.getTime() - now.getTime())
  return () => clearTimeout(tid)
}

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

  // Auto-refresh at midnight so "dies des de l'últim incident" increments by itself
  useEffect(() => {
    function scheduleNext() {
      return onNextMidnight(() => { fetchStats(); scheduleNext() })
    }
    return scheduleNext()
  }, [fetchStats])

  // ── Primary metric logic ────────────────────────────────────────
  // When there IS an active streak: show "Ratxa actual sense llum"
  // When everything is fine:        show "Dies des de l'últim incident"
  const hasActiveStreak = (stats?.current_incident_streak ?? 0) > 0

  const primaryValue: number = stats
    ? (hasActiveStreak ? stats.current_incident_streak : stats.days_since_last_incident)
    : 0

  const longestStreak: number = stats?.longest_incident_streak ?? 0

  return (
    <div className="min-h-screen bg-[#0d0d1c]">

      {/* ── Header ─────────────────────────────────────────────────── */}
      <header className="sticky top-0 z-20 bg-[#0d0d1c]/95 backdrop-blur-sm border-b border-white/8 pt-safe">
        <div className="max-w-lg mx-auto px-4 py-3 flex items-center gap-3">
          <div className="w-[3px] self-stretch bg-amber-400 rounded-full shrink-0" />
          <div className="flex-1 min-w-0">
            <p className="text-amber-400 font-bold text-[15px] leading-tight tracking-tight">
              alesfosquescat
            </p>
            <p className="text-white/35 text-[11px] leading-tight mt-0.5">
              Santa Eulàlia de Ronçana
            </p>
          </div>
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

          <div className="space-y-3">

            {/* Primary card — adapts based on whether there's an active streak */}
            <div className="status-card">
              <div className={`w-[3px] self-stretch rounded-full shrink-0 my-0.5 ${hasActiveStreak ? 'bg-red-400' : 'bg-emerald-400'}`} />

              <div className="flex-1 min-w-0">
                {statsLoading ? (
                  <div className="space-y-2">
                    <div className="h-14 w-20 bg-white/6 rounded animate-pulse" />
                    <div className="h-3  w-44 bg-white/4 rounded animate-pulse" />
                  </div>
                ) : (
                  <>
                    <div
                      className={`text-6xl sm:text-7xl font-black leading-none ${hasActiveStreak ? 'text-red-400' : 'text-emerald-400'}`}
                      style={{ fontFamily: 'Anton, sans-serif' }}
                    >
                      {primaryValue}
                    </div>
                    <p className="text-white/40 text-xs mt-1.5 tracking-wide">
                      {hasActiveStreak
                        ? 'dies consecutius sense llum'
                        : 'dies des de l\'últim incident'}
                    </p>
                  </>
                )}
              </div>

              <div className="shrink-0 self-start mt-0.5">
                {statsLoading ? (
                  <div className="h-6 w-20 bg-white/6 rounded animate-pulse" />
                ) : (
                  <div className={`flex items-center gap-1.5 px-2 py-1 rounded border text-[10px] uppercase tracking-wider whitespace-nowrap ${
                    hasActiveStreak
                      ? 'border-red-500/30 bg-red-500/8 text-red-400'
                      : 'border-emerald-500/30 bg-emerald-500/8 text-emerald-400'
                  }`}>
                    <span className={`w-1.5 h-1.5 rounded-full shrink-0 ${hasActiveStreak ? 'bg-red-400 animate-pulse' : 'bg-emerald-400'}`} />
                    {hasActiveStreak ? 'Ratxa actual' : 'Tot bé'}
                  </div>
                )}
              </div>
            </div>

            {/* Secondary card — Ratxa més llarga */}
            <div className="status-card">
              <div className="w-[3px] self-stretch bg-amber-400/50 rounded-full shrink-0 my-0.5" />

              <div className="flex-1 min-w-0">
                {statsLoading ? (
                  <div className="space-y-2">
                    <div className="h-9 w-16 bg-white/6 rounded animate-pulse" />
                    <div className="h-3 w-36 bg-white/4 rounded animate-pulse" />
                  </div>
                ) : (
                  <>
                    <div
                      className="text-4xl font-black leading-none text-amber-400/80"
                      style={{ fontFamily: 'Anton, sans-serif' }}
                    >
                      {longestStreak}
                    </div>
                    <p className="text-white/35 text-xs mt-1 tracking-wide">
                      ratxa més llarga sense llum
                    </p>
                  </>
                )}
              </div>

              <div className="shrink-0 self-start mt-0.5">
                <div className="flex items-center gap-1.5 px-2 py-1 rounded border border-amber-500/20 bg-amber-500/6">
                  <span className="text-[10px] text-amber-400/50 uppercase tracking-wider whitespace-nowrap">
                    Rècord
                  </span>
                </div>
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

        {/* ── Dades anuals ── */}
        <section className="py-5">
          <p className="section-label">Dades anuals</p>
          <Stats stats={stats} loading={statsLoading} />
        </section>

        <footer className="py-5 border-t border-white/6 text-center space-y-2">
          <p className="text-white/15 text-[10px] tracking-widest">
            alesfosquescat · Santa Eulàlia de Ronçana · {new Date().getFullYear()}
          </p>
          <p className="text-white/20 text-[10px] tracking-wide">
            Sàtira · Humor · Veïns indignats 🕯️
          </p>
          <div className="flex items-center justify-center gap-4">
            <a
              href="/legal"
              className="text-white/20 text-[10px] tracking-widest hover:text-amber-400/50 transition-colors"
            >
              Avís legal
            </a>
            <a
              href="https://www.instagram.com/ajstaeulaliaderoncana/"
              target="_blank"
              rel="noopener noreferrer"
              aria-label="Instagram de l'Ajuntament de Santa Eulàlia de Ronçana"
              className="text-white/20 hover:text-amber-400/50 transition-colors"
            >
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" className="w-3.5 h-3.5">
                <rect x="2" y="2" width="20" height="20" rx="5" ry="5"/>
                <circle cx="12" cy="12" r="4"/>
                <circle cx="17.5" cy="6.5" r="0.5" fill="currentColor" stroke="none"/>
              </svg>
            </a>
          </div>
        </footer>
      </main>
    </div>
  )
}
