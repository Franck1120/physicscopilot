import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:physicscopilot/widgets/guidance_overlay.dart';

void main() {
  group('GuidanceOverlay', () {
    testWidgets(
      'renders without crashing when text is null and not processing',
      (tester) async {
        await tester.pumpWidget(
          const MaterialApp(
            home: Scaffold(
              body: GuidanceOverlay(text: null, isProcessing: false),
            ),
          ),
        );
        expect(find.byType(GuidanceOverlay), findsOneWidget);
        expect(find.text('AI sta analizzando...'), findsNothing);
      },
    );

    testWidgets('shows provided text after animation completes', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: GuidanceOverlay(text: 'Check the extruder', isProcessing: false),
          ),
        ),
      );
      // Let the 350ms fade-in animation complete
      await tester.pump(const Duration(milliseconds: 400));
      expect(find.text('Check the extruder'), findsOneWidget);
    });

    testWidgets('shows analysing text when isProcessing is true', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: GuidanceOverlay(text: null, isProcessing: true),
          ),
        ),
      );
      await tester.pump(const Duration(milliseconds: 400));
      expect(find.text('AI sta analizzando...'), findsOneWidget);
    });

    testWidgets('processing indicator takes priority over text', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: GuidanceOverlay(text: 'Some instruction', isProcessing: true),
          ),
        ),
      );
      await tester.pump(const Duration(milliseconds: 400));
      // isProcessing renders _AnalysingRow which does not show the text prop
      expect(find.text('AI sta analizzando...'), findsOneWidget);
      expect(find.text('Some instruction'), findsNothing);
    });

    testWidgets('transitions from processing to text display', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: GuidanceOverlay(text: null, isProcessing: true),
          ),
        ),
      );
      await tester.pump(const Duration(milliseconds: 400));
      expect(find.text('AI sta analizzando...'), findsOneWidget);

      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: GuidanceOverlay(text: 'Tighten the belt', isProcessing: false),
          ),
        ),
      );
      // Cross-fade: 350ms reverse + 350ms forward
      await tester.pump(const Duration(milliseconds: 900));
      expect(find.text('Tighten the belt'), findsOneWidget);
    });
  });
}
