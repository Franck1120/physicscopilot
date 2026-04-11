import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../main.dart' show kAccent, kAccentDark, kBgPrimary, kBgCard, kBgCardBorder, kTextMuted;
import '../providers/printer_provider.dart';
import '../providers/websocket_provider.dart';
import '../services/websocket_service.dart';
import '../models/session_record.dart';
import 'history_screen.dart';

// ---------------------------------------------------------------------------
// Mock data — last 3 sessions shown on the home tab
// DateTime has no const constructor, so this list is final (not const).
// ---------------------------------------------------------------------------

final List<SessionRecord> _mockRecentSessions = [
  SessionRecord(
    id: 's1',
    date: DateTime(2026, 4, 10),
    printerName: 'HP LaserJet 1020',
    problemDescription: 'Paper jam at tray 2',
    status: SessionStatus.resolved,
    duration: const Duration(minutes: 8),
  ),
  SessionRecord(
    id: 's2',
    date: DateTime(2026, 4, 8),
    printerName: 'Canon PIXMA G3470',
    problemDescription: 'Ink cartridge not recognised',
    status: SessionStatus.unresolved,
    duration: const Duration(minutes: 14),
  ),
  SessionRecord(
    id: 's3',
    date: DateTime(2026, 4, 5),
    printerName: 'Epson EcoTank L3150',
    problemDescription: 'Stripes on printed pages',
    status: SessionStatus.resolved,
    duration: const Duration(minutes: 5),
  ),
];

// ---------------------------------------------------------------------------
// HomeScreen
// ---------------------------------------------------------------------------

class HomeScreen extends ConsumerStatefulWidget {
  const HomeScreen({
    super.key,
    required this.onChangePrinter,
    required this.onStartCamera,
  });

  final VoidCallback onChangePrinter;
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
            onChangePrinter: widget.onChangePrinter,
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
    required this.onChangePrinter,
  });

  final VoidCallback onGoToCamera;
  final VoidCallback onChangePrinter;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final printer = ref.watch(printerProvider);
    final connectionStatus = ref.watch(connectionStatusProvider);

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
            _PrinterSection(
              printerName: printer?.name,
              onChangePrinter: onChangePrinter,
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
// Printer Section
// ---------------------------------------------------------------------------

class _PrinterSection extends StatelessWidget {
  const _PrinterSection({
    required this.printerName,
    required this.onChangePrinter,
  });

  final String? printerName;
  final VoidCallback onChangePrinter;

  @override
  Widget build(BuildContext context) {
    final bool hasPrinter = printerName != null;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text(
          'STAMPANTE ATTIVA',
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
            color: const kBgCard,
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: const kBgCardBorder, width: 1),
          ),
          child: Row(
            children: [
              const Icon(
                Icons.print_outlined,
                color: kAccent,
                size: 22,
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Text(
                  printerName ?? 'Nessuna stampante selezionata',
                  style: TextStyle(
                    color: hasPrinter
                        ? Colors.white
                        : kTextMuted,
                    fontSize: 15,
                    fontWeight:
                        hasPrinter ? FontWeight.w500 : FontWeight.normal,
                  ),
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              const SizedBox(width: 8),
              GestureDetector(
                onTap: onChangePrinter,
                child: Chip(
                  label: const Text(
                    'Cambia',
                    style: TextStyle(
                      color: kAccent,
                      fontSize: 12,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                  backgroundColor: const kBgPrimary,
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

class _RecentSessionsSection extends StatelessWidget {
  const _RecentSessionsSection();

  @override
  Widget build(BuildContext context) {
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
        SingleChildScrollView(
          scrollDirection: Axis.horizontal,
          child: Row(
            children: _mockRecentSessions
                .map((s) => _SessionCompactCard(record: s))
                .toList(),
          ),
        ),
      ],
    );
  }
}

class _SessionCompactCard extends StatelessWidget {
  const _SessionCompactCard({required this.record});

  final SessionRecord record;

  static const Map<SessionStatus, Color> _statusColors = {
    SessionStatus.resolved: Color(0xFF1E8449),
    SessionStatus.unresolved: Color(0xFFC0392B),
  };

  static const Map<SessionStatus, String> _statusLabels = {
    SessionStatus.resolved: 'Risolto',
    SessionStatus.unresolved: 'Aperto',
  };

  String _formatDate(DateTime d) =>
      '${d.day.toString().padLeft(2, '0')}/'
      '${d.month.toString().padLeft(2, '0')}/'
      '${d.year}';

  @override
  Widget build(BuildContext context) {
    final Color statusColor = _statusColors[record.status]!;
    final String statusLabel = _statusLabels[record.status]!;

    return Container(
      width: 160,
      margin: const EdgeInsets.only(right: 12),
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: const kBgCard,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: const kBgCardBorder, width: 1),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                _formatDate(record.date),
                style: const TextStyle(
                  color: kTextMuted,
                  fontSize: 11,
                ),
              ),
              Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                decoration: BoxDecoration(
                  color: statusColor.withValues(alpha: 0.15),
                  borderRadius: BorderRadius.circular(4),
                ),
                child: Text(
                  statusLabel,
                  style: TextStyle(
                    color: statusColor,
                    fontSize: 10,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Text(
            record.printerName,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 13,
              fontWeight: FontWeight.w600,
            ),
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
          ),
          const SizedBox(height: 4),
          Text(
            record.problemDescription,
            style: const TextStyle(
              color: kTextMuted,
              fontSize: 11,
            ),
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
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
      backgroundColor: const kBgPrimary,
      appBar: AppBar(
        backgroundColor: const kBgCard,
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
          backgroundColor: const kBgPrimary,
          side: const BorderSide(color: kAccent, width: 1.5),
          padding: const EdgeInsets.symmetric(horizontal: 8),
        ),
      ],
    );
  }
}

class _ProfileTileList extends StatelessWidget {
  const _ProfileTileList();

  static const List<_ProfileTileData> _tiles = [
    _ProfileTileData(icon: Icons.settings_outlined, label: 'Impostazioni'),
    _ProfileTileData(icon: Icons.lock_outline, label: 'Privacy'),
    _ProfileTileData(icon: Icons.info_outline, label: 'Informazioni app'),
  ];

  @override
  Widget build(BuildContext context) {
    return Column(
      children: _tiles.map(_buildTile).toList(),
    );
  }

  Widget _buildTile(_ProfileTileData t) {
    return Container(
      margin: const EdgeInsets.only(bottom: 10),
      decoration: BoxDecoration(
        color: const kBgCard,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: const kBgCardBorder, width: 1),
      ),
      child: ListTile(
        leading: Icon(t.icon, color: kTextMuted),
        title: Text(
          t.label,
          style: const TextStyle(color: Colors.white, fontSize: 15),
        ),
        trailing: const Icon(
          Icons.arrow_forward_ios_rounded,
          color: kTextMuted,
          size: 16,
        ),
        enabled: false,
        onTap: null,
      ),
    );
  }
}

class _ProfileTileData {
  const _ProfileTileData({required this.icon, required this.label});
  final IconData icon;
  final String label;
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
