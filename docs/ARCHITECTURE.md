# PhysicsCopilot — Architecture

## Overview

PhysicsCopilot is an AI-assisted diagnostics tool for physical devices (3D printers, electronics, industrial equipment). A technician points their phone camera at a device and receives real-time guidance from Gemini.

---

## System Diagram

```
┌──────────────────────────────────────────────────────┐
│                  Flutter App (Mobile)                │
│                                                      │
│  CameraService ──frames──►  WebSocketService         │
│                              │     ▲                 │
│  Riverpod Providers ◄────────┘     │ messages        │
│  (session, camera, ws, voice)      │                 │
│                              ▼     │                 │
│  UI: SessionScreen / CameraScreen  │                 │
│       GuidancePanel / ArOverlay    │                 │
└────────────────────┬───────────────┘                 │
                     │ WSS /ws?token=<jwt>              │
                     ▼                                 │
┌──────────────────────────────────────────────────────┐
│                   Go Backend                         │
│                                                      │
│  HTTP Router (gorilla/mux)                           │
│  ├── Middleware: JWT auth, rate limiting, CORS       │
│  ├── GET  /health                                    │
│  ├── GET  /ws        ◄── WebSocket handler           │
│  ├── POST /api/sessions                             │
│  ├── GET  /api/sessions                             │
│  └── GET/DELETE /api/sessions/:id                   │
│                                                      │
│  WebSocket Handler                                   │
│  ├── Receives frames + text queries from client      │
│  ├── Calls Gemini 1.5 Vision API                    │
│  ├── Streams response chunks back (type: chunk)      │
│  └── Sends final response (type: response)           │
│                                                      │
│  Supabase Client (Postgres + Auth)                  │
└──────────────────────────────────────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────────────────┐
│              Supabase (Postgres)                     │
│  Tables: sessions, session_steps, devices, users     │
│  Auth: JWT issuer, Row-Level Security               │
└──────────────────────────────────────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────────────────┐
│              Google Gemini 1.5 Vision                │
│  Input: base64 frame + text prompt                   │
│  Output: JSON {text, voice_text, overlay, steps}     │
└──────────────────────────────────────────────────────┘
```

---

## Data Flow: Camera Frame → AI Response

```
1. CameraService captures frame (Uint8List, JPEG)
2. frame stream → WebSocketService.sendFrame()
3. Client sends: {"type":"frame","data":"<base64>","timestamp":1234567890}
4. Go handler receives frame, decodes base64
5. Gemini 1.5 Vision API call (with system prompt + frame)
6. Gemini returns structured JSON
7. Server sends streaming chunks: {"type":"chunk","text":"..."}  (0..N)
8. Server sends final: {"type":"response","text":"...","voice_text":"...",
                        "overlay":{...},"steps":[...]}
9. Flutter: SessionNotifier.updateFromResponse() updates Riverpod state
10. GuidancePanel re-renders with new text
11. VoiceService.speak() reads voice_text aloud (flutter_tts)
```

---

## Flutter App — Component Map

| Component | File | Responsibility |
|-----------|------|----------------|
| `WebSocketService` | `services/websocket_service.dart` | Manages WSS connection, exponential reconnect, message decode |
| `CameraService` | `services/camera_service.dart` | Camera init, frame capture, quality analysis, FPS stream |
| `VoiceService` | `services/voice_service.dart` | TTS (flutter_tts) + STT (speech_to_text) |
| `ApiService` | `services/api_service.dart` | REST calls to Go server (Dio) |
| `sessionProvider` | `providers/session_provider.dart` | AI response state: text, isProcessing, error, streaming |
| `webSocketServiceProvider` | `providers/websocket_provider.dart` | Singleton WebSocketService, reconnects on settings change |
| `cameraServiceProvider` | `providers/camera_provider.dart` | Singleton CameraService |
| `settingsProvider` | `providers/settings_provider.dart` | Server URL, language, voice toggle (SharedPreferences) |
| `SessionScreen` | `screens/session_screen.dart` | Main session UI: camera + guidance panel, session timer |
| `CameraScreen` | `screens/camera_screen.dart` | Full-screen AR mode: overlays + voice I/O + step progress |

---

## Go Server — Component Map

| Package | Responsibility |
|---------|----------------|
| `internal/handlers` | HTTP/WS request handlers |
| `internal/services` | Gemini API client, AI prompt construction |
| `internal/middleware` | JWT validation, rate limiting, CORS, request logging |
| `internal/models` | Session, Device, User data structures |
| `internal/db` | Supabase SQL queries |
| `cmd/server` | main(), router setup, server startup |

---

## Database Schema Overview

```sql
-- User sessions (one per repair attempt)
sessions (
  id          UUID PRIMARY KEY,
  user_id     UUID REFERENCES auth.users,
  device_brand TEXT,
  device_model TEXT,
  created_at  TIMESTAMPTZ,
  ended_at    TIMESTAMPTZ
)

-- Step-by-step procedure for a session
session_steps (
  id          UUID PRIMARY KEY,
  session_id  UUID REFERENCES sessions,
  step_number INT,
  title       TEXT,
  description TEXT,
  completed   BOOLEAN DEFAULT FALSE
)

-- Device catalog
devices (
  id           UUID PRIMARY KEY,
  manufacturer TEXT,
  model        TEXT,
  category     TEXT,
  manual_url   TEXT
)
```

RLS is enabled on all tables. Users can only access rows where `user_id = auth.uid()`.

---

## State Management (Riverpod)

The app uses **Riverpod 2.x** with:
- `StateNotifierProvider` for complex state with methods (`sessionProvider`, `voiceProvider`, `stepProvider`)
- `StreamProvider` for async streams (`connectionStatusProvider`, `frameQualityProvider`)
- `FutureProvider` for async initialization (`cameraInitProvider`)
- `StateProvider` for simple values (`settingsProvider`, `cachedResponseProvider`, `showTutorialProvider`, `lastFrameProvider`)
- `Provider` for singletons (`cameraServiceProvider`, `webSocketServiceProvider`)

---

## Deployment

- **Server**: Docker container on Render.com (`render.yaml`)
- **DB**: Supabase (managed Postgres)
- **App**: Flutter build → Play Store / App Store
- **Secrets**: `SUPABASE_JWT_SECRET`, `GEMINI_API_KEY`, `SUPABASE_URL`, `SUPABASE_SERVICE_KEY` via Render environment variables
