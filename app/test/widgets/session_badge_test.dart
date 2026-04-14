// Widget tests for SessionBadge.
//
// SessionBadge is a pure StatelessWidget that reads a SessionStatus enum and
// renders a coloured pill with a label from AppStrings.
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:physicscopilot/models/session_record.dart';
import 'package:physicscopilot/widgets/session_badge.dart';

void main() {
  group('SessionBadge', () {
    Widget wrap(Widget child) => MaterialApp(home: Scaffold(body: child));

    testWidgets('resolved status shows "Risolto" text', (tester) async {
      await tester.pumpWidget(
        wrap(const SessionBadge(status: SessionStatus.resolved)),
      );
      await tester.pump();
      expect(find.text('Risolto'), findsOneWidget);
    });

    testWidgets('unresolved status shows "Non risolto" text', (tester) async {
      await tester.pumpWidget(
        wrap(const SessionBadge(status: SessionStatus.unresolved)),
      );
      await tester.pump();
      expect(find.text('Non risolto'), findsOneWidget);
    });

    testWidgets('resolved variant renders without error', (tester) async {
      await tester.pumpWidget(
        wrap(const SessionBadge(status: SessionStatus.resolved)),
      );
      await tester.pump();
      expect(find.byType(SessionBadge), findsOneWidget);
    });

    testWidgets('unresolved variant renders without error', (tester) async {
      await tester.pumpWidget(
        wrap(const SessionBadge(status: SessionStatus.unresolved)),
      );
      await tester.pump();
      expect(find.byType(SessionBadge), findsOneWidget);
    });
  });
}
