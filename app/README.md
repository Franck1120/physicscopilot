# PhysicsCopilot — App Mobile

![Flutter](https://img.shields.io/badge/Flutter-3.x-02569B?logo=flutter)
![Dart](https://img.shields.io/badge/Dart-3.4+-0175C2?logo=dart)
![Go](https://img.shields.io/badge/Backend-Go-00ADD8?logo=go)

## Panoramica

PhysicsCopilot è un'app mobile per Android e iOS che affianca tecnici e maker nella diagnosi e manutenzione di dispositivi fisici: stampanti 3D, elettronica, apparecchi industriali. L'utente punta la camera sul dispositivo e riceve guidance in tempo reale dall'AI (Gemini), con supporto vocale bidirezionale per operare a mani libere.

## Funzionalità

- **Camera live** — anteprima in tempo reale del dispositivo da diagnosticare
- **AI guidance in real-time** — comunicazione WebSocket con il backend Go + Gemini
- **Voice guidance** — lettura vocale delle istruzioni (flutter_tts) e input vocale (speech_to_text)
- **Cronologia sessioni** — storico locale delle sessioni di diagnosi consultabile offline
- **Selezione dispositivo** — catalogo attrezzature dal server con ricerca locale
- **Impostazioni** — configurazione URL del server, override runtime senza rebuild
- **Graceful degradation** — l'app funziona in sola lettura se il server non è raggiungibile

## Architettura

```
┌─────────────────────────────────────────────────────────────────┐
│  UI Layer          screens/ + widgets/                          │
│  (Flutter widgets, GoRouter navigation, design tokens)          │
├─────────────────────────────────────────────────────────────────┤
│  State Layer       providers/ (Riverpod)                        │
│  (session, camera, websocket, equipment, settings, history)     │
├─────────────────────────────────────────────────────────────────┤
│  Service Layer     services/                                    │
│  (ApiService via Dio, WebSocketService, CameraService,          │
│   VoiceService)                                                 │
├─────────────────────────────────────────────────────────────────┤
│  Data Layer        models/ + SharedPreferences                  │
│  (Device, Session, SessionStep, SessionRecord, prefs locali)    │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
                   Backend Go (REST + WebSocket)
                   ├── GET    /health
                   ├── POST   /api/sessions
                   ├── GET    /api/sessions
                   ├── GET    /api/sessions/:id
                   ├── DELETE /api/sessions/:id
                   └── GET    /ws  (WebSocket AI guidance)
```

## Setup e build

### Prerequisiti

- Flutter SDK 3.x
- Dart 3.4+
- Android SDK (per build APK) o Xcode 15+ (per build iOS)
- Backend Go in esecuzione (vedi `/server/README.md`)

### Installazione

```bash
# Clona il repository
git clone https://github.com/tuo-org/physicscopilot.git
cd physicscopilot/app

# Installa le dipendenze
flutter pub get
```

### Avvio in sviluppo

```bash
# Server locale sulla porta 8080
flutter run --dart-define=SERVER_URL=localhost:8080
```

### Build production

```bash
# Android APK
flutter build apk --dart-define=SERVER_URL=your.domain.com

# Android App Bundle (Play Store)
flutter build appbundle --dart-define=SERVER_URL=your.domain.com

# iOS
flutter build ios --dart-define=SERVER_URL=your.domain.com
```

## Configurazione server

L'URL del backend viene iniettato a compile-time tramite `--dart-define=SERVER_URL=...` e letto in `lib/utils/constants.dart`:

```dart
class AppConstants {
  static const serverUrl = String.fromEnvironment('SERVER_URL', defaultValue: 'localhost:8080');
}
```

In alternativa, l'utente può sovrascrivere l'URL a runtime dalla schermata **Impostazioni** dell'app, senza necessità di rebuild. Il valore viene persistito in SharedPreferences e ha precedenza sul valore compilato.

## Permessi richiesti

| Permesso | Piattaforma | Motivo |
|----------|-------------|--------|
| `CAMERA` | Android + iOS | Acquisizione live per diagnosi |
| `RECORD_AUDIO` | Android + iOS | Input vocale (speech_to_text) |
| `INTERNET` | Android | Comunicazione con backend |

I permessi vengono richiesti a runtime tramite `permission_handler`. Se i permessi camera o microfono vengono negati, l'app mostra una schermata dedicata con istruzioni per le impostazioni di sistema.

## Screenshot

_Screenshot placeholder — esegui `flutter run` e acquisisci schermate delle schermate: Home, Camera, Session (guidance attiva), History, Settings._

## Test

```bash
# Tutti i test unitari e widget
flutter test

# Con coverage
flutter test --coverage
genhtml coverage/lcov.info -o coverage/html
```

La struttura di `test/` rispecchia `lib/`: `test/providers/`, `test/services/`, `test/models/`.

## Note di sviluppo

### Design tokens

Definiti in `main.dart` come costanti globali:

```dart
const kAccent    = Color(0xFF10B981); // emerald green — brand color
const kBgPrimary = Color(0xFF0A0A0A); // background scuro
```

Font: **Poppins** caricato via `google_fonts`. Il tema estende `ThemeData.dark()` con override dell'accent color e della tipografia.

### Error handling

- `ApiService` (Dio) — timeout di 10s, retry automatico su 503, errori mappati in eccezioni tipizzate
- `WebSocketService` — reconnect automatico con backoff esponenziale (1s → 2s → 4s → max 30s)
- `CameraService` — fallback graceful se la camera non è disponibile (simulatore, permesso negato)
- Tutti i provider Riverpod espongono `AsyncValue` — gli stati di loading/error sono gestiti uniformemente nelle UI senza try/catch nei widget

### Graceful degradation

L'app avvia in modalità offline se il server non risponde all'health check (`GET /health`). In questa modalità la cronologia locale rimane consultabile, mentre camera e AI guidance sono disabilitate con messaggio esplicativo. La connessione viene ritentata automaticamente ogni 30 secondi in background.
