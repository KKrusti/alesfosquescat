import type { StatsResponse } from '../types'

interface Props {
  stats: StatsResponse | null
  loading: boolean
}

interface StatItem {
  key: keyof Omit<StatsResponse, 'last_incident_date'>
  label: string
  text: string
}

const STAT_ITEMS: StatItem[] = [
  { key: 'total_this_year',           label: 'Dies sense llum',           text: 'text-amber-400' },
  { key: 'longest_incident_streak',   label: 'Racha màx. d\'apagons',     text: 'text-yellow-300' },
  { key: 'days_since_last_incident',  label: 'Dies des del darrer',        text: 'text-emerald-400' },
  { key: 'longest_no_incident_streak',label: 'Racha màx. sense apagons',  text: 'text-sky-400' },
  { key: 'current_incident_streak',   label: 'Racha actual d\'apagons',   text: 'text-red-400' },
]

function StatRow({ item, value, index }: { item: StatItem; value: number; index: number }) {
  return (
    <div
      className="flex items-center justify-between py-3.5 border-b border-white/7"
      style={{ animationDelay: `${index * 50}ms` }}
    >
      <span className="text-[13px] text-white/50 tracking-wide">{item.label}</span>
      <span
        className={`text-2xl font-black leading-none shrink-0 ml-4 ${item.text}`}
        style={{ fontFamily: 'Anton, sans-serif' }}
      >
        {value}
      </span>
    </div>
  )
}

function SkeletonRow() {
  return (
    <div className="flex items-center justify-between py-3.5 border-b border-white/7">
      <div className="h-3.5 w-44 bg-white/7 rounded animate-pulse" />
      <div className="h-6 w-8 bg-white/7 rounded animate-pulse ml-4" />
    </div>
  )
}

export function Stats({ stats, loading }: Props) {
  if (loading) {
    return (
      <div>
        {STAT_ITEMS.map((_, i) => <SkeletonRow key={i} />)}
      </div>
    )
  }

  if (!stats) return null

  return (
    <div>
      {STAT_ITEMS.map((item, i) => (
        <StatRow key={item.key} item={item} value={stats[item.key]} index={i} />
      ))}

      {stats.last_incident_date && (
        <div className="flex items-center justify-between pt-4">
          <span className="text-[11px] text-white/25 uppercase tracking-widest">
            Darrer incident
          </span>
          <span className="text-[13px] text-white/40 font-mono">
            {stats.last_incident_date}
          </span>
        </div>
      )}
    </div>
  )
}
