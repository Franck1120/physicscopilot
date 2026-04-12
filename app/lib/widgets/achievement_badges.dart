// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

import 'package:flutter/material.dart';

import 'package:physicscopilot/main.dart'
    show kAccent, kBgCard, kBgCardBorder, kTextMuted;

// ── Milestones ────────────────────────────────────────────────────────────────

const _milestones = [5, 10, 25, 50, 100];

// ── AchievementBadgesWidget ───────────────────────────────────────────────────

/// Displays a row of circular badge icons for session count milestones.
///
/// Unlocked badges (where [sessionCount] >= milestone) are filled with
/// [kAccent] and show an emoji + count label. Locked badges are rendered
/// in semi-transparent grey with a lock icon.
class AchievementBadgesWidget extends StatelessWidget {
  const AchievementBadgesWidget({super.key, required this.sessionCount});

  final int sessionCount;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: kBgCard,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: kBgCardBorder),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            'TRAGUARDI',
            style: TextStyle(
              color: kTextMuted,
              fontSize: 11,
              fontWeight: FontWeight.w600,
              letterSpacing: 0.8,
            ),
          ),
          const SizedBox(height: 14),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceAround,
            children: _milestones
                .map((m) => _Badge(milestone: m, unlocked: sessionCount >= m))
                .toList(),
          ),
          if (sessionCount >= 50) ...[
            const SizedBox(height: 16),
            const _ExpertBadge(),
          ],
        ],
      ),
    );
  }
}

// ── _ExpertBadge ──────────────────────────────────────────────────────────────

/// Shown only when [sessionCount] >= 50. Displays a gold trophy icon with
/// an "Expert" label to mark the user as a power user.
class _ExpertBadge extends StatelessWidget {
  const _ExpertBadge();

  static const _gold = Color(0xFFFFD700);

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        Container(
          width: 52,
          height: 52,
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            color: _gold.withAlpha(30),
            border: Border.all(color: _gold, width: 1.5),
          ),
          child: const Center(
            child: Icon(
              Icons.emoji_events,
              size: 26,
              color: _gold,
            ),
          ),
        ),
        const SizedBox(width: 12),
        const Text(
          'Expert',
          style: TextStyle(
            color: _gold,
            fontSize: 14,
            fontWeight: FontWeight.w700,
            letterSpacing: 0.5,
          ),
        ),
      ],
    );
  }
}

// ── _Badge ────────────────────────────────────────────────────────────────────

class _Badge extends StatelessWidget {
  const _Badge({required this.milestone, required this.unlocked});

  final int milestone;
  final bool unlocked;

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          width: 52,
          height: 52,
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            color: unlocked ? kAccent.withAlpha(30) : Colors.white.withAlpha(10),
            border: Border.all(
              color: unlocked ? kAccent : Colors.white24,
              width: unlocked ? 1.5 : 1,
            ),
          ),
          child: Center(
            child: unlocked
                ? Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      const Text('⭐', style: TextStyle(fontSize: 16)),
                      Text(
                        '$milestone',
                        style: const TextStyle(
                          color: kAccent,
                          fontSize: 11,
                          fontWeight: FontWeight.bold,
                          height: 1.1,
                        ),
                      ),
                    ],
                  )
                : const Icon(
                    Icons.lock_outline_rounded,
                    size: 20,
                    color: Colors.white24,
                  ),
          ),
        ),
        const SizedBox(height: 6),
        Text(
          '$milestone',
          style: TextStyle(
            color: unlocked ? kAccent : kTextMuted,
            fontSize: 11,
            fontWeight: unlocked ? FontWeight.w600 : FontWeight.normal,
          ),
        ),
      ],
    );
  }
}
