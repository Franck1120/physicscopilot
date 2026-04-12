import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../main.dart' show kAccent, kAccentDark, kBgPrimary, kBgCard, kBgCardBorder, kTextMuted;
import '../providers/equipment_provider.dart';
import '../providers/websocket_provider.dart';
import '../services/websocket_service.dart';
import '../services/api_service.dart';
import '../providers/session_history_provider.dart';
import '../models/session_record.dart';
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

class _HomeTab extends ConsumerWidget {
  const _HomeTab({
    required this.onGoToCamera,
    required this.onChangeEquipment,
  });

  final VoidCallback onGoToCamera;
  final VoidCallback onChangeEquipment;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final equipment = ref.watch(equipmentProvider);
    final connectionStatus = ref.watch(connectionStatusProvider);
    final serverHealth = ref.watch(serverHealthProvider);

    final bool isServerOnline = serverHealth.when(
      data: (healthy) => healthy,
      loading: () => false,
      error: (_, __) => false,
    );

    return Scaffold(
      backgroundColor: kBgPrimary,
      appBar: AppBar(
        backgroundColor: const Color(0xFF111111),
        elevation: 0,
        title: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            _ApiHealthDot(isOnline: isServerOnline),
            const SizedBox(width: 8),
            const Text(
              'PhysicsCopilot',
              style: TextStyle(
                color: Colors.white,
                fontWeight: FontWeight.bold,
                letterSpacing: 0.5,
              ),
            ),
          ],
        ),
        actions: [
          Padding(
            padding: const EdgeInsets.only(right: 16),
            child: _WsStatusChip(status: connectionStatus),
          ),
        ],
      ),
      body: SafeArea(
        child: ListView(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 20),
          children: [
            _NewSessionCard(onTap: onGoToCamera),
            const SizedBox(height: 24),
            _EquipmentSection(
              equipmentName: equipment?.name,
              onChangeEquipment: onChangeEquipment,
            ),
            const SizedBox(height: 24),
            const _RecentSessionsSection(),
          ],
        ),
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// API Health Dot — shown in AppBar title
// ---------------------------------------------------------------------------

class _ApiHealthDot extends StatefulWidget {
  const _ApiHealthDot({required this.isOnline});

  final bool isOnline;

  @override
  State<_ApiHealthDot> createState() => _ApiHealthDotState();
}

class _ApiHealthDotState extends State<_ApiHealthDot>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller;
  late final Animation<double> _opacityAnimation;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 900),
    )..repeat(reverse: true);
    _opacityAnimation = Tween<double>(begin: 0.4, end: 1.0).animate(
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
    final color = widget.isOnline ? kAccent : Colors.redAccent;

    return AnimatedBuilder(
      animation: _opacityAnimation,
      builder: (context, child) {
        return Opacity(
          opacity: widget.isOnline ? _opacityAnimation.value : 1.0,
          child: Container(
            width: 8,
            height: 8,
            decoration: BoxDecoration(
              color: color,
              shape: BoxShape.circle,
              boxShadow: widget.isOnline
                  ? [
                      BoxShadow(
                        color: color.withValues(alpha: 0.5),
                        blurRadius: 4,
                        spreadRadius: 1,
                      ),
                    ]
                  : null,
            ),
          ),
        );
      },
    );
  }
}

// ---------------------------------------------------------------------------
// New Session Card
// ---------------------------------------------------------------------------

class _NewSessionCard extends StatelessWidget {
  const _NewSessionCard({required this.onTap});

  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: () {
        HapticFeedback.mediumImpact();
        onTap();
      },
      child: Container(
        width: double.infinity,
        padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 36),
        decoration: BoxDecoration(
          gradient: const LinearGradient(
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
            colors: [Color(0xFF064E3B), Color(0xFF047857)],
          ),
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
                      fontSize: 22,
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

class _RecentSessionsSection extends ConsumerWidget {
  const _RecentSessionsSection();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final sessions = ref.watch(sessionHistoryProvider).take(3).toList();

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
            children: sessions
                .map((session) => _SessionCard(session: session))
                .toList(),
          ),
      ],
    );
  }
}

class _SessionCard extends StatelessWidget {
  const _SessionCard({required this.session});

  final SessionRecord session;

  String _formatDate(DateTime date) {
    final now = DateTime.now();
    final today = DateTime(now.year, now.month, now.day);
    final sessionDay = DateTime(date.year, date.month, date.day);
    final hour = date.hour.toString().padLeft(2, '0');
    final minute = date.minute.toString().padLeft(2, '0');

    if (sessionDay == today) {
      return 'oggi $hour:$minute';
    }
    final yesterday = today.subtract(const Duration(days: 1));
    if (sessionDay == yesterday) {
      return 'ieri $hour:$minute';
    }
    return '${date.day}/${date.month} $hour:$minute';
  }

  String _formatDuration(Duration duration) {
    final minutes = duration.inMinutes;
    final seconds = duration.inSeconds % 60;
    if (minutes > 0) {
      return '${minutes}m ${seconds}s';
    }
    return '${seconds}s';
  }

  String _truncate(String text, int maxLen) {
    if (text.length <= maxLen) return text;
    return '${text.substring(0, maxLen)}…';
  }

  @override
  Widget build(BuildContext context) {
    final isResolved = session.status == SessionStatus.resolved;
    final statusColor = isResolved ? kAccent : Colors.orangeAccent;
    final statusLabel = isResolved ? 'Risolto' : 'In corso';

    return Container(
      margin: const EdgeInsets.only(bottom: 10),
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: kBgCard,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: kBgCardBorder, width: 1),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Expanded(
                child: Text(
                  session.equipmentName,
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 13,
                    fontWeight: FontWeight.w600,
                  ),
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              const SizedBox(width: 8),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                decoration: BoxDecoration(
                  color: statusColor.withValues(alpha: 0.12),
                  borderRadius: BorderRadius.circular(20),
                  border: Border.all(
                    color: statusColor.withValues(alpha: 0.4),
                    width: 1,
                  ),
                ),
                child: Text(
                  statusLabel,
                  style: TextStyle(
                    color: statusColor,
                    fontSize: 11,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 6),
          Text(
            _truncate(session.problemDescription, 40),
            style: const TextStyle(
              color: kTextMuted,
              fontSize: 12,
            ),
          ),
          const SizedBox(height: 8),
          Row(
            children: [
              const Icon(Icons.access_time, color: kTextMuted, size: 12),
              const SizedBox(width: 4),
              Text(
                _formatDate(session.date),
                style: const TextStyle(color: kTextMuted, fontSize: 11),
              ),
              const SizedBox(width: 12),
              const Icon(Icons.timer_outlined, color: kTextMuted, size: 12),
              const SizedBox(width: 4),
              Text(
                _formatDuration(session.duration),
                style: const TextStyle(color: kTextMuted, fontSize: 11),
              ),
            ],
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
