# Frequently Asked Questions

## General

**Q: What devices does PhysicsCopilot support?**  
A: Any physical device you can point a camera at. The AI uses Gemini Vision with no pre-defined device catalog — it analyzes what it sees. The knowledge base (`kb/`) adds device-specific depth for supported models.

**Q: Does the app work offline?**  
A: Partially. The last AI response is cached and shown when offline. New analyses require an internet connection and a running server.

**Q: Is my camera footage stored?**  
A: No. Frames are processed in memory on the server and discarded. They are never written to disk or stored in the database.

**Q: Which languages are supported?**  
A: The app sends a `?lang=` parameter with each WebSocket connection. Gemini responds in the language you configure. Currently tested: Italian (it), English (en). Other BCP-47 codes should work but are untested.

---

## Setup

**Q: Why doesn't the server URL field accept my local IP?**  
A: Make sure you're using `http://` (not `https://`) for local development, and that your phone is on the same Wi-Fi network as your computer. Format: `http://192.168.x.x:8080`.

**Q: Do I need a Supabase account?**  
A: For basic usage (no auth, no session history), no. Set `ENV=development` and the server runs without Supabase. For production with authentication and session persistence, yes.

**Q: How do I get a Gemini API key?**  
A: Go to [Google AI Studio](https://aistudio.google.com) → Get API key. Free tier includes 15 requests/minute.

---

## Performance

**Q: Responses are slow (> 5 seconds). What's wrong?**  
A: See `docs/TROUBLESHOOTING.md` → "Slow AI responses". Most common causes: Gemini API rate limit, Render cold start, slow Wi-Fi.

**Q: The app shows "No response from AI" after 15 seconds.**  
A: The app has a 15-second timeout. If Gemini is slow (during high traffic), the timeout fires first. Retry — subsequent requests are usually faster. The timeout is configurable in `camera_screen.dart` (`_kAIResponseTimeout`).

---

## Development

**Q: Can I use OpenAI instead of Gemini?**  
A: Yes. Set `AI_BACKEND=openai` and provide `OPENAI_API_KEY`. The `openai_backend.go` service implements the same `AIService` interface. Note: OpenAI does not support native multimodal vision in the same way — frame analysis quality may differ.

**Q: How do I add a new AI prompt?**  
A: Edit `server/internal/services/gemini_service.go` (or `openai_backend.go`). The system prompt defines the AI's persona, output format, and device context.

**Q: Can I self-host without Docker?**  
A: Yes. Build the Go binary directly:
```bash
cd server && go build -o physicscopilot-server ./cmd/server
./physicscopilot-server
```

**Q: How do I add a device to the knowledge base?**  
A: See `docs/KB_FORMAT.md`. Create a Markdown file in `kb/` with YAML frontmatter, add tags for retrieval, and restart the server. The RAG service rebuilds its index on startup.

---

## WebSocket & Connectivity

**Q: Why does the WebSocket disconnect frequently?**  
A: Common causes: (1) Render free tier spins down after 15 min of inactivity — upgrade to Starter to avoid cold starts. (2) Unstable Wi-Fi — the app reconnects automatically with exponential back-off (1s, 2s, 4s... up to 30s). (3) Rate limit exceeded — the server closes the connection after 30 msg/min per user. Check the connection status banner in the app for details.

**Q: What happens if the Gemini API does not respond?**  
A: The server has a 30-second timeout per Gemini request. If Gemini is unreachable or exceeds the timeout, the server returns an error message to the client. The app shows "No response from AI" and the user can retry. If `GEMINI_PROXY_URL` is configured, the server falls back to the proxy endpoint automatically.

**Q: How do I export session history?**  
A: Session data is stored in Supabase Postgres when `DATABASE_URL` is set. You can query it directly via the Supabase dashboard (Table Editor or SQL Editor). A future release will add PDF export of annotated sessions from the app. For now, use the REST API: `GET /api/sessions/:id` returns the full session JSON.

---

## Permissions & Security

**Q: What permissions does the app require on Android?**  
A: Camera (for live frame analysis), Microphone (for voice commands via STT), Internet (for server communication). These are declared in `AndroidManifest.xml`. The app requests camera and microphone permissions at runtime before using them.

**Q: What permissions does the app require on iOS?**  
A: Camera (`NSCameraUsageDescription`), Microphone (`NSMicrophoneUsageDescription`), Speech Recognition (`NSSpeechRecognitionUsageDescription`). All are declared in `Info.plist` with user-facing descriptions. iOS prompts the user before granting each permission.

**Q: How do I configure Supabase Auth?**  
A: (1) Create a project on [supabase.com](https://supabase.com). (2) Copy `SUPABASE_URL`, `SUPABASE_ANON_KEY`, and `SUPABASE_JWT_SECRET` from Settings → API. (3) Set these as environment variables on your server. (4) In the Flutter app, configure `SupabaseService` with the URL and anon key. The app uses Supabase's email/password auth flow. See `app/lib/services/auth_service.dart` for the implementation.

---

## Advanced

**Q: Can I use a different AI model instead of Gemini?**  
A: Yes. Set `AI_BACKEND=openai` and provide `OPENAI_API_KEY` to use OpenAI models. The `openai_backend.go` implements the same `AIService` interface. Multimodal vision quality may differ since Gemini 2.5 Flash is optimized for frame analysis. Adding new backends requires implementing the `AIService` interface in `server/internal/services/`.

**Q: How do I run load tests?**  
A: Install [k6](https://k6.io) and run against your server. See `docs/PERFORMANCE.md` for a ready-to-use k6 script. The server comfortably handles ~50 concurrent WebSocket sessions on Render Standard tier (2GB RAM, 1 vCPU).

**Q: How do I monitor the server in production?**  
A: The server exposes Prometheus metrics at `GET /metrics` (protected with HTTP Basic Auth via `METRICS_USER`/`METRICS_PASSWORD`). See `docs/MONITORING.md` for Prometheus and Grafana setup, including alert rules for latency spikes and error rate.

**Q: Can I run PhysicsCopilot behind a reverse proxy?**  
A: Yes. The `infra/nginx.conf` file provides a production-ready nginx configuration with TLS termination, WebSocket proxy (`Upgrade` header forwarding), security headers (HSTS, CSP, X-Frame-Options), and gzip compression. Point nginx at `localhost:8080` where the Go server listens.

---

## Contributing

**Q: How do I contribute?**  
A: See `CONTRIBUTING.md`. TL;DR: fork → branch → PR → pass CI → review.

**Q: Where do I report bugs?**  
A: [GitHub Issues](https://github.com/Franck1120/physicscopilot/issues). Use the bug report template.
