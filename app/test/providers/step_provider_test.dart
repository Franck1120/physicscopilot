import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:physicscopilot/providers/step_provider.dart';
import 'package:physicscopilot/widgets/step_progress.dart';

// ── Helpers ───────────────────────────────────────────────────────────────────

List<StepInfo> _makeSteps(int count) => List.generate(
      count,
      (i) => StepInfo(description: 'Step ${i + 1}'),
    );

void main() {
  group('ProcedureState', () {
    test('default state: steps=[], currentIndex=0, isCompleted=false', () {
      const state = ProcedureState();
      expect(state.steps, isEmpty);
      expect(state.currentIndex, 0);
      expect(state.isCompleted, isFalse);
      expect(state.totalSteps, 0);
    });

    test('isCompleted=true when currentIndex == steps.length - 1', () {
      final state = ProcedureState(
        steps: _makeSteps(3),
        currentIndex: 2,
      );
      expect(state.isCompleted, isTrue);
    });

    test('isCompleted=false when currentIndex < steps.length - 1', () {
      final state = ProcedureState(
        steps: _makeSteps(3),
        currentIndex: 1,
      );
      expect(state.isCompleted, isFalse);
    });
  });

  group('StepNotifier', () {
    late ProviderContainer container;
    late StepNotifier notifier;

    setUp(() {
      container = ProviderContainer();
      notifier = container.read(stepProvider.notifier);
    });

    tearDown(() {
      container.dispose();
    });

    test('initial state: steps=[], currentIndex=0, isCompleted=false', () {
      final state = container.read(stepProvider);
      expect(state.steps, isEmpty);
      expect(state.currentIndex, 0);
      expect(state.isCompleted, isFalse);
    });

    test('loadSteps([...]) → totalSteps updated, currentIndex=0', () {
      notifier.loadSteps(_makeSteps(4));

      final state = container.read(stepProvider);
      expect(state.totalSteps, 4);
      expect(state.currentIndex, 0);
    });

    test('advance() → increments currentIndex', () {
      notifier.loadSteps(_makeSteps(3));
      notifier.advance();

      expect(container.read(stepProvider).currentIndex, 1);
    });

    test('advance() at last step → does not increment beyond last index', () {
      notifier.loadSteps(_makeSteps(3));
      // Advance to last step (index 2)
      notifier.advance();
      notifier.advance();
      expect(container.read(stepProvider).currentIndex, 2);

      // Extra advance should not move past end
      notifier.advance();
      expect(container.read(stepProvider).currentIndex, 2);
    });

    test('goTo(n) → currentIndex=n', () {
      notifier.loadSteps(_makeSteps(5));
      notifier.goTo(3);

      expect(container.read(stepProvider).currentIndex, 3);
    });

    test('goTo(-1) → does not change currentIndex', () {
      notifier.loadSteps(_makeSteps(3));
      notifier.goTo(-1);

      expect(container.read(stepProvider).currentIndex, 0);
    });

    test('goTo(steps.length) → does not change (out of range)', () {
      notifier.loadSteps(_makeSteps(3));
      notifier.goTo(3); // index 3 is out of range for 3 steps (0,1,2)

      expect(container.read(stepProvider).currentIndex, 0);
    });

    test('isCompleted=true when currentIndex == steps.length - 1', () {
      notifier.loadSteps(_makeSteps(3));
      notifier.advance();
      notifier.advance();

      expect(container.read(stepProvider).isCompleted, isTrue);
    });

    test('reset() → restores initial state', () {
      notifier.loadSteps(_makeSteps(3));
      notifier.advance();
      notifier.reset();

      final state = container.read(stepProvider);
      expect(state.steps, isEmpty);
      expect(state.currentIndex, 0);
      expect(state.isCompleted, isFalse);
    });

    test('updateFromResponse parses steps and current_step', () {
      notifier.updateFromResponse({
        'steps': [
          {'description': 'Step 1', 'estimated_seconds': 30},
          {'description': 'Step 2', 'estimated_seconds': null},
          {'description': 'Step 3'},
        ],
        'current_step': 1,
      });

      final state = container.read(stepProvider);
      expect(state.totalSteps, 3);
      expect(state.steps[0].description, 'Step 1');
      expect(state.steps[0].estimatedDuration, const Duration(seconds: 30));
      expect(state.steps[1].estimatedDuration, isNull);
      expect(state.currentIndex, 1);
    });

    test('updateFromResponse with only steps and no current_step defaults to index 0', () {
      notifier.updateFromResponse({
        'steps': [
          {'description': 'Step 1', 'estimated_seconds': 30},
        ],
      });

      final state = container.read(stepProvider);
      expect(state.totalSteps, 1);
      expect(state.currentIndex, 0);
    });
  });
}
