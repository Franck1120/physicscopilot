import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:physicscopilot/widgets/tutorial_overlay.dart';
import 'package:physicscopilot/utils/strings.dart';

void main() {
  Widget wrap(Widget child) =>
      MaterialApp(home: Scaffold(body: child));

  group('TutorialOverlay', () {
    testWidgets('renders without crashing', (WidgetTester tester) async {
      await tester.pumpWidget(
        wrap(TutorialOverlay(onDismiss: () {})),
      );
      await tester.pumpAndSettle();

      expect(find.byType(TutorialOverlay), findsOneWidget);
    });

    testWidgets('shows tutorialHint text', (WidgetTester tester) async {
      await tester.pumpWidget(
        wrap(TutorialOverlay(onDismiss: () {})),
      );
      await tester.pumpAndSettle();

      expect(find.text(AppStrings.tutorialHint), findsOneWidget);
    });

    testWidgets('shows tutorialDismiss text', (WidgetTester tester) async {
      await tester.pumpWidget(
        wrap(TutorialOverlay(onDismiss: () {})),
      );
      await tester.pumpAndSettle();

      expect(find.text(AppStrings.tutorialDismiss), findsOneWidget);
    });

    testWidgets('tapping overlay calls onDismiss callback',
        (WidgetTester tester) async {
      var dismissed = false;

      await tester.pumpWidget(
        wrap(TutorialOverlay(onDismiss: () => dismissed = true)),
      );
      await tester.pumpAndSettle();

      // Tap the GestureDetector — the outermost Container fills the scaffold.
      await tester.tap(find.byType(TutorialOverlay));
      await tester.pumpAndSettle();

      expect(dismissed, isTrue);
    });
  });
}
