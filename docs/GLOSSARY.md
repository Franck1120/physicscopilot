# Glossary

Technical terms and acronyms used throughout the codebase and documentation.

---

| Term | Definition |
|------|-----------|
| **AI Response** | The structured JSON payload returned by the server after analyzing a frame: `{text, voice_text, overlay, steps}` |
| **AR Overlay** | Augmented reality layer drawn on top of the camera preview. Renders step markers, warning regions, and component labels from the server's `overlay` response field |
| **Back-off (exponential)** | Reconnect strategy where each failure doubles the wait time: 1s, 2s, 4s, 8s... capped at 30s |
| **BCP-47** | IETF language tag format used for locale codes: `it` (Italian), `en` (English), `en-US`, etc. |
| **Camera frame** | A single JPEG-compressed image captured from the device camera, encoded as base64 before transmission |
| **Chunk message** | Server-to-client WebSocket message of type `chunk` containing a partial AI response text. Used for streaming — multiple chunks arrive before the final `response` message |
| **Connection status** | Enum with three states: `disconnected`, `connecting`, `connected`. Reflected in the UI connection banner |
| **ConsumerStatefulWidget** | Flutter widget that has both mutable local state (`State`) and access to Riverpod providers (`WidgetRef`) |
| **CQRS** | Command Query Responsibility Segregation — the project separates read (GET) and write (POST/DELETE) operations in handlers |
| **Dart define** | Build-time constant passed to Flutter via `--dart-define=KEY=VALUE`. Used for `SERVER_URL` and `LANGUAGE` |
| **Domain** | A category of repairable devices (e.g. `hvac`, `drone`, `iot`) used to scope KB queries to a per-domain `VectorStore` |
| **Equipment** | A device selected by the user from the catalog before starting a session. Provides context to the AI |
| **Frame deduplication** | Perceptual hashing of frames to skip sending near-identical images to Gemini. Implemented in `services/phash.go` |
| **Gemini** | Google's multimodal AI model used for visual analysis. PhysicsCopilot uses Gemini 1.5 Flash |
| **GoDoc** | Go's documentation convention: comments immediately preceding a function/type, starting with the identifier name |
| **LRU Cache** | Least Recently Used eviction cache used for RAG query results (128 entries, 5 min TTL). Prevents redundant TF-IDF scoring on repeated queries |
| **HPA** | Horizontal Pod Autoscaler — Kubernetes resource that scales the number of server pods based on CPU/memory |
| **JWT** | JSON Web Token — signed token used for authentication. Issued by Supabase Auth, validated by the Go server |
| **KB** | Knowledge Base — Markdown documents in `kb/` that the RAG service embeds and retrieves for device-specific context |
| **perceptual hash (phash)** | A hash of image content (not bytes) that produces similar hashes for visually similar images. Used to detect duplicate frames |
| **pHash** | 64-bit DCT-based perceptual fingerprint of a camera frame. Frames whose pHash differs by ≤ 8 bits (Hamming distance) are considered duplicates and skip the AI call |
| **Provider** | Riverpod's unit of shared state. Types used: `StateNotifierProvider`, `StreamProvider`, `FutureProvider`, `StateProvider`, `Provider` |
| **RAG** | Retrieval-Augmented Generation — technique of injecting relevant KB context into AI prompts to ground responses in device-specific knowledge |
| **TF-IDF** | Term Frequency-Inverse Document Frequency — the scoring algorithm used by `VectorStore` to rank KB entries by relevance to a query |
| **Render.com** | Cloud PaaS used to host the Go server. Configured via `render.yaml` |
| **Riverpod** | Flutter state management library. Providers are defined at module level and watched/read via `WidgetRef` |
| **RLS** | Row-Level Security — Supabase/Postgres feature that ensures users can only access their own rows |
| **session** | One repair interaction from start to end. Has: equipment, start time, Q&A messages, AI responses, and a summary |
| **SessionState** | Riverpod state class holding the current AI response, processing flag, error text, and streaming buffer |
| **STT** | Speech-to-Text — converts microphone audio to text. Used for voice commands in `CameraScreen` |
| **TTS** | Text-to-Speech — reads AI guidance aloud. Uses `flutter_tts` package |
| **Typewriter effect** | Progressive character-by-character text reveal animation in `GuidancePanel`. Uses `StreamingText` widget |
| **UserRateLimiter** | Per-JWT rate limiter (30 msg/min, burst 5) implemented in `middleware/userlimit.go` |
| **VectorStore** | In-memory TF-IDF index over `KBEntry` documents. One store per domain plus one global fallback; loaded at server startup from `KB_DATA_DIR` |
| **VoiceState** | Riverpod state class tracking microphone (`isListening`) and TTS (`isSpeaking`) status |
| **WebSocket** | Full-duplex TCP connection used for real-time frame streaming and AI response delivery. Endpoint: `wss://server/ws?token=<jwt>&lang=it` |
| **WS frame** | Binary/text message sent over a WebSocket connection. Not to be confused with a camera frame |
| **WSS** | WebSocket Secure — WebSocket over TLS (same as `https://` but for WebSocket protocol) |
