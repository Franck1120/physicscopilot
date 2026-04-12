// Widget tests for StatsScreen.
//
// StatsScreen is a ConsumerWidget that reads sessionHistoryProvider.
// We override the provider via ProviderScope so tests are fully self-contained
// (no SharedPreferences, no network).
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:physicscopilot/models/session_record.dart';
import 'package:physicscopilot/providers/prefs_provider.dart';
import 'package:physicscopilot/providers/session_history_provider.dart';
import 'package:physicscopilot/screens/stats_screen.dart';
import 'package:physicscopilot/utils/strings.dart';
import 'package:physicscopilot/widgets/achievement_badges.dart';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

SessionRecord _makeRecord({
  String id = 'id-1',
  String equipmentName = 'Prusa MK4',
  SessionStatus status = SessionStatus.resolved,
}) =>
    SessionRecord(
      id: id,
      date: DateTime(2026, 3, 10, 14, 0),
      equipmentName: equipmentName,
      problemDescription: 'Test problem',
      summary: 'Test summary',
      status: status,
      duration: const Duration(minutes: 10),
    );

/// Builds a [ProviderScope]-wrapped [StatsScreen] with [sessions] pre-loaded.
Future<Widget> _buildStatsScreen(List<SessionRecord> sessions) async {
  SharedPreferences.setMockInitialValues(
    sessions.isEmpty
        ? {}
        : {'session_history': SessionRecord.encodeList(sessions)},
  );
  final prefs = await SharedPreferences.getInstance();

  return ProviderScope(
    overrides: [
      sharedPrefsProvider.overrideWithValue(prefs),
      sessionHistoryProvider.overrideWith(
        (ref) => SessionHistoryNotifier(prefs),
      ),
    ],
    child: const MaterialApp(home: StatsScreen()),
  );
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

void main() {
  group('StatsScreen — empty state', () {
    testWidgets('empty session list → renders without crash', (tester) async {
      await tester.pumpWidget(await _buildStatsScreen([]));
      await tester.pump();
      expect(find.byType(StatsScreen), findsOneWidget);
    });

    testWidgets('empty session list → shows empty-state text', (tester) async {
      await tester.pumpWidget(await _buildStatsScreen([]));
      await tester.pump();
      expect(find.text(AppStrings.historyEmpty), findsOneWidget);
    });

    testWidgets(
        'empty session list → does NOT show AchievementBadgesWidget',
        (tester) async {
      await tester.pumpWidget(await _buildStatsScreen([]));
      await tester.pump();
      expect(find.byType(AchievementBadgesWidget), findsNothing);
    });
  });

  // -------------------------------------------------------------------------

  group('StatsScreen — with sessions', () {
    testWidgets('sessions present → does NOT show empty-state text',
        (tester) async {
      final sessions = [_makeRecord()];
      await tester.pumpWidget(await _buildStatsScreen(sessions));
      await tester.pump();
      expect(find.text(AppStrings.historyEmpty), findsNothing);
    });

    testWidgets('sessions present → shows AchievementBadgesWidget',
        (tester) async {
      final sessions = [_makeRecord()];
      await tester.pumpWidget(await _buildStatsScreen(sessions));
      await tester.pump();
      expect(find.byType(AchievementBadgesWidget), findsOneWidget);
    });

    testWidgets(
        '3 sessions → AchievementBadgesWidget receives sessionCount=3',
        (tester) async {
      final sessions = [
        _makeRecord(id: 'a', equipmentName: 'Printer A'),
        _makeRecord(id: 'b', equipmentName: 'Printer B'),
        _makeRecord(id: 'c', equipmentName: 'Printer C'),
      ];
      await tester.pumpWidget(await _buildStatsScreen(sessions));
      await tester.pump();

      final badgesWidget = tester.widget<AchievementBadgesWidget>(
        find.byType(AchievementBadgesWidget),
      );
      expect(badgesWidget.sessionCount, equals(3));
    });

    testWidgets('5 sessions → AchievementBadgesWidget receives sessionCount=5',
        (tester) async {
      final sessions = List.generate(
        5,
        (i) => _makeRecord(id: 'id-$i', equipmentName: 'Device $i'),
      );
      await tester.pumpWidget(await _buildStatsScreen(sessions));
      await tester.pump();

      final badgesWidget = tester.widget<AchievementBadgesWidget>(
        find.byType(AchievementBadgesWidget),
      );
      expect(badgesWidget.sessionCount, equals(5));
    });

    testWidgets('equipment name appears somewhere in the widget tree',
        (tester) async {
      // With a single session, its equipmentName surfaces in both the "top
      // domain" badge and the recent-sessions list — at least one match is
      // guaranteed to be rendered.
      final sessions = [_makeRecord(id: 'r1', equipmentName: 'Bambu X1C')];
      await tester.pumpWidget(await _buildStatsScreen(sessions));
      await tester.pump();

      expect(find.text('Bambu X1C'), findsWidgets);
    });

    testWidgets('shows total duration for multiple sessions', (tester) async {
      // Two sessions: 1 minute each → total 2 minutes.
      final sessions = [
        SessionRecord(
          id: 'td-1',
          date: DateTime(2026, 3, 10, 14, 0),
          equipmentName: 'Printer A',
          problemDescription: 'Problem A',
          summary: 'Summary A',
          status: SessionStatus.resolved,
          duration: const Duration(minutes: 1),
        ),
        SessionRecord(
          id: 'td-2',
          date: DateTime(2026, 3, 11, 14, 0),
          equipmentName: 'Printer B',
          problemDescription: 'Problem B',
          summary: 'Summary B',
          status: SessionStatus.resolved,
          duration: const Duration(minutes: 1),
        ),
      ];
      await tester.pumpWidget(await _buildStatsScreen(sessions));
      await tester.pump();

      // Total = 2 minutes 0 seconds → "2m 00s".
      expect(find.text('2m 00s'), findsOneWidget);
    });

    testWidgets(
        'session with empty equipmentName → "Sessione" fallback in recent list',
        (tester) async {
      // Empty equipmentName means the top-domain section is suppressed (domain
      // name is ''). The _RecentSessionRow fallback renders "Sessione".
      // We scroll the ListView to ensure the row is built.
      final sessions = [_makeRecord(id: 'x', equipmentName: '')];
      await tester.pumpWidget(await _buildStatsScreen(sessions));
      await tester.pump();

      // Scroll to bottom to ensure the recent-sessions section is rendered.
      await tester.dragFrom(const Offset(200, 400), const Offset(200, -800));
      await tester.pump();

      expect(find.text('Sessione'), findsOneWidget);
    });
  });

  // -------------------------------------------------------------------------

  group('StatsScreen — pull-to-refresh', () {
    testWidgets('RefreshIndicator is present in the widget tree (empty state)',
        (tester) async {
      await tester.pumpWidget(await _buildStatsScreen([]));
      await tester.pump();
      expect(find.byType(RefreshIndicator), findsOneWidget);
    });

    testWidgets(
        'RefreshIndicator is present in the widget tree (with sessions)',
        (tester) async {
      final sessions = [_makeRecord()];
      await tester.pumpWidget(await _buildStatsScreen(sessions));
      await tester.pump();
      expect(find.byType(RefreshIndicator), findsOneWidget);
    });
  });
}
