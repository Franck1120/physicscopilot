// Widget tests for RateLimitCountdown.
//
// The widget uses Timer.periodic internally. Tests use fakeAsync / pump to
// control time without spinning up real timers.
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:physicscopilot/widgets/rate_limit_countdown.dart';

void main() {
  group('RateLimitCountdown', () {
    Widget wrap(Widget child) => MaterialApp(home: Scaffold(body: child));

    testWidgets(
        'when remainingSeconds is 0 returns SizedBox.shrink — no text visible',
        (tester) async {
      await tester.pumpWidget(
        wrap(const RateLimitCountdown(remaining: Duration.zero)),
      );
      await tester.pump();

      // No label text should be in the tree when countdown is idle.
      expect(find.byType(Text), findsNothing);
    });

    testWidgets('when remainingSeconds is 30, shows text containing "30"',
        (tester) async {
      await tester.pumpWidget(
        wrap(
          const RateLimitCountdown(remaining: Duration(seconds: 30)),
        ),
      );
      await tester.pump();

      // Label format for < 60 s: "Attendi 30s"
      expect(
        find.textContaining('30'),
        findsOneWidget,
        reason: 'Expected the remaining seconds (30) to appear in the label',
      );
    });

    testWidgets(
        'when remainingSeconds is 30, label starts with "Attendi"',
        (tester) async {
      await tester.pumpWidget(
        wrap(
          const RateLimitCountdown(remaining: Duration(seconds: 30)),
        ),
      );
      await tester.pump();

      expect(find.textContaining('Attendi'), findsOneWidget);
    });

    testWidgets(
        'when remainingSeconds is 90, shows minutes+seconds format containing "1m"',
        (tester) async {
      await tester.pumpWidget(
        wrap(
          const RateLimitCountdown(remaining: Duration(seconds: 90)),
        ),
      );
      await tester.pump();

      // Label format for >= 60 s: "Attendi 1m 30s"
      expect(
        find.textContaining('1m'),
        findsOneWidget,
        reason: 'Expected minutes component "1m" in the label for 90 seconds',
      );
    });

    testWidgets('onExpired is NOT called during initial build', (tester) async {
      var called = false;

      await tester.pumpWidget(
        wrap(
          RateLimitCountdown(
            remaining: const Duration(seconds: 30),
            onExpired: () => called = true,
          ),
        ),
      );
      // Only pump a single frame — no timer ticks.
      await tester.pump();

      expect(called, isFalse,
          reason: 'onExpired must not fire during the initial build',);
    });
  });
}
