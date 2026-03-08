export function LegalPage() {
  return (
    <div className="min-h-screen bg-[#0d0d1c] flex flex-col">

      <header className="border-b border-white/8 pt-safe">
        <div className="max-w-lg mx-auto px-4 py-3 flex items-center gap-3">
          <div className="w-[3px] self-stretch bg-amber-400/40 rounded-full shrink-0" />
          <div className="flex-1 min-w-0">
            <a href="/" className="text-amber-400/60 font-bold text-[15px] leading-tight hover:text-amber-400 transition-colors">
              alesfosquestcat
            </a>
            <p className="text-white/20 text-[11px] leading-tight mt-0.5">Santa Eulàlia de Ronçana</p>
          </div>
        </div>
      </header>

      <main className="flex-1 max-w-lg mx-auto w-full px-4 py-8">
        <p className="section-label mb-6">Avís legal</p>

        <div className="space-y-5 text-white/50 text-[13px] leading-relaxed">

          <p>
            Aquesta web és una obra de <span className="text-amber-400/70">sàtira i humor</span> creada per un veí
            del municipi sense gaires llums. No té caràcter oficial ni informatiu.
          </p>

          <p>
            Les dades mostrades (dies sense llum, incidències) són aportacions anònimes d'usuaris
            i no han estat verificades per cap organisme oficial.
          </p>

          <p>
            Aquesta web no té ànim de lucre ni pretén difamar cap persona física o jurídica.
            Qualsevol semblança amb plans d'emergència oficials és purament humorística.
          </p>

          <p>
            El titular no es fa responsable de l'ús que tercers facin del contingut d'aquesta web.
          </p>

        </div>

        <div className="mt-10">
          <a
            href="/"
            className="px-5 py-2.5 rounded border border-amber-500/30 bg-amber-500/8 text-amber-400 text-sm font-medium hover:bg-amber-500/15 transition-colors"
          >
            ← Tornar a l&apos;inici
          </a>
        </div>
      </main>

      <footer className="py-5 border-t border-white/6 text-center">
        <p className="text-white/15 text-[10px] tracking-widest">
          alesfosquestcat · Santa Eulàlia de Ronçana · {new Date().getFullYear()}
        </p>
      </footer>
    </div>
  )
}
