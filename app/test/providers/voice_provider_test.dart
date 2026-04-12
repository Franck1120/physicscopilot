import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:physicscopilot/providers/voice_provider.dart';
import 'package:physicscopilot/services/voice_service.dart';

// ── Fake VoiceService ─────────────────────────────────────────────────────────

class FakeVoiceService extends VoiceService {
  final _speakingController = StreamController<bool>.broadcast();

  String? lastSpokenText;
  bool _isListening = false;

  @override
  Stream<bool> get speakingStream => _speakingController.stream;

  @override
  bool get isListening => _isListening;

  @override
  void speak(String text) {
    lastSpokenText = text;
  }

  @override
  Future<void> stop() async {}

  @override
  Future<void> startListening() async {
    _isListening = true;
  }

  @override
  Future<void> stopListening() async {
    _isListening = false;
  }

  @override
  Future<void> initialize() async {}

  @override
  Future<void> dispose() async {
    await _speakingController.close();
  }

  /// Emits a speaking state event so tests can trigger stream-based updates.
  void emitSpeaking(bool value) => _speakingController.add(value);
}

// ── Tests ─────────────────────────────────────────────────────────────────────

void main() {
  group('VoiceState', () {
    test('default constructor: isListening=false, isSpeaking=false, lastRecognizedText=null', () {
      const state = VoiceState();
      expect(state.isListening, isFalse);
      expect(state.isSpeaking, isFalse);
      expect(state.lastRecognizedText, isNull);
    });

    test('copyWith overrides only specified fields', () {
      const original = VoiceState(isListening: false, isSpeaking: false);
      final updated = original.copyWith(isSpeaking: true);
      expect(updated.isListening, isFalse);
      expect(updated.isSpeaking, isTrue);
      expect(updated.lastRecognizedText, isNull);
    });

    test('copyWith partial override preserves other fields', () {
      const original = VoiceState(
        isListening: true,
        isSpeaking: true,
        lastRecognizedText: 'hello',
      );
      final updated = original.copyWith(isListening: false);
      expect(updated.isListening, isFalse);
      expect(updated.isSpeaking, isTrue);
      expect(updated.lastRecognizedText, 'hello');
    });
  });

  group('VoiceNotifier', () {
    late FakeVoiceService fakeService;
    late ProviderContainer container;

    setUp(() {
      fakeService = FakeVoiceService();
      container = ProviderContainer(
        overrides: [
          voiceServiceProvider.overrideWithValue(fakeService),
        ],
      );
    });

    tearDown(() {
      container.dispose();
    });

    test('initial state: isListening=false, isSpeaking=false, lastRecognizedText=null', () {
      final state = container.read(voiceProvider);
      expect(state.isListening, isFalse);
      expect(state.isSpeaking, isFalse);
      expect(state.lastRecognizedText, isNull);
    });

    test('speak(text) → isSpeaking=true', () {
      container.read(voiceProvider.notifier).speak('test message');

      final state = container.read(voiceProvider);
      expect(state.isSpeaking, isTrue);
      expect(fakeService.lastSpokenText, 'test message');
    });

    test('stopSpeaking() → isSpeaking=false', () async {
      // First put isSpeaking into true
      container.read(voiceProvider.notifier).speak('something');
      expect(container.read(voiceProvider).isSpeaking, isTrue);

      await container.read(voiceProvider.notifier).stopSpeaking();
      expect(container.read(voiceProvider).isSpeaking, isFalse);
    });

    test('onRecognized(text) → lastRecognizedText=text', () {
      container.read(voiceProvider.notifier).onRecognized('recognized words');
      expect(container.read(voiceProvider).lastRecognizedText, 'recognized words');
    });

    test('speakingStream event true → isSpeaking=true', () async {
      // Ensure state starts as isSpeaking=false
      expect(container.read(voiceProvider).isSpeaking, isFalse);

      fakeService.emitSpeaking(true);

      // Let the stream listener process the event
      await Future<void>.delayed(Duration.zero);

      expect(container.read(voiceProvider).isSpeaking, isTrue);
    });

    test('speakingStream event false after true → isSpeaking=false', () async {
      // Set isSpeaking=true via speak()
      container.read(voiceProvider.notifier).speak('hello');
      expect(container.read(voiceProvider).isSpeaking, isTrue);

      fakeService.emitSpeaking(false);
      await Future<void>.delayed(Duration.zero);

      expect(container.read(voiceProvider).isSpeaking, isFalse);
    });
  });
}
