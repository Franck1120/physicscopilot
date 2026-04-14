import 'dart:math' as math;

import 'package:flutter/material.dart';

import 'package:physicscopilot/main.dart' show kAccent, kBgCard;

// ── ProgressRing ─────────────────────────────────────────────────────────────

/// Circular arc progress indicator with an optional centred child widget.
///
/// [value] must be in the range [0.0, 1.0]. Changes to [value] are smoothly
/// animated via [TweenAnimationBuilder].
class ProgressRing extends StatelessWidget {
  const ProgressRing({
    super.key,
    required this.value,
    this.size = 80,
    this.color,
    this.backgroundColor,
    this.strokeWidth = 6,
    this.child,
  }) : assert(value >= 0.0 && value <= 1.0, 'value must be between 0 and 1');

  /// Progress in the range [0.0, 1.0].
  final double value;

  /// Diameter of the ring in logical pixels.
  final double size;

  /// Foreground arc colour. Defaults to [kAccent].
  final Color? color;

  /// Background track colour. Defaults to a dimmed version of the foreground.
  final Color? backgroundColor;

  /// Width of the arc stroke.
  final double strokeWidth;

  /// Widget drawn at the centre of the ring (e.g. a percentage label).
  final Widget? child;

  @override
  Widget build(BuildContext context) {
    final fg = color ?? kAccent;
    final bg = backgroundColor ?? kBgCard;

    return TweenAnimationBuilder<double>(
      tween: Tween<double>(begin: 0, end: value),
      duration: const Duration(milliseconds: 600),
      curve: Curves.easeInOut,
      builder: (context, animatedValue, _) {
        return SizedBox(
          width: size,
          height: size,
          child: CustomPaint(
            painter: _RingPainter(
              value: animatedValue,
              foregroundColor: fg,
              backgroundColor: bg,
              strokeWidth: strokeWidth,
            ),
            child: child == null
                ? null
                : Center(child: child),
          ),
        );
      },
    );
  }
}

// ── Painter ───────────────────────────────────────────────────────────────────

class _RingPainter extends CustomPainter {
  const _RingPainter({
    required this.value,
    required this.foregroundColor,
    required this.backgroundColor,
    required this.strokeWidth,
  });

  final double value;
  final Color foregroundColor;
  final Color backgroundColor;
  final double strokeWidth;

  static const double _startAngle = -math.pi / 2; // 12 o'clock

  @override
  void paint(Canvas canvas, Size size) {
    final center = Offset(size.width / 2, size.height / 2);
    final radius = (size.shortestSide - strokeWidth) / 2;
    final rect = Rect.fromCircle(center: center, radius: radius);

    // Background full circle track.
    final bgPaint = Paint()
      ..color = backgroundColor
      ..style = PaintingStyle.stroke
      ..strokeWidth = strokeWidth
      ..strokeCap = StrokeCap.round;
    canvas.drawArc(rect, 0, 2 * math.pi, false, bgPaint);

    // Foreground arc proportional to value.
    if (value > 0) {
      final fgPaint = Paint()
        ..color = foregroundColor
        ..style = PaintingStyle.stroke
        ..strokeWidth = strokeWidth
        ..strokeCap = StrokeCap.round;
      canvas.drawArc(rect, _startAngle, 2 * math.pi * value, false, fgPaint);
    }
  }

  @override
  bool shouldRepaint(_RingPainter oldDelegate) =>
      oldDelegate.value != value ||
      oldDelegate.foregroundColor != foregroundColor ||
      oldDelegate.backgroundColor != backgroundColor ||
      oldDelegate.strokeWidth != strokeWidth;
}
