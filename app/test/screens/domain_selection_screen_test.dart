// Widget tests for DomainSelectionScreen.
//
// The screen is a ConsumerWidget backed by settingsProvider (SharedPreferences).
// We override both sharedPrefsProvider and settingsProvider so no real I/O
// happens, following the same pattern as settings_screen_test.dart.
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:physicscopilot/providers/prefs_provider.dart' show sharedPrefsProvider;
import 'package:physicscopilot/providers/settings_provider.dart';
import 'package:physicscopilot/screens/domain_selection_screen.dart';

void main() {
  group('DomainSelectionScreen', () {
    Future<Widget> buildTestWidget({
      Map<String, Object> prefs = const {},
      void Function(String)? onSelected,
    }) async {
      SharedPreferences.setMockInitialValues(prefs);
      final sharedPrefs = await SharedPreferences.getInstance();

      return ProviderScope(
        overrides: [
          sharedPrefsProvider.overrideWithValue(sharedPrefs),
          settingsProvider.overrideWith(
            (ref) => SettingsNotifier(sharedPrefs),
          ),
        ],
        child: MaterialApp(
          home: DomainSelectionScreen(onSelected: onSelected),
        ),
      );
    }

    testWidgets('screen renders without crash', (tester) async {
      await tester.pumpWidget(await buildTestWidget());
      await tester.pump();
      expect(find.byType(DomainSelectionScreen), findsOneWidget);
    });

    testWidgets('shows at least one known domain label "Stampanti"',
        (tester) async {
      await tester.pumpWidget(await buildTestWidget());
      await tester.pump();
      expect(find.text('Stampanti'), findsOneWidget);
    });

    testWidgets('shows AppBar title "Seleziona dominio"', (tester) async {
      await tester.pumpWidget(await buildTestWidget());
      await tester.pump();
      expect(find.text('Seleziona dominio'), findsOneWidget);
    });

    testWidgets('tapping a domain tile calls onSelected with the domain id',
        (tester) async {
      String? selected;

      await tester.pumpWidget(
        await buildTestWidget(onSelected: (id) => selected = id),
      );
      await tester.pump();

      // Tap the first domain card ("Stampanti" → id "printer").
      await tester.tap(find.text('Stampanti'));
      await tester.pumpAndSettle();

      expect(selected, equals('printer'));
    });
  });
}
