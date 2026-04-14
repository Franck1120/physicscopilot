import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:physicscopilot/widgets/streaming_text.dart';
// kAccent is used inside StreamingText; importing main.dart ensures the
// constant is available at link time (no explicit reference needed here).
// ignore: unused_import
import 'package:physicscopilot/main.dart' show kAccent;

void main() {
  Widget wrap(Widget child) =>
      MaterialApp(home: Scaffold(body: child));

  group('StreamingText', () {
    testWidgets('renders with empty text without throwing',
        (WidgetTester tester) async {
      await tester.pumpWidget(wrap(const StreamingText(text: '')));
      await tester.pumpAndSettle();

      expect(find.byType(StreamingText), findsOneWidget);
    });

    testWidgets('renders non-empty text visible in widget tree',
        (WidgetTester tester) async {
      const testText = 'Hello world';
      await tester.pumpWidget(wrap(const StreamingText(text: testText)));
      await tester.pumpAndSettle();

      expect(find.textContaining(testText), findsOneWidget);
    });

    testWidgets('shows Elaborazione… indicator',
        (WidgetTester tester) async {
      await tester.pumpWidget(
        wrap(const StreamingText(text: 'some text')),
      );
      await tester.pumpAndSettle();

      expect(find.text('Elaborazione…'), findsOneWidget);
    });

    testWidgets('find.textContaining locates the provided text',
        (WidgetTester tester) async {
      await tester.pumpWidget(
        wrap(const StreamingText(text: 'Hello world')),
      );
      await tester.pumpAndSettle();

      expect(find.textContaining('Hello world'), findsOneWidget);
    });
  });
}
