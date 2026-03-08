import { createContext, useContext, useState } from 'react'
import { translations, type Lang } from '../i18n/translations'

type Translations = Record<string, string>

interface LanguageContextValue {
  lang: Lang
  t: Translations
  toggle: () => void
}

const LanguageContext = createContext<LanguageContextValue>({
  lang: 'ca',
  t: translations.ca,
  toggle: () => {},
})

export function LanguageProvider({ children }: { children: React.ReactNode }) {
  const [lang, setLang] = useState<Lang>(() => {
    const stored = localStorage.getItem('afc_lang')
    const browserLang = navigator.language.toLowerCase()
    const initial: Lang = stored === 'es' || stored === 'ca'
      ? (stored as Lang)
      : browserLang.startsWith('ca') ? 'ca' : 'es'
    document.documentElement.lang = initial
    return initial
  })

  const toggle = () => {
    setLang(prev => {
      const next: Lang = prev === 'ca' ? 'es' : 'ca'
      localStorage.setItem('afc_lang', next)
      document.documentElement.lang = next
      return next
    })
  }

  return (
    <LanguageContext.Provider value={{ lang, t: translations[lang], toggle }}>
      {children}
    </LanguageContext.Provider>
  )
}

export const useLanguage = () => useContext(LanguageContext)
