import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../main.dart' show kAccent, kAccentDark, kBgPrimary, kBgCard, kBgCardBorder, kTextMuted;
import '../models/session_record.dart';
import '../providers/equipment_provider.dart';
import '../providers/session_history_provider.dart';
import '../providers/websocket_provider.dart';
import '../services/api_service.dart';
import '../services/websocket_service.dart';
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

    return Scaffold(
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
            padding: const EdgeInsets.only(right: 10),
            child: Center(
              child: _ServerHealthDot(health: serverHealth),
            ),
          ),
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
// Server health dot — small indicator next to WS chip
// ---------------------------------------------------------------------------

class _ServerHealthDot extends StatelessWidget {
  const _ServerHealthDot({required this.health});

  final AsyncValue<bool> health;

  @override
  Widget build(BuildContext context) {
    final (color, tooltip) = health.when(
      data: (ok) => ok
          ? (kAccent, 'Server raggiungibile')
          : (Colors.redAccent, 'Server non raggiungibile'),
      loading: () => (Colors.orangeAccent, 'Verifica server…'),
      error: (_, __) => (Colors.redAccent, 'Server non raggiungibile'),
    );

    return Tooltip(
      message: tooltip,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 400),
        width: 9,
        height: 9,
        decoration: BoxDecoration(
          color: color,
          shape: BoxShape.circle,
          boxShadow: [
            BoxShadow(color: color.withAlpha(120), blurRadius: 5),
          ],
        ),
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// New Session Card — prominent CTA
// ---------------------------------------------------------------------------

class _NewSessionCard extends StatelessWidget {
  const _NewSessionCard({required this.onTap});

  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Semantics(
      label: 'Nuova sessione — avvia analisi AI',
      button: true,
      child: GestureDetector(
        onTap: () {
          HapticFeedback.mediumImpact();
          onTap();
        },
        child: Container(
        width: double.infinity,
        padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 32),
        decoration: BoxDecoration(
          gradient: const LinearGradient(
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
            colors: [Color(0xFF065F46), Color(0xFF064E3B)],
          ),
          borderRadius: BorderRadius.circular(16),
          boxShadow: [
            BoxShadow(
              color: kAccent.withValues(alpha: 0.22),
              blurRadius: 20,
              offset: const Offset(0, 8),
            ),
          ],
        ),
        child: Row(
          children: [
            Container(
              width: 56,
              height: 56,
              decoration: BoxDecoration(
                color: kAccent.withAlpha(40),
                shape: BoxShape.circle,
                border: Border.all(color: kAccent.withAlpha(80), width: 1.5),
              ),
              child: const Icon(
                Icons.camera_alt_outlined,
                color: Colors.white,
                size: 28,
              ),
            ),
            const SizedBox(width: 20),
            const Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Nuova sessione',
                    style: TextStyle(
                      color: Colors.white,
                      fontSize: 20,
                      fontWeight: FontWeight.bold,
                      letterSpacing: 0.3,
                    ),
                  ),
                  SizedBox(height: 5),
                  Text(
                    'Punta la camera e avvia l\'analisi AI',
                    style: TextStyle(
                      color: Color(0xFF6EE7B7),
                      fontSize: 13,
                    ),
                  ),
                ],
              ),
            ),
            Container(
              width: 36,
              height: 36,
              decoration: BoxDecoration(
                color: kAccent.withAlpha(30),
                shape: BoxShape.circle,
              ),
              child: const Icon(
                Icons.arrow_forward_rounded,
                color: Color(0xFF6EE7B7),
                size: 18,
              ),
            ),
          ],
        ),
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
              Semantics(
                label: 'Cambia dispositivo attivo',
                button: true,
                child: GestureDetector(
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
              ),
            ],
          ),
        ),
      ],
    );
  }
}

// ---------------------------------------------------------------------------
// Recent Sessions Section — shows last 3 real sessions
// ---------------------------------------------------------------------------

class _RecentSessionsSection extends ConsumerWidget {
  const _RecentSessionsSection();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final allSessions = ref.watch(sessionHistoryProvider);
    final recent = allSessions.take(3).toList();

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
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
            const Spacer(),
            if (allSessions.isNotEmpty)
              Semantics(
                label: 'Vedi tutta la cronologia',
                button: true,
                child: GestureDetector(
                  onTap: () => context.push('/history'),
                  child: const Text(
                    'Vedi tutte',
                    style: TextStyle(
                      color: kAccent,
                      fontSize: 12,
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                ),
              ),
          ],
        ),
        const SizedBox(height: 10),
        if (recent.isEmpty)
          const _NoSessionsCard()
        else
          ...recent.map((s) => _RecentSessionCard(session: s)),
      ],
    );
  }
}

class _RecentSessionCard extends StatelessWidget {
  const _RecentSessionCard({required this.session});

  final SessionRecord session;

  @override
  Widget build(BuildContext context) {
    final isResolved = session.status == SessionStatus.resolved;
    final statusColor = isResolved ? kAccent : Colors.redAccent;

    return Container(
      margin: const EdgeInsets.only(bottom: 10),
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
      decoration: BoxDecoration(
        color: kBgCard,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: kBgCardBorder, width: 1),
      ),
      child: Row(
        children: [
          Container(
            width: 9,
            height: 9,
            decoration: BoxDecoration(
              color: statusColor,
              shape: BoxShape.circle,
              boxShadow: [
                BoxShadow(color: statusColor.withAlpha(80), blurRadius: 4),
              ],
            ),
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
                    fontSize: 13,
                    fontWeight: FontWeight.w500,
                  ),
                ),
                if (session.summary.isNotEmpty) ...[
                  const SizedBox(height: 2),
                  Text(
                    session.summary,
                    style: const TextStyle(color: kTextMuted, fontSize: 11),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ],
              ],
            ),
          ),
          const SizedBox(width: 8),
          Text(
            _formatSessionDate(session.date),
            style: const TextStyle(color: kTextMuted, fontSize: 11),
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

String _formatSessionDate(DateTime date) {
  const months = [
    'gen', 'feb', 'mar', 'apr', 'mag', 'giu',
    'lug', 'ago', 'set', 'ott', 'nov', 'dic',
  ];
  final now = DateTime.now();
  if (date.year == now.year &&
      date.month == now.month &&
      date.day == now.day) {
    return 'oggi';
  }
  final yesterday = now.subtract(const Duration(days: 1));
  if (date.year == yesterday.year &&
      date.month == yesterday.month &&
      date.day == yesterday.day) {
    return 'ieri';
  }
  return '${date.day} ${months[date.month - 1]}';
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
        ConnectionStatus.connected => (kAccent, 'Connesso', Icons.wifi),
        ConnectionStatus.connecting => (
          Colors.orangeAccent,
          'Connessione…',
          Icons.wifi_find,
        ),
        ConnectionStatus.disconnected => (
          Colors.redAccent,
          'Non connesso',
          Icons.wifi_off,
        ),
      },
      loading: () => (Colors.orangeAccent, 'Connessione…', Icons.wifi_find),
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
