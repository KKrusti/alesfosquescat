export function ErrorPage() {
  return (
    <div className="min-h-screen bg-[#0d0d1c] flex flex-col">

      {/* Same header as main app */}
      <header className="border-b border-white/8 pt-safe">
        <div className="max-w-lg mx-auto px-4 py-3 flex items-center gap-3">
          <div className="w-[3px] self-stretch bg-amber-400/40 rounded-full shrink-0" />
          <div>
            <p className="text-amber-400/60 font-bold text-[15px] leading-tight">alesfosquescat</p>
            <p className="text-white/20 text-[11px] leading-tight mt-0.5">Santa Eulàlia de Ronçana</p>
          </div>
        </div>
      </header>

      {/* 404 content */}
      <main className="flex-1 flex flex-col items-center justify-center px-4 text-center max-w-lg mx-auto w-full">

        {/* Error code */}
        <div
          className="text-8xl font-black text-amber-400/20 leading-none mb-4"
          style={{ fontFamily: 'Anton, sans-serif' }}
        >
          404
        </div>

        {/* Dim signal image */}
        <img
          src="/signal.png"
          alt=""
          aria-hidden="true"
          className="w-40 opacity-15 grayscale mb-6"
        />

        {/* Message */}
        <p className="text-white/60 text-lg font-semibold mb-1">
          Aquesta pàgina no existeix...
        </p>
        <p className="text-amber-400/50 text-base mb-8">
          ...com la llum
        </p>

        <a
          href="/"
          className="px-5 py-2.5 rounded border border-amber-500/30 bg-amber-500/8 text-amber-400 text-sm font-medium hover:bg-amber-500/15 transition-colors"
        >
          ← Tornar a l&apos;inici
        </a>
      </main>
    </div>
  )
}
