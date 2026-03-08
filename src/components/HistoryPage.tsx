import { useEffect, useState } from 'react'
import { useLanguage } from '../context/LanguageContext'

interface IncidentPeriod {
  start_date: string
  days: number
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

export function HistoryPage() {
  const { t } = useLanguage()
  const [periods, setPeriods] = useState<IncidentPeriod[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetch('/api/history')
      .then(r => r.ok ? r.json() : [])
      .then(setPeriods)
      .catch(() => setPeriods([]))
      .finally(() => setLoading(false))
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

      <main className="flex-1 max-w-lg mx-auto w-full px-4 py-8">
        <p className="section-label mb-6">{t.historyTitle}</p>

        <div>
          {loading ? (
            Array.from({ length: 5 }).map((_, i) => <SkeletonRow key={i} />)
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

        <div className="mt-10">
          <a
            href="/"
            className="px-5 py-2.5 rounded border border-amber-600/30 dark:border-amber-500/30 bg-amber-500/8 text-amber-700 dark:text-amber-400 text-sm font-medium hover:bg-amber-500/15 dark:hover:bg-amber-500/12 transition-colors"
          >
            {t.historyBack}
          </a>
        </div>
      </main>

      <footer className="py-5 border-t border-stone-200 dark:border-white/6 text-center">
        <p className="text-stone-400 dark:text-white/15 text-[10px] tracking-widest">
          alesfosquescat · Santa Eulàlia de Ronçana · {new Date().getFullYear()}
        </p>
      </footer>
    </div>
  )
}
