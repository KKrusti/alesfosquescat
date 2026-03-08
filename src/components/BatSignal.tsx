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

    // Client-side cookie check (fast path, no network)
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

  // ── Animation classes ──────────────────────────────────────────
  const groupAnim =
    state === 'already_voted' ? 'animate-shake' : 'animate-oscillate'

  const beamClass =
    state === 'success' ? 'animate-pulse-beam' :
    (state === 'network_error' || state === 'server_error') ? 'animate-blink' :
    ''

  const letterClass = state === 'success' ? 'bat-letter-flash' : ''

  const isClickable = state !== 'loading'

  // ── Message styling ────────────────────────────────────────────
  const msgStyle =
    state === 'already_voted'
      ? 'border border-signal-600/50 bg-signal-700/10 text-signal-400'
      : 'border border-red-500/50 bg-red-900/10 text-red-400 animate-pulse'

  return (
    <div className="relative flex flex-col items-center gap-4">
      {/* State message */}
      {message && (
        <div className={`px-4 py-2 rounded font-mono text-sm font-bold ${msgStyle}`}>
          {message}
        </div>
      )}

      {/* ── Bat signal SVG ── */}
      <svg
        viewBox="0 0 300 500"
        className={[
          'w-52 sm:w-64 md:w-72 select-none transition duration-300',
          isClickable
            ? 'cursor-pointer hover:drop-shadow-[0_0_30px_rgba(251,191,36,0.35)]'
            : 'cursor-wait opacity-80',
        ].join(' ')}
        onClick={isClickable ? handleClick : undefined}
        role="button"
        aria-label="Fes click per reportar un apagó"
        tabIndex={0}
        onKeyDown={(e) => { if (e.key === 'Enter' && isClickable) handleClick() }}
      >
        <defs>
          {/* Beam gradient: dense at bottom, fades upward */}
          <linearGradient id="beamGrad" x1="0.5" y1="1" x2="0.5" y2="0">
            <stop offset="0%"   stopColor="#fcd34d" stopOpacity="0.9" />
            <stop offset="55%"  stopColor="#fbbf24" stopOpacity="0.28" />
            <stop offset="100%" stopColor="#fbbf24" stopOpacity="0.04" />
          </linearGradient>

          {/* Projector lens glow */}
          <radialGradient id="lensGlow" cx="50%" cy="50%" r="50%">
            <stop offset="0%"   stopColor="#fde68a" stopOpacity="0.95" />
            <stop offset="50%"  stopColor="#f59e0b" stopOpacity="0.55" />
            <stop offset="100%" stopColor="#b45309" stopOpacity="0" />
          </radialGradient>

          {/* Glow filter for the A */}
          <filter id="aGlow" x="-30%" y="-20%" width="160%" height="140%">
            <feGaussianBlur stdDeviation="7" result="blur" />
            <feMerge>
              <feMergeNode in="blur" />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>

          {/* Soft glow for projector */}
          <filter id="lensFilter" x="-40%" y="-40%" width="180%" height="180%">
            <feGaussianBlur stdDeviation="4" result="blur" />
            <feMerge>
              <feMergeNode in="blur" />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>
        </defs>

        {/*
          ── Oscillating group ──────────────────────────────────────
          Pivots at the projector centre (150, 458).
          transform-origin set both via style AND transform-box:fill-box
          (see index.css .bat-signal-group)
        */}
        <g
          className={`bat-signal-group ${groupAnim}`}
          style={{ transformOrigin: '150px 458px' }}
        >
          {/* Light beam */}
          <polygon
            points="150,440 42,12 258,12"
            fill="url(#beamGrad)"
            className={beamClass}
          />

          {/* Inner beam highlight */}
          <polygon
            points="150,440 110,100 190,100"
            fill="#fcd34d"
            opacity="0.06"
          />

          {/* ── The "A" — the clickable element ── */}
          <text
            x="150"
            y="300"
            textAnchor="middle"
            fontSize="168"
            fontWeight="900"
            fontFamily="'Anton', Impact, sans-serif"
            fill="#fcd34d"
            filter="url(#aGlow)"
            className={letterClass}
            style={{ letterSpacing: '-4px' }}
          >
            A
          </text>

          {/* Projector housing rect */}
          <rect
            x="112" y="444"
            width="76" height="26"
            rx="7"
            fill="#0e0e22"
            stroke="#2e2e50"
            strokeWidth="1.5"
          />

          {/* Projector outer lens ring */}
          <ellipse
            cx="150" cy="458"
            rx="40" ry="17"
            fill="#181830"
            stroke="#3a3a6a"
            strokeWidth="2"
          />

          {/* Projector mid lens */}
          <ellipse
            cx="150" cy="458"
            rx="28" ry="12"
            fill="#0d0d20"
            stroke="#4444a0"
            strokeWidth="1"
          />

          {/* Projector lens glow */}
          <ellipse
            cx="150" cy="458"
            rx="18" ry="7"
            fill="url(#lensGlow)"
            filter="url(#lensFilter)"
          />

          {/* Lens highlight specular */}
          <ellipse
            cx="142" cy="454"
            rx="7" ry="2.5"
            fill="white"
            opacity="0.28"
            transform="rotate(-18 142 454)"
          />

          {/* Projector base */}
          <rect
            x="128" y="469"
            width="44" height="9"
            rx="3"
            fill="#0a0a1c"
            stroke="#222244"
            strokeWidth="1"
          />
        </g>
      </svg>

      {/* Hint label */}
      <p className={[
        'text-xs font-mono uppercase tracking-[0.25em] transition-opacity duration-300',
        state === 'loading' ? 'text-signal-500/60 animate-pulse' : 'text-white/25',
      ].join(' ')}>
        {state === 'loading' ? 'enviant...' : '— fes click per reportar —'}
      </p>
    </div>
  )
}
