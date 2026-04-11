import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../services/voice_service.dart';

// ── Service provider ──────────────────────────────────────────────────────────

/// Singleton [VoiceService]; initialised on creation and disposed with the scope.
final voiceServiceProvider = Provider<VoiceService>((ref) {
  final service = VoiceService();
  // Fire-and-forget: initialization is async but non-blocking.
  service.initialize();
  ref.onDispose(service.dispose);
  return service;
});

// ── State ─────────────────────────────────────────────────────────────────────

class VoiceState {
  final bool isListening;
  final bool isSpeaking;
  final String? lastRecognizedText;

  const VoiceState({
    this.isListening = false,
    this.isSpeaking = false,
    this.lastRecognizedText,
  });

  VoiceState copyWith({
    bool? isListening,
    bool? isSpeaking,
    String? lastRecognizedText,
  }) =>
      VoiceState(
        isListening: isListening ?? this.isListening,
        isSpeaking: isSpeaking ?? this.isSpeaking,
        lastRecognizedText: lastRecognizedText ?? this.lastRecognizedText,
      );
}

// ── Notifier ──────────────────────────────────────────────────────────────────

class VoiceNotifier extends StateNotifier<VoiceState> {
  final VoiceService _service;

  VoiceNotifier(this._service) : super(const VoiceState());

  Future<void> toggleListening() async {
    if (state.isListening) {
      await _service.stopListening();
      state = state.copyWith(isListening: false);
    } else {
      await _service.startListening();
      state = state.copyWith(isListening: _service.isListening);
    }
  }

  void speak(String text) {
    _service.speak(text);
    state = state.copyWith(isSpeaking: true);
  }

  Future<void> stopSpeaking() async {
    await _service.stop();
    state = state.copyWith(isSpeaking: false);
  }

  void onRecognized(String text) {
    state = state.copyWith(lastRecognizedText: text);
  }
}

// ── Provider ──────────────────────────────────────────────────────────────────

final voiceProvider =
    StateNotifierProvider<VoiceNotifier, VoiceState>((ref) {
  final service = ref.watch(voiceServiceProvider);
  return VoiceNotifier(service);
});
