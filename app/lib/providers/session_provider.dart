// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

import 'dart:typed_data';

import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'prefs_provider.dart';

/// Immutable snapshot of the current AI session output.
class SessionState {
  final String? responseText;
  final String? voiceText; // TTS-optimised version of responseText (no markdown)
  final String? audioUrl;
  final Map<String, dynamic>? overlay;
  final bool isProcessing;

  /// Non-null when the last AI request failed or timed out.
  final String? errorText;

  /// Accumulated text received so far when the server is streaming chunks.
  /// Non-null only while [isStreaming] is true.
  final String? streamingText;

  /// True while the server is actively sending `chunk` messages.
  final bool isStreaming;

  const SessionState({
    this.responseText,
    this.voiceText,
    this.audioUrl,
    this.overlay,
    this.isProcessing = false,
    this.errorText,
    this.streamingText,
    this.isStreaming = false,
  });

  SessionState copyWith({
    String? responseText,
    String? voiceText,
    String? audioUrl,
    Map<String, dynamic>? overlay,
    bool? isProcessing,
    // Pass null explicitly to clear errorText / streamingText.
    String? errorText,
    String? streamingText,
    bool? isStreaming,
  }) {
    return SessionState(
      responseText: responseText ?? this.responseText,
      voiceText: voiceText ?? this.voiceText,
      audioUrl: audioUrl ?? this.audioUrl,
      overlay: overlay ?? this.overlay,
      isProcessing: isProcessing ?? this.isProcessing,
      errorText: errorText,
      streamingText: streamingText,
      isStreaming: isStreaming ?? this.isStreaming,
    );
  }
}

/// Manages session state driven by server messages.
class SessionNotifier extends Notifier<SessionState> {
  @override
  SessionState build() => const SessionState();

  /// Applies a decoded `{"type":"response", ...}` payload from the backend.
  ///
  /// Also ends any active stream, replacing the accumulated [streamingText]
  /// with the authoritative full text from the server.
  void updateFromResponse(Map<String, dynamic> json) {
    state = SessionState(
      responseText: json['text'] as String?,
      voiceText: json['voice_text'] as String?,
      audioUrl: json['audio_url'] as String?,
      overlay: json['overlay'] as Map<String, dynamic>?,
      streamingText: null,
      isStreaming: false,
    );
  }

  /// Appends [chunk] to the streaming buffer.
  ///
  /// Called for each `{"type":"chunk","text":"…"}` message.
  void appendChunk(String chunk) {
    state = SessionState(
      responseText: state.responseText,
      audioUrl: state.audioUrl,
      overlay: state.overlay,
      isProcessing: false,
      errorText: null,
      streamingText: (state.streamingText ?? '') + chunk,
      isStreaming: true,
    );
  }

  /// Marks the session as waiting for a server response (clears previous error).
  void setProcessing() {
    state = state.copyWith(
      isProcessing: true,
      errorText: null,
      streamingText: null,
      isStreaming: false,
    );
  }

  /// Sets a user-visible error message and clears the processing state.
  void setError(String message) {
    state = state.copyWith(
      isProcessing: false,
      errorText: message,
      streamingText: null,
      isStreaming: false,
    );
  }

  /// Called when the 15-second AI response timeout fires.
  void setAITimeout() => setError('Nessuna risposta dall\'AI. Riprova.');

  /// Resets all session state (e.g. on reconnect or new session).
  void reset() => state = const SessionState();
}

final sessionProvider =
    NotifierProvider<SessionNotifier, SessionState>(SessionNotifier.new);

// ── Cached response (offline fallback) ───────────────────────────────────────

const _kCachedResponseKey = 'offline_last_ai_response';

/// Last AI response persisted to SharedPreferences for offline fallback.
/// Initialised from storage; updated whenever a new response arrives.
class _CachedResponseNotifier extends Notifier<String?> {
  @override
  String? build() =>
      ref.read(sharedPrefsProvider).getString(_kCachedResponseKey);

  void set(String? value) => state = value;
}

final cachedResponseProvider =
    NotifierProvider<_CachedResponseNotifier, String?>(_CachedResponseNotifier.new);

// ── Tutorial visibility ───────────────────────────────────────────────────────

const _kTutorialKey = 'session_tutorial_shown';

/// True when the session tutorial overlay should be displayed.
class _ShowTutorialNotifier extends Notifier<bool> {
  @override
  bool build() =>
      !(ref.read(sharedPrefsProvider).getBool(_kTutorialKey) ?? false);

  void dismiss() => state = false;
}

final showTutorialProvider =
    NotifierProvider<_ShowTutorialNotifier, bool>(_ShowTutorialNotifier.new);

// ── Last captured frame ───────────────────────────────────────────────────────

/// The most-recently captured camera frame; non-null once the user has
/// tapped the capture button at least once.
class _LastFrameNotifier extends Notifier<Uint8List?> {
  @override
  Uint8List? build() => null;

  void set(Uint8List? frame) => state = frame;
}

final lastFrameProvider =
    NotifierProvider<_LastFrameNotifier, Uint8List?>(_LastFrameNotifier.new);
