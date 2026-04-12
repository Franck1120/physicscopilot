// Widget tests for AchievementBadgesWidget.
//
// Milestones are [5, 10, 25, 50, 100].
// A badge is "unlocked" when sessionCount >= milestone: it renders a ⭐ emoji.
// A locked badge renders a lock icon (Icons.lock_outline_rounded) instead.
//
// Strategy: count ⭐ text widgets and lock-icon widgets in the tree to assert
// exactly how many badges are unlocked/locked for a given sessionCount.
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:physicscopilot/widgets/achievement_badges.dart';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

Widget _wrap(int sessionCount) => MaterialApp(
      home: Scaffold(
        body: AchievementBadgesWidget(sessionCount: sessionCount),
      ),
    );

// Count how many ⭐ emoji text widgets are in the tree.
int _starCount(WidgetTester tester) =>
    tester.widgetList(find.text('⭐')).length;

// Count how many lock icons are in the tree.
int _lockCount(WidgetTester tester) =>
    tester
        .widgetList(find.byIcon(Icons.lock_outline_rounded))
        .length;

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

void main() {
  group('AchievementBadgesWidget — total badge count', () {
    testWidgets('always renders exactly 5 badges regardless of sessionCount',
        (tester) async {
      for (final count in [0, 4, 5, 10, 100, 999]) {
        await tester.pumpWidget(_wrap(count));
        await tester.pump();
        final total = _starCount(tester) + _lockCount(tester);
        expect(
          total,
          equals(5),
          reason: 'sessionCount=$count should yield 5 badges total',
        );
      }
    });
  });

  // -------------------------------------------------------------------------

  group('AchievementBadgesWidget — boundary conditions', () {
    testWidgets('sessionCount=0 → all 5 badges locked, zero stars',
        (tester) async {
      await tester.pumpWidget(_wrap(0));
      await tester.pump();
      expect(_starCount(tester), equals(0));
      expect(_lockCount(tester), equals(5));
    });

    testWidgets('sessionCount=4 → all locked (one below first milestone 5)',
        (tester) async {
      await tester.pumpWidget(_wrap(4));
      await tester.pump();
      expect(_starCount(tester), equals(0));
      expect(_lockCount(tester), equals(5));
    });

    testWidgets('sessionCount=5 → exactly 1 badge unlocked (exact boundary)',
        (tester) async {
      await tester.pumpWidget(_wrap(5));
      await tester.pump();
      expect(_starCount(tester), equals(1));
      expect(_lockCount(tester), equals(4));
    });

    testWidgets('sessionCount=6 → still only 1 badge unlocked (above first, below second)',
        (tester) async {
      await tester.pumpWidget(_wrap(6));
      await tester.pump();
      expect(_starCount(tester), equals(1));
      expect(_lockCount(tester), equals(4));
    });

    testWidgets('sessionCount=9 → still only 1 badge unlocked (one below milestone 10)',
        (tester) async {
      await tester.pumpWidget(_wrap(9));
      await tester.pump();
      expect(_starCount(tester), equals(1));
      expect(_lockCount(tester), equals(4));
    });

    testWidgets('sessionCount=10 → first 2 badges unlocked (exact boundary)',
        (tester) async {
      await tester.pumpWidget(_wrap(10));
      await tester.pump();
      expect(_starCount(tester), equals(2));
      expect(_lockCount(tester), equals(3));
    });

    testWidgets('sessionCount=25 → first 3 badges unlocked (exact boundary)',
        (tester) async {
      await tester.pumpWidget(_wrap(25));
      await tester.pump();
      expect(_starCount(tester), equals(3));
      expect(_lockCount(tester), equals(2));
    });

    testWidgets('sessionCount=50 → first 4 badges unlocked (exact boundary)',
        (tester) async {
      await tester.pumpWidget(_wrap(50));
      await tester.pump();
      expect(_starCount(tester), equals(4));
      expect(_lockCount(tester), equals(1));
    });

    testWidgets('sessionCount=100 → all 5 badges unlocked (exact boundary)',
        (tester) async {
      await tester.pumpWidget(_wrap(100));
      await tester.pump();
      expect(_starCount(tester), equals(5));
      expect(_lockCount(tester), equals(0));
    });

    testWidgets('sessionCount=101 → all 5 unlocked (beyond last milestone)',
        (tester) async {
      await tester.pumpWidget(_wrap(101));
      await tester.pump();
      expect(_starCount(tester), equals(5));
      expect(_lockCount(tester), equals(0));
    });
  });

  // -------------------------------------------------------------------------

  group('AchievementBadgesWidget — milestone label display', () {
    testWidgets('renders milestone numbers as text labels (5, 10, 25, 50, 100)',
        (tester) async {
      await tester.pumpWidget(_wrap(0));
      await tester.pump();
      // Each milestone value appears as a label below the badge circle.
      // Note: the unlocked badge also shows the number inside the circle
      // so with sessionCount=0 each milestone number appears exactly once
      // (only in the label row).
      for (final label in ['5', '10', '25', '50', '100']) {
        expect(
          find.text(label),
          findsWidgets,
          reason: 'milestone label "$label" should be present',
        );
      }
    });

    testWidgets('unlocked badge shows ⭐ symbol, locked badge does not',
        (tester) async {
      // With sessionCount=5, first badge unlocked, rest locked.
      await tester.pumpWidget(_wrap(5));
      await tester.pump();
      expect(find.text('⭐'), findsOneWidget);
      expect(find.byIcon(Icons.lock_outline_rounded), findsNWidgets(4));
    });
  });
}
