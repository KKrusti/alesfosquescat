import { useState, useEffect, useCallback } from 'react'

const TOKEN_KEY = 'admin_token'

type DeleteState = 'idle' | 'confirming' | 'loading' | 'done' | 'error'

interface IncidentRowProps {
  date: string
  token: string
  onDeleted: (date: string) => void
}

function IncidentRow({ date, token, onDeleted }: IncidentRowProps) {
  const [deleteState, setDeleteState] = useState<DeleteState>('idle')

  const handleDelete = useCallback(async () => {
    setDeleteState('loading')
    try {
      const res = await fetch(`/api/admin?t=${encodeURIComponent(token)}&date=${date}`, {
        method: 'DELETE',
      })
      if (res.ok) {
        setDeleteState('done')
        setTimeout(() => onDeleted(date), 500)
      } else {
        setDeleteState('error')
        setTimeout(() => setDeleteState('idle'), 3000)
      }
    } catch {
      setDeleteState('error')
      setTimeout(() => setDeleteState('idle'), 3000)
    }
  }, [date, token, onDeleted])

  return (
    <div className={[
      'flex items-center justify-between px-4 py-3 rounded-lg border transition-all duration-300',
      deleteState === 'done'
        ? 'opacity-0 scale-95 border-emerald-500/20 bg-emerald-500/5'
        : 'border-stone-200 dark:border-white/8 bg-white dark:bg-white/3',
    ].join(' ')}>
      <span className="font-mono text-sm text-stone-700 dark:text-white/70 tabular-nums">
        {date}
      </span>

      <div className="flex items-center gap-2">
        {deleteState === 'idle' && (
          <button
            onClick={() => setDeleteState('confirming')}
            className="px-3 py-1 rounded text-[11px] font-mono uppercase tracking-widest border border-red-400/30 text-red-500 dark:text-red-400/70 hover:bg-red-500/10 hover:border-red-400/50 transition-colors duration-150 cursor-pointer"
          >
            Borrar
          </button>
        )}

        {deleteState === 'confirming' && (
          <div className="flex items-center gap-2">
            <span className="text-[11px] font-mono text-stone-500 dark:text-white/40">¿Seguro?</span>
            <button
              onClick={handleDelete}
              className="px-3 py-1 rounded text-[11px] font-mono uppercase tracking-widest bg-red-500/15 border border-red-500/40 text-red-600 dark:text-red-400 hover:bg-red-500/25 transition-colors duration-150 cursor-pointer"
            >
              Sí
            </button>
            <button
              onClick={() => setDeleteState('idle')}
              className="px-3 py-1 rounded text-[11px] font-mono uppercase tracking-widest border border-stone-200 dark:border-white/10 text-stone-400 dark:text-white/30 hover:bg-stone-100 dark:hover:bg-white/5 transition-colors duration-150 cursor-pointer"
            >
              No
            </button>
          </div>
        )}

        {deleteState === 'loading' && (
          <span className="text-[11px] font-mono text-stone-400 dark:text-white/30 animate-pulse">
            borrando...
          </span>
        )}

        {deleteState === 'done' && (
          <span className="text-[11px] font-mono text-emerald-600 dark:text-emerald-400">
            ✓ borrado
          </span>
        )}

        {deleteState === 'error' && (
          <span className="text-[11px] font-mono text-red-500">
            error
          </span>
        )}
      </div>
    </div>
  )
}

export function AdminPage() {
  const [token, setToken] = useState<string>(() => sessionStorage.getItem(TOKEN_KEY) ?? '')
  const [input, setInput] = useState('')
  const [incidents, setIncidents] = useState<string[]>([])
  const [loading, setLoading] = useState(false)
  const [authError, setAuthError] = useState(false)

  const fetchIncidents = useCallback(async (t: string) => {
    setLoading(true)
    setAuthError(false)
    try {
      const res = await fetch(`/api/admin?t=${encodeURIComponent(t)}`)
      if (res.status === 401) {
        setAuthError(true)
        setToken('')
        sessionStorage.removeItem(TOKEN_KEY)
        return
      }
      if (res.ok) {
        const data = await res.json()
        setIncidents(data.incidents ?? [])
      }
    } catch { /* non-critical */ }
    finally { setLoading(false) }
  }, [])

  useEffect(() => {
    if (token) fetchIncidents(token)
  }, [token, fetchIncidents])

  const handleLogin = useCallback((e: React.FormEvent) => {
    e.preventDefault()
    const t = input.trim()
    if (!t) return
    sessionStorage.setItem(TOKEN_KEY, t)
    setToken(t)
    setInput('')
  }, [input])

  const handleDeleted = useCallback((date: string) => {
    setIncidents(prev => prev.filter(d => d !== date))
  }, [])

  // ── Not authenticated ─────────────────────────────────────────────────────
  if (!token) {
    return (
      <div className="min-h-screen bg-stone-100 dark:bg-[#0d0d1c] flex items-center justify-center px-4">
        <div className="w-full max-w-sm space-y-6">
          <div className="text-center space-y-1">
            <p className="font-mono text-xs uppercase tracking-[0.3em] text-stone-400 dark:text-white/20">
              admin
            </p>
            <p className="font-mono text-xs text-stone-400 dark:text-white/15">
              alesfosquescat
            </p>
          </div>

          {authError && (
            <p className="text-center text-xs font-mono text-red-500 dark:text-red-400">
              Token incorrecto
            </p>
          )}

          <form onSubmit={handleLogin} className="space-y-3">
            <input
              type="password"
              value={input}
              onChange={e => setInput(e.target.value)}
              placeholder="token"
              autoFocus
              className="w-full px-4 py-3 rounded-lg border border-stone-200 dark:border-white/10 bg-white dark:bg-white/5 text-stone-800 dark:text-white/80 font-mono text-sm placeholder-stone-300 dark:placeholder-white/20 focus:outline-none focus:border-amber-500/50 dark:focus:border-amber-400/40 transition-colors"
            />
            <button
              type="submit"
              className="w-full py-2.5 rounded-lg border border-amber-500/40 bg-amber-500/10 hover:bg-amber-500/20 text-amber-700 dark:text-amber-400 font-mono text-xs uppercase tracking-widest transition-colors duration-150 cursor-pointer"
            >
              Entrar
            </button>
          </form>
        </div>
      </div>
    )
  }

  // ── Authenticated ─────────────────────────────────────────────────────────
  return (
    <div className="min-h-screen bg-stone-100 dark:bg-[#0d0d1c] px-4 py-8">
      <div className="max-w-sm mx-auto space-y-6">

        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <p className="font-mono text-xs uppercase tracking-[0.3em] text-stone-400 dark:text-white/20">
              admin
            </p>
            <p className="font-mono text-sm font-bold text-stone-700 dark:text-white/60 mt-0.5">
              Incidencias
            </p>
          </div>
          <div className="flex items-center gap-3">
            <button
              onClick={() => fetchIncidents(token)}
              className="text-[10px] font-mono uppercase tracking-widest text-stone-400 dark:text-white/25 hover:text-amber-600 dark:hover:text-amber-400 transition-colors cursor-pointer"
            >
              Actualizar
            </button>
            <button
              onClick={() => { sessionStorage.removeItem(TOKEN_KEY); setToken('') }}
              className="text-[10px] font-mono uppercase tracking-widest text-stone-400 dark:text-white/25 hover:text-red-500 dark:hover:text-red-400 transition-colors cursor-pointer"
            >
              Salir
            </button>
          </div>
        </div>

        {/* Incident list */}
        {loading ? (
          <div className="space-y-2">
            {[...Array(5)].map((_, i) => (
              <div key={i} className="h-12 rounded-lg bg-stone-200 dark:bg-white/5 animate-pulse" />
            ))}
          </div>
        ) : incidents.length === 0 ? (
          <p className="text-center font-mono text-xs text-stone-400 dark:text-white/20 py-8">
            Sin incidencias registradas
          </p>
        ) : (
          <div className="space-y-2">
            {incidents.map(date => (
              <IncidentRow
                key={date}
                date={date}
                token={token}
                onDeleted={handleDeleted}
              />
            ))}
          </div>
        )}

        <p className="text-center font-mono text-[10px] text-stone-300 dark:text-white/10">
          {incidents.length} incidencia{incidents.length !== 1 ? 's' : ''}
        </p>
      </div>
    </div>
  )
}
