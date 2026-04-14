import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:physicscopilot/widgets/ar_overlay.dart';

void main() {
  // ── Data model tests (no Flutter rendering needed) ──────────────────────────

  group('OverlayData.fromJson', () {
    test('empty JSON produces isEmpty == true', () {
      final data = OverlayData.fromJson({});
      expect(data.isEmpty, isTrue);
    });

    test('JSON with one box produces boxes.length == 1', () {
      final data = OverlayData.fromJson({
        'boxes': [
          {
            'x': 0.1,
            'y': 0.2,
            'w': 0.3,
            'h': 0.4,
            'label': 'test',
            'color': 'red',
          }
        ],
      });
      expect(data.boxes.length, 1);
    });

    test('isEmpty is false when boxes list is non-empty', () {
      final data = OverlayData.fromJson({
        'boxes': [
          {'x': 0.0, 'y': 0.0, 'w': 0.5, 'h': 0.5, 'label': 'A', 'color': 'blue'},
        ],
      });
      expect(data.isEmpty, isFalse);
    });
  });

  group('OverlayBox.fromJson', () {
    test('integer values in JSON are converted to double', () {
      final box = OverlayBox.fromJson(
        {'x': 1, 'y': 2, 'w': 3, 'h': 4, 'label': 'item', 'color': 'red'},
      );
      expect(box.x, isA<double>());
      expect(box.y, isA<double>());
      expect(box.w, isA<double>());
      expect(box.h, isA<double>());
      expect(box.x, 1.0);
      expect(box.y, 2.0);
    });

    test('"green" color string maps to Colors.greenAccent', () {
      final box = OverlayBox.fromJson(
        {'x': 0, 'y': 0, 'w': 0.1, 'h': 0.1, 'label': '', 'color': 'green'},
      );
      expect(box.color, equals(Colors.greenAccent));
    });
  });

  group('OverlayArrow.fromJson', () {
    test('from and to coordinates are parsed correctly', () {
      final arrow = OverlayArrow.fromJson({
        'from_x': 0.1,
        'from_y': 0.2,
        'to_x': 0.8,
        'to_y': 0.9,
        'label': 'force',
      });
      expect(arrow.fromX, closeTo(0.1, 1e-9));
      expect(arrow.fromY, closeTo(0.2, 1e-9));
      expect(arrow.toX, closeTo(0.8, 1e-9));
      expect(arrow.toY, closeTo(0.9, 1e-9));
      expect(arrow.label, 'force');
    });

    test('missing fields fall back to 0.0 and empty label', () {
      final arrow = OverlayArrow.fromJson({});
      expect(arrow.fromX, 0.0);
      expect(arrow.label, '');
    });
  });

  // ── Widget tests — only null/empty cases (no hardware painter needed) ───────

  group('ArOverlay widget', () {
    testWidgets('null data renders SizedBox.shrink — no CustomPaint',
        (WidgetTester tester) async {
      await tester.pumpWidget(
        const MaterialApp(home: Scaffold(body: ArOverlay())),
      );
      await tester.pumpAndSettle();

      expect(find.byType(CustomPaint), findsNothing);
    });

    testWidgets('empty OverlayData renders SizedBox.shrink — no CustomPaint',
        (WidgetTester tester) async {
      final emptyData = OverlayData.fromJson({});
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(body: ArOverlay(data: emptyData)),
        ),
      );
      await tester.pumpAndSettle();

      expect(find.byType(CustomPaint), findsNothing);
    });
  });
}
