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
  });
}
