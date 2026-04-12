// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

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

/// VoiceState - immutable snapshot of the voice I/O state (STT listening + TTS speaking).
class VoiceState {
  /// True while the STT engine is actively recording.
  final bool isListening;

  /// True while the TTS engine is speaking.
  final bool isSpeaking;

  /// The most-recently recognised phrase from the microphone, or `null`.
  final String? lastRecognizedText;

  const VoiceState({
    this.isListening = false,
    this.isSpeaking = false,
    this.lastRecognizedText,
  });

  /// Returns a copy with the given fields replaced.
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

/// Bridges [VoiceService] events into Riverpod state for the UI.
///
/// Subscribes to [VoiceService.speakingStream] so [VoiceState.isSpeaking]
/// stays in sync with TTS completion/error callbacks.
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

  /// Starts listening if idle, or stops if already listening.
  Future<void> toggleListening() async {
    if (state.isListening) {
      await _service.stopListening();
      state = state.copyWith(isListening: false);
    } else {
      await _service.startListening();
      state = state.copyWith(isListening: _service.isListening);
    }
  }

  /// Passes [text] to TTS and marks [VoiceState.isSpeaking] true.
  void speak(String text) {
    _service.speak(text);
    state = state.copyWith(isSpeaking: true);
  }

  /// Stops TTS immediately and clears the speaking flag.
  Future<void> stopSpeaking() async {
    await _service.stop();
    // Stream subscription handles the state reset; explicit call for clarity.
    state = state.copyWith(isSpeaking: false);
  }

  /// Called by the session screen when the STT engine emits a final result.
  void onRecognized(String text) {
    state = state.copyWith(lastRecognizedText: text);
  }
}

// ── Provider ──────────────────────────────────────────────────────────────────

/// Provides the [VoiceNotifier] and current [VoiceState].
///
/// Depends on [voiceServiceProvider]; rebuilds when the service instance changes.
final voiceProvider =
    StateNotifierProvider<VoiceNotifier, VoiceState>((ref) {
  final service = ref.watch(voiceServiceProvider);
  return VoiceNotifier(service);
});
