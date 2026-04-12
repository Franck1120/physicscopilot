# PhysicsCopilot — Flutter App

Flutter mobile app (Android & iOS) that gives technicians real-time AI guidance while they work on physical devices.

## What it does

Point your phone camera at a device — a 3D printer, a circuit board, an appliance — and describe the problem. The app streams live frames to the backend over WebSocket. The server runs them through Gemini 1.5 Vision and returns step-by-step repair instructions, which the app reads aloud so you can work hands-free.

## Architecture

```
CameraService → WebSocketService → Go Backend → Gemini 1.5
      ↑                                               ↓
  preview UI          ←──── AI guidance text ─────────┘
                      ←──── AR overlay data ──────────
                      ←──── voice TTS ────────────────
```

State management: Riverpod 2.x. Navigation: go_router. Full architecture in [`docs/ARCHITECTURE.md`](../docs/ARCHITECTURE.md).

## Prerequisites

- Flutter 3.x (`flutter --version`)
- Android Studio / Xcode for device targets
- A running backend (see `server/` or use the hosted instance)

## Setup

```bash
# Install dependencies
flutter pub get

# Run on a connected device (not emulator — needs real camera)
flutter run --dart-define=SERVER_URL=https://your-server.onrender.com

# Or set the URL at runtime via the app Settings screen
```

## Configuration

All runtime config lives in `lib/utils/constants.dart` and is overridable via:
- **Dart defines** at build time: `--dart-define=SERVER_URL=...`
- **Settings screen** at runtime (stored in SharedPreferences)

| Setting | Default | Description |
|---------|---------|-------------|
| `SERVER_URL` | `http://10.0.2.2:8080` (emulator) | Backend base URL |
| `LANGUAGE` | `it` | BCP-47 language code for AI responses |
| Voice guidance | off | Reads AI responses aloud (flutter_tts) |

## Key files

| File | What it does |
|------|-------------|
| `lib/screens/session_screen.dart` | Main session UI (camera + guidance panel) |
| `lib/screens/camera_screen.dart` | Full-screen AR mode with voice I/O |
| `lib/services/websocket_service.dart` | WebSocket client with exponential reconnect |
| `lib/services/camera_service.dart` | Camera init, JPEG capture, quality analysis |
| `lib/services/voice_service.dart` | TTS + STT bridge |
| `lib/providers/session_provider.dart` | AI session state (Riverpod) |
| `lib/providers/websocket_provider.dart` | WebSocket singleton + reconnect on URL change |
| `lib/utils/strings.dart` | All UI strings (AppStrings) — single source of truth |

## Tests

```bash
cd app
flutter test              # all unit + widget tests
flutter test --coverage   # with coverage report
```

Tests live in `test/`. Key test files:
- `test/services/websocket_service_test.dart` — connection lifecycle, reconnect, back-off
- `test/widgets/session_screen_test.dart` — screen rendering

## Build

```bash
# Android release APK
flutter build apk --release

# iOS (requires macOS + Xcode)
flutter build ios --release
```

See [`../CONTRIBUTING.md`](../CONTRIBUTING.md) for the full release checklist.
