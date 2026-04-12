import 'package:flutter_test/flutter_test.dart';

import 'package:physicscopilot/widgets/ar_overlay.dart';

// Tests cover only the pure data-model layer (no widgets, no providers).

void main() {
  group('OverlayBox.fromJson', () {
    test('parses double values correctly', () {
      final box = OverlayBox.fromJson({
        'x': 0.1,
        'y': 0.2,
        'w': 0.5,
        'h': 0.3,
        'label': 'Motor',
        'color': 'green',
      });

      expect(box.x, closeTo(0.1, 1e-9));
      expect(box.y, closeTo(0.2, 1e-9));
      expect(box.w, closeTo(0.5, 1e-9));
      expect(box.h, closeTo(0.3, 1e-9));
      expect(box.label, 'Motor');
    });

    test('parses int values (num coercion) correctly', () {
      final box = OverlayBox.fromJson({
        'x': 0,
        'y': 1,
        'w': 1,
        'h': 1,
        'label': 'Part',
        'color': 'blue',
      });

      expect(box.x, 0.0);
      expect(box.y, 1.0);
    });

    test('falls back to 0.0 and empty label when fields are null', () {
      final box = OverlayBox.fromJson({});

      expect(box.x, 0.0);
      expect(box.y, 0.0);
      expect(box.w, 0.0);
      expect(box.h, 0.0);
      expect(box.label, '');
    });
  });

  group('OverlayArrow.fromJson', () {
    test('parses all coordinate fields', () {
      final arrow = OverlayArrow.fromJson({
        'from_x': 0.1,
        'from_y': 0.2,
        'to_x': 0.8,
        'to_y': 0.9,
        'label': 'direction',
      });

      expect(arrow.fromX, closeTo(0.1, 1e-9));
      expect(arrow.fromY, closeTo(0.2, 1e-9));
      expect(arrow.toX, closeTo(0.8, 1e-9));
      expect(arrow.toY, closeTo(0.9, 1e-9));
      expect(arrow.label, 'direction');
    });

    test('falls back to 0.0 and empty label when fields are null', () {
      final arrow = OverlayArrow.fromJson({});

      expect(arrow.fromX, 0.0);
      expect(arrow.fromY, 0.0);
      expect(arrow.toX, 0.0);
      expect(arrow.toY, 0.0);
      expect(arrow.label, '');
    });
  });

  group('OverlayTextItem.fromJson', () {
    test('parses x, y, content, and size', () {
      final item = OverlayTextItem.fromJson({
        'x': 0.3,
        'y': 0.4,
        'content': 'Label A',
        'size': 18.0,
      });

      expect(item.x, closeTo(0.3, 1e-9));
      expect(item.y, closeTo(0.4, 1e-9));
      expect(item.content, 'Label A');
      expect(item.size, closeTo(18.0, 1e-9));
    });

    test('falls back to default size 16 and empty content when fields are null', () {
      final item = OverlayTextItem.fromJson({'x': 0.0, 'y': 0.0});

      expect(item.content, '');
      expect(item.size, 16.0);
    });
  });

  group('OverlayData.fromJson', () {
    test('parses boxes, arrows, and texts from full JSON', () {
      final data = OverlayData.fromJson({
        'boxes': [
          {'x': 0.1, 'y': 0.1, 'w': 0.2, 'h': 0.2, 'label': 'Box1', 'color': 'red'},
        ],
        'arrows': [
          {'from_x': 0.0, 'from_y': 0.0, 'to_x': 1.0, 'to_y': 1.0, 'label': 'Arrow1'},
        ],
        'text': [
          {'x': 0.5, 'y': 0.5, 'content': 'Text1', 'size': 14.0},
        ],
      });

      expect(data.boxes, hasLength(1));
      expect(data.boxes.first.label, 'Box1');
      expect(data.arrows, hasLength(1));
      expect(data.arrows.first.label, 'Arrow1');
      expect(data.texts, hasLength(1));
      expect(data.texts.first.content, 'Text1');
    });

    test('fromJson with empty JSON → all lists empty', () {
      final data = OverlayData.fromJson({});

      expect(data.boxes, isEmpty);
      expect(data.arrows, isEmpty);
      expect(data.texts, isEmpty);
    });

    test('fromJson with null list fields → all lists empty', () {
      final data = OverlayData.fromJson({
        'boxes': null,
        'arrows': null,
        'text': null,
      });

      expect(data.boxes, isEmpty);
      expect(data.arrows, isEmpty);
      expect(data.texts, isEmpty);
    });
  });

  group('OverlayData.isEmpty', () {
    test('returns true when all lists are empty', () {
      const data = OverlayData(boxes: [], arrows: [], texts: []);
      expect(data.isEmpty, isTrue);
    });

    test('returns false when boxes has items', () {
      final data = OverlayData.fromJson({
        'boxes': [
          {'x': 0.0, 'y': 0.0, 'w': 0.1, 'h': 0.1, 'label': 'X', 'color': 'red'},
        ],
        'arrows': [],
        'text': [],
      });

      expect(data.isEmpty, isFalse);
    });

    test('returns false when arrows has items', () {
      final data = OverlayData.fromJson({
        'boxes': [],
        'arrows': [
          {'from_x': 0.0, 'from_y': 0.0, 'to_x': 1.0, 'to_y': 1.0, 'label': ''},
        ],
        'text': [],
      });

      expect(data.isEmpty, isFalse);
    });

    test('returns false when texts has items', () {
      final data = OverlayData.fromJson({
        'boxes': [],
        'arrows': [],
        'text': [
          {'x': 0.0, 'y': 0.0, 'content': 'hello', 'size': 14.0},
        ],
      });

      expect(data.isEmpty, isFalse);
    });
  });
}
