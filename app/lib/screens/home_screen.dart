import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../main.dart' show kAccent, kAccentDark, kBgPrimary, kBgCard, kBgCardBorder, kTextMuted;
import '../models/session_record.dart';
import '../providers/equipment_provider.dart';
import '../providers/session_history_provider.dart';
import '../providers/websocket_provider.dart';
import '../services/api_service.dart' show serverOnlineProvider;
import '../services/websocket_service.dart';
import '../utils/extensions.dart';
import '../widgets/confetti_overlay.dart';
import '../widgets/session_skeleton.dart';
import 'history_screen.dart';


// ---------------------------------------------------------------------------
// HomeScreen
// ---------------------------------------------------------------------------

class HomeScreen extends ConsumerStatefulWidget {
  const HomeScreen({
    super.key,
    required this.onChangeEquipment,
    required this.onStartCamera,
  });

  final VoidCallback onChangeEquipment;
  final VoidCallback onStartCamera;

  @override
  ConsumerState<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends ConsumerState<HomeScreen> {
  int _selectedIndex = 0;

  static const Color _navBarBackground = Color(0xFF111111);
  static const Color _selectedItemColor = kAccent;
  static const Color _unselectedItemColor = Color(0xFF6B7280);
  static const Color _scaffoldBackground = kBgPrimary;

  void _onItemTapped(int index) {
    if (index == 1) {
      // Camera tab: delegate navigation to the router callback, stay at home.
      widget.onStartCamera();
      return;
    }
    setState(() => _selectedIndex = index);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: _scaffoldBackground,
      body: IndexedStack(
        index: _selectedIndex,
        children: [
          _HomeTab(
            onGoToCamera: () => _onItemTapped(1),
            onChangeEquipment: widget.onChangeEquipment,
          ),
          // Tab 1 — Camera: never rendered; handled by onStartCamera callback.
          const SizedBox.shrink(),
          const HistoryScreen(),
          const _ProfileTab(),
        ],
      ),
      bottomNavigationBar: BottomNavigationBar(
        backgroundColor: _navBarBackground,
        selectedItemColor: _selectedItemColor,
        unselectedItemColor: _unselectedItemColor,
        type: BottomNavigationBarType.fixed,
        currentIndex: _selectedIndex,
        onTap: _onItemTapped,
        items: const [
          BottomNavigationBarItem(
            icon: Icon(Icons.home_outlined),
            activeIcon: Icon(Icons.home),
            label: 'Home',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.camera_alt_outlined),
            activeIcon: Icon(Icons.camera_alt),
            label: 'Camera',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.history_outlined),
            activeIcon: Icon(Icons.history),
            label: 'Cronologia',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.person_outline),
            activeIcon: Icon(Icons.person),
            label: 'Profilo',
          ),
        ],
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Tab 0 — Home
// ---------------------------------------------------------------------------

class _HomeTab extends ConsumerStatefulWidget {
  const _HomeTab({
    required this.onGoToCamera,
    required this.onChangeEquipment,
  });

  final VoidCallback onGoToCamera;
  final VoidCallback onChangeEquipment;

  @override
  ConsumerState<_HomeTab> createState() => _HomeTabState();
}

class _HomeTabState extends ConsumerState<_HomeTab> {
  bool _initialLoading = true;
  bool _showConfetti = false;
  int _lastSessionCount = 0;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) setState(() => _initialLoading = false);
    });
  }

  @override
  Widget build(BuildContext context) {
    final equipment = ref.watch(equipmentProvider);
    final connectionStatus = ref.watch(connectionStatusProvider);
    final sessions = ref.watch(sessionHistoryProvider);

    // Trigger confetti when a 10-session milestone is crossed.
    if (sessions.length != _lastSessionCount) {
      final prev = _lastSessionCount;
      _lastSessionCount = sessions.length;
      if (sessions.length > 0 &&
          sessions.length % 10 == 0 &&
          sessions.length > prev) {
        WidgetsBinding.instance.addPostFrameCallback((_) {
          if (mounted) setState(() => _showConfetti = true);
        });
      }
    }

    final scaffold = Scaffold(
      backgroundColor: kBgPrimary,
      appBar: AppBar(
        backgroundColor: const Color(0xFF111111),
        elevation: 0,
        title: const Text(
          'PhysicsCopilot',
          style: TextStyle(
            color: Colors.white,
            fontWeight: FontWeight.bold,
            letterSpacing: 0.5,
          ),
        ),
        actions: [
          Padding(
            padding: const EdgeInsets.only(right: 16),
            child: _WsStatusChip(status: connectionStatus),
          ),
        ],
      ),
      body: SafeArea(
        child: RefreshIndicator(
          color: kAccent,
          backgroundColor: const Color(0xFF1E1E1E),
          onRefresh: () async {
            ref.invalidate(sessionHistoryProvider);
            ref.invalidate(serverOnlineProvider);
          },
          child: ListView(
            physics: const AlwaysScrollableScrollPhysics(),
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 20),
            children: [
              _NewSessionCard(onTap: widget.onGoToCamera),
              const SizedBox(height: 24),
              _EquipmentSection(
                equipmentName: equipment?.name,
                onChangeEquipment: widget.onChangeEquipment,
              ),
              const SizedBox(height: 16),
              const _ServerStatusBanner(),
              const SizedBox(height: 8),
              const SizedBox(height: 16),
              if (_initialLoading)
                const SessionSkeleton(itemCount: 3)
              else
                const _RecentSessionsSection(),
            ],
          ),
        ),
      ),
    );

    if (!_showConfetti) return scaffold;

    return Stack(
      children: [
        scaffold,
        Positioned.fill(
          child: ConfettiOverlay(
            onComplete: () {
              if (mounted) setState(() => _showConfetti = false);
            },
          ),
        ),
      ],
    );
  }
}

// ---------------------------------------------------------------------------
// Server Status Banner
// ---------------------------------------------------------------------------

/// Shows a warning banner when the server is offline.
///
/// Returns [SizedBox.shrink] when the server is reachable so it takes no
/// space in the layout.
class _ServerStatusBanner extends ConsumerWidget {
  const _ServerStatusBanner();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final isOnline = ref.watch(serverOnlineProvider);

    if (isOnline) return const SizedBox.shrink();

    return Container(
      decoration: BoxDecoration(
        color: const Color(0xFF1A0000),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: Colors.redAccent.withAlpha(80)),
      ),
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
      child: const Row(
        children: [
          Icon(Icons.warning_amber_outlined, color: Colors.redAccent, size: 16),
          SizedBox(width: 8),
          Text(
            'Server non raggiungibile',
            style: TextStyle(color: Colors.redAccent, fontSize: 13),
          ),
        ],
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// New Session Card
// ---------------------------------------------------------------------------

class _NewSessionCard extends StatelessWidget {
  const _NewSessionCard({required this.onTap});

  final VoidCallback onTap;

  static const Color _cardBackground = Color(0xFF064E3B);

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: () {
        HapticFeedback.mediumImpact();
        onTap();
      },
      child: Container(
        width: double.infinity,
        padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 28),
        decoration: BoxDecoration(
          color: _cardBackground,
          borderRadius: BorderRadius.circular(16),
          boxShadow: [
            BoxShadow(
              color: kAccent.withValues(alpha: 0.15),
              blurRadius: 16,
              offset: const Offset(0, 6),
            ),
          ],
        ),
        child: const Row(
          children: [
            Icon(
              Icons.camera_alt_outlined,
              color: Colors.white,
              size: 36,
            ),
            SizedBox(width: 20),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Inizia sessione',
                    style: TextStyle(
                      color: Colors.white,
                      fontSize: 20,
                      fontWeight: FontWeight.bold,
                      letterSpacing: 0.3,
                    ),
                  ),
                  SizedBox(height: 4),
                  Text(
                    'Nuova sessione di analisi',
                    style: TextStyle(
                      color: Color(0xFF6EE7B7),
                      fontSize: 13,
                    ),
                  ),
                ],
              ),
            ),
            Icon(
              Icons.arrow_forward_ios_rounded,
              color: Color(0xFF6EE7B7),
              size: 18,
            ),
          ],
        ),
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Equipment Section
// ---------------------------------------------------------------------------

class _EquipmentSection extends StatelessWidget {
  const _EquipmentSection({
    required this.equipmentName,
    required this.onChangeEquipment,
  });

  final String? equipmentName;
  final VoidCallback onChangeEquipment;

  @override
  Widget build(BuildContext context) {
    final bool hasEquipment = equipmentName != null;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text(
          'DISPOSITIVO ATTIVO',
          style: TextStyle(
            color: kTextMuted,
            fontSize: 11,
            fontWeight: FontWeight.w600,
            letterSpacing: 0.8,
          ),
        ),
        const SizedBox(height: 10),
        Container(
          width: double.infinity,
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
          decoration: BoxDecoration(
            color: kBgCard,
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: kBgCardBorder, width: 1),
          ),
          child: Row(
            children: [
              const Icon(
                Icons.build_outlined,
                color: kAccent,
                size: 22,
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Text(
                  equipmentName ?? 'Nessun dispositivo selezionato',
                  style: TextStyle(
                    color: hasEquipment ? Colors.white : kTextMuted,
                    fontSize: 15,
                    fontWeight:
                        hasEquipment ? FontWeight.w500 : FontWeight.normal,
                  ),
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              const SizedBox(width: 8),
              GestureDetector(
                onTap: onChangeEquipment,
                child: Chip(
                  label: const Text(
                    'Cambia',
                    style: TextStyle(
                      color: kAccent,
                      fontSize: 12,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                  backgroundColor: kBgPrimary,
                  side: const BorderSide(color: kAccent, width: 1),
                  padding: const EdgeInsets.symmetric(horizontal: 4),
                  visualDensity: VisualDensity.compact,
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

// ---------------------------------------------------------------------------
// Recent Sessions Section
// ---------------------------------------------------------------------------

/// Reads the last 3 sessions from [sessionHistoryProvider] and renders them.
///
/// Renders [_NoSessionsCard] when the history is empty, or a list of
/// [_SessionMiniCard] widgets with an optional "Vedi tutte" link otherwise.
class _RecentSessionsSection extends ConsumerWidget {
  const _RecentSessionsSection();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final sessions = ref.watch(sessionHistoryProvider);
    final recent = sessions.take(3).toList();

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text(
          'ULTIME SESSIONI',
          style: TextStyle(
            color: kTextMuted,
            fontSize: 11,
            fontWeight: FontWeight.w600,
            letterSpacing: 0.8,
          ),
        ),
        const SizedBox(height: 10),
        if (sessions.isEmpty)
          const _NoSessionsCard()
        else
          Column(
            children: [
              for (final session in recent)
                Padding(
                  padding: const EdgeInsets.only(bottom: 8),
                  child: _SessionMiniCard(session: session),
                ),
              if (sessions.length > 3)
                Align(
                  alignment: Alignment.centerRight,
                  child: TextButton(
                    onPressed: () => context.push('/history'),
                    child: const Text(
                      'Vedi tutte →',
                      style: TextStyle(
                        color: kAccent,
                        fontSize: 13,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ),
                ),
            ],
          ),
      ],
    );
  }
}

// ---------------------------------------------------------------------------
// Session Mini Card
// ---------------------------------------------------------------------------

/// Compact card for a single [SessionRecord] shown in the recent sessions list.
class _SessionMiniCard extends StatelessWidget {
  const _SessionMiniCard({required this.session});

  final SessionRecord session;

  @override
  Widget build(BuildContext context) {
    final isResolved = session.status == SessionStatus.resolved;

    return Container(
      decoration: BoxDecoration(
        color: kBgCard,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: kBgCardBorder, width: 1),
      ),
      padding: const EdgeInsets.all(14),
      child: Row(
        children: [
          Icon(
            isResolved ? Icons.check_circle_outline : Icons.cancel_outlined,
            color: isResolved ? kAccent : Colors.redAccent,
            size: 20,
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  session.equipmentName.isEmpty
                      ? 'Sessione'
                      : session.equipmentName,
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 14,
                    fontWeight: FontWeight.w500,
                  ),
                ),
                Text(
                  session.date.relativeLabel,
                  style: const TextStyle(color: kTextMuted, fontSize: 12),
                ),
              ],
            ),
          ),
          const SizedBox(width: 8),
          Text(
            session.duration.formatted,
            style: const TextStyle(color: kTextMuted, fontSize: 12),
          ),
        ],
      ),
    );
  }
}

class _NoSessionsCard extends StatelessWidget {
  const _NoSessionsCard();

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 24),
      decoration: BoxDecoration(
        color: kBgCard,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: kBgCardBorder, width: 1),
      ),
      child: const Column(
        children: [
          Icon(Icons.history_outlined, color: kTextMuted, size: 32),
          SizedBox(height: 10),
          Text(
            'Nessuna sessione recente',
            style: TextStyle(
              color: kTextMuted,
              fontSize: 13,
            ),
          ),
          SizedBox(height: 4),
          Text(
            'Le tue sessioni di diagnosi appariranno qui.',
            style: TextStyle(color: kTextMuted, fontSize: 11),
            textAlign: TextAlign.center,
          ),
        ],
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Tab 3 — Profile
// ---------------------------------------------------------------------------

class _ProfileTab extends StatelessWidget {
  const _ProfileTab();

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: kBgPrimary,
      appBar: AppBar(
        backgroundColor: kBgCard,
        elevation: 0,
        title: const Text(
          'Profilo',
          style: TextStyle(
            color: Colors.white,
            fontWeight: FontWeight.bold,
          ),
        ),
      ),
      body: const SafeArea(
        child: _ProfileBody(),
      ),
    );
  }
}

class _ProfileBody extends StatelessWidget {
  const _ProfileBody();

  @override
  Widget build(BuildContext context) {
    return ListView(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 24),
      children: const [
        _ProfileHeader(),
        SizedBox(height: 32),
        _ProfileTileList(),
      ],
    );
  }
}

class _ProfileHeader extends StatelessWidget {
  const _ProfileHeader();

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        const CircleAvatar(
          radius: 42,
          backgroundColor: kAccentDark,
          child: Text(
            'U',
            style: TextStyle(
              color: Colors.white,
              fontSize: 32,
              fontWeight: FontWeight.bold,
            ),
          ),
        ),
        const SizedBox(height: 14),
        const Text(
          'Utente',
          style: TextStyle(
            color: Colors.white,
            fontSize: 20,
            fontWeight: FontWeight.bold,
          ),
        ),
        const SizedBox(height: 8),
        Chip(
          label: const Text(
            'Free',
            style: TextStyle(
              color: kAccent,
              fontWeight: FontWeight.w700,
              fontSize: 13,
            ),
          ),
          backgroundColor: kBgPrimary,
          side: const BorderSide(color: kAccent, width: 1.5),
          padding: const EdgeInsets.symmetric(horizontal: 8),
        ),
      ],
    );
  }
}

class _ProfileTileList extends StatelessWidget {
  const _ProfileTileList();

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        _buildTile(
          context,
          icon: Icons.settings_outlined,
          label: 'Impostazioni',
          onTap: () => context.push('/settings'),
        ),
        _buildTile(
          context,
          icon: Icons.lock_outline,
          label: 'Privacy',
          onTap: null,
        ),
        _buildTile(
          context,
          icon: Icons.info_outline,
          label: 'Informazioni app',
          onTap: null,
        ),
      ],
    );
  }

  Widget _buildTile(
    BuildContext context, {
    required IconData icon,
    required String label,
    required VoidCallback? onTap,
  }) {
    return Container(
      margin: const EdgeInsets.only(bottom: 10),
      decoration: BoxDecoration(
        color: kBgCard,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: kBgCardBorder, width: 1),
      ),
      child: ListTile(
        leading: Icon(icon, color: onTap != null ? kAccent : kTextMuted),
        title: Text(
          label,
          style: TextStyle(
            color: onTap != null ? Colors.white : Colors.white54,
            fontSize: 15,
          ),
        ),
        trailing: Icon(
          Icons.arrow_forward_ios_rounded,
          color: onTap != null ? kTextMuted : kTextMuted.withAlpha(80),
          size: 16,
        ),
        enabled: onTap != null,
        onTap: onTap,
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// WebSocket Status Chip — shown in AppBar
// ---------------------------------------------------------------------------

class _WsStatusChip extends StatelessWidget {
  const _WsStatusChip({required this.status});

  final AsyncValue<ConnectionStatus> status;

  @override
  Widget build(BuildContext context) {
    final (color, label, icon) = status.when(
      data: (s) => switch (s) {
        ConnectionStatus.connected => (kAccent, 'Online', Icons.wifi),
        ConnectionStatus.connecting => (
          Colors.orangeAccent,
          'Connessione...',
          Icons.wifi_off,
        ),
        ConnectionStatus.disconnected => (
          Colors.redAccent,
          'Offline',
          Icons.wifi_off,
        ),
      },
      loading: () => (Colors.orangeAccent, 'Connessione...', Icons.wifi_off),
      error: (_, __) => (Colors.redAccent, 'Errore', Icons.error_outline),
    );

    return AnimatedContainer(
      duration: const Duration(milliseconds: 300),
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
      decoration: BoxDecoration(
        color: color.withAlpha(20),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: color.withAlpha(80), width: 1),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 12, color: color),
          const SizedBox(width: 5),
          Text(
            label,
            style: TextStyle(
              color: color,
              fontSize: 11,
              fontWeight: FontWeight.w600,
            ),
          ),
        ],
      ),
    );
  }
}
