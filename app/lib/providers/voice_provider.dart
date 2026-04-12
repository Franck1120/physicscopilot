import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../services/voice_service.dart';

// ── Service provider ──────────────────────────────────────────────────────────

/// Singleton [VoiceService]; initialised on creation and disposed with the scope.
/// Initialization errors are logged but non-fatal — the app degrades
/// gracefully (voice features unavailable, camera still works).
final voiceServiceProvider = Provider<VoiceService>((ref) {
  final service = VoiceService();
  service.initialize().catchError((e) {
    // Non-fatal: STT/TTS may be unavailable but the app remains usable.
    assert(() { print('VoiceService init failed: $e'); return true; }());
  });
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
  late final StreamSubscription<bool> _speakingSub;

  VoiceNotifier(this._service) : super(const VoiceState()) {
    // Keep isSpeaking in sync with TTS completion/error callbacks from the
    // service layer. Without this subscription, speak() sets isSpeaking=true
    // but the state is never reset when TTS finishes naturally.
    _speakingSub = _service.speakingStream.listen((speaking) {
      if (state.isSpeaking != speaking) {
        state = state.copyWith(isSpeaking: speaking);
      }
    });
  }

  @override
  void dispose() {
    _speakingSub.cancel();
    super.dispose();
  }

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
    // Stream subscription handles the state reset; explicit call for clarity.
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
