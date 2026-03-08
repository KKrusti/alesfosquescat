import { useState, useCallback } from 'react'
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

export function BatSignal({ onSuccess }: Props) {
  const [state, setState] = useState<AnimState>('idle')
  const [message, setMessage] = useState<string | null>(null)

  const handleClick = useCallback(async () => {
    if (state === 'loading') return

    if (getCookie('afc_voted') === getToday()) {
      setState('already_voted')
      setMessage('Ja ho sabem, tio 🕯️')
      setTimeout(() => { setState('idle'); setMessage(null) }, 2500)
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
        setState('already_voted')
        setMessage('Ja ho sabem, tio 🕯️')
        setTimeout(() => { setState('idle'); setMessage(null) }, 2500)
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

  const isClickable = state !== 'loading'

  // Container animation (shake / blink applied to the whole image)
  const containerAnim =
    state === 'already_voted' ? 'animate-shake' :
    (state === 'network_error' || state === 'server_error') ? 'animate-blink' :
    ''

  const msgStyle =
    state === 'already_voted'
      ? 'border border-signal-600/50 bg-signal-700/10 text-signal-400'
      : 'border border-red-500/50 bg-red-900/10 text-red-400'

  return (
    <div className="flex flex-col items-center gap-3">

      {/* ── Image wrapper ── */}
      <div className={`relative touch-signal ${containerAnim}`}>

        {/* Success: white flash overlay */}
        {state === 'success' && (
          <div className="absolute inset-0 rounded-[18px] bg-white/30 animate-pulse pointer-events-none z-10" />
        )}

        <img
          src="/signal.png"
          alt="Foco proyectando la A en el cielo — toca per reportar un apagó"
          draggable={false}
          role="button"
          tabIndex={0}
          aria-label="Toca per reportar un apagó"
          className={[
            'block select-none',
            // Size: ~65% viewport width on mobile, capped at 300px on larger screens
            'w-[min(52vw,220px)] sm:w-56',
            // Press feedback
            'transition-[opacity,transform] duration-150',
            isClickable
              ? 'cursor-pointer active:opacity-70 active:scale-[0.96]'
              : 'cursor-wait opacity-50',
            // Loading: subtle pulse
            state === 'loading' ? 'animate-pulse' : '',
          ].join(' ')}
          onClick={isClickable ? handleClick : undefined}
          onKeyDown={(e) => { if (e.key === 'Enter' && isClickable) handleClick() }}
        />
      </div>

      {/* State message */}
      {message && (
        <div className={`px-4 py-2.5 rounded font-mono text-sm font-bold text-center max-w-[260px] ${msgStyle}`}>
          {message}
        </div>
      )}

      {/* Hint */}
      <p className={[
        'text-[11px] font-mono uppercase tracking-[0.25em] transition-opacity duration-300',
        state === 'loading' ? 'text-signal-500/70 animate-pulse' : 'text-white/25',
      ].join(' ')}>
        {state === 'loading' ? 'enviant...' : '— toca per reportar —'}
      </p>
    </div>
  )
}
