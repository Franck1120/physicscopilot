// Copyright (c) 2026 Franck1120. All rights reserved.
// Use of this source code is governed by a MIT license that can be
// found in the LICENSE file.

import 'dart:math' as math;

import 'package:flutter/material.dart';

// ── Data models ──────────────────────────────────────────────────────────────

/// A bounding box from the backend overlay payload.
class OverlayBox {
  final double x, y, w, h;
  final String label;
  final Color color;

  const OverlayBox({
    required this.x,
    required this.y,
    required this.w,
    required this.h,
    required this.label,
    required this.color,
  });

  factory OverlayBox.fromJson(Map<String, dynamic> json) => OverlayBox(
        x: (json['x'] as num?)?.toDouble() ?? 0.0,
        y: (json['y'] as num?)?.toDouble() ?? 0.0,
        w: (json['w'] as num?)?.toDouble() ?? 0.0,
        h: (json['h'] as num?)?.toDouble() ?? 0.0,
        label: (json['label'] as String?) ?? '',
        color: _parseColor((json['color'] as String?) ?? 'red'),
      );
}

/// A directional arrow from the backend overlay payload.
class OverlayArrow {
  final double fromX, fromY, toX, toY;
  final String label;

  const OverlayArrow({
    required this.fromX,
    required this.fromY,
    required this.toX,
    required this.toY,
    required this.label,
  });

  factory OverlayArrow.fromJson(Map<String, dynamic> json) => OverlayArrow(
        fromX: (json['from_x'] as num?)?.toDouble() ?? 0.0,
        fromY: (json['from_y'] as num?)?.toDouble() ?? 0.0,
        toX: (json['to_x'] as num?)?.toDouble() ?? 0.0,
        toY: (json['to_y'] as num?)?.toDouble() ?? 0.0,
        label: (json['label'] as String?) ?? '',
      );
}

/// An annotated text label from the backend overlay payload.
class OverlayTextItem {
  final double x, y;
  final String content;
  final double size;

  const OverlayTextItem({
    required this.x,
    required this.y,
    required this.content,
    required this.size,
  });

  factory OverlayTextItem.fromJson(Map<String, dynamic> json) => OverlayTextItem(
        x: (json['x'] as num).toDouble(),
        y: (json['y'] as num).toDouble(),
        content: (json['content'] as String?) ?? '',
        size: (json['size'] as num?)?.toDouble() ?? 16.0,
      );
}

/// Parsed backend overlay payload containing boxes, arrows, and text labels.
/// All coordinates are normalised 0–1.
class OverlayData {
  final List<OverlayBox> boxes;
  final List<OverlayArrow> arrows;
  final List<OverlayTextItem> texts;

  const OverlayData({
    required this.boxes,
    required this.arrows,
    required this.texts,
  });

  factory OverlayData.fromJson(Map<String, dynamic> json) => OverlayData(
        boxes: (json['boxes'] as List<dynamic>? ?? [])
            .map((e) => OverlayBox.fromJson(e as Map<String, dynamic>))
            .toList(),
        arrows: (json['arrows'] as List<dynamic>? ?? [])
            .map((e) => OverlayArrow.fromJson(e as Map<String, dynamic>))
            .toList(),
        texts: (json['text'] as List<dynamic>? ?? [])
            .map((e) => OverlayTextItem.fromJson(e as Map<String, dynamic>))
            .toList(),
      );

  bool get isEmpty => boxes.isEmpty && arrows.isEmpty && texts.isEmpty;
}

// ── Color helper ─────────────────────────────────────────────────────────────

Color _parseColor(String name) => switch (name.toLowerCase()) {
      'red' => Colors.red,
      'green' => Colors.greenAccent,
      'blue' => Colors.blueAccent,
      'yellow' => Colors.yellowAccent,
      'orange' => Colors.orange,
      'purple' => Colors.purpleAccent,
      'cyan' => Colors.cyanAccent,
      'white' => Colors.white,
      _ => Colors.red,
    };

// ── Widget ───────────────────────────────────────────────────────────────────

/// Renders AR overlay elements (boxes, arrows, text) on top of the camera feed.
///
/// Animates with a fade-in/out transition whenever [data] changes.
/// Coordinates in [data] are normalised 0–1 and mapped to the widget's size.
class ArOverlay extends StatefulWidget {
  final OverlayData? data;

  const ArOverlay({super.key, this.data});

  @override
  State<ArOverlay> createState() => _ArOverlayState();
}

class _ArOverlayState extends State<ArOverlay>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller;
  late final Animation<double> _opacity;
  OverlayData? _currentData;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 300),
    );
    _opacity = CurvedAnimation(parent: _controller, curve: Curves.easeInOut);
    _currentData = widget.data;
    if (_currentData != null && !_currentData!.isEmpty) {
      _controller.forward();
    }
  }

  @override
  void didUpdateWidget(ArOverlay oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.data != widget.data) {
      // Fade out → swap data → fade in.
      _controller.reverse().then((_) {
        if (mounted) {
          setState(() => _currentData = widget.data);
          if (widget.data != null && !widget.data!.isEmpty) {
            _controller.forward();
          }
        }
      });
    }
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final data = _currentData;
    if (data == null || data.isEmpty) return const SizedBox.shrink();

    return FadeTransition(
      opacity: _opacity,
      child: CustomPaint(
        painter: _ArPainter(data: data),
        child: const SizedBox.expand(),
      ),
    );
  }
}

// ── Painter ───────────────────────────────────────────────────────────────────

class _ArPainter extends CustomPainter {
  final OverlayData data;

  const _ArPainter({required this.data});

  @override
  void paint(Canvas canvas, Size size) {
    for (final box in data.boxes) {
      _drawBox(canvas, size, box);
    }
    for (final arrow in data.arrows) {
      _drawArrow(canvas, size, arrow);
    }
    for (final text in data.texts) {
      _drawText(canvas, size, text);
    }
  }

  // ── Bounding box ───────────────────────────────────────────────────────────

  void _drawBox(Canvas canvas, Size size, OverlayBox box) {
    final rect = Rect.fromLTWH(
      box.x * size.width,
      box.y * size.height,
      box.w * size.width,
      box.h * size.height,
    );

    canvas.drawRect(
      rect,
      Paint()
        ..color = box.color
        ..style = PaintingStyle.stroke
        ..strokeWidth = 2.5,
    );

    if (box.label.isNotEmpty) {
      _drawLabel(canvas, box.label, Offset(rect.left, rect.top), box.color);
    }
  }

  // ── Arrow ─────────────────────────────────────────────────────────────────

  void _drawArrow(Canvas canvas, Size size, OverlayArrow arrow) {
    final from = Offset(arrow.fromX * size.width, arrow.fromY * size.height);
    final to = Offset(arrow.toX * size.width, arrow.toY * size.height);

    final paint = Paint()
      ..color = Colors.yellowAccent
      ..strokeWidth = 2.5
      ..style = PaintingStyle.stroke
      ..strokeCap = StrokeCap.round;

    canvas.drawLine(from, to, paint);

    // Arrowhead
    final angle = math.atan2(to.dy - from.dy, to.dx - from.dx);
    const headLen = 14.0;
    const headAngle = 0.4;

    canvas.drawPath(
      Path()
        ..moveTo(to.dx, to.dy)
        ..lineTo(
          to.dx - headLen * math.cos(angle - headAngle),
          to.dy - headLen * math.sin(angle - headAngle),
        )
        ..moveTo(to.dx, to.dy)
        ..lineTo(
          to.dx - headLen * math.cos(angle + headAngle),
          to.dy - headLen * math.sin(angle + headAngle),
        ),
      paint,
    );

    if (arrow.label.isNotEmpty) {
      final mid = Offset((from.dx + to.dx) / 2, (from.dy + to.dy) / 2);
      _drawLabel(canvas, arrow.label, mid, Colors.yellowAccent);
    }
  }

  // ── Text label ────────────────────────────────────────────────────────────

  void _drawText(Canvas canvas, Size size, OverlayTextItem item) {
    final origin = Offset(item.x * size.width, item.y * size.height);

    final painter = TextPainter(
      text: TextSpan(
        text: item.content,
        style: TextStyle(
          color: Colors.white,
          fontSize: item.size,
          fontWeight: FontWeight.w600,
          shadows: const [Shadow(blurRadius: 4, color: Colors.black)],
        ),
      ),
      textDirection: TextDirection.ltr,
    )..layout(maxWidth: size.width * 0.8);

    // Semi-transparent background pill
    canvas.drawRRect(
      RRect.fromRectAndRadius(
        Rect.fromLTWH(
          origin.dx - 4,
          origin.dy - 4,
          painter.width + 8,
          painter.height + 8,
        ),
        const Radius.circular(4),
      ),
      Paint()..color = const Color(0x99000000),
    );

    painter.paint(canvas, origin);
  }

  // ── Shared label helper ───────────────────────────────────────────────────

  void _drawLabel(Canvas canvas, String label, Offset position, Color color) {
    final painter = TextPainter(
      text: TextSpan(
        text: label,
        style: const TextStyle(
          color: Colors.white,
          fontSize: 12,
          fontWeight: FontWeight.w600,
        ),
      ),
      textDirection: TextDirection.ltr,
    )..layout();

    final bgRect = Rect.fromLTWH(
      position.dx,
      position.dy - painter.height - 2,
      painter.width + 8,
      painter.height + 4,
    );

    canvas.drawRRect(
      RRect.fromRectAndRadius(bgRect, const Radius.circular(3)),
      Paint()..color = color.withAlpha(200),
    );

    painter.paint(
      canvas,
      Offset(position.dx + 4, position.dy - painter.height),
    );
  }

  @override
  bool shouldRepaint(_ArPainter oldDelegate) => oldDelegate.data != data;
}
