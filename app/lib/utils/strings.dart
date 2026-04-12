/// Central repository of every user-visible string in PhysicsCopilot.
///
/// Keeping strings here makes future localization a one-step migration:
/// replace references to [AppStrings] with the generated
/// `AppLocalizations` class produced by Flutter's gen-l10n tooling.
///
/// Convention:
///   - Group constants by screen / feature.
///   - Use `$variable` interpolation at the call site, not here.
///   - Do NOT put format logic in this file.
abstract final class AppStrings {
  // ── App ─────────────────────────────────────────────────────────────────────
  static const appName = 'PhysicsCopilot';
  static const appTagline = 'Il tuo assistente AI per le riparazioni';
  static const appVersion = 'v1.0.0';

  // ── Common ───────────────────────────────────────────────────────────────────
  static const save = 'Salva';
  static const cancel = 'Annulla';
  static const delete = 'Elimina';
  static const close = 'Chiudi';
  static const reset = 'Reset';
  static const retry = 'Riprova';
  static const errorGeneric = 'Qualcosa è andato storto';
  static const restartApp = 'Riavvia l\'app per riprendere.';

  // ── Session screen ───────────────────────────────────────────────────────────
  static const sessionTitle = 'Sessione attiva';
  static const sessionEndSession = 'Termina sessione';
  static const sessionNewAnalysis = 'Nuova analisi';
  static const sessionIdle =
      'Punta la camera sull\'oggetto\nper avviare l\'analisi AI.';
  static const sessionAiThinking = 'L\'AI sta analizzando…';
  static const sessionResponseCopied = 'Risposta copiata negli appunti';
  static const sessionCopyResponse = 'Copia risposta';
  static const sessionServerUnreachable =
      'Server non raggiungibile — attendi la riconnessione.';
  static const sessionFrameError = 'Impossibile acquisire il frame';
  static const sessionAiTimeout = 'Nessuna risposta dall\'AI. Riprova.';
  static const sessionErrorFromServer = 'Errore dal server.';
  static const sessionInvalidResponse = 'Risposta non valida dal server.';
  static const sessionErrorUnknown = 'Errore sconosciuto';
  static const sessionDescribeProblem = 'Descrivi il problema…';
  static const sessionSend = 'Invia messaggio';
  static const sessionNoResponseYet = 'Nessuna risposta ancora.';
  static const sessionCameraInit = 'Inizializzazione camera…';
  static const sessionCameraUnavailable = 'Camera non disponibile';

  // ── Tutorial ─────────────────────────────────────────────────────────────────
  static const tutorialHint = 'Tocca qui per analizzare';
  static const tutorialDismiss = 'Tocca ovunque per chiudere';

  // ── Home ─────────────────────────────────────────────────────────────────────
  static const homeNewSession = 'Nuova sessione';
  static const homeNewSessionSub = 'Punta la camera e avvia l\'analisi AI';
  static const homeSectionDevice = 'DISPOSITIVO ATTIVO';
  static const homeSectionSessions = 'ULTIME SESSIONI';
  static const homeSeeAll = 'Vedi tutte';
  static const homeNoDevice = 'Nessun dispositivo selezionato';
  static const homeChangeDevice = 'Cambia';
  static const homeChangeDeviceLabel = 'Cambia dispositivo attivo';
  static const homeEmptyTitle = 'Pronto per iniziare?';
  static const homeEmptySubtitle =
      'Avvia la tua prima sessione\ne lascia che l\'AI ti guidi nella diagnosi.';
  static const homeServerReachable = 'Server raggiungibile';
  static const homeServerUnreachable = 'Server non raggiungibile';
  static const homeServerChecking = 'Verifica server…';

  // ── History ──────────────────────────────────────────────────────────────────
  static const historyTitle = 'Sessioni';
  static const historyClearAll = 'Cancella tutto';
  static const historyClearConfirm =
      'Eliminare tutta la cronologia delle sessioni?';
  static const historyEmpty = 'Nessuna sessione';
  static const historyEmptySub = 'Le sessioni completate appariranno qui.';
  static const historyStatusResolved = 'Risolto';
  static const historyStatusUnresolved = 'Non risolto';
  static const historySectionAI = 'Analisi AI';

  // ── Settings ─────────────────────────────────────────────────────────────────
  static const settingsTitle = 'Impostazioni';
  static const settingsSectionConnection = 'CONNESSIONE';
  static const settingsSectionFeatures = 'FUNZIONALITÀ';
  static const settingsSectionServerInfo = 'CONNESSIONE SERVER';
  static const settingsSectionAbout = 'INFORMAZIONI APP';
  static const settingsServerUrl = 'URL Server';
  static const settingsServerUrlHint = 'es. wss://your-tunnel.trycloudflare.com';
  static const settingsServerUrlLabel =
      'URL del server, lascia vuoto per usare il default';
  static const settingsVoice = 'Guida vocale';
  static const settingsVoiceSub = 'Legge le istruzioni AI ad alta voce.';
  static const settingsUrlSaved =
      'URL server aggiornato — riavvia la sessione.';
  static const settingsUrlReset = 'URL ripristinato al valore di default.';

  // ── About ────────────────────────────────────────────────────────────────────
  static const aboutPoweredBy = 'Powered by';
  static const aboutEngine = 'Google Gemini AI';
  static const aboutPrivacy = 'Privacy Policy';
  static const aboutTerms = 'Termini di Servizio';
  static const aboutDetails = 'Dettagli e licenze';

  // ── Camera / connection labels ────────────────────────────────────────────────
  static const connConnected = 'Connesso';
  static const connConnecting = 'Connessione…';
  static const connDisconnected = 'Non connesso';
  static const connError = 'Errore';
  static const connBannerConnecting = 'Connessione al server in corso…';
  static const connBannerUnreachable =
      'Server non raggiungibile — i dati non vengono inviati';
  static const cameraError = 'Errore camera';
  static const cameraOfflineBanner = 'Offline — Riconnessione in corso…';
  static const cameraWriteHere = 'Scrivi una domanda…';

  // ── Onboarding ───────────────────────────────────────────────────────────────
  static const onboardingNext = 'Avanti';
  static const onboardingStart = 'Inizia';
  static const onboardingSlide1Title = 'Punta la camera';
  static const onboardingSlide2Title = 'L\'AI analizza';

  // ── Profile ──────────────────────────────────────────────────────────────────
  static const profileTitle = 'Profilo';
  static const profileSettings = 'Impostazioni';
  static const profilePrivacy = 'Privacy';
  static const profileAbout = 'Informazioni app';
}
