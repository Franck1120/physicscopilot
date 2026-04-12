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
A: See `docs/KB_FORMAT.md`.

---

## Contributing

**Q: How do I contribute?**  
A: See `CONTRIBUTING.md`. TL;DR: fork → branch → PR → pass CI → review.

**Q: Where do I report bugs?**  
A: [GitHub Issues](https://github.com/Franck1120/physicscopilot/issues). Use the bug report template.
