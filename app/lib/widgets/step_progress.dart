// Copyright (c) 2026 Franck1120. All rights reserved.
// Use of this source code is governed by a MIT license that can be
// found in the LICENSE file.

import 'package:flutter/material.dart';

// ── Data model ───────────────────────────────────────────────────────────────

/// Metadata for a single procedural step.
class StepInfo {
  final String description;
  final Duration? estimatedDuration;

  const StepInfo({required this.description, this.estimatedDuration});
}

enum _StepStatus { completed, current, upcoming }

// ── Widget ───────────────────────────────────────────────────────────────────

/// Progress bar shown at the bottom of the camera screen.
///
/// Renders dot indicators, an animated linear bar, and the current step
/// description. Completed dots are green, the current dot is yellow, and
/// future dots are grey.
class StepProgress extends StatelessWidget {
  final List<StepInfo> steps;

  /// Zero-based index of the active step.
  final int currentStep;

  const StepProgress({
    super.key,
    required this.steps,
    required this.currentStep,
  });

  @override
  Widget build(BuildContext context) {
    if (steps.isEmpty) return const SizedBox.shrink();

    final safeIndex = currentStep.clamp(0, steps.length - 1);
    final current = steps[safeIndex];
    final total = steps.length;
    final progress = (safeIndex + 1) / total;

    return Container(
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 16),
      decoration: const BoxDecoration(
        color: Color(0xCC000000),
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _StepDots(total: total, currentIndex: safeIndex),
          const SizedBox(height: 8),
          _AnimatedProgressBar(progress: progress),
          const SizedBox(height: 10),
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Expanded(
                child: Text(
                  current.description,
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 14,
                    fontWeight: FontWeight.w500,
                    height: 1.3,
                  ),
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              if (current.estimatedDuration != null) ...[
                const SizedBox(width: 8),
                _DurationBadge(duration: current.estimatedDuration!),
              ],
            ],
          ),
          const SizedBox(height: 4),
          Text(
            'Step ${safeIndex + 1} of $total',
            style: const TextStyle(color: Colors.white54, fontSize: 11),
          ),
        ],
      ),
    );
  }
}

// ── Dot row ──────────────────────────────────────────────────────────────────

class _StepDots extends StatelessWidget {
  final int total;
  final int currentIndex;

  const _StepDots({required this.total, required this.currentIndex});

  @override
  Widget build(BuildContext context) {
    return Row(
      children: List.generate(total, (i) {
        final status = i < currentIndex
            ? _StepStatus.completed
            : i == currentIndex
                ? _StepStatus.current
                : _StepStatus.upcoming;
        return _AnimatedDot(status: status);
      }),
    );
  }
}

class _AnimatedDot extends StatelessWidget {
  final _StepStatus status;

  const _AnimatedDot({required this.status});

  @override
  Widget build(BuildContext context) {
    final color = switch (status) {
      _StepStatus.completed => Colors.greenAccent,
      _StepStatus.current => Colors.yellowAccent,
      _StepStatus.upcoming => Colors.white24,
    };
    final size = status == _StepStatus.current ? 12.0 : 8.0;

    return AnimatedContainer(
      duration: const Duration(milliseconds: 300),
      curve: Curves.easeInOut,
      margin: const EdgeInsets.only(right: 6),
      width: size,
      height: size,
      decoration: BoxDecoration(
        color: color,
        shape: BoxShape.circle,
        boxShadow: status == _StepStatus.current
            ? [
                BoxShadow(
                  color: Colors.yellowAccent.withAlpha(120),
                  blurRadius: 6,
                  spreadRadius: 1,
                ),
              ]
            : null,
      ),
    );
  }
}

// ── Progress bar ─────────────────────────────────────────────────────────────

class _AnimatedProgressBar extends StatelessWidget {
  final double progress;

  const _AnimatedProgressBar({required this.progress});

  @override
  Widget build(BuildContext context) {
    return ClipRRect(
      borderRadius: BorderRadius.circular(4),
      child: TweenAnimationBuilder<double>(
        tween: Tween<double>(begin: 0, end: progress),
        duration: const Duration(milliseconds: 400),
        curve: Curves.easeInOut,
        builder: (context, value, _) => LinearProgressIndicator(
          value: value,
          minHeight: 6,
          backgroundColor: Colors.white12,
          valueColor: AlwaysStoppedAnimation<Color>(
            value >= 1.0 ? Colors.greenAccent : Colors.yellowAccent,
          ),
        ),
      ),
    );
  }
}

// ── Duration badge ────────────────────────────────────────────────────────────

class _DurationBadge extends StatelessWidget {
  final Duration duration;

  const _DurationBadge({required this.duration});

  String get _label {
    final mins = duration.inMinutes;
    if (mins > 0) return '~$mins min';
    return '~${duration.inSeconds} s';
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
        color: Colors.white12,
        borderRadius: BorderRadius.circular(10),
      ),
      child: Text(
        _label,
        style: const TextStyle(color: Colors.white70, fontSize: 11),
      ),
    );
  }
}
