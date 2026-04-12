import 'dart:math' as math;

import 'package:flutter/material.dart';

/// Particle data for the confetti animation.
class _Particle {
  _Particle({
    required this.x,
    required this.y,
    required this.speed,
    required this.angle,
    required this.color,
    required this.size,
    required this.spin,
  });

  double x;
  double y;
  final double speed;
  final double angle;
  final Color color;
  final double size;
  final double spin;
  double rotation = 0;
}

/// A fullscreen confetti overlay that plays once and then removes itself.
///
/// Renders [particleCount] coloured particles that fall from the top of the
/// screen. The overlay runs for [duration], fades out over the final 25 % of
/// that time, and then calls [onComplete] so the parent can remove the widget.
///
/// Uses a custom [CustomPainter] — no external packages required.
///
/// Example:
/// ```dart
/// if (_showConfetti)
///   ConfettiOverlay(
///     onComplete: () => setState(() => _showConfetti = false),
///   )
/// ```
class ConfettiOverlay extends StatefulWidget {
  /// Creates the confetti overlay.
  ///
  /// [duration] controls total playback time; [particleCount] controls visual
  /// density; [onComplete] is called when the animation finishes.
  const ConfettiOverlay({
    super.key,
    this.duration = const Duration(seconds: 3),
    this.particleCount = 80,
    this.onComplete,
  });

  /// Total animation duration, including the fade-out phase.
  final Duration duration;

  /// Number of confetti particles rendered simultaneously.
  final int particleCount;

  /// Callback invoked once when the overlay animation finishes.
  ///
  /// Use this to remove the overlay from the widget tree:
  /// ```dart
  /// onComplete: () => setState(() => _showConfetti = false),
  /// ```
  final VoidCallback? onComplete;

  @override
  State<ConfettiOverlay> createState() => _ConfettiOverlayState();
}

class _ConfettiOverlayState extends State<ConfettiOverlay>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller;
  late final Animation<double> _fade;
  late final List<_Particle> _particles;
  final math.Random _random = math.Random();

  static const List<Color> _palette = [
    Color(0xFF10B981),
    Color(0xFF3B82F6),
    Color(0xFFF59E0B),
    Color(0xFFEF4444),
    Color(0xFF8B5CF6),
    Color(0xFFEC4899),
    Color(0xFF06B6D4),
    Color(0xFFFBBF24),
  ];

  @override
  void initState() {
    super.initState();

    _controller = AnimationController(
      vsync: this,
      duration: widget.duration,
    );

    _fade = Tween<double>(begin: 1.0, end: 0.0).animate(
      CurvedAnimation(
        parent: _controller,
        curve: const Interval(0.75, 1.0, curve: Curves.easeOut),
      ),
    );

    _particles = List.generate(widget.particleCount, (_) => _makeParticle());

    _controller.forward().then((_) {
      widget.onComplete?.call();
    });
  }

  _Particle _makeParticle() {
    return _Particle(
      x: _random.nextDouble(),
      y: -_random.nextDouble() * 0.3,
      speed: 0.002 + _random.nextDouble() * 0.004,
      angle: (_random.nextDouble() - 0.5) * 0.04,
      color: _palette[_random.nextInt(_palette.length)],
      size: 6 + _random.nextDouble() * 8,
      spin: (_random.nextDouble() - 0.5) * 0.15,
    );
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return FadeTransition(
      opacity: _fade,
      child: IgnorePointer(
        child: AnimatedBuilder(
          animation: _controller,
          builder: (context, _) {
            // Update particle positions each frame.
            for (final p in _particles) {
              p.y += p.speed;
              p.x += p.angle;
              p.rotation += p.spin;
            }
            return CustomPaint(
              painter: _ConfettiPainter(particles: _particles),
              size: Size.infinite,
            );
          },
        ),
      ),
    );
  }
}

class _ConfettiPainter extends CustomPainter {
  const _ConfettiPainter({required this.particles});

  final List<_Particle> particles;

  @override
  void paint(Canvas canvas, Size size) {
    for (final p in particles) {
      if (p.y > 1.2) continue; // Off-screen

      final cx = p.x * size.width;
      final cy = p.y * size.height;
      final paint = Paint()..color = p.color;

      canvas.save();
      canvas.translate(cx, cy);
      canvas.rotate(p.rotation);

      // Draw a small rectangle as a confetti piece.
      canvas.drawRect(
        Rect.fromCenter(
          center: Offset.zero,
          width: p.size,
          height: p.size * 0.5,
        ),
        paint,
      );

      canvas.restore();
    }
  }

  @override
  bool shouldRepaint(_ConfettiPainter oldDelegate) => true;
}
