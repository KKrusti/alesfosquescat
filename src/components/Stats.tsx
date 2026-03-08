import type { StatsResponse } from '../types'

interface Props {
  stats: StatsResponse | null
  loading: boolean
}

interface StatItem {
  key: keyof Omit<StatsResponse, 'last_incident_date'>
  label: string
  icon: string
  border: string
  text: string
  bg: string
}

const STAT_ITEMS: StatItem[] = [
  {
    key: 'total_this_year',
    label: 'Dies sense llum',
    icon: '🕯',
    border: 'border-signal-600/40',
    text: 'text-signal-400',
    bg: 'bg-signal-500/5',
  },
  {
    key: 'longest_incident_streak',
    label: 'Racha màx. apagons',
    icon: '⚡',
    border: 'border-yellow-500/40',
    text: 'text-yellow-300',
    bg: 'bg-yellow-500/5',
  },
  {
    key: 'days_since_last_incident',
    label: 'Dies des del darrer',
    icon: '✓',
    border: 'border-green-500/40',
    text: 'text-green-400',
    bg: 'bg-green-500/5',
  },
  {
    key: 'longest_no_incident_streak',
    label: 'Racha màx. sense llum',
    icon: '★',
    border: 'border-sky-500/40',
    text: 'text-sky-400',
    bg: 'bg-sky-500/5',
  },
  {
    key: 'current_incident_streak',
    label: 'Racha actual apagons',
    icon: '▲',
    border: 'border-red-500/40',
    text: 'text-red-400',
    bg: 'bg-red-500/5',
  },
]

function StatCard({ item, value, index }: { item: StatItem; value: number; index: number }) {
  return (
    <div
      className={[
        'rounded border p-4 backdrop-blur-sm animate-slide-up',
        item.border, item.bg,
      ].join(' ')}
      style={{ animationDelay: `${index * 80}ms`, animationFillMode: 'both' }}
    >
      <div className="text-lg mb-1 font-mono text-white/40">{item.icon}</div>
      <div className={`text-3xl font-black leading-none ${item.text}`} style={{ fontFamily: 'Anton, sans-serif' }}>
        {value}
      </div>
      <div className="text-xs text-white/35 mt-1 font-mono uppercase tracking-wider leading-tight">
        {item.label}
      </div>
    </div>
  )
}

function SkeletonCard() {
  return (
    <div className="rounded border border-white/10 p-4 animate-pulse bg-white/3 h-[96px]" />
  )
}

export function Stats({ stats, loading }: Props) {
  if (loading) {
    return (
      <div className="grid grid-cols-2 sm:grid-cols-3 gap-3 w-full">
        {STAT_ITEMS.map((_, i) => <SkeletonCard key={i} />)}
      </div>
    )
  }

  if (!stats) return null

  return (
    <div className="w-full space-y-3">
      <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
        {STAT_ITEMS.map((item, i) => (
          <StatCard
            key={item.key}
            item={item}
            value={stats[item.key]}
            index={i}
          />
        ))}
      </div>

      {stats.last_incident_date && (
        <div
          className="border border-white/10 rounded p-3 bg-white/3 backdrop-blur-sm animate-slide-up"
          style={{ animationDelay: '450ms', animationFillMode: 'both' }}
        >
          <span className="text-xs font-mono text-white/30 uppercase tracking-widest">
            Darrer incident registrat
          </span>
          <span className="ml-3 font-mono text-white/60 text-sm">
            {stats.last_incident_date}
          </span>
        </div>
      )}
    </div>
  )
}
