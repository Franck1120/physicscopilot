import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Immutable snapshot of the current AI session output.
class SessionState {
  final String? responseText;
  final String? audioUrl;
  final Map<String, dynamic>? overlay;
  final bool isProcessing;

  const SessionState({
    this.responseText,
    this.audioUrl,
    this.overlay,
    this.isProcessing = false,
  });

  SessionState copyWith({
    String? responseText,
    String? audioUrl,
    Map<String, dynamic>? overlay,
    bool? isProcessing,
  }) {
    return SessionState(
      responseText: responseText ?? this.responseText,
      audioUrl: audioUrl ?? this.audioUrl,
      overlay: overlay ?? this.overlay,
      isProcessing: isProcessing ?? this.isProcessing,
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
      audioUrl: json['audio_url'] as String?,
      overlay: json['overlay'] as Map<String, dynamic>?,
    );
  }

  /// Marks the session as waiting for a server response.
  void setProcessing() {
    state = state.copyWith(isProcessing: true);
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
