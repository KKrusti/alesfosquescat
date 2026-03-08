import { useEffect, useState } from 'react'
import { useLanguage } from '../context/LanguageContext'
import type { WeatherResponse } from '../types'

export function WeatherAlert() {
  const { t } = useLanguage()
  const [data, setData] = useState<WeatherResponse | null>(null)

  useEffect(() => {
    if (new URLSearchParams(window.location.search).get('mock') === '1') {
      setData({ alert: true, days_until: 3, mm: 18.5, prob: 75 })
      return
    }
    fetch('/api/weather')
      .then(r => (r.ok ? r.json() : null))
      .then((d: WeatherResponse | null) => setData(d))
      .catch(() => {})
  }, [])

  if (!data?.alert) return null

  const message =
    data.days_until === 0
      ? t.weatherAlertToday
      : t.weatherAlertDays.replace('{days}', String(data.days_until))

  return (
    <div
      role="alert"
      aria-live="polite"
      className="status-card border-sky-400/35 dark:border-sky-500/20 bg-sky-500/6 dark:bg-sky-500/4"
    >
      <div className="w-[3px] self-stretch rounded-full shrink-0 my-0.5 bg-sky-500 dark:bg-sky-400" />
      <div className="flex-1 min-w-0">
        <p className="text-sky-700 dark:text-sky-300 text-sm leading-snug">
          {message}
        </p>
        <p className="text-sky-600/70 dark:text-sky-400/50 text-[11px] mt-1 font-mono tracking-wide">
          {data.prob > 0 ? `${data.prob}% · ` : ''}{data.mm.toFixed(1)} mm
        </p>
      </div>
    </div>
  )
}
