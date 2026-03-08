import type { StatsResponse } from '../types'
import { useLanguage } from '../context/LanguageContext'

interface Props {
  stats: StatsResponse | null
  loading: boolean
}

function SkeletonRow() {
  return (
    <div className="flex items-center justify-between py-3.5 border-b border-stone-200 dark:border-white/7">
      <div className="h-3.5 w-44 bg-stone-200 dark:bg-white/7 rounded animate-pulse" />
      <div className="h-6 w-8 bg-stone-200 dark:bg-white/7 rounded animate-pulse ml-4" />
    </div>
  )
}

interface StatRowProps {
  label: string
  value: number
  text: string
  index: number
}

function StatRow({ label, value, text, index }: StatRowProps) {
  return (
    <div
      className="flex items-center justify-between py-3.5 border-b border-stone-200 dark:border-white/7"
      style={{ animationDelay: `${index * 50}ms` }}
    >
      <span className="text-[13px] text-stone-500 dark:text-white/50 tracking-wide">{label}</span>
      <span
        className={`text-2xl font-black leading-none shrink-0 ml-4 ${text}`}
        style={{ fontFamily: 'Anton, sans-serif' }}
      >
        {value}
      </span>
    </div>
  )
}

export function Stats({ stats, loading }: Props) {
  const { t } = useLanguage()

  // Stat rows definition — built at render time so labels react to language changes
  const rows = [
    { key: 'total_this_year'        as const, label: t.statNights,  text: 'text-amber-700 dark:text-amber-400' },
    { key: 'longest_incident_streak'as const, label: t.statLongest, text: 'text-amber-600 dark:text-yellow-300' },
    { key: 'normal_days_this_year'  as const, label: t.statNormal,  text: 'text-sky-600 dark:text-sky-400' },
  ]

  if (loading) {
    return (
      <div>
        {rows.map((_, i) => <SkeletonRow key={i} />)}
      </div>
    )
  }

  if (!stats) return null

  return (
    <div>
      {rows.map((row, i) => (
        <StatRow
          key={row.key}
          label={row.label}
          value={stats[row.key]}
          text={row.text}
          index={i}
        />
      ))}

      {/* Percentage split bar — incident nights vs normal nights */}
      {(() => {
        const daysElapsed = stats.total_this_year + stats.normal_days_this_year
        if (daysElapsed === 0) return null
        const pctIncident = Math.round(stats.total_this_year / daysElapsed * 100)
        const pctNormal   = 100 - pctIncident
        return (
          <div className="pt-4 pb-1">
            <div className="h-1.5 flex rounded-full overflow-hidden bg-stone-200 dark:bg-white/8">
              <div
                className="bg-amber-500/70 dark:bg-amber-400/55 transition-all duration-500"
                style={{ width: `${pctIncident}%` }}
              />
            </div>
            <div className="flex justify-between mt-1.5">
              <span className="text-[10px] text-amber-600/80 dark:text-amber-400/50 font-mono tabular-nums">
                {pctIncident}% {t.statNights.toLowerCase()}
              </span>
              <span className="text-[10px] text-sky-600/70 dark:text-sky-400/50 font-mono tabular-nums">
                {pctNormal}% {t.statNormal.toLowerCase()}
              </span>
            </div>
          </div>
        )
      })()}

      {stats.last_incident_date && (
        <div className="flex items-center justify-between pt-4">
          <span className="text-[11px] text-stone-400 dark:text-white/25 uppercase tracking-widest">
            {t.statLastIncident}
          </span>
          <span className="text-[13px] text-stone-500 dark:text-white/40 font-mono">
            {stats.last_incident_date.split('-').reverse().join('-')}
          </span>
        </div>
      )}
    </div>
  )
}
