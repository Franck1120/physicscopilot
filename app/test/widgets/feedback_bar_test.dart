import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:physicscopilot/providers/prefs_provider.dart';
import 'package:physicscopilot/widgets/feedback_bar.dart';

Widget _buildSubject(SharedPreferences prefs) => ProviderScope(
      overrides: [sharedPrefsProvider.overrideWithValue(prefs)],
      child: const MaterialApp(
        home: Scaffold(
          body: FeedbackBar(responseText: 'test response'),
        ),
      ),
    );

void main() {
  group('FeedbackBar widget', () {
    late SharedPreferences prefs;

    setUp(() async {
      SharedPreferences.setMockInitialValues({});
      prefs = await SharedPreferences.getInstance();
    });

    testWidgets('initially both thumbs are outlined (not selected)',
        (tester) async {
      await tester.pumpWidget(_buildSubject(prefs));

      expect(find.byIcon(Icons.thumb_up_outlined), findsOneWidget);
      expect(find.byIcon(Icons.thumb_down_outlined), findsOneWidget);
      // Rounded (selected) variants should NOT be present
      expect(find.byIcon(Icons.thumb_up_rounded), findsNothing);
      expect(find.byIcon(Icons.thumb_down_rounded), findsNothing);
    });

    testWidgets('tap thumbs-up → thumb_up_rounded shown, thumb_down still outlined',
        (tester) async {
      await tester.pumpWidget(_buildSubject(prefs));

      await tester.tap(find.byIcon(Icons.thumb_up_outlined));
      await tester.pump();

      expect(find.byIcon(Icons.thumb_up_rounded), findsOneWidget);
      expect(find.byIcon(Icons.thumb_down_outlined), findsOneWidget);
      // Rounded down should not appear
      expect(find.byIcon(Icons.thumb_down_rounded), findsNothing);
    });

    testWidgets('after voting thumbs-up, tapping again does nothing (vote locked)',
        (tester) async {
      await tester.pumpWidget(_buildSubject(prefs));

      await tester.tap(find.byIcon(Icons.thumb_up_outlined));
      await tester.pump();

      // Both buttons are disabled after voting — tapping should not change state
      // The thumbs-up button's onPressed should now be null
      final thumbUpBtn = tester.widget<IconButton>(
        find.widgetWithIcon(IconButton, Icons.thumb_up_rounded),
      );
      expect(thumbUpBtn.onPressed, isNull);

      // Verify icons remain unchanged
      expect(find.byIcon(Icons.thumb_up_rounded), findsOneWidget);
      expect(find.byIcon(Icons.thumb_down_outlined), findsOneWidget);
    });

    testWidgets('tap thumbs-down → thumb_down_rounded shown, thumb_up still outlined',
        (tester) async {
      await tester.pumpWidget(_buildSubject(prefs));

      await tester.tap(find.byIcon(Icons.thumb_down_outlined));
      await tester.pump();

      expect(find.byIcon(Icons.thumb_down_rounded), findsOneWidget);
      expect(find.byIcon(Icons.thumb_up_outlined), findsOneWidget);
      // Rounded up should not appear
      expect(find.byIcon(Icons.thumb_up_rounded), findsNothing);
    });

    testWidgets('after voting thumbs-down, tapping again does nothing (vote locked)',
        (tester) async {
      await tester.pumpWidget(_buildSubject(prefs));

      await tester.tap(find.byIcon(Icons.thumb_down_outlined));
      await tester.pump();

      final thumbDownBtn = tester.widget<IconButton>(
        find.widgetWithIcon(IconButton, Icons.thumb_down_rounded),
      );
      expect(thumbDownBtn.onPressed, isNull);

      expect(find.byIcon(Icons.thumb_down_rounded), findsOneWidget);
      expect(find.byIcon(Icons.thumb_up_outlined), findsOneWidget);
    });
  });
}
