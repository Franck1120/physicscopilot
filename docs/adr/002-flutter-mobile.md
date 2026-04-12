# ADR-002: Flutter for the Mobile Client

**Status:** Accepted
**Date:** 2026-04-12

---

## Context

PhysicsCopilot's primary user experience is a mobile camera feed with
real-time overlays. The client must:

1. Access the hardware camera at ~1 fps.
2. Maintain a persistent WebSocket connection to the Go server.
3. Render bounding-box and arrow overlays on top of the live feed using a
   `CustomPainter`.
4. Play text-to-speech audio (`flutter_tts`) for hands-free use.
5. Ship to both Android and iOS from a single codebase.

Candidate frameworks:

| Framework | Language | Rendering | Camera access | iOS/Android parity |
|-----------|----------|-----------|---------------|--------------------|
| **Flutter** | Dart | Own engine (Skia/Impeller) | Good | Excellent |
| React Native | TypeScript | Native components | Variable | Moderate |
| Native (Swift + Kotlin) | Swift / Kotlin | Native | Perfect | None (two codebases) |

React Native's bridge architecture (or JSI in the new architecture) adds
serialization overhead on the hot path. The overlay drawing requirement
— custom canvas painting on every frame — is a first-class use case for
Flutter's `CustomPainter`, while in React Native it requires a native module
or a third-party canvas library.

Writing separate native apps would double the maintenance burden and require
two distinct language skill sets.

---

## Decision

Use **Flutter 3.41** with **Dart** for the mobile client, targeting Android
as the primary platform with iOS compilation supported.

State management uses **Riverpod** (code-generated providers via
`riverpod_annotation`) for its compile-time safety and testability.

---

## Consequences

### Positive

- **Single codebase, two platforms.** One `flutter build apk` / `flutter
  build ipa` invocation covers both stores.
- **CustomPainter for overlays.** Bounding boxes and arrows are drawn
  natively in the Flutter render tree, with no bridge overhead on each frame.
- **Impeller renderer (Android/iOS).** Smooth 60+ fps rendering of overlays
  even on mid-range devices.
- **Strong camera package.** The `camera: ^0.11` plugin exposes
  `CameraController` with `startImageStream` for frame capture without
  needing platform channels.
- **`flutter_tts` for hands-free audio.** Single-line TTS calls work
  identically on both platforms.

### Negative

- **Dart learning curve.** Dart is not widely known outside the Flutter
  ecosystem. New contributors familiar with JS/TS need onboarding.
- **Not truly native.** Flutter renders into its own canvas; OS-provided
  UI components (e.g. native date pickers, system fonts) are not used by
  default. This is acceptable for a tool-focused app.
- **App size.** Flutter apps include the engine (~10 MB baseline), larger
  than a minimal React Native app, though comparable in practice once
  JavaScript bundles are counted.
- **Camera plugin fragmentation.** The `camera` package has known quirks on
  certain Android OEM camera stacks; additional testing on target devices
  is required.
