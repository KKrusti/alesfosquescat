import { useLanguage } from '../context/LanguageContext'

export function ErrorPage() {
  const { t } = useLanguage()

  return (
    <div className="min-h-screen bg-stone-100 dark:bg-[#0d0d1c] flex flex-col transition-colors duration-200">

      <header className="border-b border-stone-200 dark:border-white/8 pt-safe">
        <div className="max-w-lg mx-auto px-4 py-3 flex items-center gap-3">
          <div className="w-[3px] self-stretch bg-amber-600/40 dark:bg-amber-400/40 rounded-full shrink-0" />
          <div>
            <p className="text-amber-700/60 dark:text-amber-400/60 font-bold text-[15px] leading-tight">alesfosquescat</p>
            <p className="text-stone-400 dark:text-white/20 text-[11px] leading-tight mt-0.5">Santa Eulàlia de Ronçana</p>
          </div>
        </div>
      </header>

      <main className="flex-1 flex flex-col items-center justify-center px-4 text-center max-w-lg mx-auto w-full">

        <div
          className="text-8xl font-black text-amber-700/20 dark:text-amber-400/20 leading-none mb-4"
          style={{ fontFamily: 'Anton, sans-serif' }}
        >
          404
        </div>

        <img
          src="/signal.png"
          alt=""
          aria-hidden="true"
          className="w-40 opacity-15 grayscale mb-6"
        />

        <p className="text-stone-700 dark:text-white/60 text-lg font-semibold mb-1">
          {t.error404}
        </p>
        <p className="text-amber-600 dark:text-amber-400/50 text-base mb-8">
          {t.errorSub}
        </p>

        <a
          href="/"
          className="px-5 py-2.5 rounded border border-amber-600/30 dark:border-amber-500/30 bg-amber-500/8 text-amber-700 dark:text-amber-400 text-sm font-medium hover:bg-amber-500/15 dark:hover:bg-amber-500/12 transition-colors"
        >
          {t.backHome}
        </a>
      </main>
    </div>
  )
}
