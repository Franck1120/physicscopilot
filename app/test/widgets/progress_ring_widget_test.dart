// Widget tests for ProgressRing.
//
// ProgressRing is a pure StatelessWidget that delegates animation to
// TweenAnimationBuilder, so all assertions are made after pump() settles the
// first frame.
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:physicscopilot/widgets/progress_ring_widget.dart';

void main() {
  group('ProgressRing', () {
    Widget _wrap(Widget child) => MaterialApp(home: Scaffold(body: child));

    testWidgets('renders without child and does not throw', (tester) async {
      await tester.pumpWidget(_wrap(const ProgressRing(value: 0.5)));
      await tester.pump();
      expect(find.byType(ProgressRing), findsOneWidget);
    });

    testWidgets('renders with a Text child inside the ring', (tester) async {
      await tester.pumpWidget(
        _wrap(const ProgressRing(value: 0.5, child: Text('50%'))),
      );
      await tester.pump();
      expect(find.text('50%'), findsOneWidget);
    });

    testWidgets('value=0.0 does not throw', (tester) async {
      await tester.pumpWidget(_wrap(const ProgressRing(value: 0.0)));
      await tester.pump();
      expect(find.byType(ProgressRing), findsOneWidget);
    });

    testWidgets('value=1.0 does not throw', (tester) async {
      await tester.pumpWidget(_wrap(const ProgressRing(value: 1.0)));
      await tester.pump();
      expect(find.byType(ProgressRing), findsOneWidget);
    });

    testWidgets('custom size parameter produces a SizedBox with given dimensions',
        (tester) async {
      const double customSize = 120;
      await tester.pumpWidget(
        _wrap(const ProgressRing(value: 0.4, size: customSize)),
      );
      await tester.pump();

      // TweenAnimationBuilder wraps its output in the builder return value,
      // which is a SizedBox(width: size, height: size).
      final sizedBoxes = tester.widgetList<SizedBox>(find.byType(SizedBox));
      final match = sizedBoxes.any(
        (box) => box.width == customSize && box.height == customSize,
      );
      expect(match, isTrue,
          reason: 'Expected a SizedBox with width=$customSize and height=$customSize');
    });
  });
}
