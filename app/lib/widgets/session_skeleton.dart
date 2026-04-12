import 'package:flutter/material.dart';

/// A shimmer-style skeleton placeholder shown while session data is loading.
///
/// Renders [itemCount] placeholder rows that animate between two shades,
/// mimicking the shape of a [_SessionCard] without any real content.
/// Uses a custom opacity-based animation — no external packages required.
class SessionSkeleton extends StatefulWidget {
  const SessionSkeleton({super.key, this.itemCount = 5});

  /// Number of skeleton rows to display.
  final int itemCount;

  @override
  State<SessionSkeleton> createState() => _SessionSkeletonState();
}

class _SessionSkeletonState extends State<SessionSkeleton>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller;
  late final Animation<double> _opacity;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 900),
    )..repeat(reverse: true);

    _opacity = Tween<double>(begin: 0.3, end: 0.7).animate(
      CurvedAnimation(parent: _controller, curve: Curves.easeInOut),
    );
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: _opacity,
      builder: (context, _) {
        return ListView.builder(
          physics: const NeverScrollableScrollPhysics(),
          shrinkWrap: true,
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
          itemCount: widget.itemCount,
          itemBuilder: (context, index) =>
              _SkeletonCard(opacity: _opacity.value),
        );
      },
    );
  }
}

class _SkeletonCard extends StatelessWidget {
  const _SkeletonCard({required this.opacity});

  final double opacity;

  @override
  Widget build(BuildContext context) {
    return Opacity(
      opacity: opacity,
      child: Container(
        margin: const EdgeInsets.only(bottom: 12),
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: const Color(0xFF1E1E1E),
          borderRadius: BorderRadius.circular(12),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                _Pill(width: 140, height: 14),
                const Spacer(),
                _Pill(width: 60, height: 20),
              ],
            ),
            const SizedBox(height: 10),
            _Pill(width: double.infinity, height: 12),
            const SizedBox(height: 6),
            _Pill(width: 200, height: 12),
            const SizedBox(height: 14),
            Row(
              children: [
                _Pill(width: 80, height: 10),
                const Spacer(),
                _Pill(width: 60, height: 10),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

/// A rounded rectangle placeholder block used inside skeleton cards.
class _Pill extends StatelessWidget {
  const _Pill({required this.width, required this.height});

  final double width;
  final double height;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: width,
      height: height,
      decoration: BoxDecoration(
        color: const Color(0xFF2A2A2A),
        borderRadius: BorderRadius.circular(6),
      ),
    );
  }
}
