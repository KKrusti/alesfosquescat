export function ErrorPage() {
  return (
    <div className="min-h-screen bg-[#050510] flex flex-col items-center justify-center px-4 text-white">
      {/* Dimmed, broken bat signal */}
      <svg
        viewBox="0 0 300 500"
        className="w-40 opacity-15 mb-8 grayscale"
        aria-hidden="true"
      >
        <defs>
          <linearGradient id="offBeam" x1="0.5" y1="1" x2="0.5" y2="0">
            <stop offset="0%"   stopColor="#555" stopOpacity="0.5" />
            <stop offset="100%" stopColor="#333" stopOpacity="0" />
          </linearGradient>
        </defs>
        <polygon points="150,440 42,12 258,12" fill="url(#offBeam)" />
        <text
          x="150" y="300"
          textAnchor="middle"
          fontSize="168"
          fontWeight="900"
          fontFamily="'Anton', Impact, sans-serif"
          fill="#333"
          style={{ letterSpacing: '-4px' }}
        >
          A
        </text>
        <rect x="112" y="444" width="76" height="26" rx="7" fill="#111" stroke="#222" strokeWidth="1.5" />
        <ellipse cx="150" cy="458" rx="40" ry="17" fill="#111" stroke="#222" strokeWidth="2" />
        <ellipse cx="150" cy="458" rx="28" ry="12" fill="#0a0a0a" />
        <ellipse cx="150" cy="458" rx="18" ry="7"  fill="#181818" />
        <rect x="128" y="469" width="44" height="9" rx="3" fill="#0a0a0a" />
      </svg>

      {/* 404 number */}
      <h1
        className="text-8xl sm:text-9xl font-black text-signal-600/25 leading-none mb-4"
        style={{ fontFamily: 'Anton, sans-serif' }}
      >
        404
      </h1>

      {/* Messages */}
      <p className="text-2xl sm:text-3xl font-bold text-white/60 text-center mb-2">
        Aquesta pàgina no existeix...
      </p>
      <p
        className="text-xl sm:text-2xl text-signal-500/50 text-center mb-10"
        style={{ fontFamily: 'Anton, sans-serif' }}
      >
        ...com la llum
      </p>

      {/* Back link */}
      <a
        href="/"
        className={[
          'px-6 py-3 rounded border border-signal-600/40 bg-signal-500/8',
          'font-mono text-signal-400 text-sm uppercase tracking-widest',
          'hover:bg-signal-500/15 hover:border-signal-500/60',
          'transition-colors duration-200 cursor-pointer',
        ].join(' ')}
      >
        ← Tornar a l&apos;inici
      </a>
    </div>
  )
}
