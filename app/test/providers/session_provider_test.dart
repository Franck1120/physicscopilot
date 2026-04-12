import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:physicscopilot/providers/session_provider.dart';

void main() {
  group('SessionNotifier', () {
    late ProviderContainer container;
    late SessionNotifier notifier;

    setUp(() {
      container = ProviderContainer();
      notifier = container.read(sessionProvider.notifier);
    });

    tearDown(() {
      container.dispose();
    });

    test('initial state: all fields null/false', () {
      final state = container.read(sessionProvider);
      expect(state.responseText, isNull);
      expect(state.audioUrl, isNull);
      expect(state.overlay, isNull);
      expect(state.isProcessing, isFalse);
      expect(state.errorText, isNull);
      expect(state.streamingText, isNull);
      expect(state.isStreaming, isFalse);
    });

    test('setProcessing() → isProcessing=true, clears errorText/streamingText/isStreaming', () {
      // Arrange: set some prior state
      notifier.setError('previous error');
      notifier.appendChunk('some chunk');

      // Act
      notifier.setProcessing();

      final state = container.read(sessionProvider);
      expect(state.isProcessing, isTrue);
      expect(state.errorText, isNull);
      expect(state.streamingText, isNull);
      expect(state.isStreaming, isFalse);
    });

    test('setError("msg") → errorText="msg", isProcessing=false, isStreaming=false', () {
      // Arrange: set processing state first
      notifier.setProcessing();

      // Act
      notifier.setError('msg');

      final state = container.read(sessionProvider);
      expect(state.errorText, equals('msg'));
      expect(state.isProcessing, isFalse);
      expect(state.isStreaming, isFalse);
    });

    test('setAITimeout() → errorText contains "AI"', () {
      notifier.setAITimeout();

      final state = container.read(sessionProvider);
      expect(state.errorText, contains('AI'));
      expect(state.isProcessing, isFalse);
      expect(state.isStreaming, isFalse);
    });

    test('appendChunk("hello") → streamingText="hello", isStreaming=true', () {
      notifier.appendChunk('hello');

      final state = container.read(sessionProvider);
      expect(state.streamingText, equals('hello'));
      expect(state.isStreaming, isTrue);
    });

    test('appendChunk("hello") then appendChunk(" world") → streamingText="hello world"', () {
      notifier.appendChunk('hello');
      notifier.appendChunk(' world');

      final state = container.read(sessionProvider);
      expect(state.streamingText, equals('hello world'));
      expect(state.isStreaming, isTrue);
    });

    test('updateFromResponse → responseText set, isStreaming=false, streamingText=null', () {
      // Arrange: simulate active streaming
      notifier.appendChunk('partial');

      // Act
      notifier.updateFromResponse({
        'text': 'final',
        'audio_url': null,
        'overlay': null,
      });

      final state = container.read(sessionProvider);
      expect(state.responseText, equals('final'));
      expect(state.isStreaming, isFalse);
      expect(state.streamingText, isNull);
      expect(state.audioUrl, isNull);
      expect(state.overlay, isNull);
    });

    test('reset() → all fields back to initial values', () {
      // Arrange: set various state values
      notifier.setProcessing();
      notifier.appendChunk('some text');
      notifier.updateFromResponse({
        'text': 'response',
        'audio_url': 'http://example.com/audio.mp3',
        'overlay': {'type': 'highlight'},
      });

      // Act
      notifier.reset();

      final state = container.read(sessionProvider);
      expect(state.responseText, isNull);
      expect(state.audioUrl, isNull);
      expect(state.overlay, isNull);
      expect(state.isProcessing, isFalse);
      expect(state.errorText, isNull);
      expect(state.streamingText, isNull);
      expect(state.isStreaming, isFalse);
    });
  });
}
