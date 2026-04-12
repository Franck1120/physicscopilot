import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:physicscopilot/widgets/step_progress.dart';

void main() {
  Widget wrap(Widget child) =>
      MaterialApp(home: Scaffold(body: child));

  // Helper that builds three generic steps.
  List<StepInfo> threeSteps() => const [
        StepInfo(description: 'Primo passo'),
        StepInfo(description: 'Secondo passo'),
        StepInfo(description: 'Terzo passo'),
      ];

  group('StepProgress', () {
    testWidgets('empty steps list renders SizedBox.shrink — nothing visible',
        (WidgetTester tester) async {
      await tester.pumpWidget(
        wrap(const StepProgress(steps: [], currentStep: 0)),
      );
      await tester.pumpAndSettle();

      // No Text, no progress bar rendered when steps is empty.
      expect(find.text('Step 1 of 0'), findsNothing);
      expect(find.byType(LinearProgressIndicator), findsNothing);
    });

    testWidgets('three steps with currentStep=0 shows "Step 1 of 3"',
        (WidgetTester tester) async {
      await tester.pumpWidget(
        wrap(StepProgress(steps: threeSteps(), currentStep: 0)),
      );
      await tester.pumpAndSettle();

      expect(find.text('Step 1 of 3'), findsOneWidget);
    });

    testWidgets('three steps with currentStep=2 shows "Step 3 of 3"',
        (WidgetTester tester) async {
      await tester.pumpWidget(
        wrap(StepProgress(steps: threeSteps(), currentStep: 2)),
      );
      await tester.pumpAndSettle();

      expect(find.text('Step 3 of 3'), findsOneWidget);
    });

    testWidgets('shows description of the current step',
        (WidgetTester tester) async {
      await tester.pumpWidget(
        wrap(StepProgress(steps: threeSteps(), currentStep: 1)),
      );
      await tester.pumpAndSettle();

      expect(find.text('Secondo passo'), findsOneWidget);
    });

    testWidgets('currentStep out-of-bounds does not crash (clamping)',
        (WidgetTester tester) async {
      // Below lower bound
      await tester.pumpWidget(
        wrap(StepProgress(steps: threeSteps(), currentStep: -5)),
      );
      await tester.pumpAndSettle();
      expect(find.text('Step 1 of 3'), findsOneWidget);

      // Above upper bound
      await tester.pumpWidget(
        wrap(StepProgress(steps: threeSteps(), currentStep: 100)),
      );
      await tester.pumpAndSettle();
      expect(find.text('Step 3 of 3'), findsOneWidget);
    });

    testWidgets('step with seconds duration shows duration badge with "s"',
        (WidgetTester tester) async {
      const steps = [
        StepInfo(
          description: 'Quick step',
          estimatedDuration: Duration(seconds: 30),
        ),
      ];
      await tester.pumpWidget(
        wrap(const StepProgress(steps: steps, currentStep: 0)),
      );
      await tester.pumpAndSettle();

      expect(find.text('~30 s'), findsOneWidget);
    });

    testWidgets('step with minute duration shows duration badge with "min"',
        (WidgetTester tester) async {
      const steps = [
        StepInfo(
          description: 'Long step',
          estimatedDuration: Duration(minutes: 1),
        ),
      ];
      await tester.pumpWidget(
        wrap(const StepProgress(steps: steps, currentStep: 0)),
      );
      await tester.pumpAndSettle();

      expect(find.text('~1 min'), findsOneWidget);
    });
  });
}
