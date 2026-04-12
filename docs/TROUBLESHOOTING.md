# Troubleshooting Guide

## The app won't connect to the server

**Symptoms:** Connection banner shows "Disconnected" indefinitely; no AI responses.

**Checklist:**
1. Verify the server URL in Settings → Server URL (default: `https://physicscopilot.onrender.com`)
   - Must be `https://` (not `http://`), no trailing slash
   - WebSocket uses `wss://` automatically (derived from the base URL)
2. Test the server health endpoint from a browser or `curl`:
   ```
   curl https://your-server-url/health
   # Expected: {"status":"ok"}
   ```
3. If running a local server:
   - Use your machine's LAN IP (e.g., `http://192.168.1.10:8080`), not `localhost` (doesn't reach the phone)
   - Ensure the server is bound to `0.0.0.0`, not `127.0.0.1`
4. Check that Render (or your hosting provider) hasn't put the service to sleep — free-tier services spin down after 15 minutes of inactivity. First request after wake-up takes ~30s.
5. Check phone network: some corporate Wi-Fi blocks WebSocket connections on non-standard ports.

---

## Camera not working / black screen

**Symptoms:** Camera preview is black; "Camera unavailable" message.

**Checklist:**
1. **Permissions**: Go to phone Settings → Apps → PhysicsCopilot → Permissions → Camera → Allow
2. **Another app is using the camera**: Close all other apps that might hold the camera (Teams, Zoom, QR scanner). On Android, restart the app.
3. **Hardware issue**: Test with the default camera app. If that also fails, the camera hardware may be faulty.
4. **Debug build on emulator**: The Android emulator's virtual camera has limitations. Test on a physical device for best results.
5. **iOS**: Check that you granted camera permission at the first-run prompt. If you denied it, go to Settings → Privacy → Camera → PhysicsCopilot → toggle On.

---

## Slow AI responses / long wait

**Symptoms:** Spinner shows for >10 seconds after sending a frame or text.

**Checklist:**
1. **Rate limiting**: If you've sent many frames quickly, the server may be enforcing rate limits. Wait 60 seconds and try again. Look for `429` errors in device logs.
2. **Gemini API status**: Check [Google AI Studio status page](https://status.cloud.google.com). Gemini occasionally has degraded performance.
3. **Server cold start**: On Render free tier, the server takes 20–30 seconds to wake up from sleep. The first request after inactivity will be slow; subsequent requests are fast.
4. **Network latency**: Camera frames are large (compressed JPEG). On slow connections (3G), transmission takes longer. Switch to Wi-Fi.
5. **Frame quality**: Very dark or overexposed frames take longer for Gemini to analyze. Ensure good lighting on the device.
6. **App timeout**: The app surfaces a "No response from AI" error after 15 seconds. This is configurable in `camera_screen.dart` (`_kAIResponseTimeout`).

---

## Onboarding loop / app stuck on tutorial

**Symptoms:** The tutorial overlay reappears every time the app is opened; can't dismiss it permanently.

**Solution:**
1. **Clear app data** (Android): Settings → Apps → PhysicsCopilot → Storage → Clear Data
2. **Clear app data** (iOS): Delete and reinstall the app
3. **Developer workaround**: The tutorial state is stored in `SharedPreferences` under the key `session_tutorial_shown`. You can inspect it with a Flutter DevTools session:
   ```dart
   // In Flutter DevTools → App State → SharedPreferences
   // Look for: session_tutorial_shown = true
   ```
4. If the problem persists after clearing data, there may be a bug in the SharedPreferences write (check `_dismissTutorial()` in `session_screen.dart`).

---

## Voice guidance not working

**Symptoms:** No audio when AI responds; microphone button does nothing.

**Checklist:**
1. **Voice enabled in settings**: Settings → Voice Guidance → toggle On
2. **System volume**: Check that media volume is not muted
3. **TTS engine** (Android): Settings → General Management → Language → Text-to-Speech → confirm a TTS engine is installed
4. **Microphone permissions**: Settings → Apps → PhysicsCopilot → Permissions → Microphone → Allow
5. **Background audio**: On iOS, background audio sessions can conflict. Restart the app.

---

## "Error: server" or unknown errors

**Symptoms:** Red error banner with generic message.

**Checklist:**
1. Check server logs on Render dashboard → PhysicsCopilot service → Logs
2. Common causes:
   - `GEMINI_API_KEY` not set or expired → server returns `error` type message
   - `SUPABASE_JWT_SECRET` mismatch → JWT validation fails → 401
   - Supabase database unreachable → 500 on session endpoints
3. Enable debug logging on the app by building with `--dart-define=DEBUG_WS=true` to see raw WebSocket messages in the console.

---

## Reporting bugs

Open an issue at [GitHub Issues](https://github.com/your-org/physicscopilot/issues) with:
- Device model and OS version
- App version (Settings → About)
- Steps to reproduce
- Screenshot or screen recording if possible
