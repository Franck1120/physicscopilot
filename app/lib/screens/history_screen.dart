import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:share_plus/share_plus.dart';

import '../models/session_record.dart';
import '../providers/session_history_provider.dart';
import '../screens/session_detail_screen.dart';
import '../services/api_service.dart';
import '../utils/strings.dart';
import '../utils/transitions.dart';
import '../widgets/safe_screen.dart';
import '../widgets/session_skeleton.dart';

// ---------------------------------------------------------------------------
// Server sessions provider — fetches from REST API and converts to SessionRecord.
// Falls back gracefully to an empty list on any error.
// ---------------------------------------------------------------------------

final _serverSessionsProvider = FutureProvider<List<SessionRecord>>((ref) async {
  final api = ref.watch(apiServiceProvider);
  final remotes = await api.listSessions();
  return remotes.map((r) => SessionRecord(
    id: r.sessionId,
    date: r.createdAt,
    equipmentName: r.deviceName,
    problemDescription: r.problemDetected,
    summary: r.problemDetected,
    status: SessionStatus.resolved,
    duration: Duration.zero,
  )).toList();
});

/// Screen that displays the user's session history.
///
/// Merges locally stored sessions ([sessionHistoryProvider]) with sessions
/// fetched from the REST API ([_serverSessionsProvider]). Server-only sessions
/// (not yet in local storage) appear at the top of the list.
///
/// Features:
/// - Full-text search across device name, problem description, and summary.
/// - Filter chips for resolved / unresolved status and by domain (equipment name).
/// - Infinite scroll with client-side pagination ([_pageSize] items per page).
/// - Pull-to-refresh that re-fetches server sessions.
/// - Swipe-to-delete with undo via [ScaffoldMessenger] snack bar.
/// - Tap on a card navigates to [SessionDetailScreen] via [slideFromRight].
///
/// See also:
/// - [SessionDetailScreen] for the per-session detail view.
/// - [sessionHistoryProvider] for the local persistence layer.
class HistoryScreen extends ConsumerStatefulWidget {
  /// Creates the history screen.
  const HistoryScreen({super.key});

  @override
  ConsumerState<HistoryScreen> createState() => _HistoryScreenState();
}

class _HistoryScreenState extends ConsumerState<HistoryScreen> {
  static const int _pageSize = 20;

  final ScrollController _scrollController = ScrollController();
  final TextEditingController _searchController = TextEditingController();
  int _page = 1;
  bool _hasMore = true;
  bool _isLoadingMore = false;
  String _searchQuery = '';
  // null = all, 'resolved', 'unresolved'
  String? _statusFilter;
  // null = all domains
  String? _domainFilter;

  @override
  void initState() {
    super.initState();
    _scrollController.addListener(_onScroll);
    _searchController.addListener(() {
      setState(() => _searchQuery = _searchController.text);
    });
  }

  @override
  void dispose() {
    _scrollController.dispose();
    _searchController.dispose();
    super.dispose();
  }

  void _onScroll() {
    if (!_scrollController.hasClients) return;
    final maxExtent = _scrollController.position.maxScrollExtent;
    final current = _scrollController.offset;
    if (current >= maxExtent - 200 && _hasMore && !_isLoadingMore) {
      _loadNextPage();
    }
  }

  Future<void> _loadNextPage() async {
    setState(() => _isLoadingMore = true);
    // Simulate async page load (local pagination — no server paging API).
    await Future.delayed(const Duration(milliseconds: 300));
    if (!mounted) return;
    setState(() {
      _page += 1;
      _isLoadingMore = false;
    });
  }

  void _resetPagination() {
    setState(() {
      _page = 1;
      _hasMore = true;
      _isLoadingMore = false;
    });
  }

  @override
  Widget build(BuildContext context) {
    try {
      return _buildContent(context);
    } catch (e) {
      return screenError(e, context);
    }
  }

  Widget _buildContent(BuildContext context) {
    final localSessions = ref.watch(sessionHistoryProvider);
    final serverAsync = ref.watch(_serverSessionsProvider);

    // Merge: server sessions not already in local appear at the top.
    final allSessions = serverAsync.when(
      data: (serverSessions) {
        final localIds = localSessions.map((s) => s.id).toSet();
        final extra = serverSessions
            .where((s) => !localIds.contains(s.id))
            .toList();
        return [...extra, ...localSessions];
      },
      loading: () => localSessions,
      error: (_, __) => localSessions,
    );

    final isSyncing = serverAsync.isLoading;

    // Collect unique domain names for the domain filter chips.
    final domains = allSessions
        .map((s) => s.equipmentName)
        .where((n) => n.isNotEmpty)
        .toSet()
        .toList()
      ..sort();

    // Apply search + status + domain filters.
    final filteredSessions = allSessions.where((s) {
      // Search query
      if (_searchQuery.isNotEmpty) {
        final q = _searchQuery.toLowerCase();
        final matches = s.equipmentName.toLowerCase().contains(q) ||
            s.problemDescription.toLowerCase().contains(q) ||
            s.summary.toLowerCase().contains(q);
        if (!matches) return false;
      }
      // Status filter
      if (_statusFilter != null) {
        final isResolved = s.status == SessionStatus.resolved;
        if (_statusFilter == 'resolved' && !isResolved) return false;
        if (_statusFilter == 'unresolved' && isResolved) return false;
      }
      // Domain filter
      if (_domainFilter != null && s.equipmentName != _domainFilter) {
        return false;
      }
      return true;
    }).toList();

    // Paginate locally: show up to _page * _pageSize items.
    final visibleCount = (_page * _pageSize).clamp(0, filteredSessions.length);
    final sessions = filteredSessions.take(visibleCount).toList();
    // Update _hasMore without triggering an extra setState during build.
    final hasMore = visibleCount < filteredSessions.length;
    if (_hasMore != hasMore) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted) setState(() => _hasMore = hasMore);
      });
    }

    return Scaffold(
      backgroundColor: const Color(0xFF121212),
      appBar: AppBar(
        backgroundColor: const Color(0xFF1E1E1E),
        title: Row(
          children: [
            const Text(
              AppStrings.historyTitle,
              style: TextStyle(color: Colors.white, fontWeight: FontWeight.w600),
            ),
            if (isSyncing) ...[
              const SizedBox(width: 10),
              const SizedBox(
                width: 14,
                height: 14,
                child: CircularProgressIndicator(
                  strokeWidth: 2,
                  color: Colors.white54,
                ),
              ),
            ],
          ],
        ),
        elevation: 0,
        actions: localSessions.isEmpty
            ? null
            : [
                IconButton(
                  icon: const Icon(Icons.delete_sweep_outlined,
                      color: Colors.white54),
                  tooltip: AppStrings.historyClearAll,
                  onPressed: () => _confirmClearAll(context),
                ),
              ],
      ),
      body: Column(
        children: [
          _SearchBar(controller: _searchController),
          _FilterChipsRow(
            statusFilter: _statusFilter,
            domainFilter: _domainFilter,
            domains: domains,
            onStatusChanged: (v) => setState(() {
              _statusFilter = v;
              _resetPagination();
            }),
            onDomainChanged: (v) => setState(() {
              _domainFilter = v;
              _resetPagination();
            }),
          ),
          Expanded(
            child: RefreshIndicator(
        color: const Color(0xFF10B981),
        backgroundColor: const Color(0xFF1E1E1E),
        onRefresh: () async {
          _resetPagination();
          ref.invalidate(_serverSessionsProvider);
          await ref.read(_serverSessionsProvider.future).catchError((_) {});
        },
        child: isSyncing && localSessions.isEmpty
            // Show skeleton while loading and no cached data.
            ? const SessionSkeleton(itemCount: 5)
            : sessions.isEmpty
            ? const SingleChildScrollView(
                physics: AlwaysScrollableScrollPhysics(),
                child: SizedBox(
                  height: 500,
                  child: _EmptyState(),
                ),
              )
            : ListView.builder(
                controller: _scrollController,
                physics: const AlwaysScrollableScrollPhysics(),
                padding:
                    const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                itemCount: sessions.length + (_hasMore ? 1 : 0),
                itemBuilder: (context, index) {
                  if (index == sessions.length) {
                    return const Padding(
                      padding: EdgeInsets.symmetric(vertical: 16),
                      child: Center(
                        child: SizedBox(
                          width: 24,
                          height: 24,
                          child: CircularProgressIndicator(
                            strokeWidth: 2,
                            color: Color(0xFF10B981),
                          ),
                        ),
                      ),
                    );
                  }
                  final session = sessions[index];
                  return _DismissibleCard(
                    session: session,
                    onDismissed: () => _deleteWithUndo(context, session),
                    onTap: () => _showDetailSheet(context, session),
                  );
                },
              ),
            ),
          ),
        ],
      ),
    );
  }

  void _deleteWithUndo(BuildContext context, SessionRecord session) {
    // Remove immediately from local provider.
    ref.read(sessionHistoryProvider.notifier).remove(session.id);

    final messenger = ScaffoldMessenger.of(context);
    messenger.clearSnackBars();
    messenger.showSnackBar(
      SnackBar(
        content: Text(
          'Sessione eliminata',
          style: const TextStyle(color: Colors.white),
        ),
        backgroundColor: const Color(0xFF1E1E1E),
        behavior: SnackBarBehavior.floating,
        duration: const Duration(seconds: 4),
        action: SnackBarAction(
          label: 'Annulla',
          textColor: const Color(0xFF10B981),
          onPressed: () {
            // Re-insert the session into the provider list.
            ref.read(sessionHistoryProvider.notifier).add(session);
          },
        ),
      ),
    );
  }

  void _showDetailSheet(BuildContext context, SessionRecord session) {
    Navigator.of(context).push(
      slideFromRight(SessionDetailScreen(session: session)),
    );
  }

  void _confirmClearAll(BuildContext context) {
    showDialog<void>(
      context: context,
      builder: (ctx) => AlertDialog(
        backgroundColor: const Color(0xFF1E1E1E),
        title: const Text(AppStrings.historyClearAll,
            style: TextStyle(color: Colors.white)),
        content: const Text(
          AppStrings.historyClearConfirm,
          style: TextStyle(color: Colors.white70),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(),
            child: const Text(AppStrings.cancel,
                style: TextStyle(color: Colors.white54)),
          ),
          TextButton(
            onPressed: () {
              ref.read(sessionHistoryProvider.notifier).clearAll();
              Navigator.of(ctx).pop();
            },
            child: const Text(AppStrings.delete,
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
    final name =
        session.equipmentName.isEmpty ? 'Sessione' : session.equipmentName;
    return Semantics(
      label: name,
      hint: 'Scorri verso sinistra per eliminare. Tocca per i dettagli.',
      child: Dismissible(
        key: ValueKey(session.id),
        direction: DismissDirection.endToStart,
        background: Container(
          alignment: Alignment.centerRight,
          padding: const EdgeInsets.only(right: 24),
          margin: const EdgeInsets.only(bottom: 12),
          decoration: BoxDecoration(
            color: const Color(0xFFB71C1C),
            borderRadius: BorderRadius.circular(12),
          ),
          child: const Row(
            mainAxisAlignment: MainAxisAlignment.end,
            children: [
              Icon(Icons.delete_outline, color: Colors.white, size: 22),
              SizedBox(width: 6),
              Text(
                'Elimina',
                style: TextStyle(
                  color: Colors.white,
                  fontWeight: FontWeight.w600,
                  fontSize: 13,
                ),
              ),
            ],
          ),
        ),
        onDismissed: (_) => onDismissed(),
        child: _SessionCard(session: session, onTap: onTap),
      ),
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
        isResolved ? AppStrings.historyStatusResolved : AppStrings.historyStatusUnresolved,
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

  String _buildReport() {
    final buf = StringBuffer();
    final deviceName =
        session.equipmentName.isEmpty ? 'N/D' : session.equipmentName;
    final status =
        session.status == SessionStatus.resolved ? AppStrings.historyStatusResolved : AppStrings.historyStatusUnresolved;

    buf.writeln('PhysicsCopilot — Report Sessione');
    buf.writeln('=================================');
    buf.writeln('Data:       ${_formatDate(session.date)}');
    buf.writeln('Durata:     ${_formatDuration(session.duration)}');
    buf.writeln('Dispositivo: $deviceName');
    buf.writeln('Stato:      $status');

    if (session.problemDescription.isNotEmpty) {
      buf.writeln();
      buf.writeln('PROBLEMA RILEVATO');
      buf.writeln('-----------------');
      buf.writeln(session.problemDescription);
    }

    if (session.summary.isNotEmpty) {
      buf.writeln();
      buf.writeln('ANALISI AI');
      buf.writeln('----------');
      buf.writeln(session.summary);
    }

    buf.writeln();
    buf.writeln('--- Generato da PhysicsCopilot ---');
    return buf.toString();
  }

  Future<void> _export(BuildContext context) async {
    final report = _buildReport();
    final box = context.findRenderObject() as RenderBox?;
    await Share.share(
      report,
      subject: 'Report Sessione — ${session.equipmentName}',
      sharePositionOrigin:
          box == null ? null : box.localToGlobal(Offset.zero) & box.size,
    );
  }

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
          const SizedBox(height: 24),
          SizedBox(
            width: double.infinity,
            child: OutlinedButton.icon(
              onPressed: () => _export(context),
              icon: const Icon(Icons.ios_share_outlined, size: 18),
              label: const Text('Esporta sessione'),
              style: OutlinedButton.styleFrom(
                foregroundColor: const Color(0xFF10B981),
                side: const BorderSide(color: Color(0xFF10B981)),
                padding: const EdgeInsets.symmetric(vertical: 12),
              ),
            ),
          ),
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
            AppStrings.historyEmpty,
            style: TextStyle(
                color: Colors.white.withValues(alpha: 0.4),
                fontSize: 16,
                fontWeight: FontWeight.w500),
          ),
          const SizedBox(height: 8),
          Text(
            AppStrings.historyEmptySub,
            style: TextStyle(
                color: Colors.white.withValues(alpha: 0.3), fontSize: 13),
          ),
        ],
      ),
    );
  }
}

// ── Filter chips ──────────────────────────────────────────────────────────────

class _FilterChipsRow extends StatelessWidget {
  const _FilterChipsRow({
    required this.statusFilter,
    required this.domainFilter,
    required this.domains,
    required this.onStatusChanged,
    required this.onDomainChanged,
  });

  final String? statusFilter;
  final String? domainFilter;
  final List<String> domains;
  final ValueChanged<String?> onStatusChanged;
  final ValueChanged<String?> onDomainChanged;

  @override
  Widget build(BuildContext context) {
    const chipColor = Color(0xFF1E1E1E);
    const selectedColor = Color(0xFF10B981);

    Widget statusChip(String label, String? value) {
      final selected = statusFilter == value;
      return Padding(
        padding: const EdgeInsets.only(right: 6),
        child: FilterChip(
          label: Text(label),
          selected: selected,
          onSelected: (_) => onStatusChanged(selected ? null : value),
          backgroundColor: chipColor,
          selectedColor: selectedColor.withAlpha(40),
          checkmarkColor: selectedColor,
          labelStyle: TextStyle(
            color: selected ? selectedColor : Colors.white70,
            fontSize: 12,
          ),
          side: BorderSide(
            color: selected ? selectedColor : Colors.white24,
          ),
          padding: const EdgeInsets.symmetric(horizontal: 4),
          visualDensity: VisualDensity.compact,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(20),
          ),
        ),
      );
    }

    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      padding: const EdgeInsets.fromLTRB(16, 4, 16, 8),
      child: Row(
        children: [
          statusChip('Tutti', null),
          statusChip('Risolti', 'resolved'),
          statusChip('Non risolti', 'unresolved'),
          if (domains.isNotEmpty) ...[
            const SizedBox(width: 8),
            const VerticalDivider(width: 1, color: Colors.white24, indent: 4, endIndent: 4),
            const SizedBox(width: 8),
            for (final domain in domains.take(5))
              Padding(
                padding: const EdgeInsets.only(right: 6),
                child: FilterChip(
                  label: Text(domain, overflow: TextOverflow.ellipsis),
                  selected: domainFilter == domain,
                  onSelected: (_) =>
                      onDomainChanged(domainFilter == domain ? null : domain),
                  backgroundColor: chipColor,
                  selectedColor: selectedColor.withAlpha(40),
                  checkmarkColor: selectedColor,
                  labelStyle: TextStyle(
                    color: domainFilter == domain ? selectedColor : Colors.white70,
                    fontSize: 12,
                  ),
                  side: BorderSide(
                    color: domainFilter == domain ? selectedColor : Colors.white24,
                  ),
                  padding: const EdgeInsets.symmetric(horizontal: 4),
                  visualDensity: VisualDensity.compact,
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(20),
                  ),
                ),
              ),
          ],
        ],
      ),
    );
  }
}

// ── Search bar ────────────────────────────────────────────────────────────────

class _SearchBar extends StatelessWidget {
  const _SearchBar({required this.controller});

  final TextEditingController controller;

  @override
  Widget build(BuildContext context) {
    return Container(
      color: const Color(0xFF121212),
      padding: const EdgeInsets.fromLTRB(16, 10, 16, 6),
      child: TextField(
        controller: controller,
        style: const TextStyle(color: Colors.white, fontSize: 14),
        decoration: InputDecoration(
          hintText: 'Cerca per dispositivo o problema...',
          hintStyle: const TextStyle(color: Colors.white38, fontSize: 14),
          prefixIcon: const Icon(Icons.search, color: Colors.white38, size: 20),
          suffixIcon: controller.text.isNotEmpty
              ? IconButton(
                  icon: const Icon(Icons.close, color: Colors.white38, size: 18),
                  onPressed: controller.clear,
                )
              : null,
          filled: true,
          fillColor: const Color(0xFF1E1E1E),
          contentPadding: const EdgeInsets.symmetric(vertical: 10),
          border: OutlineInputBorder(
            borderRadius: BorderRadius.circular(10),
            borderSide: BorderSide.none,
          ),
        ),
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
