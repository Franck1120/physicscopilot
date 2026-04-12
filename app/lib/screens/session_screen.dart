import 'dart:async';
import 'dart:math' as math;

import 'package:camera/camera.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../main.dart' show kAccent, kBgPrimary, kBgCard, kBgCardBorder, kTextMuted;
import '../models/session_record.dart';
import '../providers/camera_provider.dart';
import '../providers/equipment_provider.dart';
import '../providers/session_history_provider.dart';
import '../providers/session_provider.dart';
import '../providers/websocket_provider.dart';
import '../services/websocket_service.dart';

/// Active repair session screen.
///
/// Layout:
///   - Connection banner (shown only when WS is not connected)
///   - Top 60 %: live camera preview with manual capture button
///   - Bottom 40 %: AI guidance panel (response text + text input)
class SessionScreen extends ConsumerStatefulWidget {
  const SessionScreen({super.key});

  @override
  ConsumerState<SessionScreen> createState() => _SessionScreenState();
}

class _SessionScreenState extends ConsumerState<SessionScreen> {
  StreamSubscription<Uint8List>? _frameSubscription;
  StreamSubscription<Map<String, dynamic>>? _messageSubscription;
  final TextEditingController _textController = TextEditingController();

  final DateTime _sessionStart = DateTime.now();
  Duration _elapsed = Duration.zero;
  Timer? _ticker;
  String? _firstUserMessage;

  @override
  void initState() {
    super.initState();
    _startListening();
    _ticker = Timer.periodic(const Duration(seconds: 1), (_) {
      if (mounted) {
        setState(() => _elapsed = DateTime.now().difference(_sessionStart));
      }
    });
  }

  @override
  void dispose() {
    _ticker?.cancel();
    _frameSubscription?.cancel();
    _messageSubscription?.cancel();
    _textController.dispose();
    super.dispose();
  }

  void _startListening() {
    final wsService = ref.read(webSocketServiceProvider);
    final cameraService = ref.read(cameraServiceProvider);
    _frameSubscription = cameraService.frames.listen(wsService.sendFrame);
    _messageSubscription = wsService.messages.listen(_onServerMessage);
  }

  void _onServerMessage(Map<String, dynamic> msg) {
    final type = msg['type'] as String?;
    if (type == 'response') {
      ref.read(sessionProvider.notifier).updateFromResponse(msg);
    } else if (type == 'error') {
      ref.read(sessionProvider.notifier).setError(
        (msg['error'] as String?) ?? 'Errore sconosciuto',
      );
    }
  }

  bool get _isConnected {
    final status = ref.read(connectionStatusProvider).value;
    return status == ConnectionStatus.connected;
  }

  Future<void> _captureAndSend() async {
    if (!_isConnected) {
      ref.read(sessionProvider.notifier).setError(
        'Server non raggiungibile — attendi la riconnessione.',
      );
      return;
    }
    HapticFeedback.mediumImpact();
    final cameraService = ref.read(cameraServiceProvider);
    final wsService = ref.read(webSocketServiceProvider);
    ref.read(sessionProvider.notifier).setProcessing();
    try {
      final frame = await cameraService.captureFrame();
      if (frame != null) wsService.sendFrame(frame);
    } catch (_) {
      ref
          .read(sessionProvider.notifier)
          .setError('Impossibile acquisire il frame');
    }
  }

  void _sendText() {
    if (!_isConnected) {
      ref.read(sessionProvider.notifier).setError(
        'Server non raggiungibile — attendi la riconnessione.',
      );
      return;
    }
    final text = _textController.text.trim();
    if (text.isEmpty) return;
    _firstUserMessage ??= text;
    final wsService = ref.read(webSocketServiceProvider);
    ref.read(sessionProvider.notifier).setProcessing();
    wsService.sendText(text);
    _textController.clear();
    FocusScope.of(context).unfocus();
  }

  void _saveSessionIfNeeded() {
    final sessionState = ref.read(sessionProvider);
    final summary = sessionState.responseText;
    if (summary == null || summary.isEmpty) return;

    final equipment = ref.read(equipmentProvider);
    final duration = DateTime.now().difference(_sessionStart);
    final record = SessionRecord(
      id: _sessionStart.millisecondsSinceEpoch.toString(),
      date: _sessionStart,
      equipmentName: equipment?.name ?? '',
      problemDescription: _firstUserMessage ?? '',
      summary: summary,
      status: SessionStatus.resolved,
      duration: duration,
    );
    ref.read(sessionHistoryProvider.notifier).add(record);
  }

  String _formatElapsed(Duration d) {
    final m = d.inMinutes.remainder(60).toString().padLeft(2, '0');
    final s = d.inSeconds.remainder(60).toString().padLeft(2, '0');
    return '${d.inHours > 0 ? '${d.inHours}:' : ''}$m:$s';
  }

  @override
  Widget build(BuildContext context) {
    final wsStatus = ref.watch(connectionStatusProvider);

    return Scaffold(
      backgroundColor: kBgPrimary,
      appBar: AppBar(
        backgroundColor: const Color(0xFF111111),
        elevation: 0,
        title: const Text(
          'Sessione attiva',
          style: TextStyle(
            color: Colors.white,
            fontWeight: FontWeight.bold,
            letterSpacing: 0.4,
          ),
        ),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios_new,
              color: Colors.white, size: 20),
          onPressed: () {
            _saveSessionIfNeeded();
            Navigator.of(context).pop();
          },
        ),
        actions: [
          Padding(
            padding: const EdgeInsets.only(right: 16),
            child: Center(
              child: Text(
                _formatElapsed(_elapsed),
                style: const TextStyle(
                  color: kAccent,
                  fontSize: 13,
                  fontWeight: FontWeight.w600,
                  letterSpacing: 1,
                  fontFeatures: [FontFeature.tabularFigures()],
                ),
              ),
            ),
          ),
        ],
      ),
      body: Column(
        children: [
          wsStatus.when(
            data: (s) => s != ConnectionStatus.connected
                ? _ConnectionBanner(status: s)
                : const SizedBox.shrink(),
            loading: () =>
                const _ConnectionBanner(status: ConnectionStatus.connecting),
            error: (_, __) =>
                const _ConnectionBanner(status: ConnectionStatus.disconnected),
          ),
          Expanded(
            flex: 6,
            child: _CameraSection(onCapture: _captureAndSend),
          ),
          Expanded(
            flex: 4,
            child: _GuidancePanel(
              textController: _textController,
              onSendText: _sendText,
            ),
          ),
        ],
      ),
    );
  }
}

// ── Connection banner ────────────────────────────────────────────────────────

class _ConnectionBanner extends StatelessWidget {
  const _ConnectionBanner({required this.status});
  final ConnectionStatus status;

  @override
  Widget build(BuildContext context) {
    final isConnecting = status == ConnectionStatus.connecting;
    final color = isConnecting ? Colors.orangeAccent : Colors.redAccent;
    final icon = isConnecting ? Icons.wifi_find : Icons.wifi_off;
    final message = isConnecting
        ? 'Connessione al server in corso…'
        : 'Server non raggiungibile — i dati non vengono inviati';

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 7),
      color: color.withAlpha(30),
      child: Row(
        children: [
          Icon(icon, color: color, size: 14),
          const SizedBox(width: 8),
          Expanded(
            child: Text(message,
                style: TextStyle(
                    color: color, fontSize: 12, fontWeight: FontWeight.w500)),
          ),
        ],
      ),
    );
  }
}

// ── Camera section ──────────────────────────────────────────────────────────

class _CameraSection extends ConsumerWidget {
  const _CameraSection({required this.onCapture});
  final VoidCallback onCapture;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final cameraInit = ref.watch(cameraInitProvider);
    final cameraService = ref.watch(cameraServiceProvider);

    return Stack(
      fit: StackFit.expand,
      children: [
        cameraInit.when(
          data: (_) {
            final controller = cameraService.controller;
            if (controller == null || !controller.value.isInitialized) {
              return const _CameraPlaceholder();
            }
            return ClipRect(
              child: OverflowBox(
                maxWidth: double.infinity,
                maxHeight: double.infinity,
                child: FittedBox(
                  fit: BoxFit.cover,
                  child: SizedBox(
                    width: controller.value.previewSize?.height ?? 1,
                    height: controller.value.previewSize?.width ?? 1,
                    child: CameraPreview(controller),
                  ),
                ),
              ),
            );
          },
          loading: () => const _CameraPlaceholder(),
          error: (_, __) => const _CameraError(),
        ),
        Positioned(
          bottom: 16,
          right: 16,
          child: FloatingActionButton(
            heroTag: 'session_capture',
            backgroundColor: kAccent,
            foregroundColor: Colors.white,
            onPressed: onCapture,
            child: const Icon(Icons.camera_alt),
          ),
        ),
      ],
    );
  }
}

class _CameraPlaceholder extends StatelessWidget {
  const _CameraPlaceholder();
  @override
  Widget build(BuildContext context) => Container(
        color: const Color(0xFF0D0D0D),
        child: const Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              CircularProgressIndicator(color: kAccent),
              SizedBox(height: 16),
              Text('Inizializzazione camera…',
                  style: TextStyle(color: kTextMuted, fontSize: 13)),
            ],
          ),
        ),
      );
}

class _CameraError extends StatelessWidget {
  const _CameraError();
  @override
  Widget build(BuildContext context) => Container(
        color: const Color(0xFF0D0D0D),
        child: const Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(Icons.camera_alt_outlined, color: kTextMuted, size: 48),
              SizedBox(height: 12),
              Text('Camera non disponibile',
                  style: TextStyle(color: kTextMuted, fontSize: 13)),
            ],
          ),
        ),
      );
}

// ── Guidance panel ──────────────────────────────────────────────────────────

class _GuidancePanel extends ConsumerWidget {
  const _GuidancePanel({
    required this.textController,
    required this.onSendText,
  });
  final TextEditingController textController;
  final VoidCallback onSendText;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final session = ref.watch(sessionProvider);
    return Container(
      decoration: const BoxDecoration(
        color: kBgCard,
        border: Border(top: BorderSide(color: kBgCardBorder, width: 1)),
      ),
      child: Column(
        children: [
          Expanded(child: _ResponseArea(session: session)),
          _TextInputRow(controller: textController, onSend: onSendText),
        ],
      ),
    );
  }
}

// ── Response area with animations ───────────────────────────────────────────

class _ResponseArea extends StatelessWidget {
  const _ResponseArea({required this.session});
  final SessionState session;

  @override
  Widget build(BuildContext context) {
    Widget child;

    if (session.isProcessing) {
      child = const _ThinkingIndicator();
    } else if (session.errorText != null) {
      child = Padding(
        key: const ValueKey('error'),
        padding: const EdgeInsets.all(16),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Icon(Icons.warning_amber_rounded,
                color: Colors.orangeAccent, size: 18),
            const SizedBox(width: 8),
            Expanded(
              child: Text(session.errorText!,
                  style: const TextStyle(
                      color: Colors.orangeAccent, fontSize: 13)),
            ),
          ],
        ),
      );
    } else if (session.responseText != null) {
      child = SingleChildScrollView(
        key: ValueKey(session.responseText),
        padding: const EdgeInsets.all(16),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Icon(Icons.auto_fix_high, color: kAccent, size: 18),
            const SizedBox(width: 10),
            Expanded(
              child: Text(
                session.responseText!,
                style: const TextStyle(
                    color: Colors.white, fontSize: 14, height: 1.5),
              ),
            ),
          ],
        ),
      );
    } else {
      child = const Center(
        key: ValueKey('idle'),
        child: Text(
          'Punta la camera sull\'oggetto\nper avviare l\'analisi AI.',
          textAlign: TextAlign.center,
          style: TextStyle(color: kTextMuted, fontSize: 13, height: 1.5),
        ),
      );
    }

    return AnimatedSwitcher(
      duration: const Duration(milliseconds: 350),
      switchInCurve: Curves.easeOut,
      switchOutCurve: Curves.easeIn,
      transitionBuilder: (child, animation) => FadeTransition(
        opacity: animation,
        child: child,
      ),
      child: child,
    );
  }
}

// ── Thinking indicator — three pulsing dots ──────────────────────────────────

class _ThinkingIndicator extends StatefulWidget {
  const _ThinkingIndicator();
  @override
  State<_ThinkingIndicator> createState() => _ThinkingIndicatorState();
}

class _ThinkingIndicatorState extends State<_ThinkingIndicator>
    with SingleTickerProviderStateMixin {
  late final AnimationController _ctrl;

  @override
  void initState() {
    super.initState();
    _ctrl = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1200),
    )..repeat();
  }

  @override
  void dispose() {
    _ctrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Row(
            mainAxisSize: MainAxisSize.min,
            children: List.generate(3, (i) => _Dot(ctrl: _ctrl, index: i)),
          ),
          const SizedBox(height: 12),
          const Text(
            'L\'AI sta analizzando…',
            style: TextStyle(color: kTextMuted, fontSize: 13),
          ),
        ],
      ),
    );
  }
}

class _Dot extends StatelessWidget {
  const _Dot({required this.ctrl, required this.index});
  final AnimationController ctrl;
  final int index;

  @override
  Widget build(BuildContext context) {
    // Each dot is offset by 0.2 of the animation cycle.
    final offsetAnimation = Tween<double>(begin: 0, end: 1).animate(
      CurvedAnimation(
        parent: ctrl,
        curve: Interval(
          index * 0.2,
          math.min(index * 0.2 + 0.6, 1.0),
          curve: Curves.easeInOut,
        ),
      ),
    );
    return AnimatedBuilder(
      animation: offsetAnimation,
      builder: (_, __) {
        final t = offsetAnimation.value;
        final dy = -6.0 * math.sin(t * math.pi);
        return Transform.translate(
          offset: Offset(0, dy),
          child: Container(
            margin: const EdgeInsets.symmetric(horizontal: 4),
            width: 8,
            height: 8,
            decoration: BoxDecoration(
              color: kAccent.withValues(alpha: 0.4 + 0.6 * (1 - (dy / -6).abs())),
              shape: BoxShape.circle,
            ),
          ),
        );
      },
    );
  }
}

// ── Text input row ───────────────────────────────────────────────────────────

class _TextInputRow extends StatelessWidget {
  const _TextInputRow({required this.controller, required this.onSend});
  final TextEditingController controller;
  final VoidCallback onSend;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.fromLTRB(12, 8, 8, 12),
      decoration: const BoxDecoration(
        border: Border(top: BorderSide(color: kBgCardBorder, width: 1)),
      ),
      child: Row(
        children: [
          Expanded(
            child: TextField(
              controller: controller,
              style: const TextStyle(color: Colors.white, fontSize: 14),
              decoration: InputDecoration(
                hintText: 'Descrivi il problema…',
                hintStyle: const TextStyle(color: kTextMuted, fontSize: 14),
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(24),
                  borderSide: const BorderSide(color: kBgCardBorder),
                ),
                enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(24),
                  borderSide: const BorderSide(color: kBgCardBorder),
                ),
                focusedBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(24),
                  borderSide: const BorderSide(color: kAccent),
                ),
                contentPadding:
                    const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
                filled: true,
                fillColor: const Color(0xFF111111),
              ),
              onSubmitted: (_) => onSend(),
              textInputAction: TextInputAction.send,
            ),
          ),
          const SizedBox(width: 8),
          IconButton(
            onPressed: onSend,
            icon: const Icon(Icons.send_rounded),
            color: kAccent,
            style: IconButton.styleFrom(
              backgroundColor: kAccent.withAlpha(20),
              shape: const CircleBorder(),
            ),
          ),
        ],
      ),
    );
  }
}
