import type { StatsResponse } from '../types'

interface Props {
  stats: StatsResponse | null
  loading: boolean
}

interface StatItem {
  key: keyof Omit<StatsResponse, 'last_incident_date'>
  label: string
  icon: string
  text: string
}

const STAT_ITEMS: StatItem[] = [
  {
    key: 'total_this_year',
    label: 'Dies sense llum',
    icon: '🕯',
    text: 'text-signal-400',
  },
  {
    key: 'longest_incident_streak',
    label: 'Racha màx. apagons',
    icon: '⚡',
    text: 'text-yellow-300',
  },
  {
    key: 'days_since_last_incident',
    label: 'Dies des del darrer',
    icon: '✓',
    text: 'text-green-400',
  },
  {
    key: 'longest_no_incident_streak',
    label: 'Racha màx. sense apagons',
    icon: '★',
    text: 'text-sky-400',
  },
  {
    key: 'current_incident_streak',
    label: 'Racha actual apagons',
    icon: '▲',
    text: 'text-red-400',
  },
]

function StatRow({ item, value, index }: { item: StatItem; value: number; index: number }) {
  return (
    <div
      className="flex items-center justify-between py-3.5 border-b border-white/8 animate-slide-up"
      style={{ animationDelay: `${index * 60}ms`, animationFillMode: 'both' }}
    >
      {/* Left: icon + label */}
      <div className="flex items-center gap-3 min-w-0">
        <span className="text-base w-5 shrink-0 text-center leading-none">{item.icon}</span>
        <span className="text-xs font-mono text-white/45 uppercase tracking-wide truncate">
          {item.label}
        </span>
      </div>

      {/* Right: value */}
      <div
        className={`text-3xl font-black leading-none shrink-0 ml-3 ${item.text}`}
        style={{ fontFamily: 'Anton, sans-serif' }}
      >
        {value}
      </div>
    </div>
  )
}

function SkeletonRow() {
  return (
    <div className="flex items-center justify-between py-3.5 border-b border-white/8">
      <div className="h-4 w-40 bg-white/8 rounded animate-pulse" />
      <div className="h-8 w-10 bg-white/8 rounded animate-pulse ml-3" />
    </div>
  )
}

export function Stats({ stats, loading }: Props) {
  if (loading) {
    return (
      <div className="w-full">
        {STAT_ITEMS.map((_, i) => <SkeletonRow key={i} />)}
      </div>
    )
  }

  if (!stats) return null

  return (
    <div className="w-full">
      {STAT_ITEMS.map((item, i) => (
        <StatRow
          key={item.key}
          item={item}
          value={stats[item.key]}
          index={i}
        />
      ))}

      {stats.last_incident_date && (
        <div
          className="flex items-center justify-between pt-4 animate-slide-up"
          style={{ animationDelay: '320ms', animationFillMode: 'both' }}
        >
          <span className="text-[10px] font-mono text-white/25 uppercase tracking-widest">
            Darrer incident
          </span>
          <span className="font-mono text-white/45 text-xs">
            {stats.last_incident_date}
          </span>
        </div>
      )}
    </div>
  )
}
