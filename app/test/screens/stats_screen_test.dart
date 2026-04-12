// Widget tests for StatsScreen.
//
// The screen reads from sessionHistoryProvider (SharedPreferences-backed) and
// earnedMilestonesProvider (derived). We override sharedPrefsProvider so no
// real I/O happens.
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:physicscopilot/providers/prefs_provider.dart'
    show sharedPrefsProvider;
import 'package:physicscopilot/providers/session_history_provider.dart';
import 'package:physicscopilot/models/session_record.dart';
import 'package:physicscopilot/screens/stats_screen.dart';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/// Builds a minimal [ProviderScope]-wrapped [StatsScreen] with SharedPreferences
/// pre-populated with [records].
Future<Widget> _buildTestWidget(List<SessionRecord> records) async {
  final encoded = SessionRecord.encodeList(records);
  SharedPreferences.setMockInitialValues({'session_history': encoded});
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

/// Generates [n] identical dummy [SessionRecord] objects.
List<SessionRecord> _records(int n) => List.generate(
      n,
      (i) => SessionRecord(
        id: 'id-$i',
        date: DateTime(2026, 1, i + 1),
        equipmentName: 'Dispositivo',
        problemDescription: 'Problema',
        summary: 'Riassunto',
        status: i.isEven ? SessionStatus.resolved : SessionStatus.unresolved,
        duration: const Duration(minutes: 5),
      ),
    );

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

void main() {
  group('StatsScreen — empty state', () {
    testWidgets('shows Nessuna sessione when history is empty', (tester) async {
      await tester.pumpWidget(await _buildTestWidget([]));
      await tester.pump();
      // _EmptyStats shows AppStrings.historyEmpty which contains "Nessuna".
      expect(find.textContaining('Nessuna'), findsOneWidget);
    });
  });

  group('StatsScreen — with sessions', () {
    testWidgets('shows total session count', (tester) async {
      await tester.pumpWidget(await _buildTestWidget(_records(3)));
      await tester.pump();
      expect(find.text('3'), findsAtLeast(1));
    });

    testWidgets('shows Sessioni totali label', (tester) async {
      await tester.pumpWidget(await _buildTestWidget(_records(3)));
      await tester.pump();
      expect(find.text('Sessioni totali'), findsOneWidget);
    });

    testWidgets('shows Risoluzione section', (tester) async {
      await tester.pumpWidget(await _buildTestWidget(_records(4)));
      await tester.pump();
      expect(find.text('Risoluzione'), findsAtLeast(1));
    });

    testWidgets('shows Ultime sessioni section', (tester) async {
      await tester.pumpWidget(await _buildTestWidget(_records(3)));
      await tester.pump();
      expect(find.text('Ultime sessioni'), findsOneWidget);
    });
  });

  group('StatsScreen — milestones', () {
    testWidgets('no milestone section when fewer than 5 sessions', (tester) async {
      await tester.pumpWidget(await _buildTestWidget(_records(4)));
      await tester.pump();
      expect(find.text('Traguardi raggiunti'), findsNothing);
    });

    testWidgets('shows milestone section when 5+ sessions reached', (tester) async {
      await tester.pumpWidget(await _buildTestWidget(_records(5)));
      await tester.pump();
      expect(find.text('Traguardi raggiunti'), findsOneWidget);
      expect(find.text('5 sessioni'), findsOneWidget);
    });

    testWidgets('shows two milestones when 10+ sessions reached', (tester) async {
      await tester.pumpWidget(await _buildTestWidget(_records(10)));
      await tester.pump();
      expect(find.text('5 sessioni'), findsOneWidget);
      expect(find.text('10 sessioni'), findsOneWidget);
    });
  });
}
