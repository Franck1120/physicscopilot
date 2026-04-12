import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Immutable snapshot of the current AI session output.
class SessionState {
  final String? responseText;
  final String? voiceText; // TTS-optimised version of responseText (no markdown)
  final String? audioUrl;
  final Map<String, dynamic>? overlay;
  final bool isProcessing;
  /// Non-null when the last AI request failed or timed out.
  final String? errorText;

  const SessionState({
    this.responseText,
    this.voiceText,
    this.audioUrl,
    this.overlay,
    this.isProcessing = false,
    this.errorText,
  });

  SessionState copyWith({
    String? responseText,
    String? voiceText,
    String? audioUrl,
    Map<String, dynamic>? overlay,
    bool? isProcessing,
    String? errorText,
  }) {
    return SessionState(
      responseText: responseText ?? this.responseText,
      voiceText: voiceText ?? this.voiceText,
      audioUrl: audioUrl ?? this.audioUrl,
      overlay: overlay ?? this.overlay,
      isProcessing: isProcessing ?? this.isProcessing,
      // Pass null explicitly to clear errorText; copyWith cannot distinguish
      // "not provided" from "clear it", so callers pass null to clear.
      errorText: errorText,
    );
  }
}

/// Manages session state driven by server [response] messages.
class SessionNotifier extends StateNotifier<SessionState> {
  SessionNotifier() : super(const SessionState());

  /// Applies a decoded `{"type":"response", ...}` payload from the backend.
  void updateFromResponse(Map<String, dynamic> json) {
    state = SessionState(
      responseText: json['text'] as String?,
      voiceText: json['voice_text'] as String?,
      audioUrl: json['audio_url'] as String?,
      overlay: json['overlay'] as Map<String, dynamic>?,
    );
  }

  /// Marks the session as waiting for a server response (clears previous error).
  void setProcessing() {
    state = state.copyWith(isProcessing: true, errorText: null);
  }

  /// Sets a user-visible error message and clears the processing state.
  void setError(String message) {
    state = state.copyWith(isProcessing: false, errorText: message);
  }

  /// Called when the 15-second AI response timeout fires.
  void setAITimeout() {
    setError('Nessuna risposta dall\'AI. Riprova.');
  }

  /// Resets all session state (e.g. on reconnect or new session).
  void reset() {
    state = const SessionState();
  }
}

final sessionProvider =
    StateNotifierProvider<SessionNotifier, SessionState>(
  (ref) => SessionNotifier(),
);
