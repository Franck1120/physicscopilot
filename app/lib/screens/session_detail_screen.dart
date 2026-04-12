// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

import 'package:flutter/material.dart';
import 'package:share_plus/share_plus.dart';

import 'package:physicscopilot/main.dart'
    show kAccent, kBgCard, kBgCardBorder, kBgPrimary, kTextMuted;
import 'package:physicscopilot/models/session_record.dart';

// ── SessionDetailScreen ───────────────────────────────────────────────────────

/// Displays the full detail of a past [SessionRecord].
///
/// Shows equipment name, formatted date, duration, status badge, problem
/// description, and the AI-generated summary. A share button in the AppBar
/// lets the user export the session info via the platform share sheet.
class SessionDetailScreen extends StatelessWidget {
  const SessionDetailScreen({super.key, required this.session});

  final SessionRecord session;

  // ── Helpers ───────────────────────────────────────────────────────────────

  String _formatDate(DateTime d) {
    const months = [
      'gen', 'feb', 'mar', 'apr', 'mag', 'giu',
      'lug', 'ago', 'set', 'ott', 'nov', 'dic',
    ];
    final h = d.hour.toString().padLeft(2, '0');
    final m = d.minute.toString().padLeft(2, '0');
    return '${d.day} ${months[d.month - 1]} ${d.year}, $h:$m';
  }

  String _formatDuration(Duration d) {
    final min = d.inMinutes;
    final sec = d.inSeconds % 60;
    if (min > 0) return '${min}m ${sec.toString().padLeft(2, '0')}s';
    return '${d.inSeconds}s';
  }

  String _buildShareText() {
    final status =
        session.status == SessionStatus.resolved ? 'Risolto' : 'Non risolto';
    return 'PhysicsCopilot — Sessione\n'
        'Apparecchio: ${session.equipmentName}\n'
        'Data: ${_formatDate(session.date)}\n'
        'Durata: ${_formatDuration(session.duration)}\n'
        'Stato: $status\n\n'
        'Problema:\n${session.problemDescription}\n\n'
        'Sommario AI:\n${session.summary}';
  }

  // ── Build ─────────────────────────────────────────────────────────────────

  @override
  Widget build(BuildContext context) {
    final isResolved = session.status == SessionStatus.resolved;

    return Scaffold(
      backgroundColor: kBgPrimary,
      appBar: AppBar(
        title: Text(
          session.equipmentName.isNotEmpty ? session.equipmentName : 'Sessione',
          overflow: TextOverflow.ellipsis,
        ),
        backgroundColor: const Color(0xFF111111),
        actions: [
          IconButton(
            icon: const Icon(Icons.share_rounded),
            tooltip: 'Condividi',
            onPressed: () {
              Share.share(_buildShareText());
            },
          ),
        ],
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          // ── Meta card ─────────────────────────────────────────────────────
          _InfoCard(
            children: [
              _MetaRow(
                icon: Icons.calendar_today_outlined,
                label: 'Data',
                value: _formatDate(session.date),
              ),
              const SizedBox(height: 12),
              _MetaRow(
                icon: Icons.timer_outlined,
                label: 'Durata',
                value: _formatDuration(session.duration),
              ),
              const SizedBox(height: 12),
              Row(
                children: [
                  Icon(
                    isResolved
                        ? Icons.check_circle_rounded
                        : Icons.cancel_rounded,
                    size: 16,
                    color: isResolved ? kAccent : Colors.redAccent,
                  ),
                  const SizedBox(width: 8),
                  const Text(
                    'Stato',
                    style: TextStyle(color: kTextMuted, fontSize: 12),
                  ),
                  const Spacer(),
                  _StatusBadge(resolved: isResolved),
                ],
              ),
            ],
          ),
          const SizedBox(height: 12),

          // ── Problem description ───────────────────────────────────────────
          _SectionCard(
            title: 'Problema',
            child: Text(
              session.problemDescription.isNotEmpty
                  ? session.problemDescription
                  : '—',
              style: const TextStyle(
                color: Colors.white,
                fontSize: 14,
                height: 1.5,
              ),
            ),
          ),
          const SizedBox(height: 12),

          // ── AI summary ────────────────────────────────────────────────────
          _SectionCard(
            title: 'Sommario AI',
            child: Text(
              session.summary.isNotEmpty ? session.summary : '—',
              style: const TextStyle(
                color: Colors.white,
                fontSize: 14,
                height: 1.5,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

// ── Sub-widgets ───────────────────────────────────────────────────────────────

class _InfoCard extends StatelessWidget {
  const _InfoCard({required this.children});

  final List<Widget> children;

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
        children: children,
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
          const SizedBox(height: 12),
          child,
        ],
      ),
    );
  }
}

class _MetaRow extends StatelessWidget {
  const _MetaRow({
    required this.icon,
    required this.label,
    required this.value,
  });

  final IconData icon;
  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Icon(icon, size: 16, color: kAccent),
        const SizedBox(width: 8),
        Text(
          label,
          style: const TextStyle(color: kTextMuted, fontSize: 12),
        ),
        const Spacer(),
        Text(
          value,
          style: const TextStyle(
            color: Colors.white,
            fontSize: 13,
            fontWeight: FontWeight.w500,
          ),
        ),
      ],
    );
  }
}

class _StatusBadge extends StatelessWidget {
  const _StatusBadge({required this.resolved});

  final bool resolved;

  @override
  Widget build(BuildContext context) {
    final color = resolved ? kAccent : Colors.redAccent;
    final label = resolved ? 'Risolto' : 'Non risolto';

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(
        color: color.withAlpha(30),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: color.withAlpha(80)),
      ),
      child: Text(
        label,
        style: TextStyle(
          color: color,
          fontSize: 12,
          fontWeight: FontWeight.w600,
        ),
      ),
    );
  }
}
