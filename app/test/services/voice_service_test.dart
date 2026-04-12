// Unit tests for VoiceService.
//
// VoiceService depends on flutter_tts and speech_to_text, both of which
// require real hardware (speaker / microphone).  These tests exercise only
// the observable public API without calling initialize(), so that they run
// safely in a headless CI environment.
import 'package:flutter_test/flutter_test.dart';
import 'package:physicscopilot/services/voice_service.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  group('VoiceService — structural / no-hardware tests', () {
    late VoiceService service;

    setUp(() {
      service = VoiceService();
    });

    tearDown(() async {
      // dispose() must be safe to call even without initialize().
      await service.dispose();
    });

    test('creates instance without throwing', () {
      expect(service, isNotNull);
    });

    test('isListening is false before initialize()', () {
      expect(service.isListening, isFalse);
    });

    test('isSpeaking is false before initialize()', () {
      expect(service.isSpeaking, isFalse);
    });

    test('stopListening() when not listening does not throw', () async {
      // Guard clause in stopListening() returns early when _isListening==false.
      await expectLater(service.stopListening(), completes);
    });

    test('stop() sets isSpeaking to false', () async {
      await service.stop();
      expect(service.isSpeaking, isFalse);
    });

    test('stop() does not throw when called without initialize()', () async {
      await expectLater(service.stop(), completes);
    });

    test('dispose() can be called without prior initialize()', () async {
      await expectLater(service.dispose(), completes);
    });

    test('dispose() can be called multiple times without throwing', () async {
      await service.dispose();
      // Second dispose — streams are already closed; should not crash.
      await expectLater(service.dispose(), completes);
    });

    test('recognizedText is a Stream<String>', () {
      expect(service.recognizedText, isA<Stream<String>>());
    });

    test('speakingStream is a Stream<bool>', () {
      expect(service.speakingStream, isA<Stream<bool>>());
    });

    test('speakingStream emits false when stop() is called', () async {
      final values = <bool>[];
      final sub = service.speakingStream.listen(values.add);
      await service.stop();
      await sub.cancel();
      expect(values, contains(false));
    });

    test('isSpeaking remains false after repeated stop() calls', () async {
      await service.stop();
      await service.stop();
      expect(service.isSpeaking, isFalse);
    });
  });
}
