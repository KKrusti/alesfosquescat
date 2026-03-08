export type Lang = 'ca' | 'es'

export const translations = {
  ca: {
    // Header
    live:        'en directe',
    toggleDark:  'Canviar a mode fosc',
    toggleLight: 'Canviar a mode clar',

    // Section labels
    sectionStatus: 'Estat actual',
    sectionReport: 'Reportar incidència',
    sectionStats:  'Dades anuals',

    // Status cards
    streakLabel:      'nits consecutives sense llum',
    daysSinceLabel:   "nits des de l'últim incident",
    badgeActive:      'Ratxa actual',
    badgeOk:          'Tot bé',
    longestLabel:     'ratxa més llarga sense llum',
    recordBadge:      'Rècord',

    // Stats rows
    statNights:       'Nits sense llum',
    statLongest:      'Ratxa màx. sense llum',
    statNormal:       'Nits amb normalitat',
    statLastIncident: 'Darrer incident',

    // BatSignal — report button
    reportAlt:       'Focus projectant la A — toca per reportar una apagada',
    reportAriaLabel: 'Toca per reportar una apagada',
    sending:         'enviant...',
    reportLabel:     '— reportar —',

    // BatSignal — resolve button
    resolveAlt:       'Bombeta encesa — toca per marcar que la llum ha tornat',
    resolveAriaLabel: 'Toca per marcar que la llum ha tornat',
    resolveSending:   'enviant...',
    resolveLabel:     '— resoldre —',
    resolvedLabel:    '✓ resolt',

    // BatSignal — feedback messages
    msgAlreadyActive:  'La ratxa ja està activa, gràcies',
    msgRestored:       'Reprenent la ratxa anterior',
    msgRateLimit:      'Massa peticions, espera un minut',
    msgServerError:    "El servidor també s'ha apagat",
    msgNetworkError:   "S'ha anat la llum... irònic",
    msgTimeout:        "Timeout — s'ha anat la llum... irònic",
    msgResolveError:   'Error al marcar la resolució',
    msgAlreadyResolved:"L'alcalde ja sap que és un miracle, no cal insistir",

    // Footer / navigation
    legalLink:          'Avís legal',
    instagramAriaLabel: "Instagram de l'Ajuntament de Santa Eulàlia de Ronçana",

    // 404 page
    error404:  'Aquesta pàgina no existeix...',
    errorSub:  '...com la llum',
    backHome:  "← Tornar a l'inici",

    // History page
    historyTitle:      'Historial d\'incidències',
    historyEmpty:      'Sense incidències registrades aquest any',
    historyDaySingle:  'nit',
    historyDayPlural:  'nits',
    historyLinkLabel:  'Mur de la vergonya →',
    historyBack:       '← Tornar',

    // Legal page
    legalTitle:    'Avís legal',
    legalSatire:   'sàtira i humor',
    legalP1Before: 'Aquesta web és una obra de ',
    legalP1After:  ' creada per un veí del municipi sense gaires llums. No té caràcter oficial ni informatiu.',
    legalP2:       "Les dades mostrades (dies sense llum, incidències) són aportacions anònimes d'usuaris i no han estat verificades per cap organisme oficial.",
    legalP3:       "Aquesta web no té ànim de lucre ni pretén difamar cap persona física o jurídica. Qualsevol semblança amb plans d'emergència oficials és purament humorística.",
    legalP4:       "El titular no es fa responsable de l'ús que tercers facin del contingut d'aquesta web.",
  },

  es: {
    // Header
    live:        'en directo',
    toggleDark:  'Cambiar a modo oscuro',
    toggleLight: 'Cambiar a modo claro',

    // Section labels
    sectionStatus: 'Estado actual',
    sectionReport: 'Reportar incidencia',
    sectionStats:  'Datos anuales',

    // Status cards
    streakLabel:    'noches consecutivas sin luz',
    daysSinceLabel: 'noches desde el último incidente',
    badgeActive:    'Racha actual',
    badgeOk:        'Todo bien',
    longestLabel:   'racha más larga sin luz',
    recordBadge:    'Récord',

    // Stats rows
    statNights:       'Noches sin luz',
    statLongest:      'Racha máx. sin luz',
    statNormal:       'Noches con normalidad',
    statLastIncident: 'Último incidente',

    // BatSignal — report button
    reportAlt:       'Foco proyectando la A — toca para reportar un apagón',
    reportAriaLabel: 'Toca para reportar un apagón',
    sending:         'enviando...',
    reportLabel:     '— reportar —',

    // BatSignal — resolve button
    resolveAlt:       'Bombilla encendida — toca para marcar que ha vuelto la luz',
    resolveAriaLabel: 'Toca para marcar que ha vuelto la luz',
    resolveSending:   'enviando...',
    resolveLabel:     '— resolver —',
    resolvedLabel:    '✓ resuelto',

    // BatSignal — feedback messages
    msgAlreadyActive:  'La racha ya está activa, gracias',
    msgRestored:       'Retomando la racha anterior',
    msgRateLimit:      'Demasiadas peticiones, espera un minuto',
    msgServerError:    'El servidor también se ha apagado',
    msgNetworkError:   'Se ha ido la luz... irónico',
    msgTimeout:        'Timeout — se ha ido la luz... irónico',
    msgResolveError:   'Error al marcar la resolución',
    msgAlreadyResolved:'El alcalde ya sabe que es un milagro, no hace falta insistir',

    // Footer / navigation
    legalLink:          'Aviso legal',
    instagramAriaLabel: 'Instagram del Ayuntamiento de Santa Eulàlia de Ronçana',

    // 404 page
    error404: 'Esta página no existe...',
    errorSub: '...como la luz',
    backHome: '← Volver al inicio',

    // History page
    historyTitle:      'Historial de incidencias',
    historyEmpty:      'Sin incidencias registradas este año',
    historyDaySingle:  'noche',
    historyDayPlural:  'noches',
    historyLinkLabel:  'Muro de la vergüenza →',
    historyBack:       '← Volver',

    // Legal page
    legalTitle:    'Aviso legal',
    legalSatire:   'sátira y humor',
    legalP1Before: 'Esta web es una obra de ',
    legalP1After:  ' creada por un vecino del municipio sin demasiadas luces. No tiene carácter oficial ni informativo.',
    legalP2:       'Los datos mostrados (días sin luz, incidencias) son aportaciones anónimas de usuarios y no han sido verificados por ningún organismo oficial.',
    legalP3:       'Esta web no tiene ánimo de lucro ni pretende difamar a ninguna persona física o jurídica. Cualquier semejanza con planes de emergencia oficiales es puramente humorística.',
    legalP4:       'El titular no se hace responsable del uso que terceros hagan del contenido de esta web.',
  },
}

export type Translations = typeof translations['ca']
