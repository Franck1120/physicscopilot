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
| **Equipment** | A device selected by the user from the catalog before starting a session. Provides context to the AI |
| **Frame deduplication** | Perceptual hashing of frames to skip sending near-identical images to Gemini. Implemented in `services/phash.go` |
| **Gemini** | Google's multimodal AI model used for visual analysis. PhysicsCopilot uses Gemini 1.5 Flash |
| **GoDoc** | Go's documentation convention: comments immediately preceding a function/type, starting with the identifier name |
| **HPA** | Horizontal Pod Autoscaler — Kubernetes resource that scales the number of server pods based on CPU/memory |
| **JWT** | JSON Web Token — signed token used for authentication. Issued by Supabase Auth, validated by the Go server |
| **KB** | Knowledge Base — Markdown documents in `kb/` that the RAG service embeds and retrieves for device-specific context |
| **perceptual hash (phash)** | A hash of image content (not bytes) that produces similar hashes for visually similar images. Used to detect duplicate frames |
| **Provider** | Riverpod's unit of shared state. Types used: `StateNotifierProvider`, `StreamProvider`, `FutureProvider`, `StateProvider`, `Provider` |
| **RAG** | Retrieval-Augmented Generation — technique of retrieving relevant KB chunks and including them in the Gemini prompt |
| **Render.com** | Cloud PaaS used to host the Go server. Configured via `render.yaml` |
| **Riverpod** | Flutter state management library. Providers are defined at module level and watched/read via `WidgetRef` |
| **RLS** | Row-Level Security — Supabase/Postgres feature that ensures users can only access their own rows |
| **session** | One repair interaction from start to end. Has: equipment, start time, Q&A messages, AI responses, and a summary |
| **SessionState** | Riverpod state class holding the current AI response, processing flag, error text, and streaming buffer |
| **STT** | Speech-to-Text — converts microphone audio to text. Used for voice commands in `CameraScreen` |
| **TTS** | Text-to-Speech — reads AI guidance aloud. Uses `flutter_tts` package |
| **Typewriter effect** | Progressive character-by-character text reveal animation in `GuidancePanel`. Uses `StreamingText` widget |
| **UserRateLimiter** | Per-JWT rate limiter (30 msg/min, burst 5) implemented in `middleware/userlimit.go` |
| **VoiceState** | Riverpod state class tracking microphone (`isListening`) and TTS (`isSpeaking`) status |
| **WebSocket** | Full-duplex TCP connection used for real-time frame streaming and AI response delivery. Endpoint: `wss://server/ws?token=<jwt>&lang=it` |
| **WS frame** | Binary/text message sent over a WebSocket connection. Not to be confused with a camera frame |
| **WSS** | WebSocket Secure — WebSocket over TLS (same as `https://` but for WebSocket protocol) |
| **TF-IDF** | Term Frequency–Inverse Document Frequency — statistical measure used by the RAG service (`rag_service.go`) to rank KB documents by keyword relevance to a user query |
| **Knowledge Base (KB)** | Collection of domain-specific Markdown documents in `kb/` that the RAG service indexes and retrieves to enrich AI prompts with device-specific context |
| **Domain** | A category of physical equipment in the knowledge base (e.g., Printer, HVAC, Automotive). Each domain has its own `*_problems.json` file in `kb/` |
| **Session** | One repair interaction from start to end. Contains: equipment selection, start time, Q&A messages, AI responses, steps, and an optional summary |
| **pHash (perceptual hash)** | A content-based image hash that produces similar digests for visually similar images. Used in `phash.go` to deduplicate camera frames and avoid redundant Gemini API calls |
| **Frame deduplication** | Process of comparing incoming camera frames by perceptual hash to skip sending near-identical images to the AI. Saves 30–60% of API calls for static scenes |
| **DBBackend** | Go interface in `services/db_service.go` that abstracts database operations (`SaveSession`, `SaveFeedback`, etc.). Implemented by `DBService` (Postgres) and can be stubbed in tests |
| **AIService** | Go interface that abstracts the AI inference backend. Implemented by `GeminiService` and `OpenAIBackend`, allowing the server to swap AI providers via `AI_BACKEND` env var |
| **Fiber** | Go web framework used for the HTTP/WebSocket server. Chosen for its low-overhead routing and built-in WebSocket support. See ADR-001 |
| **Riverpod** | Flutter state management library used throughout the app. Providers are defined at module level and consumed via `WidgetRef.watch()` / `WidgetRef.read()` |
| **HPA** | Horizontal Pod Autoscaler — Kubernetes resource in `infra/k8s/` that scales server pods based on CPU/memory utilization |
| **RLS** | Row-Level Security — Supabase/Postgres feature that restricts data access so users can only query their own rows. Enabled on every PhysicsCopilot table |
| **Render.com** | Cloud PaaS where the Go server is deployed. Auto-deploys from `main` branch via `render.yaml`. Free tier has cold starts after 15 min of inactivity |
| **Air** | Go live-reload tool used in `Dockerfile.dev` for hot-reloading the server during development. Watches `.go` files and rebuilds on change |
| **k6** | Open-source load testing tool used to benchmark the server. See `docs/PERFORMANCE.md` for ready-to-use k6 scripts |
| **gosec** | Go security linter that scans for common vulnerabilities (SQL injection, hardcoded credentials, weak crypto). Run in CI on every PR |
| **govulncheck** | Go tool that checks dependencies against the Go vulnerability database. Run weekly in CI to detect known CVEs |
| **Supabase** | Open-source Firebase alternative providing Postgres database, authentication, and real-time subscriptions. PhysicsCopilot uses it for auth (JWT) and optional session persistence |
