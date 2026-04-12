import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/session_record.dart';
import '../providers/session_history_provider.dart';

class HistoryScreen extends ConsumerWidget {
  const HistoryScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final sessions = ref.watch(sessionHistoryProvider);

    return Scaffold(
      backgroundColor: const Color(0xFF121212),
      appBar: AppBar(
        backgroundColor: const Color(0xFF1E1E1E),
        title: const Text(
          'Sessioni',
          style: TextStyle(color: Colors.white, fontWeight: FontWeight.w600),
        ),
        elevation: 0,
        actions: sessions.isEmpty
            ? null
            : [
                IconButton(
                  icon: const Icon(Icons.delete_sweep_outlined,
                      color: Colors.white54),
                  tooltip: 'Cancella tutto',
                  onPressed: () => _confirmClearAll(context, ref),
                ),
              ],
      ),
      body: sessions.isEmpty
          ? const _EmptyState()
          : ListView.builder(
              padding:
                  const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
              itemCount: sessions.length,
              itemBuilder: (context, index) {
                final session = sessions[index];
                return _DismissibleCard(
                  session: session,
                  onDismissed: () =>
                      ref.read(sessionHistoryProvider.notifier).remove(session.id),
                  onTap: () => _showDetailSheet(context, session),
                );
              },
            ),
    );
  }

  void _showDetailSheet(BuildContext context, SessionRecord session) {
    showModalBottomSheet<void>(
      context: context,
      backgroundColor: const Color(0xFF1E1E1E),
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      isScrollControlled: true,
      builder: (_) => _SessionDetailSheet(session: session),
    );
  }

  void _confirmClearAll(BuildContext context, WidgetRef ref) {
    showDialog<void>(
      context: context,
      builder: (ctx) => AlertDialog(
        backgroundColor: const Color(0xFF1E1E1E),
        title: const Text('Cancella tutto',
            style: TextStyle(color: Colors.white)),
        content: const Text(
          'Eliminare tutta la cronologia delle sessioni?',
          style: TextStyle(color: Colors.white70),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(),
            child: const Text('Annulla',
                style: TextStyle(color: Colors.white54)),
          ),
          TextButton(
            onPressed: () {
              ref.read(sessionHistoryProvider.notifier).clearAll();
              Navigator.of(ctx).pop();
            },
            child: const Text('Elimina',
                style: TextStyle(color: Colors.redAccent)),
          ),
        ],
      ),
    );
  }
}

// ── Dismissible card ─────────────────────────────────────────────────────────

class _DismissibleCard extends StatelessWidget {
  const _DismissibleCard({
    required this.session,
    required this.onDismissed,
    required this.onTap,
  });

  final SessionRecord session;
  final VoidCallback onDismissed;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Dismissible(
      key: ValueKey(session.id),
      direction: DismissDirection.endToStart,
      background: Container(
        alignment: Alignment.centerRight,
        padding: const EdgeInsets.only(right: 20),
        margin: const EdgeInsets.only(bottom: 12),
        decoration: BoxDecoration(
          color: Colors.redAccent.withAlpha(40),
          borderRadius: BorderRadius.circular(12),
        ),
        child: const Icon(Icons.delete_outline, color: Colors.redAccent),
      ),
      onDismissed: (_) => onDismissed(),
      child: _SessionCard(session: session, onTap: onTap),
    );
  }
}

// ── Session card ─────────────────────────────────────────────────────────────

class _SessionCard extends StatelessWidget {
  const _SessionCard({required this.session, required this.onTap});

  final SessionRecord session;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Card(
      color: const Color(0xFF1E1E1E),
      elevation: 2,
      margin: const EdgeInsets.only(bottom: 12),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Expanded(
                    child: Text(
                      session.equipmentName.isEmpty
                          ? 'Sessione'
                          : session.equipmentName,
                      style: const TextStyle(
                        color: Colors.white,
                        fontWeight: FontWeight.w600,
                        fontSize: 15,
                      ),
                    ),
                  ),
                  _StatusBadge(status: session.status),
                ],
              ),
              if (session.summary.isNotEmpty) ...[
                const SizedBox(height: 6),
                Text(
                  session.summary,
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                  style: TextStyle(
                    color: Colors.white.withValues(alpha: 0.7),
                    fontSize: 13,
                    height: 1.4,
                  ),
                ),
              ],
              const SizedBox(height: 12),
              Row(
                children: [
                  Icon(Icons.calendar_today_outlined,
                      size: 13,
                      color: Colors.white.withValues(alpha: 0.45)),
                  const SizedBox(width: 4),
                  Text(
                    _formatDate(session.date),
                    style: TextStyle(
                        color: Colors.white.withValues(alpha: 0.45),
                        fontSize: 12),
                  ),
                  const Spacer(),
                  Icon(Icons.timer_outlined,
                      size: 13,
                      color: Colors.white.withValues(alpha: 0.45)),
                  const SizedBox(width: 4),
                  Text(
                    _formatDuration(session.duration),
                    style: TextStyle(
                        color: Colors.white.withValues(alpha: 0.45),
                        fontSize: 12),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}

// ── Status badge ─────────────────────────────────────────────────────────────

class _StatusBadge extends StatelessWidget {
  const _StatusBadge({required this.status});

  final SessionStatus status;

  @override
  Widget build(BuildContext context) {
    final isResolved = status == SessionStatus.resolved;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(
        color: isResolved
            ? const Color(0xFF1B5E20).withValues(alpha: 0.6)
            : const Color(0xFF7F1D1D).withValues(alpha: 0.6),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(
          color: isResolved
              ? const Color(0xFF4CAF50).withValues(alpha: 0.5)
              : const Color(0xFFEF5350).withValues(alpha: 0.5),
          width: 1,
        ),
      ),
      child: Text(
        isResolved ? 'Risolto' : 'Non risolto',
        style: TextStyle(
          color: isResolved
              ? const Color(0xFF81C784)
              : const Color(0xFFEF9A9A),
          fontSize: 11,
          fontWeight: FontWeight.w600,
        ),
      ),
    );
  }
}

// ── Detail sheet ─────────────────────────────────────────────────────────────

class _SessionDetailSheet extends StatelessWidget {
  const _SessionDetailSheet({required this.session});

  final SessionRecord session;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(24, 20, 24, 32),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Center(
            child: Container(
              width: 40,
              height: 4,
              decoration: BoxDecoration(
                color: Colors.white24,
                borderRadius: BorderRadius.circular(2),
              ),
            ),
          ),
          const SizedBox(height: 20),
          Row(
            children: [
              Expanded(
                child: Text(
                  session.equipmentName.isEmpty
                      ? 'Sessione'
                      : session.equipmentName,
                  style: const TextStyle(
                    color: Colors.white,
                    fontWeight: FontWeight.w700,
                    fontSize: 18,
                  ),
                ),
              ),
              _StatusBadge(status: session.status),
            ],
          ),
          const SizedBox(height: 16),
          _DetailRow(
            icon: Icons.calendar_today_outlined,
            label: 'Data',
            value: _formatDate(session.date),
          ),
          const SizedBox(height: 10),
          _DetailRow(
            icon: Icons.timer_outlined,
            label: 'Durata',
            value: _formatDuration(session.duration),
          ),
          if (session.summary.isNotEmpty) ...[
            const SizedBox(height: 16),
            Text(
              'Analisi AI',
              style: TextStyle(
                color: Colors.white.withValues(alpha: 0.5),
                fontSize: 12,
                fontWeight: FontWeight.w600,
                letterSpacing: 0.5,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              session.summary,
              style: const TextStyle(
                color: Colors.white,
                fontSize: 14,
                height: 1.5,
              ),
            ),
          ],
        ],
      ),
    );
  }
}

// ── Detail row ────────────────────────────────────────────────────────────────

class _DetailRow extends StatelessWidget {
  const _DetailRow({
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
        Icon(icon, size: 16, color: Colors.white54),
        const SizedBox(width: 8),
        Text('$label: ',
            style: const TextStyle(color: Colors.white54, fontSize: 13)),
        Text(value,
            style: const TextStyle(
                color: Colors.white,
                fontSize: 13,
                fontWeight: FontWeight.w500)),
      ],
    );
  }
}

// ── Empty state ───────────────────────────────────────────────────────────────

class _EmptyState extends StatelessWidget {
  const _EmptyState();

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.history,
              size: 80, color: Colors.white.withValues(alpha: 0.2)),
          const SizedBox(height: 16),
          Text(
            'Nessuna sessione',
            style: TextStyle(
                color: Colors.white.withValues(alpha: 0.4),
                fontSize: 16,
                fontWeight: FontWeight.w500),
          ),
          const SizedBox(height: 8),
          Text(
            'Le sessioni completate appariranno qui.',
            style: TextStyle(
                color: Colors.white.withValues(alpha: 0.3), fontSize: 13),
          ),
        ],
      ),
    );
  }
}

// ── Helpers ───────────────────────────────────────────────────────────────────

String _formatDate(DateTime date) {
  const months = [
    'gen', 'feb', 'mar', 'apr', 'mag', 'giu',
    'lug', 'ago', 'set', 'ott', 'nov', 'dic',
  ];
  return '${date.day} ${months[date.month - 1]} ${date.year}';
}

String _formatDuration(Duration duration) {
  final minutes = duration.inMinutes;
  if (minutes < 60) return '$minutes min';
  final hours = duration.inHours;
  final remaining = minutes % 60;
  return remaining == 0 ? '${hours}h' : '${hours}h ${remaining}min';
}
