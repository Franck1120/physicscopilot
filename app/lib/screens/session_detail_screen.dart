import 'package:flutter/material.dart';

import '../main.dart' show kAccent, kBgCard, kBgCardBorder, kBgPrimary, kTextMuted;
import '../models/session_record.dart';

/// Full-detail view for a single [SessionRecord].
///
/// Displays all available fields: ID, device name, status, timestamps,
/// problem description, AI summary and a [LinearProgressIndicator] for
/// step progress (simulated as 0 % for unresolved, 100 % for resolved).
class SessionDetailScreen extends StatelessWidget {
  const SessionDetailScreen({super.key, required this.session});

  /// The session record to display.
  final SessionRecord session;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: kBgPrimary,
      appBar: AppBar(
        backgroundColor: const Color(0xFF1E1E1E),
        title: const Text(
          'Dettaglio sessione',
          style: TextStyle(color: Colors.white, fontWeight: FontWeight.w600),
        ),
        iconTheme: const IconThemeData(color: Colors.white),
        elevation: 0,
      ),
      body: SafeArea(
        child: ListView(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 20),
          children: [
            _HeaderCard(session: session),
            const SizedBox(height: 16),
            _ProgressSection(session: session),
            const SizedBox(height: 16),
            if (session.problemDescription.isNotEmpty) ...[
              _InfoSection(
                title: 'PROBLEMA RILEVATO',
                child: Text(
                  session.problemDescription,
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 14,
                    height: 1.5,
                  ),
                ),
              ),
              const SizedBox(height: 16),
            ],
            if (session.summary.isNotEmpty) ...[
              _InfoSection(
                title: 'ANALISI AI',
                child: Text(
                  session.summary,
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 14,
                    height: 1.5,
                  ),
                ),
              ),
              const SizedBox(height: 16),
            ],
            _MetadataCard(session: session),
          ],
        ),
      ),
    );
  }
}

// ── Header card ───────────────────────────────────────────────────────────────

class _HeaderCard extends StatelessWidget {
  const _HeaderCard({required this.session});

  final SessionRecord session;

  @override
  Widget build(BuildContext context) {
    final isResolved = session.status == SessionStatus.resolved;
    final deviceName =
        session.equipmentName.isEmpty ? 'Sessione' : session.equipmentName;

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: kBgCard,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: kBgCardBorder),
      ),
      child: Row(
        children: [
          Icon(
            isResolved ? Icons.check_circle_outline : Icons.cancel_outlined,
            color: isResolved ? kAccent : Colors.redAccent,
            size: 28,
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  deviceName,
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 17,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  isResolved ? 'Risolta' : 'Non risolta',
                  style: TextStyle(
                    color: isResolved ? kAccent : Colors.redAccent,
                    fontSize: 13,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

// ── Progress section ──────────────────────────────────────────────────────────

/// Shows step progress as a [LinearProgressIndicator].
///
/// Resolved sessions show 100 %, unresolved 50 % to indicate work in progress.
class _ProgressSection extends StatelessWidget {
  const _ProgressSection({required this.session});

  final SessionRecord session;

  @override
  Widget build(BuildContext context) {
    final isResolved = session.status == SessionStatus.resolved;
    final progress = isResolved ? 1.0 : 0.5;
    final label = isResolved ? 'Completato' : 'In corso';

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
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              const Text(
                'PROGRESSO',
                style: TextStyle(
                  color: kTextMuted,
                  fontSize: 11,
                  fontWeight: FontWeight.w600,
                  letterSpacing: 0.8,
                ),
              ),
              Text(
                label,
                style: const TextStyle(
                  color: kTextMuted,
                  fontSize: 12,
                ),
              ),
            ],
          ),
          const SizedBox(height: 10),
          ClipRRect(
            borderRadius: BorderRadius.circular(4),
            child: LinearProgressIndicator(
              value: progress,
              backgroundColor: const Color(0xFF2A2A2A),
              valueColor: AlwaysStoppedAnimation<Color>(
                isResolved ? kAccent : Colors.orangeAccent,
              ),
              minHeight: 6,
            ),
          ),
        ],
      ),
    );
  }
}

// ── Info section ──────────────────────────────────────────────────────────────

class _InfoSection extends StatelessWidget {
  const _InfoSection({required this.title, required this.child});

  final String title;
  final Widget child;

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
          Text(
            title,
            style: const TextStyle(
              color: kTextMuted,
              fontSize: 11,
              fontWeight: FontWeight.w600,
              letterSpacing: 0.8,
            ),
          ),
          const SizedBox(height: 10),
          child,
        ],
      ),
    );
  }
}

// ── Metadata card ─────────────────────────────────────────────────────────────

class _MetadataCard extends StatelessWidget {
  const _MetadataCard({required this.session});

  final SessionRecord session;

  @override
  Widget build(BuildContext context) {
    final months = const [
      'gen', 'feb', 'mar', 'apr', 'mag', 'giu',
      'lug', 'ago', 'set', 'ott', 'nov', 'dic',
    ];
    final d = session.date;
    final dateStr = '${d.day} ${months[d.month - 1]} ${d.year}';
    final minutes = session.duration.inMinutes;
    final durationStr = minutes < 60
        ? '$minutes min'
        : '${session.duration.inHours}h ${minutes % 60}min';

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
          const Text(
            'DETTAGLI',
            style: TextStyle(
              color: kTextMuted,
              fontSize: 11,
              fontWeight: FontWeight.w600,
              letterSpacing: 0.8,
            ),
          ),
          const SizedBox(height: 12),
          _MetaRow(icon: Icons.tag, label: 'ID sessione', value: session.id),
          const SizedBox(height: 8),
          _MetaRow(
              icon: Icons.calendar_today_outlined,
              label: 'Data',
              value: dateStr),
          const SizedBox(height: 8),
          _MetaRow(
              icon: Icons.timer_outlined,
              label: 'Durata',
              value: durationStr),
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
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Icon(icon, size: 15, color: Colors.white38),
        const SizedBox(width: 8),
        Text(
          '$label: ',
          style: const TextStyle(color: Colors.white54, fontSize: 13),
        ),
        Expanded(
          child: Text(
            value,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 13,
              fontWeight: FontWeight.w500,
            ),
            overflow: TextOverflow.ellipsis,
          ),
        ),
      ],
    );
  }
}
