// Widget tests for SessionDetailScreen.
//
// Since _formatDate, _formatDuration, and _buildShareText are private instance
// methods on the StatelessWidget, we exercise them through the rendered widget
// tree: pump a SessionDetailScreen with a known SessionRecord and verify the
// text that gets rendered.
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:physicscopilot/models/session_record.dart';
import 'package:physicscopilot/screens/session_detail_screen.dart';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

Widget _wrap(SessionRecord session) => MaterialApp(
      home: SessionDetailScreen(session: session),
    );

SessionRecord _makeRecord({
  String id = 'test-id',
  String equipmentName = 'Prusa MK4',
  String problemDescription = 'Layer shift',
  String summary = 'Belt tension resolved the issue.',
  SessionStatus status = SessionStatus.resolved,
  DateTime? date,
  Duration duration = const Duration(minutes: 5),
}) =>
    SessionRecord(
      id: id,
      date: date ?? DateTime(2026, 6, 15, 10, 30),
      equipmentName: equipmentName,
      problemDescription: problemDescription,
      summary: summary,
      status: status,
      duration: duration,
    );

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

void main() {
  group('SessionDetailScreen — _formatDate', () {
    testWidgets('formats January with zero-padded minutes (gen)', (tester) async {
      final session = _makeRecord(date: DateTime(2026, 1, 5, 9, 3));
      await tester.pumpWidget(_wrap(session));
      await tester.pump();
      // Expected: "5 gen 2026, 09:03"
      expect(find.text('5 gen 2026, 09:03'), findsOneWidget);
    });

    testWidgets('formats December correctly (dic)', (tester) async {
      final session = _makeRecord(date: DateTime(2026, 12, 31, 23, 59));
      await tester.pumpWidget(_wrap(session));
      await tester.pump();
      expect(find.text('31 dic 2026, 23:59'), findsOneWidget);
    });

    testWidgets('formats all 12 Italian month abbreviations — spot check mar, giu, set',
        (tester) async {
      for (final entry in {
        3: '15 mar 2026, 12:00',
        6: '15 giu 2026, 12:00',
        9: '15 set 2026, 12:00',
      }.entries) {
        final session = _makeRecord(date: DateTime(2026, entry.key, 15, 12, 0));
        await tester.pumpWidget(_wrap(session));
        await tester.pump();
        expect(
          find.text(entry.value),
          findsOneWidget,
          reason: 'Month ${entry.key} should render as "${entry.value}"',
        );
      }
    });

    testWidgets('zero-pads hours and minutes correctly', (tester) async {
      final session = _makeRecord(date: DateTime(2026, 3, 1, 0, 0));
      await tester.pumpWidget(_wrap(session));
      await tester.pump();
      expect(find.text('1 mar 2026, 00:00'), findsOneWidget);
    });
  });

  // -------------------------------------------------------------------------

  group('SessionDetailScreen — _formatDuration', () {
    testWidgets('45 seconds → "45s" (no minutes)', (tester) async {
      final session = _makeRecord(duration: const Duration(seconds: 45));
      await tester.pumpWidget(_wrap(session));
      await tester.pump();
      expect(find.text('45s'), findsOneWidget);
    });

    testWidgets('90 seconds → "1m 30s"', (tester) async {
      final session = _makeRecord(duration: const Duration(seconds: 90));
      await tester.pumpWidget(_wrap(session));
      await tester.pump();
      expect(find.text('1m 30s'), findsOneWidget);
    });

    testWidgets('2 minutes 5 seconds → "2m 05s" (zero-padded seconds)',
        (tester) async {
      final session =
          _makeRecord(duration: const Duration(minutes: 2, seconds: 5));
      await tester.pumpWidget(_wrap(session));
      await tester.pump();
      expect(find.text('2m 05s'), findsOneWidget);
    });

    testWidgets('Duration.zero → "0s"', (tester) async {
      final session = _makeRecord(duration: Duration.zero);
      await tester.pumpWidget(_wrap(session));
      await tester.pump();
      expect(find.text('0s'), findsOneWidget);
    });

    testWidgets('exactly 60 seconds → "1m 00s"', (tester) async {
      final session = _makeRecord(duration: const Duration(seconds: 60));
      await tester.pumpWidget(_wrap(session));
      await tester.pump();
      expect(find.text('1m 00s'), findsOneWidget);
    });

    testWidgets('59 seconds stays as seconds (boundary below minute)',
        (tester) async {
      final session = _makeRecord(duration: const Duration(seconds: 59));
      await tester.pumpWidget(_wrap(session));
      await tester.pump();
      expect(find.text('59s'), findsOneWidget);
    });
  });

  // -------------------------------------------------------------------------

  group('SessionDetailScreen — problem & summary sections', () {
    testWidgets('non-empty problemDescription renders its text', (tester) async {
      final session = _makeRecord(problemDescription: 'Extruder clicking noise');
      await tester.pumpWidget(_wrap(session));
      await tester.pump();
      expect(find.text('Extruder clicking noise'), findsOneWidget);
    });

    testWidgets('empty problemDescription → shows "—"', (tester) async {
      final session = _makeRecord(problemDescription: '');
      await tester.pumpWidget(_wrap(session));
      await tester.pump();
      // Both problem and summary use "—" for empty — filter by finding at least one.
      expect(find.text('—'), findsWidgets);
    });

    testWidgets('empty summary → shows "—"', (tester) async {
      final session = _makeRecord(summary: '');
      await tester.pumpWidget(_wrap(session));
      await tester.pump();
      expect(find.text('—'), findsWidgets);
    });

    testWidgets('non-empty summary renders its text', (tester) async {
      final session = _makeRecord(summary: 'Replaced PTFE tube, issue resolved.');
      await tester.pumpWidget(_wrap(session));
      await tester.pump();
      expect(find.text('Replaced PTFE tube, issue resolved.'), findsOneWidget);
    });
  });

  // -------------------------------------------------------------------------

  group('SessionDetailScreen — AppBar title (equipmentName edge case)', () {
    testWidgets('non-empty equipmentName shows as AppBar title', (tester) async {
      final session = _makeRecord(equipmentName: 'Bambu X1C');
      await tester.pumpWidget(_wrap(session));
      await tester.pump();
      expect(find.text('Bambu X1C'), findsWidgets);
    });

    testWidgets('empty equipmentName → AppBar shows "Sessione"', (tester) async {
      final session = _makeRecord(equipmentName: '');
      await tester.pumpWidget(_wrap(session));
      await tester.pump();
      expect(find.text('Sessione'), findsOneWidget);
    });
  });

  // -------------------------------------------------------------------------

  group('SessionDetailScreen — status badge', () {
    testWidgets('resolved session shows "Risolto" badge', (tester) async {
      final session = _makeRecord(status: SessionStatus.resolved);
      await tester.pumpWidget(_wrap(session));
      await tester.pump();
      expect(find.text('Risolto'), findsOneWidget);
    });

    testWidgets('unresolved session shows "Non risolto" badge', (tester) async {
      final session = _makeRecord(status: SessionStatus.unresolved);
      await tester.pumpWidget(_wrap(session));
      await tester.pump();
      expect(find.text('Non risolto'), findsOneWidget);
    });
  });

  // -------------------------------------------------------------------------

  group('SessionDetailScreen — share button present', () {
    testWidgets('share icon button is in the AppBar', (tester) async {
      final session = _makeRecord();
      await tester.pumpWidget(_wrap(session));
      await tester.pump();
      expect(find.byIcon(Icons.share_rounded), findsOneWidget);
    });
  });
}
