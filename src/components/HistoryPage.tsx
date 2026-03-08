import { useEffect, useState } from 'react'
import { useLanguage } from '../context/LanguageContext'

interface IncidentPeriod {
  start_date: string
  days: number
}

interface InteractionEntry {
  action: 'report' | 'resolve'
  at: string // "DD-MM-YYYY HH:mm" already formatted server-side
}

function formatDate(iso: string): string {
  return iso.split('-').reverse().join('-')
}

function SkeletonRow() {
  return (
    <div className="flex items-center justify-between py-3.5 border-b border-stone-200 dark:border-white/7">
      <div className="h-3.5 w-28 bg-stone-200 dark:bg-white/7 rounded animate-pulse" />
      <div className="h-3.5 w-16 bg-stone-200 dark:bg-white/7 rounded animate-pulse" />
    </div>
  )
}

function ActivitySkeletonRow() {
  return (
    <div className="flex items-center gap-3 py-3 px-3 border-b border-stone-200 dark:border-white/7">
      <div className="w-2 h-2 rounded-full bg-stone-200 dark:bg-white/7 shrink-0 animate-pulse" />
      <div className="h-3 w-36 bg-stone-200 dark:bg-white/7 rounded animate-pulse" />
      <div className="h-3 w-24 bg-stone-200 dark:bg-white/7 rounded animate-pulse ml-auto" />
    </div>
  )
}

export function HistoryPage() {
  const { t } = useLanguage()
  const [periods, setPeriods]           = useState<IncidentPeriod[]>([])
  const [interactions, setInteractions] = useState<InteractionEntry[]>([])
  const [loadingPeriods, setLoadingPeriods]           = useState(true)
  const [loadingInteractions, setLoadingInteractions] = useState(true)

  useEffect(() => {
    fetch('/api/history')
      .then(r => r.ok ? r.json() : [])
      .then(setPeriods)
      .catch(() => setPeriods([]))
      .finally(() => setLoadingPeriods(false))

    fetch('/api/interactions')
      .then(r => r.ok ? r.json() : [])
      .then(setInteractions)
      .catch(() => setInteractions([]))
      .finally(() => setLoadingInteractions(false))
  }, [])

  return (
    <div className="min-h-screen bg-stone-100 dark:bg-[#0d0d1c] flex flex-col transition-colors duration-200">

      <header className="border-b border-stone-200 dark:border-white/8 pt-safe">
        <div className="max-w-lg mx-auto px-4 py-3 flex items-center gap-3">
          <div className="w-[3px] self-stretch bg-amber-600/40 dark:bg-amber-400/40 rounded-full shrink-0" />
          <div className="flex-1 min-w-0">
            <a href="/" className="text-amber-700/70 dark:text-amber-400/60 font-bold text-[15px] leading-tight hover:text-amber-700 dark:hover:text-amber-400 transition-colors">
              alesfosquescat
            </a>
            <p className="text-stone-400 dark:text-white/20 text-[11px] leading-tight mt-0.5">Santa Eulàlia de Ronçana</p>
          </div>
        </div>
      </header>

      <main className="flex-1 max-w-lg mx-auto w-full px-4 py-8 space-y-10">

        {/* ── Incident periods ── */}
        <section>
          <p className="section-label mb-4">{t.historyTitle}</p>
          <div>
            {loadingPeriods ? (
              Array.from({ length: 4 }).map((_, i) => <SkeletonRow key={i} />)
            ) : periods.length === 0 ? (
              <p className="text-stone-400 dark:text-white/25 text-[13px] py-6 text-center">
                {t.historyEmpty}
              </p>
            ) : (
              periods.map((p, i) => {
                const label = p.days === 1 ? t.historyDaySingle : t.historyDayPlural
                return (
                  <div
                    key={i}
                    className="flex items-center justify-between py-3.5 border-b border-stone-200 dark:border-white/7"
                  >
                    <span className="text-[13px] text-stone-500 dark:text-white/50 font-mono">
                      {formatDate(p.start_date)}
                    </span>
                    <span
                      className="text-xl font-black leading-none text-amber-700 dark:text-amber-400 shrink-0 ml-4"
                      style={{ fontFamily: 'Anton, sans-serif' }}
                    >
                      {p.days} <span className="text-[11px] font-normal tracking-wide">{label}</span>
                    </span>
                  </div>
                )
              })
            )}
          </div>
        </section>

        {/* ── Activity feed ── */}
        <section>
          <p className="section-label mb-4">{t.historyActivityTitle}</p>
          <div className="max-h-[336px] overflow-y-auto overscroll-contain rounded-lg border border-stone-200 dark:border-white/7">
            {loadingInteractions ? (
              Array.from({ length: 6 }).map((_, i) => <ActivitySkeletonRow key={i} />)
            ) : interactions.length === 0 ? (
              <p className="text-stone-400 dark:text-white/25 text-[13px] py-6 text-center">
                {t.historyActivityEmpty}
              </p>
            ) : (
              interactions.map((entry, i) => {
                const isReport = entry.action === 'report'
                return (
                  <div
                    key={i}
                    className={`flex items-center gap-3 py-3 px-3 ${i < interactions.length - 1 ? 'border-b border-stone-200 dark:border-white/7' : ''}`}
                  >
                    <span className={`w-2 h-2 rounded-full shrink-0 ${
                      isReport
                        ? 'bg-amber-500 dark:bg-amber-400'
                        : 'bg-emerald-500 dark:bg-emerald-400'
                    }`} />
                    <span className={`text-[13px] ${
                      isReport
                        ? 'text-amber-700 dark:text-amber-400'
                        : 'text-emerald-700 dark:text-emerald-400'
                    }`}>
                      {isReport ? t.historyActionReport : t.historyActionResolve}
                    </span>
                    <span className="text-[11px] text-stone-400 dark:text-white/25 font-mono tabular-nums ml-auto">
                      {entry.at}
                    </span>
                  </div>
                )
              })
            )}
          </div>
        </section>

        <a
          href="/"
          className="inline-block px-5 py-2.5 rounded border border-amber-600/30 dark:border-amber-500/30 bg-amber-500/8 text-amber-700 dark:text-amber-400 text-sm font-medium hover:bg-amber-500/15 dark:hover:bg-amber-500/12 transition-colors"
        >
          {t.historyBack}
        </a>

      </main>

      <footer className="py-5 border-t border-stone-200 dark:border-white/6 text-center">
        <p className="text-stone-400 dark:text-white/15 text-[10px] tracking-widest">
          alesfosquescat · Santa Eulàlia de Ronçana · {new Date().getFullYear()}
        </p>
      </footer>
    </div>
  )
}
