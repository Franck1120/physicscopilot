import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:physicscopilot/widgets/multi_step_view.dart';

void main() {
  group('parseSteps()', () {
    test('parses 2+ numbered steps with dot separator', () {
      final steps = parseSteps('1. Do this\n2. Do that');
      expect(steps, isNotNull);
      expect(steps, equals(['Do this', 'Do that']));
    });

    test('returns null when text has no numbered steps', () {
      final steps = parseSteps('no steps here');
      expect(steps, isNull);
    });

    test('returns null for a single step (fewer than 2)', () {
      final steps = parseSteps('1. Only one');
      expect(steps, isNull);
    });

    test('parses steps with parenthesis separator "1) First"', () {
      final steps = parseSteps('1) First\n2) Second\n3) Third');
      expect(steps, isNotNull);
      expect(steps, hasLength(3));
      expect(steps, equals(['First', 'Second', 'Third']));
    });

    test('trims leading/trailing whitespace from each step', () {
      final steps = parseSteps('1.   Trim me   \n2.  And me  ');
      expect(steps, isNotNull);
      expect(steps![0], equals('Trim me'));
      expect(steps[1], equals('And me'));
    });
  });

  group('MultiStepView widget', () {
    const steps = ['Open the panel', 'Unscrew the bolts', 'Replace the part'];

    Widget _buildSubject() => MaterialApp(
          home: Scaffold(
            body: MultiStepView(steps: steps),
          ),
        );

    testWidgets('renders with 3 steps and shows "0 / 3 completati"',
        (tester) async {
      await tester.pumpWidget(_buildSubject());

      expect(find.text('0 / 3 completati'), findsOneWidget);
      for (final step in steps) {
        expect(find.text(step), findsOneWidget);
      }
    });

    testWidgets('tap first step card → shows as checked, progress updates to "1 / 3 completati"',
        (tester) async {
      await tester.pumpWidget(_buildSubject());

      // Tap the first step text (inside GestureDetector)
      await tester.tap(find.text(steps[0]));
      await tester.pump();

      expect(find.text('1 / 3 completati'), findsOneWidget);
      // The check icon should now appear for the tapped step
      expect(find.byIcon(Icons.check), findsOneWidget);
    });

    testWidgets('tap same step again → unchecks, progress goes back to "0 / 3 completati"',
        (tester) async {
      await tester.pumpWidget(_buildSubject());

      await tester.tap(find.text(steps[0]));
      await tester.pump();
      expect(find.text('1 / 3 completati'), findsOneWidget);

      await tester.tap(find.text(steps[0]));
      await tester.pump();
      expect(find.text('0 / 3 completati'), findsOneWidget);
    });

    testWidgets('checking all steps shows completion banner', (tester) async {
      await tester.pumpWidget(_buildSubject());

      for (final step in steps) {
        await tester.tap(find.text(step));
        await tester.pump();
      }
      await tester.pumpAndSettle();

      expect(find.text('3 / 3 completati'), findsOneWidget);
      expect(find.text('Tutti i passi completati!'), findsOneWidget);
    });
  });
}
