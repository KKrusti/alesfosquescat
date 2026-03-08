import { useState, useCallback, useRef } from 'react'
import type { AnimState } from '../types'

interface Props {
  onSuccess: () => void
}

function getCookie(name: string): string {
  const value = `; ${document.cookie}`
  const parts = value.split(`; ${name}=`)
  if (parts.length === 2) return parts.pop()!.split(';').shift() ?? ''
  return ''
}

function setCookie(name: string, value: string) {
  const d = new Date()
  d.setDate(d.getDate() + 1)
  document.cookie = `${name}=${value}; expires=${d.toUTCString()}; path=/; SameSite=Strict`
}

function getToday(): string {
  return new Date().toISOString().split('T')[0]
}

type ResolveState = 'idle' | 'loading' | 'success' | 'error'

export function BatSignal({ onSuccess }: Props) {
  const [state, setState] = useState<AnimState>('idle')
  const [message, setMessage] = useState<string | null>(null)
  const [resolveState, setResolveState] = useState<ResolveState>('idle')
  const alreadyVotedHits = useRef(0)

  const handleClick = useCallback(async () => {
    if (state === 'loading') return

    if (getCookie('afc_voted') === getToday()) {
      alreadyVotedHits.current++
      const msg = alreadyVotedHits.current > 1
        ? 'Per molt que insisteixis, la llum no tornarà abans (per desgràcia)'
        : 'Gràcies, un altre veí ja ha reportat l\'incidència d\'avui'
      setState('already_voted')
      setMessage(msg)
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

      if (res.status === 409) {
        setCookie('afc_voted', getToday())
        alreadyVotedHits.current = 1
        setState('already_voted')
        setMessage('Gràcies, un altre veí ja ha reportat l\'incidència d\'avui')
        setTimeout(() => { setState('idle'); setMessage(null) }, 2800)
      } else if (res.ok) {
        setCookie('afc_voted', getToday())
        setState('success')
        setMessage(null)
        onSuccess()
        setTimeout(() => setState('idle'), 3000)
      } else if (res.status >= 500) {
        setState('server_error')
        setMessage('El servidor també s\'ha apagat')
        setTimeout(() => { setState('idle'); setMessage(null) }, 3500)
      } else {
        setState('network_error')
        setMessage('S\'ha anat la llum... irònic')
        setTimeout(() => { setState('idle'); setMessage(null) }, 3500)
      }
    } catch (err) {
      const isAbort = (err as Error).name === 'AbortError'
      setState('network_error')
      setMessage(isAbort ? 'Timeout — s\'ha anat la llum... irònic' : 'S\'ha anat la llum... irònic')
      setTimeout(() => { setState('idle'); setMessage(null) }, 3500)
    }
  }, [state, onSuccess])

  const handleResolve = useCallback(async () => {
    if (resolveState === 'loading') return
    setResolveState('loading')
    try {
      const res = await fetch('/api/resolve', { method: 'POST' })
      if (res.ok) {
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
  }, [resolveState, onSuccess])

  const isClickable = state !== 'loading'
  const isResolvable = resolveState !== 'loading'

  const containerAnim =
    state === 'already_voted' ? 'animate-shake' :
    (state === 'network_error' || state === 'server_error') ? 'animate-blink' :
    ''

  const msgStyle =
    state === 'already_voted'
      ? 'border border-amber-500/50 bg-amber-900/10 text-amber-400'
      : 'border border-red-500/50 bg-red-900/10 text-red-400'

  return (
    <div className="flex flex-col items-center gap-3">

      {/* ── Both buttons side by side ── */}
      <div className="flex items-end justify-center gap-6">

        {/* Report button */}
        <div className="flex flex-col items-center gap-1.5">
          <div className={`relative touch-signal ${containerAnim}`}>
            {state === 'success' && (
              <div className="absolute inset-0 rounded-[18px] bg-white/30 animate-pulse pointer-events-none z-10" />
            )}
            <img
              src="/signal.png"
              alt="Foco projectant la A — toca per reportar un apagó"
              draggable={false}
              role="button"
              tabIndex={0}
              aria-label="Toca per reportar un apagó"
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
              onKeyDown={(e) => { if (e.key === 'Enter' && isClickable) handleClick() }}
            />
          </div>
          <p className={[
            'text-[10px] font-mono uppercase tracking-[0.2em] transition-opacity duration-300',
            state === 'loading' ? 'text-amber-400/60 animate-pulse' : 'text-white/25',
          ].join(' ')}>
            {state === 'loading' ? 'enviant...' : '— reportar —'}
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
              alt="Bombeta encesa — toca per marcar que la llum ha tornat"
              draggable={false}
              role="button"
              tabIndex={0}
              aria-label="Toca per marcar que la llum ha tornat"
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
              onKeyDown={(e) => { if (e.key === 'Enter' && isResolvable) handleResolve() }}
            />
          </div>
          <p className={[
            'text-[10px] font-mono uppercase tracking-[0.2em] transition-opacity duration-300',
            resolveState === 'loading' ? 'text-emerald-400/60 animate-pulse' : 'text-white/25',
          ].join(' ')}>
            {resolveState === 'loading' ? 'enviant...' :
             resolveState === 'success' ? '✓ resolt' :
             '— resolta —'}
          </p>
        </div>

      </div>

      {/* Report state message */}
      {message && (
        <div className={`px-4 py-2.5 rounded font-mono text-sm font-bold text-center max-w-[280px] ${msgStyle}`}>
          {message}
        </div>
      )}

      {/* Resolve error message */}
      {resolveState === 'error' && (
        <div className="px-4 py-2.5 rounded font-mono text-sm font-bold text-center max-w-[280px] border border-red-500/50 bg-red-900/10 text-red-400">
          Error al marcar la resolució
        </div>
      )}

    </div>
  )
}
