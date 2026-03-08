import { useState, useCallback, useRef } from 'react'
import type { AnimState } from '../types'
import { useLanguage } from '../context/LanguageContext'

interface Props {
  onSuccess: () => void
  hasActiveStreak: boolean
}

type ResolveState = 'idle' | 'loading' | 'success' | 'error' | 'already_resolved'

export function BatSignal({ onSuccess, hasActiveStreak }: Props) {
  const { t } = useLanguage()
  const [state, setState] = useState<AnimState>('idle')
  const [message, setMessage] = useState<string | null>(null)
  const [resolveState, setResolveState] = useState<ResolveState>('idle')
  const [resolveMessage, setResolveMessage] = useState<string | null>(null)
  const resolvedToday = useRef(false)

  const handleClick = useCallback(async () => {
    if (state === 'loading') return

    // If streak is already active, nothing to report
    if (hasActiveStreak) {
      setState('already_voted')
      setMessage(t.msgAlreadyActive)
      setTimeout(() => { setState('idle'); setMessage(null) }, 2800)
      return
    }

    setState('loading')

    try {
      const controller = new AbortController()
      const tid = setTimeout(() => controller.abort(), 8000)

      const res = await fetch('/api/report', {
        method: 'POST',
        signal: controller.signal,
      })
      clearTimeout(tid)

      if (res.ok) {
        const data = await res.json()
        if (data.already_active) {
          // Backend safety net: streak became active between the check and the call
          setState('already_voted')
          setMessage(t.msgAlreadyActive)
          setTimeout(() => { setState('idle'); setMessage(null) }, 2800)
        } else {
          setState('success')
          setMessage(null)
          onSuccess()
          setTimeout(() => setState('idle'), 3000)
        }
      } else if (res.status >= 500) {
        setState('server_error')
        setMessage(t.msgServerError)
        setTimeout(() => { setState('idle'); setMessage(null) }, 3500)
      } else {
        setState('network_error')
        setMessage(t.msgNetworkError)
        setTimeout(() => { setState('idle'); setMessage(null) }, 3500)
      }
    } catch (err) {
      const isAbort = (err as Error).name === 'AbortError'
      setState('network_error')
      setMessage(isAbort ? t.msgTimeout : t.msgNetworkError)
      setTimeout(() => { setState('idle'); setMessage(null) }, 3500)
    }
  }, [state, hasActiveStreak, onSuccess, t])

  const handleResolve = useCallback(async () => {
    if (resolveState === 'loading') return

    // Client-side guard: already resolved today in this session
    if (resolvedToday.current) {
      setResolveState('already_resolved')
      setResolveMessage(t.msgAlreadyResolved)
      setTimeout(() => { setResolveState('idle'); setResolveMessage(null) }, 3500)
      return
    }

    setResolveState('loading')
    try {
      const res = await fetch('/api/resolve', { method: 'POST' })
      if (res.ok) {
        resolvedToday.current = true
        setResolveState('success')
        onSuccess()
        setTimeout(() => setResolveState('idle'), 3000)
      } else {
        setResolveState('error')
        setTimeout(() => setResolveState('idle'), 3000)
      }
    } catch {
      setResolveState('error')
      setTimeout(() => setResolveState('idle'), 3000)
    }
  }, [resolveState, onSuccess, t])

  const isClickable  = state !== 'loading'
  const isResolvable = resolveState !== 'loading'

  const containerAnim =
    state === 'already_voted' ? 'animate-shake' :
    (state === 'network_error' || state === 'server_error') ? 'animate-blink' :
    ''

  const msgStyle =
    state === 'already_voted'
      ? 'border border-amber-600/40 dark:border-amber-500/50 bg-amber-500/10 dark:bg-amber-900/10 text-amber-700 dark:text-amber-400'
      : 'border border-red-500/50 bg-red-500/10 dark:bg-red-900/10 text-red-600 dark:text-red-400'

  return (
    <div className="flex flex-col items-center gap-3">

      {/* ── Both buttons side by side ── */}
      <div className="flex items-end justify-center gap-6">

        {/* Report button */}
        <div className="flex flex-col items-center gap-1.5">
          <div className={`relative touch-signal ${containerAnim}`}>
            {state === 'success' && (
              <div className="absolute inset-0 rounded-[18px] bg-white/30 dark:bg-amber-400/15 animate-pulse pointer-events-none z-10" />
            )}
            <img
              src="/signal.png"
              alt={t.reportAlt}
              draggable={false}
              role="button"
              tabIndex={0}
              aria-label={t.reportAriaLabel}
              className={[
                'block select-none',
                'w-[min(42vw,180px)] sm:w-44',
                'transition-[opacity,transform] duration-150',
                isClickable
                  ? 'cursor-pointer active:opacity-70 active:scale-[0.96]'
                  : 'cursor-wait opacity-50',
                state === 'loading' ? 'animate-pulse' : '',
              ].join(' ')}
              onClick={isClickable ? handleClick : undefined}
              onKeyDown={(e) => { if ((e.key === 'Enter' || e.key === ' ') && isClickable) handleClick() }}
            />
          </div>
          <p className={[
            'text-[10px] font-mono uppercase tracking-[0.2em] transition-opacity duration-300',
            state === 'loading' ? 'text-amber-600 dark:text-amber-400/60 animate-pulse' : 'text-stone-400 dark:text-white/25',
          ].join(' ')}>
            {state === 'loading' ? t.sending : t.reportLabel}
          </p>
        </div>

        {/* Resolve button */}
        <div className="flex flex-col items-center gap-1.5">
          <div className="relative touch-signal">
            {resolveState === 'success' && (
              <div className="absolute inset-0 rounded-[18px] bg-emerald-400/20 animate-pulse pointer-events-none z-10" />
            )}
            <img
              src="/resolve.jpg"
              alt={t.resolveAlt}
              draggable={false}
              role="button"
              tabIndex={0}
              aria-label={t.resolveAriaLabel}
              className={[
                'block select-none rounded-xl',
                'w-[min(21vw,90px)] sm:w-[88px]',
                'transition-[opacity,transform] duration-150',
                isResolvable
                  ? 'cursor-pointer active:opacity-70 active:scale-[0.96]'
                  : 'cursor-wait opacity-50',
                resolveState === 'loading' ? 'animate-pulse' : '',
              ].join(' ')}
              onClick={isResolvable ? handleResolve : undefined}
              onKeyDown={(e) => { if ((e.key === 'Enter' || e.key === ' ') && isResolvable) handleResolve() }}
            />
          </div>
          <p className={[
            'text-[10px] font-mono uppercase tracking-[0.2em] transition-opacity duration-300',
            resolveState === 'loading' ? 'text-emerald-600 dark:text-emerald-400/60 animate-pulse' : 'text-stone-400 dark:text-white/25',
          ].join(' ')}>
            {resolveState === 'loading' ? t.resolveSending :
             resolveState === 'success' ? t.resolvedLabel :
             t.resolveLabel}
          </p>
        </div>

      </div>

      {/* Report state message */}
      {message && (
        <div className={`px-4 py-2.5 rounded font-mono text-sm font-bold text-center max-w-[280px] ${msgStyle}`}>
          {message}
        </div>
      )}

      {/* Resolve state message */}
      {resolveState === 'error' && (
        <div className="px-4 py-2.5 rounded font-mono text-sm font-bold text-center max-w-[280px] border border-red-500/50 bg-red-500/10 dark:bg-red-900/10 text-red-600 dark:text-red-400">
          {t.msgResolveError}
        </div>
      )}
      {resolveState === 'already_resolved' && resolveMessage && (
        <div className="px-4 py-2.5 rounded font-mono text-sm font-bold text-center max-w-[280px] border border-amber-600/40 dark:border-amber-500/50 bg-amber-500/10 dark:bg-amber-900/10 text-amber-700 dark:text-amber-400">
          {resolveMessage}
        </div>
      )}

    </div>
  )
}
