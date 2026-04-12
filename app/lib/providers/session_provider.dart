import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Immutable snapshot of the current AI session output.
class SessionState {
  final String? responseText;
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
    this.audioUrl,
    this.overlay,
    this.isProcessing = false,
    this.errorText,
    this.streamingText,
    this.isStreaming = false,
  });

  SessionState copyWith({
    String? responseText,
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
class SessionNotifier extends StateNotifier<SessionState> {
  SessionNotifier() : super(const SessionState());

  /// Applies a decoded `{"type":"response", ...}` payload from the backend.
  ///
  /// Also ends any active stream, replacing the accumulated [streamingText]
  /// with the authoritative full text from the server.
  void updateFromResponse(Map<String, dynamic> json) {
    state = SessionState(
      responseText: json['text'] as String?,
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
    StateNotifierProvider<SessionNotifier, SessionState>(
  (ref) => SessionNotifier(),
);
