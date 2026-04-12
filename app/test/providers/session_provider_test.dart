import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:physicscopilot/providers/session_provider.dart';

void main() {
  group('SessionState', () {
    test('default constructor has null fields and isProcessing=false', () {
      const state = SessionState();
      expect(state.responseText, isNull);
      expect(state.audioUrl, isNull);
      expect(state.overlay, isNull);
      expect(state.isProcessing, isFalse);
      expect(state.errorText, isNull);
    });

    test('copyWith updates only specified fields', () {
      const original = SessionState(responseText: 'hello', isProcessing: false);
      final updated = original.copyWith(isProcessing: true);
      expect(updated.responseText, 'hello');
      expect(updated.isProcessing, isTrue);
    });

    test('copyWith with explicit null clears errorText', () {
      const original = SessionState(errorText: 'some error');
      final updated = original.copyWith(errorText: null);
      expect(updated.errorText, isNull);
    });

    test('copyWith partial override preserves unspecified fields', () {
      const original = SessionState(
        responseText: 'AI answer',
        audioUrl: 'https://example.com/audio.mp3',
        isProcessing: false,
        errorText: null,
        isStreaming: false,
      );
      final updated = original.copyWith(isProcessing: true);

      expect(updated.responseText, 'AI answer');
      expect(updated.audioUrl, 'https://example.com/audio.mp3');
      expect(updated.isProcessing, isTrue);
      expect(updated.errorText, isNull);
      expect(updated.isStreaming, isFalse);
    });

    test('copyWith with streamingText preserves other fields', () {
      const original = SessionState(responseText: 'prev', isStreaming: false);
      final updated = original.copyWith(
        streamingText: 'chunk…',
        isStreaming: true,
      );

      expect(updated.responseText, 'prev');
      expect(updated.streamingText, 'chunk…');
      expect(updated.isStreaming, isTrue);
    });

    test('isStreaming defaults to false', () {
      const state = SessionState();
      expect(state.isStreaming, isFalse);
    });

    test('isProcessing and isStreaming are independent flags', () {
      const processing = SessionState(isProcessing: true, isStreaming: false);
      const streaming = SessionState(isProcessing: false, isStreaming: true);

      expect(processing.isProcessing, isTrue);
      expect(processing.isStreaming, isFalse);
      expect(streaming.isProcessing, isFalse);
      expect(streaming.isStreaming, isTrue);
    });

    test('errorText and isProcessing are mutually exclusive in practice', () {
      // Verify that setting an error also means not processing.
      const withError = SessionState(errorText: 'boom', isProcessing: false);
      // Setting processing clears error through notifier, but state itself
      // allows both — the test ensures copyWith respects explicit null for error.
      final processing = withError.copyWith(isProcessing: true, errorText: null);
      expect(processing.errorText, isNull);
      expect(processing.isProcessing, isTrue);
    });
  });

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

    test('initial state is blank SessionState', () {
      final state = container.read(sessionProvider);
      expect(state.responseText, isNull);
      expect(state.isProcessing, isFalse);
      expect(state.errorText, isNull);
    });

    test('updateFromResponse sets responseText from json text key', () {
      notifier.updateFromResponse({'text': 'Check the nozzle', 'overlay': null});
      expect(container.read(sessionProvider).responseText, 'Check the nozzle');
    });

    test('updateFromResponse with null text keeps responseText null', () {
      notifier.updateFromResponse({'text': null, 'overlay': null});
      expect(container.read(sessionProvider).responseText, isNull);
    });

    test('updateFromResponse sets overlay map', () {
      final overlay = <String, dynamic>{'boxes': <dynamic>[], 'arrows': <dynamic>[]};
      notifier.updateFromResponse({'text': 'ok', 'overlay': overlay});
      expect(container.read(sessionProvider).overlay, overlay);
    });

    test('setProcessing sets isProcessing=true and clears errorText', () {
      notifier.setError('previous error');
      notifier.setProcessing();
      final state = container.read(sessionProvider);
      expect(state.isProcessing, isTrue);
      expect(state.errorText, isNull);
    });

    test('setError sets errorText and clears isProcessing', () {
      notifier.setProcessing();
      notifier.setError('Something went wrong');
      final state = container.read(sessionProvider);
      expect(state.errorText, 'Something went wrong');
      expect(state.isProcessing, isFalse);
    });

    test('setAITimeout sets Italian timeout error message', () {
      notifier.setAITimeout();
      final state = container.read(sessionProvider);
      expect(state.errorText, contains('AI'));
      expect(state.errorText, contains('Riprova'));
      expect(state.isProcessing, isFalse);
    });

    test('reset clears all state to initial values', () {
      notifier.updateFromResponse({'text': 'Some text', 'overlay': null});
      notifier.setProcessing();
      notifier.reset();
      final state = container.read(sessionProvider);
      expect(state.responseText, isNull);
      expect(state.isProcessing, isFalse);
      expect(state.errorText, isNull);
    });

    test('appendChunk accumulates streaming text', () {
      notifier.appendChunk('Hello');
      notifier.appendChunk(', world');
      final state = container.read(sessionProvider);
      expect(state.streamingText, 'Hello, world');
      expect(state.isStreaming, isTrue);
      expect(state.isProcessing, isFalse);
    });

    test('appendChunk sets isStreaming=true and errorText=null', () {
      notifier.setError('previous error');
      notifier.appendChunk('new chunk');
      final state = container.read(sessionProvider);
      expect(state.isStreaming, isTrue);
      expect(state.errorText, isNull);
    });

    test('updateFromResponse clears streamingText and isStreaming', () {
      notifier.appendChunk('partial');
      notifier.updateFromResponse({'text': 'Full response', 'overlay': null});
      final state = container.read(sessionProvider);
      expect(state.streamingText, isNull);
      expect(state.isStreaming, isFalse);
      expect(state.responseText, 'Full response');
    });

    test('setProcessing clears streamingText', () {
      notifier.appendChunk('partial');
      notifier.setProcessing();
      final state = container.read(sessionProvider);
      expect(state.streamingText, isNull);
      expect(state.isStreaming, isFalse);
      expect(state.isProcessing, isTrue);
    });

    test('sessionProvider initial state after container recreate is clean', () {
      // Mutate state in the first container.
      notifier.updateFromResponse({'text': 'cached', 'overlay': null});
      expect(container.read(sessionProvider).responseText, 'cached');

      // Dispose and create a fresh container — state must not bleed over.
      container.dispose();
      final freshContainer = ProviderContainer();
      addTearDown(freshContainer.dispose);

      final freshState = freshContainer.read(sessionProvider);
      expect(freshState.responseText, isNull);
      expect(freshState.isProcessing, isFalse);
      expect(freshState.errorText, isNull);
      expect(freshState.isStreaming, isFalse);
    });
  });
}
