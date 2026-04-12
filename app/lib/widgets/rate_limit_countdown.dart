// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

import 'dart:async';

import 'package:flutter/material.dart';

import '../../main.dart' show kAccent, kBgCard, kBgCardBorder, kTextMuted;

// ── RateLimitCountdown ────────────────────────────────────────────────────────

/// Shows a countdown timer while a rate-limit cooldown is active.
///
/// Ticks every second using [Timer.periodic]. Calls [onExpired] when the
/// countdown reaches zero and removes itself from the tree (returns
/// [SizedBox.shrink]).
class RateLimitCountdown extends StatefulWidget {
  const RateLimitCountdown({
    super.key,
    required this.remaining,
    this.onExpired,
  });

  /// How long is left in the cooldown period at widget creation time.
  final Duration remaining;

  /// Called once when the countdown expires.
  final VoidCallback? onExpired;

  @override
  State<RateLimitCountdown> createState() => _RateLimitCountdownState();
}

class _RateLimitCountdownState extends State<RateLimitCountdown> {
  late Duration _remaining;
  Timer? _timer;

  @override
  void initState() {
    super.initState();
    _remaining = widget.remaining;
    _startTimer();
  }

  @override
  void didUpdateWidget(RateLimitCountdown oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.remaining != widget.remaining) {
      _remaining = widget.remaining;
      _timer?.cancel();
      _startTimer();
    }
  }

  void _startTimer() {
    if (_remaining <= Duration.zero) return;
    _timer = Timer.periodic(const Duration(seconds: 1), (_) {
      if (!mounted) return;
      setState(() {
        _remaining -= const Duration(seconds: 1);
      });
      if (_remaining <= Duration.zero) {
        _timer?.cancel();
        widget.onExpired?.call();
      }
    });
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }

  String get _label {
    final secs = _remaining.inSeconds;
    if (secs <= 0) return '';
    if (secs >= 60) {
      final m = secs ~/ 60;
      final s = secs % 60;
      return 'Attendi ${m}m ${s.toString().padLeft(2, '0')}s';
    }
    return 'Attendi ${secs}s';
  }

  @override
  Widget build(BuildContext context) {
    if (_remaining <= Duration.zero) return const SizedBox.shrink();

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
      decoration: BoxDecoration(
        color: kBgCard,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: kBgCardBorder),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Icon(Icons.timer_outlined, size: 16, color: kAccent),
          const SizedBox(width: 8),
          Text(
            _label,
            style: const TextStyle(
              color: kTextMuted,
              fontSize: 13,
              fontWeight: FontWeight.w500,
            ),
          ),
        ],
      ),
    );
  }
}
