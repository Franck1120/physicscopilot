// Widget tests for DomainSelectionScreen.
//
// The screen is a ConsumerWidget backed by settingsProvider (SharedPreferences).
// We override both sharedPrefsProvider and settingsProvider so no real I/O
// happens, following the same pattern as settings_screen_test.dart.
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:physicscopilot/models/session_record.dart';
import 'package:physicscopilot/providers/prefs_provider.dart' show sharedPrefsProvider;
import 'package:physicscopilot/providers/session_history_provider.dart';
import 'package:physicscopilot/providers/settings_provider.dart';
import 'package:physicscopilot/screens/domain_selection_screen.dart';

void main() {
  group('DomainSelectionScreen', () {
    Future<Widget> buildTestWidget({
      Map<String, Object> prefs = const {},
      List<SessionRecord> sessions = const [],
      void Function(String)? onSelected,
    }) async {
      final prefsMap = Map<String, Object>.from(prefs);
      if (sessions.isNotEmpty) {
        prefsMap['session_history'] = SessionRecord.encodeList(sessions);
      }
      SharedPreferences.setMockInitialValues(prefsMap);
      final sharedPrefs = await SharedPreferences.getInstance();

      return ProviderScope(
        overrides: [
          sharedPrefsProvider.overrideWithValue(sharedPrefs),
          settingsProvider.overrideWith(
            (ref) => SettingsNotifier(sharedPrefs),
          ),
          sessionHistoryProvider.overrideWith(
            (ref) => SessionHistoryNotifier(sharedPrefs),
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

    testWidgets('shows session count badge when domain has past sessions',
        (tester) async {
      // A session whose equipmentName contains the domain id "printer"
      // so _sessionCountForDomain returns 1 for the "printer" tile.
      final sessions = [
        SessionRecord(
          id: 'badge-1',
          date: DateTime(2026, 3, 10, 14, 0),
          equipmentName: 'printer',
          problemDescription: 'Paper jam',
          summary: 'Fixed paper jam.',
          status: SessionStatus.resolved,
          duration: const Duration(minutes: 5),
        ),
      ];

      await tester.pumpWidget(
        await buildTestWidget(sessions: sessions),
      );
      await tester.pump();

      // The badge chip shows the count "1" for the printer domain.
      expect(find.text('1'), findsWidgets);
    });
  });
}
