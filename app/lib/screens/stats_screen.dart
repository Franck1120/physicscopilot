// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/session_record.dart';
import '../providers/milestone_provider.dart';
import '../providers/session_history_provider.dart';
import '../utils/strings.dart';
import '../widgets/progress_ring_widget.dart';
import '../../main.dart'
    show kAccent, kBgCard, kBgCardBorder, kBgPrimary, kBgSurface, kTextMuted;

// ── StatsScreen ───────────────────────────────────────────────────────────────

/// Displays aggregated statistics derived from the persisted session history.
///
/// Reads from [sessionHistoryProvider] — no local state required.
class StatsScreen extends ConsumerWidget {
  const StatsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final sessions = ref.watch(sessionHistoryProvider);

    final earned = ref.watch(earnedMilestonesProvider);

    return Scaffold(
      backgroundColor: kBgPrimary,
      appBar: AppBar(
        title: const Text('Statistiche'),
        backgroundColor: const Color(0xFF111111),
      ),
      body: sessions.isEmpty
          ? _EmptyStats()
          : _StatsList(sessions: sessions, earnedMilestones: earned),
    );
  }
}

// ── Empty state ───────────────────────────────────────────────────────────────

class _EmptyStats extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Icon(Icons.bar_chart_rounded, size: 52, color: kTextMuted),
          const SizedBox(height: 16),
          Text(
            AppStrings.historyEmpty,
            style: Theme.of(context)
                .textTheme
                .titleMedium
                ?.copyWith(color: Colors.white),
          ),
          const SizedBox(height: 8),
          Text(
            AppStrings.historyEmptySub,
            style: Theme.of(context)
                .textTheme
                .bodySmall
                ?.copyWith(color: kTextMuted),
            textAlign: TextAlign.center,
          ),
        ],
      ),
    );
  }
}

// ── Stats list ────────────────────────────────────────────────────────────────

class _StatsList extends StatelessWidget {
  const _StatsList({
    required this.sessions,
    required this.earnedMilestones,
  });

  final List<SessionRecord> sessions;
  final List<int> earnedMilestones;

  // ── Computed metrics ──────────────────────────────────────────────────────

  int get _totalSessions => sessions.length;

  int get _resolvedCount =>
      sessions.where((s) => s.status == SessionStatus.resolved).length;

  double get _resolvedRatio =>
      _totalSessions == 0 ? 0 : _resolvedCount / _totalSessions;

  Duration get _averageDuration {
    if (_totalSessions == 0) return Duration.zero;
    final totalSecs =
        sessions.fold<int>(0, (sum, s) => sum + s.duration.inSeconds);
    return Duration(seconds: totalSecs ~/ _totalSessions);
  }

  String? get _topDomain {
    if (sessions.isEmpty) return null;
    final freq = <String, int>{};
    for (final s in sessions) {
      freq[s.equipmentName] = (freq[s.equipmentName] ?? 0) + 1;
    }
    return freq.entries
        .reduce((a, b) => a.value >= b.value ? a : b)
        .key;
  }

  List<SessionRecord> get _recent => sessions.take(5).toList();

  String _formatDuration(Duration d) {
    final m = d.inMinutes;
    final s = d.inSeconds % 60;
    if (m > 0) return '${m}m ${s.toString().padLeft(2, '0')}s';
    return '${d.inSeconds}s';
  }

  // ── Build ─────────────────────────────────────────────────────────────────

  @override
  Widget build(BuildContext context) {
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        // ── Top row: total sessions + avg duration ──────────────────────────
        Row(
          children: [
            Expanded(
              child: _StatCard(
                label: 'Sessioni totali',
                value: '$_totalSessions',
                icon: Icons.history_rounded,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: _StatCard(
                label: 'Durata media',
                value: _formatDuration(_averageDuration),
                icon: Icons.timer_outlined,
              ),
            ),
          ],
        ),
        const SizedBox(height: 12),

        // ── Resolution ring ─────────────────────────────────────────────────
        _SectionCard(
          title: 'Risoluzione',
          child: Row(
            children: [
              ProgressRing(
                value: _resolvedRatio,
                size: 80,
                strokeWidth: 7,
                child: Text(
                  '${(_resolvedRatio * 100).round()}%',
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 14,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
              const SizedBox(width: 24),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    _LegendRow(
                      color: kAccent,
                      label: AppStrings.historyStatusResolved,
                      count: _resolvedCount,
                    ),
                    const SizedBox(height: 8),
                    _LegendRow(
                      color: Colors.redAccent,
                      label: AppStrings.historyStatusUnresolved,
                      count: _totalSessions - _resolvedCount,
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 12),

        // ── Top domain badge ────────────────────────────────────────────────
        if (_topDomain != null) ...[
          _SectionCard(
            title: 'Dominio più utilizzato',
            child: Row(
              children: [
                const Icon(Icons.star_rounded, size: 18, color: kAccent),
                const SizedBox(width: 8),
                Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                  decoration: BoxDecoration(
                    color: kAccent.withAlpha(26),
                    borderRadius: BorderRadius.circular(20),
                    border: Border.all(color: kAccent.withAlpha(60)),
                  ),
                  child: Text(
                    _topDomain!,
                    style: const TextStyle(
                      color: kAccent,
                      fontSize: 13,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 12),
        ],

        // ── Milestone badges ────────────────────────────────────────────────
        if (earnedMilestones.isNotEmpty) ...[
          _SectionCard(
            title: 'Traguardi raggiunti',
            child: Wrap(
              spacing: 10,
              runSpacing: 10,
              children: earnedMilestones.map((t) => _MilestoneBadge(threshold: t)).toList(),
            ),
          ),
          const SizedBox(height: 12),
        ],

        // ── Recent sessions ─────────────────────────────────────────────────
        _SectionCard(
          title: 'Ultime sessioni',
          child: Column(
            children: _recent
                .map((s) => _RecentSessionRow(session: s, formatDuration: _formatDuration))
                .toList(),
          ),
        ),
      ],
    );
  }
}

// ── Sub-widgets ───────────────────────────────────────────────────────────────

class _StatCard extends StatelessWidget {
  const _StatCard({
    required this.label,
    required this.value,
    required this.icon,
  });

  final String label;
  final String value;
  final IconData icon;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: kBgCard,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: kBgCardBorder),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(icon, size: 20, color: kAccent),
          const SizedBox(height: 10),
          Text(
            value,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 22,
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 2),
          Text(label,
              style: const TextStyle(color: kTextMuted, fontSize: 12)),
        ],
      ),
    );
  }
}

class _SectionCard extends StatelessWidget {
  const _SectionCard({required this.title, required this.child});

  final String title;
  final Widget child;

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
          Text(
            title.toUpperCase(),
            style: const TextStyle(
              color: kTextMuted,
              fontSize: 11,
              fontWeight: FontWeight.w600,
              letterSpacing: 0.8,
            ),
          ),
          const SizedBox(height: 14),
          child,
        ],
      ),
    );
  }
}

class _LegendRow extends StatelessWidget {
  const _LegendRow({
    required this.color,
    required this.label,
    required this.count,
  });

  final Color color;
  final String label;
  final int count;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Container(
          width: 10,
          height: 10,
          decoration: BoxDecoration(color: color, shape: BoxShape.circle),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: Text(label,
              style:
                  const TextStyle(color: Colors.white70, fontSize: 13)),
        ),
        Text(
          '$count',
          style: TextStyle(
              color: color, fontSize: 13, fontWeight: FontWeight.bold),
        ),
      ],
    );
  }
}

// ── Milestone badge chip ──────────────────────────────────────────────────────

/// Trophy chip shown when the user has reached a session-count [threshold].
class _MilestoneBadge extends StatelessWidget {
  const _MilestoneBadge({required this.threshold});

  final int threshold;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
      decoration: BoxDecoration(
        color: const Color(0xFFFACC15).withAlpha(26),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: const Color(0xFFFACC15).withAlpha(80)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Icon(Icons.emoji_events_rounded,
              size: 14, color: Color(0xFFFACC15)),
          const SizedBox(width: 6),
          Text(
            '$threshold sessioni',
            style: const TextStyle(
              color: Color(0xFFFACC15),
              fontSize: 12,
              fontWeight: FontWeight.w600,
            ),
          ),
        ],
      ),
    );
  }
}

class _RecentSessionRow extends StatelessWidget {
  const _RecentSessionRow({
    required this.session,
    required this.formatDuration,
  });

  final SessionRecord session;
  final String Function(Duration) formatDuration;

  @override
  Widget build(BuildContext context) {
    final isResolved = session.status == SessionStatus.resolved;
    final dotColor = isResolved ? kAccent : Colors.redAccent;

    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 6),
      child: Row(
        children: [
          Container(
            width: 8,
            height: 8,
            decoration: BoxDecoration(color: dotColor, shape: BoxShape.circle),
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Text(
              session.equipmentName.isNotEmpty
                  ? session.equipmentName
                  : 'Sessione',
              style: const TextStyle(color: Colors.white, fontSize: 13),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
          ),
          Text(
            formatDuration(session.duration),
            style: const TextStyle(color: kTextMuted, fontSize: 12),
          ),
        ],
      ),
    );
  }
}
