import { useState, useEffect, useCallback } from 'react'
import { BatSignal } from './components/BatSignal'
import { Stats } from './components/Stats'
import { useTheme } from './context/ThemeContext'
import { useLanguage } from './context/LanguageContext'
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

function SunIcon() {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="w-4 h-4">
      <circle cx="12" cy="12" r="4"/>
      <path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41"/>
    </svg>
  )
}

function MoonIcon() {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="w-4 h-4">
      <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
    </svg>
  )
}

export default function App() {
  const [stats, setStats] = useState<StatsResponse | null>(null)
  const [statsLoading, setStatsLoading] = useState(true)
  const { theme, toggle } = useTheme()
  const { lang, t, toggle: toggleLang } = useLanguage()

  const fetchStats = useCallback(async () => {
    try {
      const res = await fetch('/api/stats')
      if (res.ok) setStats(await res.json())
    } catch { /* non-critical */ }
    finally { setStatsLoading(false) }
  }, [])

  useEffect(() => { fetchStats() }, [fetchStats])

  // Auto-refresh at midnight so the "days since last incident" counter increments on its own
  useEffect(() => {
    function scheduleNext() {
      return onNextMidnight(() => { fetchStats(); scheduleNext() })
    }
    return scheduleNext()
  }, [fetchStats])

  const hasActiveStreak = (stats?.current_incident_streak ?? 0) > 0

  const primaryValue: number = stats
    ? (hasActiveStreak ? stats.current_incident_streak : stats.days_since_last_incident)
    : 0

  const longestStreak: number = stats?.longest_incident_streak ?? 0

  return (
    <div className="min-h-screen bg-stone-100 dark:bg-[#0d0d1c] transition-colors duration-200">

      {/* ── Header ─────────────────────────────────────────────────── */}
      <header className="sticky top-0 z-20 bg-stone-100/95 dark:bg-[#0d0d1c]/95 backdrop-blur-sm border-b border-stone-200 dark:border-white/8 pt-safe transition-colors duration-200">
        <div className="max-w-lg mx-auto px-4 py-3 flex items-center gap-3">
          <div className="w-[3px] self-stretch bg-amber-600 dark:bg-amber-400 rounded-full shrink-0" />
          <div className="flex-1 min-w-0">
            <p className="text-amber-700 dark:text-amber-400 font-bold text-[15px] leading-tight tracking-tight">
              alesfosquescat
            </p>
            <p className="text-stone-500 dark:text-white/35 text-[11px] leading-tight mt-0.5">
              Santa Eulàlia de Ronçana
            </p>
          </div>
          <div className="flex items-center gap-2">
            {/* Live badge */}
            <div className="flex items-center gap-1.5 px-2.5 py-1 rounded border border-amber-600/30 dark:border-amber-500/25 bg-amber-500/10 dark:bg-amber-500/8 shrink-0">
              <span className="w-1.5 h-1.5 rounded-full bg-amber-600 dark:bg-amber-400 animate-pulse" />
              <span className="text-[10px] text-amber-700 dark:text-amber-400/80 uppercase tracking-widest font-medium">
                {t.live}
              </span>
            </div>
            {/* Language toggle */}
            <button
              onClick={toggleLang}
              aria-label={lang === 'ca' ? 'Cambiar a castellano' : 'Canviar a català'}
              className="px-2 py-1.5 rounded-lg border border-stone-200 dark:border-white/10 bg-white/60 dark:bg-white/5 text-stone-500 dark:text-white/40 hover:text-amber-700 dark:hover:text-amber-400 hover:border-amber-600/30 dark:hover:border-amber-500/30 transition-colors duration-150 cursor-pointer text-[10px] font-bold tracking-widest"
            >
              {lang === 'ca' ? 'ES' : 'CA'}
            </button>
            {/* Theme toggle */}
            <button
              onClick={toggle}
              aria-label={theme === 'dark' ? t.toggleLight : t.toggleDark}
              className="p-2 rounded-lg border border-stone-200 dark:border-white/10 bg-white/60 dark:bg-white/5 text-stone-500 dark:text-white/40 hover:text-amber-700 dark:hover:text-amber-400 hover:border-amber-600/30 dark:hover:border-amber-500/30 transition-colors duration-150 cursor-pointer"
            >
              {theme === 'dark' ? <SunIcon /> : <MoonIcon />}
            </button>
          </div>
        </div>
      </header>

      {/* ── Content ─────────────────────────────────────────────────── */}
      <main className="max-w-lg mx-auto px-4 pb-safe">

        {/* ── Current status ── */}
        <section className="pt-5 pb-4">
          <p className="section-label">{t.sectionStatus}</p>

          <div className="space-y-3">

            {/* Primary card */}
            <div className="status-card">
              <div className={`w-[3px] self-stretch rounded-full shrink-0 my-0.5 ${hasActiveStreak ? 'bg-red-500 dark:bg-red-400' : 'bg-emerald-600 dark:bg-emerald-400'}`} />

              <div className="flex-1 min-w-0">
                {statsLoading ? (
                  <div className="space-y-2">
                    <div className="h-14 w-20 bg-stone-200 dark:bg-white/6 rounded animate-pulse" />
                    <div className="h-3  w-44 bg-stone-100 dark:bg-white/4 rounded animate-pulse" />
                  </div>
                ) : (
                  <>
                    <div
                      className={`text-6xl sm:text-7xl font-black leading-none ${hasActiveStreak ? 'text-red-600 dark:text-red-400' : 'text-emerald-700 dark:text-emerald-400'}`}
                      style={{ fontFamily: 'Anton, sans-serif' }}
                    >
                      {primaryValue}
                    </div>
                    <p className="text-stone-500 dark:text-white/40 text-xs mt-1.5 tracking-wide">
                      {hasActiveStreak ? t.streakLabel : t.daysSinceLabel}
                    </p>
                  </>
                )}
              </div>

              <div className="shrink-0 self-start mt-0.5">
                {statsLoading ? (
                  <div className="h-6 w-20 bg-stone-200 dark:bg-white/6 rounded animate-pulse" />
                ) : (
                  <div className={`flex items-center gap-1.5 px-2 py-1 rounded border text-[10px] uppercase tracking-wider whitespace-nowrap ${
                    hasActiveStreak
                      ? 'border-red-500/40 dark:border-red-500/30 bg-red-500/10 dark:bg-red-500/8 text-red-600 dark:text-red-400'
                      : 'border-emerald-600/30 dark:border-emerald-500/30 bg-emerald-500/10 dark:bg-emerald-500/8 text-emerald-700 dark:text-emerald-400'
                  }`}>
                    <span className={`w-1.5 h-1.5 rounded-full shrink-0 ${hasActiveStreak ? 'bg-red-500 dark:bg-red-400 animate-pulse' : 'bg-emerald-600 dark:bg-emerald-400'}`} />
                    {hasActiveStreak ? t.badgeActive : t.badgeOk}
                  </div>
                )}
              </div>
            </div>

            {/* Secondary card — longest streak */}
            <div className="status-card">
              <div className="w-[3px] self-stretch bg-amber-600/50 dark:bg-amber-400/50 rounded-full shrink-0 my-0.5" />

              <div className="flex-1 min-w-0">
                {statsLoading ? (
                  <div className="space-y-2">
                    <div className="h-9 w-16 bg-stone-200 dark:bg-white/6 rounded animate-pulse" />
                    <div className="h-3 w-36 bg-stone-100 dark:bg-white/4 rounded animate-pulse" />
                  </div>
                ) : (
                  <>
                    <div
                      className="text-4xl font-black leading-none text-amber-700 dark:text-amber-400/80"
                      style={{ fontFamily: 'Anton, sans-serif' }}
                    >
                      {longestStreak}
                    </div>
                    <p className="text-stone-500 dark:text-white/35 text-xs mt-1 tracking-wide">
                      {t.longestLabel}
                    </p>
                  </>
                )}
              </div>

              <div className="shrink-0 self-start mt-0.5">
                <div className="flex items-center gap-1.5 px-2 py-1 rounded border border-amber-600/25 dark:border-amber-500/20 bg-amber-500/8 dark:bg-amber-500/6">
                  <span className="text-[10px] text-amber-700 dark:text-amber-400/50 uppercase tracking-wider whitespace-nowrap">
                    {t.recordBadge}
                  </span>
                </div>
              </div>
            </div>

          </div>
        </section>

        <div className="section-divider" />

        {/* ── Report incident ── */}
        <section className="py-5">
          <p className="section-label">{t.sectionReport}</p>
          <div className="flex flex-col items-center">
            <BatSignal onSuccess={fetchStats} />
          </div>
        </section>

        <div className="section-divider" />

        {/* ── Annual data ── */}
        <section className="py-5">
          <p className="section-label">{t.sectionStats}</p>
          <Stats stats={stats} loading={statsLoading} />
        </section>

        <footer className="py-5 border-t border-stone-200 dark:border-white/6 text-center space-y-2">
          <p className="text-stone-400 dark:text-white/15 text-[10px] tracking-widest">
            alesfosquescat · Santa Eulàlia de Ronçana · {new Date().getFullYear()}
          </p>
          <div className="flex items-center justify-center gap-4">
            <a
              href="/legal"
              className="text-stone-400 dark:text-white/20 text-[10px] tracking-widest hover:text-amber-700 dark:hover:text-amber-400/50 transition-colors"
            >
              {t.legalLink}
            </a>
            <a
              href="https://www.instagram.com/ajstaeulaliaderoncana/"
              target="_blank"
              rel="noopener noreferrer"
              aria-label={t.instagramAriaLabel}
              className="text-stone-400 dark:text-white/20 hover:text-amber-700 dark:hover:text-amber-400/50 transition-colors"
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
