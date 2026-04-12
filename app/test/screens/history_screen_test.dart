// Widget tests for HistoryScreen.
//
// The screen depends on:
//   - sessionHistoryProvider  (backed by SharedPreferences)
//   - _serverSessionsProvider (private FutureProvider that calls
//     apiServiceProvider.listSessions)
//
// Strategy: override apiServiceProvider with an implementation whose
// listSessions() throws immediately. The screen's serverAsync.error branch
// already falls back gracefully to localSessions, so the tests only see the
// local-sessions path — which is all we need to verify UI correctness.
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:physicscopilot/models/session_record.dart';
import 'package:physicscopilot/providers/prefs_provider.dart';
import 'package:physicscopilot/providers/session_history_provider.dart';
import 'package:physicscopilot/screens/history_screen.dart';
import 'package:physicscopilot/services/api_service.dart';
import 'package:physicscopilot/utils/strings.dart';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

SessionRecord _makeRecord({
  String id = 'test-id-1',
  String equipmentName = 'Prusa MK4',
  String problemDescription = 'Layer shift',
  SessionStatus status = SessionStatus.resolved,
}) =>
    SessionRecord(
      id: id,
      date: DateTime(2025, 6, 15),
      equipmentName: equipmentName,
      problemDescription: problemDescription,
      summary: 'Layer shift caused by loose belt.',
      status: status,
      duration: const Duration(minutes: 20),
    );

/// Builds a [ProviderScope]-wrapped [HistoryScreen] with the given local
/// sessions pre-loaded in SharedPreferences.
///
/// The apiServiceProvider is overridden with an [ApiService] pointed at a
/// non-existent server so that _serverSessionsProvider's Future fails
/// fast and the screen falls back to [localSessions] via the error branch.
Future<Widget> buildHistoryScreen({
  List<SessionRecord> localSessions = const [],
}) async {
  SharedPreferences.setMockInitialValues(
    localSessions.isEmpty
        ? {}
        : {'session_history': SessionRecord.encodeList(localSessions)},
  );
  final prefs = await SharedPreferences.getInstance();

  return ProviderScope(
    overrides: [
      sharedPrefsProvider.overrideWithValue(prefs),
      sessionHistoryProvider.overrideWith(
        (ref) => SessionHistoryNotifier(prefs),
      ),
      // Point apiServiceProvider at a dead server; the screen's error
      // branch falls back to localSessions gracefully.
      apiServiceProvider.overrideWithValue(
        ApiService(baseUrl: 'http://127.0.0.1:19999'),
      ),
    ],
    child: const MaterialApp(
      home: HistoryScreen(),
    ),
  );
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

void main() {
  group('HistoryScreen', () {
    testWidgets('renders without crash', (tester) async {
      await tester.pumpWidget(await buildHistoryScreen());
      await tester.pump();
      expect(find.byType(HistoryScreen), findsOneWidget);
    });

    testWidgets('shows AppBar title from AppStrings.historyTitle', (tester) async {
      await tester.pumpWidget(await buildHistoryScreen());
      await tester.pump();
      expect(find.text(AppStrings.historyTitle), findsOneWidget);
    });

    testWidgets(
        'empty local sessions → shows AppStrings.historyEmpty message',
        (tester) async {
      await tester.pumpWidget(await buildHistoryScreen());
      // First pump renders with loading state (FutureProvider not yet settled).
      await tester.pump();
      // Second pump: FutureProvider error branch resolves; screen shows empty state.
      await tester.pump();
      expect(find.text(AppStrings.historyEmpty), findsOneWidget);
    });

    testWidgets('with local sessions → shows session item cards',
        (tester) async {
      final sessions = [
        _makeRecord(id: 'r1', equipmentName: 'Prusa MK4'),
        _makeRecord(id: 'r2', equipmentName: 'Bambu X1C'),
      ];
      await tester.pumpWidget(await buildHistoryScreen(localSessions: sessions));
      await tester.pump();
      await tester.pump();

      expect(find.text('Prusa MK4'), findsOneWidget);
      expect(find.text('Bambu X1C'), findsOneWidget);
    });

    testWidgets('with local sessions → does NOT show empty state', (tester) async {
      final sessions = [_makeRecord()];
      await tester.pumpWidget(await buildHistoryScreen(localSessions: sessions));
      await tester.pump();
      await tester.pump();

      expect(find.text(AppStrings.historyEmpty), findsNothing);
    });

    testWidgets(
        'with local sessions → shows delete sweep icon button in AppBar',
        (tester) async {
      final sessions = [_makeRecord()];
      await tester.pumpWidget(await buildHistoryScreen(localSessions: sessions));
      await tester.pump();
      await tester.pump();

      expect(find.byIcon(Icons.delete_sweep_outlined), findsOneWidget);
    });

    testWidgets(
        'with no sessions → does NOT show delete sweep icon button',
        (tester) async {
      await tester.pumpWidget(await buildHistoryScreen());
      await tester.pump();
      await tester.pump();

      expect(find.byIcon(Icons.delete_sweep_outlined), findsNothing);
    });
  });

  // ---------------------------------------------------------------------------
  // Search filtering tests
  // ---------------------------------------------------------------------------

  group('HistoryScreen — search filtering', () {
    testWidgets('search bar TextField is present', (tester) async {
      final sessions = [_makeRecord()];
      await tester.pumpWidget(await buildHistoryScreen(localSessions: sessions));
      await tester.pump();
      await tester.pump();

      expect(find.byType(TextField), findsWidgets);
    });

    testWidgets('typing in search bar shows matching session', (tester) async {
      final sessions = [
        _makeRecord(id: 'r1', equipmentName: 'Prusa MK4'),
        _makeRecord(id: 'r2', equipmentName: 'Bambu X1C'),
      ];
      await tester.pumpWidget(await buildHistoryScreen(localSessions: sessions));
      await tester.pump();
      await tester.pump();

      // Both sessions visible initially.
      expect(find.text('Prusa MK4'), findsOneWidget);
      expect(find.text('Bambu X1C'), findsOneWidget);

      // Type in search bar.
      await tester.enterText(find.byType(TextField).first, 'Prusa');
      await tester.pump();

      // Only the matching session should be visible.
      expect(find.text('Prusa MK4'), findsOneWidget);
      expect(find.text('Bambu X1C'), findsNothing);
    });

    testWidgets('clearing search bar restores full list', (tester) async {
      final sessions = [
        _makeRecord(id: 'r1', equipmentName: 'Prusa MK4'),
        _makeRecord(id: 'r2', equipmentName: 'Bambu X1C'),
      ];
      await tester.pumpWidget(await buildHistoryScreen(localSessions: sessions));
      await tester.pump();
      await tester.pump();

      // Filter to one result.
      await tester.enterText(find.byType(TextField).first, 'Bambu');
      await tester.pump();

      expect(find.text('Prusa MK4'), findsNothing);
      expect(find.text('Bambu X1C'), findsOneWidget);

      // Clear the search.
      await tester.enterText(find.byType(TextField).first, '');
      await tester.pump();

      expect(find.text('Prusa MK4'), findsOneWidget);
      expect(find.text('Bambu X1C'), findsOneWidget);
    });

    testWidgets('search with no match shows empty state', (tester) async {
      final sessions = [_makeRecord(equipmentName: 'Prusa MK4')];
      await tester.pumpWidget(await buildHistoryScreen(localSessions: sessions));
      await tester.pump();
      await tester.pump();

      await tester.enterText(find.byType(TextField).first, 'zzznomatch');
      await tester.pump();

      expect(find.text(AppStrings.historyEmpty), findsOneWidget);
    });
  });
}
