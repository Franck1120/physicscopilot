// Widget tests for screenError() in safe_screen.dart.
//
// screenError() is a free function that returns a Scaffold with an error
// message and a "Vai alla home" button.  Tests verify the expected UI
// elements appear and that long messages are truncated at 180 chars.
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:physicscopilot/widgets/safe_screen.dart';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/// Wraps [screenError] in a minimal MaterialApp so Navigator is available.
Widget _buildErrorScreen(Object error) {
  return MaterialApp(
    home: Builder(
      builder: (context) => screenError(error, context),
    ),
  );
}

void main() {
  group('screenError()', () {
    testWidgets('renders without crash', (tester) async {
      await tester.pumpWidget(_buildErrorScreen(Exception('test error')));
      await tester.pump();
      expect(find.byType(Scaffold), findsOneWidget);
    });

    testWidgets('shows "Qualcosa è andato storto" heading', (tester) async {
      await tester.pumpWidget(_buildErrorScreen(Exception('test error')));
      await tester.pump();
      expect(find.text('Qualcosa è andato storto'), findsOneWidget);
    });

    testWidgets('shows the error message text', (tester) async {
      await tester.pumpWidget(
          _buildErrorScreen(Exception('something broke badly')));
      await tester.pump();
      expect(find.textContaining('something broke badly'), findsOneWidget);
    });

    testWidgets('shows warning icon', (tester) async {
      await tester.pumpWidget(_buildErrorScreen(Exception('err')));
      await tester.pump();
      expect(find.byIcon(Icons.warning_amber_rounded), findsOneWidget);
    });

    testWidgets('shows "Vai alla home" button', (tester) async {
      await tester.pumpWidget(_buildErrorScreen(Exception('err')));
      await tester.pump();
      expect(find.text('Vai alla home'), findsOneWidget);
    });

    testWidgets('truncates error message longer than 180 chars with ellipsis',
        (tester) async {
      final longMessage = 'A' * 200; // 200 chars > 180
      await tester.pumpWidget(_buildErrorScreen(Exception(longMessage)));
      await tester.pump();

      // The preview must end with '…' and be at most 181 visible chars
      // (180 from message + ellipsis character).
      final textWidget = tester.widget<Text>(
        find
            .byWidgetPredicate(
              (w) => w is Text && (w.data ?? '').endsWith('…'),
            )
            .first,
      );
      final displayed = textWidget.data ?? '';
      expect(displayed.endsWith('…'), isTrue);
      // Displayed text is Exception prefix + 180 chars + '…'
      expect(displayed.contains('A' * 10), isTrue);
    });

    testWidgets('does not truncate error message of exactly 180 chars',
        (tester) async {
      final exactMessage = 'B' * 180;
      await tester.pumpWidget(_buildErrorScreen(Exception(exactMessage)));
      await tester.pump();
      // No truncation — full message visible, no trailing '…'.
      expect(find.textContaining(exactMessage), findsOneWidget);
    });

    testWidgets('does not show "Torna indietro" when navigator cannot pop',
        (tester) async {
      // Root route → canPop() == false → OutlinedButton not rendered.
      await tester.pumpWidget(_buildErrorScreen(Exception('err')));
      await tester.pump();
      expect(find.text('Torna indietro'), findsNothing);
    });

    testWidgets('"Vai alla home" button triggers Navigator.popUntil',
        (tester) async {
      // Push a second route so "Vai alla home" has something to pop to.
      await tester.pumpWidget(
        MaterialApp(
          home: Builder(
            builder: (context) => ElevatedButton(
              onPressed: () => Navigator.of(context).push(
                MaterialPageRoute<void>(
                  builder: (ctx) => screenError(Exception('err'), ctx),
                ),
              ),
              child: const Text('Go'),
            ),
          ),
        ),
      );
      await tester.tap(find.text('Go'));
      await tester.pumpAndSettle();

      // Now on the error screen — tap "Vai alla home".
      await tester.tap(find.text('Vai alla home'));
      await tester.pumpAndSettle();

      // Should have navigated back to the first route.
      expect(find.text('Go'), findsOneWidget);
    });

    testWidgets('shows "Torna indietro" when navigator can pop', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Builder(
            builder: (context) => ElevatedButton(
              onPressed: () => Navigator.of(context).push(
                MaterialPageRoute<void>(
                  builder: (ctx) => screenError(Exception('err'), ctx),
                ),
              ),
              child: const Text('Go'),
            ),
          ),
        ),
      );
      await tester.tap(find.text('Go'));
      await tester.pumpAndSettle();
      expect(find.text('Torna indietro'), findsOneWidget);
    });

    testWidgets('displays string representation of arbitrary objects',
        (tester) async {
      await tester.pumpWidget(_buildErrorScreen('plain string error'));
      await tester.pump();
      expect(find.textContaining('plain string error'), findsOneWidget);
    });
  });
}
