import { useLanguage } from '../context/LanguageContext'

export function LegalPage() {
  const { t } = useLanguage()

  return (
    <div className="min-h-screen bg-stone-100 dark:bg-[#0d0d1c] flex flex-col transition-colors duration-200">

      <header className="border-b border-stone-200 dark:border-white/8 pt-safe">
        <div className="max-w-lg mx-auto px-4 py-3 flex items-center gap-3">
          <div className="w-[3px] self-stretch bg-amber-600/40 dark:bg-amber-400/40 rounded-full shrink-0" />
          <div className="flex-1 min-w-0">
            <a href="/" className="text-amber-700/70 dark:text-amber-400/60 font-bold text-[15px] leading-tight hover:text-amber-700 dark:hover:text-amber-400 transition-colors">
              alesfosquescat
            </a>
            <p className="text-stone-400 dark:text-white/20 text-[11px] leading-tight mt-0.5">Santa Eulàlia de Ronçana</p>
          </div>
        </div>
      </header>

      <main className="flex-1 max-w-lg mx-auto w-full px-4 py-8">
        <p className="section-label mb-6">{t.legalTitle}</p>

        <div className="space-y-5 text-stone-600 dark:text-white/50 text-[13px] leading-relaxed">

          <p>
            {t.legalP1Before}
            <span className="text-amber-700 dark:text-amber-400/70">{t.legalSatire}</span>
            {t.legalP1After}
          </p>

          <p>{t.legalP2}</p>

          <p>{t.legalP3}</p>

          <p>{t.legalP4}</p>

        </div>

        <div className="mt-10">
          <a
            href="/"
            className="px-5 py-2.5 rounded border border-amber-600/30 dark:border-amber-500/30 bg-amber-500/8 text-amber-700 dark:text-amber-400 text-sm font-medium hover:bg-amber-500/15 dark:hover:bg-amber-500/12 transition-colors"
          >
            {t.backHome}
          </a>
        </div>
      </main>

      <footer className="py-5 border-t border-stone-200 dark:border-white/6 text-center">
        <p className="text-stone-400 dark:text-white/15 text-[10px] tracking-widest">
          alesfosquescat · Santa Eulàlia de Ronçana · {new Date().getFullYear()}
        </p>
      </footer>
    </div>
  )
}
