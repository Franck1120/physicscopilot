import 'dart:async';
import 'dart:math' as math;

import 'package:flutter/foundation.dart' show kDebugMode;
import 'package:flutter/scheduler.dart';

import 'package:camera/camera.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../main.dart' show kAccent, kBgPrimary, kBgCard, kBgCardBorder, kTextMuted;
import '../services/camera_service.dart' show FrameQuality;
import '../models/session_record.dart';
import '../providers/camera_provider.dart';
import '../providers/equipment_provider.dart';
import '../providers/prefs_provider.dart';
import '../providers/session_history_provider.dart';
import '../providers/session_provider.dart';
import '../providers/websocket_provider.dart';
import '../services/api_service.dart';
import '../services/websocket_service.dart';
import '../services/notification_service.dart';
import '../utils/strings.dart';

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

class _SessionScreenState extends ConsumerState<SessionScreen>
    with WidgetsBindingObserver {
  StreamSubscription<Uint8List>? _frameSubscription;
  StreamSubscription<Map<String, dynamic>>? _messageSubscription;
  final TextEditingController _textController = TextEditingController();

  final DateTime _sessionStart = DateTime.now();
  Duration _elapsed = Duration.zero;
  Timer? _ticker;
  String? _firstUserMessage;
  bool _showTutorial = false;

  static const _kTutorialKey = 'session_tutorial_shown';

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _startListening();
    _ticker = Timer.periodic(const Duration(seconds: 1), (_) {
      if (mounted) {
        setState(() => _elapsed = DateTime.now().difference(_sessionStart));
      }
    });
    // Check after first frame so we don't call setState during build.
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      final shown =
          ref.read(sharedPrefsProvider).getBool(_kTutorialKey) ?? false;
      if (!shown) setState(() => _showTutorial = true);
    });
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    NotificationService.cancelSessionNotification();
    _ticker?.cancel();
    _frameSubscription?.cancel();
    _messageSubscription?.cancel();
    _textController.dispose();
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    final session = ref.read(sessionProvider);
    final isActive = session.responseText != null || session.isProcessing;
    switch (state) {
      case AppLifecycleState.paused:
      case AppLifecycleState.inactive:
        if (isActive) NotificationService.showSessionRunning();
      case AppLifecycleState.resumed:
        NotificationService.cancelSessionNotification();
      default:
        break;
    }
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
    // Persist locally (always).
    ref.read(sessionHistoryProvider.notifier).add(record);
    // Best-effort sync to server (fire-and-forget, no UI blocking).
    _syncSessionToServer(
      deviceBrand: equipment?.manufacturer ?? '',
      deviceModel: equipment?.name ?? '',
    );
  }

  /// Posts the finished session to the server for server-side tracking.
  /// Failures are silently swallowed — local storage is the source of truth.
  Future<void> _syncSessionToServer({
    required String deviceBrand,
    required String deviceModel,
  }) async {
    try {
      final api = ref.read(apiServiceProvider);
      await api.createSession(
        deviceBrand: deviceBrand,
        deviceModel: deviceModel,
      );
    } catch (_) {}
  }

  void _dismissTutorial() {
    setState(() => _showTutorial = false);
    // Fire-and-forget; we don't need to await the write.
    ref.read(sharedPrefsProvider).setBool(_kTutorialKey, true);
  }

  void _resetSession() {
    HapticFeedback.mediumImpact();
    ref.read(sessionProvider.notifier).reset();
    setState(() => _firstUserMessage = null);
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
          tooltip: AppStrings.sessionEndSession,
          onPressed: () {
            _saveSessionIfNeeded();
            Navigator.of(context).pop();
          },
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh_rounded,
                color: Colors.white54, size: 20),
            tooltip: AppStrings.sessionNewAnalysis,
            onPressed: _resetSession,
          ),
          Padding(
            padding: const EdgeInsets.only(right: 16),
            child: Center(
              child: Semantics(
                label: 'Durata sessione: ${_formatElapsed(_elapsed)}',
                child: ExcludeSemantics(
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
            ),
          ),
        ],
      ),
      body: Stack(
        children: [
          Column(
            children: [
              wsStatus.when(
                data: (s) => s != ConnectionStatus.connected
                    ? _ConnectionBanner(status: s)
                    : const SizedBox.shrink(),
                loading: () => const _ConnectionBanner(
                    status: ConnectionStatus.connecting),
                error: (_, __) => const _ConnectionBanner(
                    status: ConnectionStatus.disconnected),
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
          if (_showTutorial)
            _TutorialOverlay(onDismiss: _dismissTutorial),
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

    return Semantics(
      liveRegion: true,
      label: message,
      child: Container(
        width: double.infinity,
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 7),
        color: color.withAlpha(30),
        child: ExcludeSemantics(
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
        ),
      ),
    );
  }
}

// ── Camera section — flash, zoom, tap-focus, quality badge ──────────────────

class _CameraSection extends ConsumerStatefulWidget {
  const _CameraSection({required this.onCapture});
  final VoidCallback onCapture;

  @override
  ConsumerState<_CameraSection> createState() => _CameraSectionState();
}

class _CameraSectionState extends ConsumerState<_CameraSection> {
  bool _torchOn = false;
  double _currentZoom = 1.0;
  double _minZoom = 1.0;
  double _maxZoom = 1.0;
  double _baseZoom = 1.0;
  Offset? _focusPoint; // screen-space position for the focus ring
  FrameQuality _quality = FrameQuality.ok;
  StreamSubscription<FrameQuality>? _qualitySub;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) => _setup());
  }

  Future<void> _setup() async {
    final service = ref.read(cameraServiceProvider);
    _qualitySub = service.quality.listen((q) {
      if (mounted) setState(() => _quality = q);
    });
    final controller = service.controller;
    if (controller == null || !controller.value.isInitialized) return;
    try {
      final min = await controller.getMinZoomLevel();
      final max = await controller.getMaxZoomLevel();
      if (mounted) setState(() { _minZoom = min; _maxZoom = max; });
    } catch (_) {}
  }

  @override
  void dispose() {
    _qualitySub?.cancel();
    super.dispose();
  }

  Future<void> _toggleTorch() async {
    final controller = ref.read(cameraServiceProvider).controller;
    if (controller == null) return;
    try {
      await controller.setFlashMode(
        _torchOn ? FlashMode.off : FlashMode.torch,
      );
      if (mounted) setState(() => _torchOn = !_torchOn);
      HapticFeedback.selectionClick();
    } catch (_) {}
  }

  void _onScaleStart(ScaleStartDetails _) => _baseZoom = _currentZoom;

  Future<void> _onScaleUpdate(ScaleUpdateDetails details) async {
    if (details.pointerCount < 2) return; // only pinch, not single-finger pan
    final controller = ref.read(cameraServiceProvider).controller;
    if (controller == null) return;
    final target = (_baseZoom * details.scale).clamp(_minZoom, _maxZoom);
    try {
      await controller.setZoomLevel(target);
      if (mounted) setState(() => _currentZoom = target);
    } catch (_) {}
  }

  Future<void> _onTapUp(TapUpDetails details, BoxConstraints box) async {
    final controller = ref.read(cameraServiceProvider).controller;
    if (controller == null) return;
    final norm = Offset(
      details.localPosition.dx / box.maxWidth,
      details.localPosition.dy / box.maxHeight,
    );
    try {
      await controller.setFocusPoint(norm);
      await controller.setExposurePoint(norm);
    } catch (_) {}
    if (!mounted) return;
    setState(() => _focusPoint = details.localPosition);
    Future.delayed(const Duration(milliseconds: 1500), () {
      if (mounted) setState(() => _focusPoint = null);
    });
  }

  @override
  Widget build(BuildContext context) {
    final cameraInit = ref.watch(cameraInitProvider);
    final cameraService = ref.watch(cameraServiceProvider);

    return Stack(
      fit: StackFit.expand,
      children: [
        // ── Camera preview with gestures ──────────────────────────────────
        cameraInit.when(
          data: (_) {
            final controller = cameraService.controller;
            if (controller == null || !controller.value.isInitialized) {
              return const _CameraPlaceholder();
            }
            return LayoutBuilder(
              builder: (context, constraints) => GestureDetector(
                onScaleStart: _onScaleStart,
                onScaleUpdate: _onScaleUpdate,
                onTapUp: (d) => _onTapUp(d, constraints),
                child: ClipRect(
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
                ),
              ),
            );
          },
          loading: () => const _CameraPlaceholder(),
          error: (_, __) => const _CameraError(),
        ),

        // ── Focus ring ────────────────────────────────────────────────────
        if (_focusPoint != null)
          Positioned(
            left: _focusPoint!.dx - 28,
            top: _focusPoint!.dy - 28,
            child: const _FocusRing(),
          ),

        // ── Flash toggle ──────────────────────────────────────────────────
        Positioned(
          top: 12,
          left: 12,
          child: _TorchButton(isOn: _torchOn, onTap: _toggleTorch),
        ),

        // ── FPS counter (debug builds only) ───────────────────────────────
        if (kDebugMode)
          const Positioned(
            top: 12,
            right: 12,
            child: _FpsOverlay(),
          ),

        // ── Frame quality badge ───────────────────────────────────────────
        if (_quality != FrameQuality.ok)
          Positioned(
            top: 12,
            left: 0,
            right: 0,
            child: Center(child: _QualityBadge(quality: _quality)),
          ),

        // ── Zoom level indicator ──────────────────────────────────────────
        if (_currentZoom > 1.05)
          Positioned(
            bottom: 70,
            left: 0,
            right: 0,
            child: Center(
              child: Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                decoration: BoxDecoration(
                  color: Colors.black54,
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Text(
                  '${_currentZoom.toStringAsFixed(1)}×',
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 13,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ),
            ),
          ),

        // ── Capture FAB ───────────────────────────────────────────────────
        Positioned(
          bottom: 16,
          right: 16,
          child: FloatingActionButton(
            heroTag: 'session_capture',
            backgroundColor: kAccent,
            foregroundColor: Colors.white,
            tooltip: 'Cattura frame e invia all\'AI',
            onPressed: widget.onCapture,
            child: const Icon(Icons.camera_alt),
          ),
        ),
      ],
    );
  }
}

// ── Focus ring ────────────────────────────────────────────────────────────────

class _FocusRing extends StatefulWidget {
  const _FocusRing();
  @override
  State<_FocusRing> createState() => _FocusRingState();
}

class _FocusRingState extends State<_FocusRing>
    with SingleTickerProviderStateMixin {
  late final AnimationController _ctrl;
  late final Animation<double> _scale;
  late final Animation<double> _opacity;

  @override
  void initState() {
    super.initState();
    _ctrl = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 600),
    )..forward();
    _scale = Tween<double>(begin: 1.4, end: 1.0).animate(
      CurvedAnimation(parent: _ctrl, curve: Curves.easeOut),
    );
    _opacity = Tween<double>(begin: 1.0, end: 0.3).animate(
      CurvedAnimation(
        parent: _ctrl,
        curve: const Interval(0.5, 1.0, curve: Curves.easeIn),
      ),
    );
  }

  @override
  void dispose() {
    _ctrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: _ctrl,
      builder: (_, __) => Opacity(
        opacity: _opacity.value,
        child: Transform.scale(
          scale: _scale.value,
          child: Container(
            width: 56,
            height: 56,
            decoration: BoxDecoration(
              shape: BoxShape.rectangle,
              borderRadius: BorderRadius.circular(6),
              border: Border.all(color: kAccent, width: 2),
            ),
          ),
        ),
      ),
    );
  }
}

// ── Torch button ──────────────────────────────────────────────────────────────

class _TorchButton extends StatelessWidget {
  const _TorchButton({required this.isOn, required this.onTap});
  final bool isOn;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        width: 38,
        height: 38,
        decoration: BoxDecoration(
          color: isOn ? kAccent : Colors.black54,
          shape: BoxShape.circle,
        ),
        child: Icon(
          isOn ? Icons.flash_on : Icons.flash_off,
          color: Colors.white,
          size: 18,
        ),
      ),
    );
  }
}

// ── Frame quality badge ───────────────────────────────────────────────────────

class _QualityBadge extends StatelessWidget {
  const _QualityBadge({required this.quality});
  final FrameQuality quality;

  @override
  Widget build(BuildContext context) {
    final (icon, label, color) = switch (quality) {
      FrameQuality.tooDark => (Icons.brightness_2_outlined, 'Troppo scuro', Colors.orangeAccent),
      FrameQuality.tooBright => (Icons.brightness_7, 'Troppo luminoso', Colors.amberAccent),
      FrameQuality.ok => (Icons.check_circle_outline, '', Colors.greenAccent),
    };
    if (label.isEmpty) return const SizedBox.shrink();
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
      decoration: BoxDecoration(
        color: Colors.black.withAlpha(180),
        borderRadius: BorderRadius.circular(20),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, color: color, size: 14),
          const SizedBox(width: 6),
          Text(label,
              style: TextStyle(
                  color: color, fontSize: 12, fontWeight: FontWeight.w600)),
        ],
      ),
    );
  }
}

// ── FPS overlay (debug mode only) ────────────────────────────────────────────

/// Displays the current UI frame rate in the top-right corner of the camera
/// preview.  Only compiled into debug builds (guarded by [kDebugMode]).
class _FpsOverlay extends StatefulWidget {
  const _FpsOverlay();
  @override
  State<_FpsOverlay> createState() => _FpsOverlayState();
}

class _FpsOverlayState extends State<_FpsOverlay> {
  double _fps = 0;
  final _timestamps = <int>[];

  @override
  void initState() {
    super.initState();
    SchedulerBinding.instance.addTimingsCallback(_onTimings);
  }

  @override
  void dispose() {
    SchedulerBinding.instance.removeTimingsCallback(_onTimings);
    super.dispose();
  }

  void _onTimings(List<FrameTiming> timings) {
    if (!mounted) return;
    final now = DateTime.now().microsecondsSinceEpoch;
    // Add one entry per frame timing batch.
    for (final _ in timings) {
      _timestamps.add(now);
    }
    // Keep only timestamps within the last second.
    _timestamps.removeWhere((t) => now - t > 1000000);
    setState(() => _fps = _timestamps.length.toDouble());
  }

  @override
  Widget build(BuildContext context) {
    final color = _fps >= 55
        ? Colors.greenAccent
        : _fps >= 30
            ? Colors.orangeAccent
            : Colors.redAccent;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: Colors.black.withAlpha(160),
        borderRadius: BorderRadius.circular(6),
      ),
      child: Text(
        '${_fps.toStringAsFixed(0)} fps',
        style: TextStyle(
          color: color,
          fontSize: 11,
          fontWeight: FontWeight.w700,
          fontFeatures: const [FontFeature.tabularFigures()],
        ),
      ),
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
      // _TypewriterResponse carries the ValueKey so AnimatedSwitcher recreates
      // it (and restarts the animation) whenever the text changes.
      child = _TypewriterResponse(
        key: ValueKey(session.responseText),
        text: session.responseText!,
      );
    } else {
      child = const Center(
        key: ValueKey('idle'),
        child: Text(
          AppStrings.sessionIdle,
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

// ── Typewriter response — animates text char-by-char + copy button ────────────

class _TypewriterResponse extends StatefulWidget {
  const _TypewriterResponse({super.key, required this.text});

  final String text;

  @override
  State<_TypewriterResponse> createState() => _TypewriterResponseState();
}

class _TypewriterResponseState extends State<_TypewriterResponse> {
  int _length = 0;
  Timer? _timer;

  // ~100 chars/sec feels snappy without losing readability.
  static const _charInterval = Duration(milliseconds: 10);

  @override
  void initState() {
    super.initState();
    _timer = Timer.periodic(_charInterval, (_) {
      if (_length < widget.text.length) {
        if (mounted) setState(() => _length++);
      } else {
        _timer?.cancel();
        _timer = null;
      }
    });
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final displayed = widget.text.substring(0, _length);
    final done = _length >= widget.text.length;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Expanded(
          child: SingleChildScrollView(
            padding: const EdgeInsets.fromLTRB(16, 16, 16, 4),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Icon(Icons.auto_fix_high, color: kAccent, size: 18),
                const SizedBox(width: 10),
                Expanded(
                  child: Text(
                    displayed,
                    style: const TextStyle(
                        color: Colors.white, fontSize: 14, height: 1.5),
                  ),
                ),
              ],
            ),
          ),
        ),
        // Copy button appears once typing is complete.
        AnimatedOpacity(
          opacity: done ? 1.0 : 0.0,
          duration: const Duration(milliseconds: 300),
          child: Align(
            alignment: Alignment.centerRight,
            child: Padding(
              padding: const EdgeInsets.only(right: 8, bottom: 4),
              child: Semantics(
                label: 'Copia risposta AI',
                button: true,
                child: IconButton(
                  icon: const Icon(Icons.copy_outlined,
                      size: 16, color: kTextMuted),
                  tooltip: 'Copia risposta',
                  onPressed: done
                      ? () {
                          HapticFeedback.selectionClick();
                          Clipboard.setData(
                              ClipboardData(text: widget.text));
                          ScaffoldMessenger.of(context).showSnackBar(
                            const SnackBar(
                              content: Text(AppStrings.sessionResponseCopied),
                            ),
                          );
                        }
                      : null,
                ),
              ),
            ),
          ),
        ),
      ],
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
    return Semantics(
      label: "L'AI sta analizzando",
      child: Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ExcludeSemantics(
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: List.generate(3, (i) => _Dot(ctrl: _ctrl, index: i)),
              ),
            ),
            const SizedBox(height: 12),
            const ExcludeSemantics(
              child: Text(
                "L'AI sta analizzando…",
                style: TextStyle(color: kTextMuted, fontSize: 13),
              ),
            ),
          ],
        ),
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
            child: Semantics(
              label: 'Descrivi il problema',
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
          ),
          const SizedBox(width: 8),
          IconButton(
            onPressed: onSend,
            icon: const Icon(Icons.send_rounded),
            color: kAccent,
            tooltip: 'Invia messaggio',
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

// ── Tutorial overlay ──────────────────────────────────────────────────────────

/// Shown once on the first use of the session screen.
/// Tapping anywhere dismisses it and marks it as seen in SharedPreferences.
class _TutorialOverlay extends StatefulWidget {
  const _TutorialOverlay({required this.onDismiss});

  final VoidCallback onDismiss;

  @override
  State<_TutorialOverlay> createState() => _TutorialOverlayState();
}

class _TutorialOverlayState extends State<_TutorialOverlay>
    with SingleTickerProviderStateMixin {
  late final AnimationController _pulse;
  late final Animation<double> _pulseScale;
  late final Animation<double> _arrowBounce;

  @override
  void initState() {
    super.initState();
    _pulse = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 900),
    )..repeat(reverse: true);
    _pulseScale = Tween<double>(begin: 0.92, end: 1.0).animate(
      CurvedAnimation(parent: _pulse, curve: Curves.easeInOut),
    );
    _arrowBounce = Tween<double>(begin: 0, end: 8).animate(
      CurvedAnimation(parent: _pulse, curve: Curves.easeInOut),
    );
  }

  @override
  void dispose() {
    _pulse.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final size = MediaQuery.of(context).size;
    // Camera section occupies the top ~60 % of the body.
    // The capture FAB is near the bottom-right of that section.
    final fabAreaTop = size.height * 0.55;

    return GestureDetector(
      onTap: widget.onDismiss,
      child: Container(
        color: Colors.black.withAlpha(160),
        width: double.infinity,
        height: double.infinity,
        child: Stack(
          children: [
            // Hint badge + bouncing arrow anchored near the capture FAB
            Positioned(
              top: fabAreaTop - 110,
              right: 20,
              child: AnimatedBuilder(
                animation: _pulse,
                builder: (_, __) {
                  return Transform.translate(
                    offset: Offset(0, -_arrowBounce.value),
                    child: ScaleTransition(
                      scale: _pulseScale,
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        crossAxisAlignment: CrossAxisAlignment.end,
                        children: [
                          Container(
                            padding: const EdgeInsets.symmetric(
                                horizontal: 14, vertical: 10),
                            decoration: BoxDecoration(
                              color: kAccent,
                              borderRadius: BorderRadius.circular(12),
                              boxShadow: [
                                BoxShadow(
                                  color: kAccent.withAlpha(100),
                                  blurRadius: 16,
                                  spreadRadius: 2,
                                ),
                              ],
                            ),
                            child: const Text(
                              AppStrings.tutorialHint,
                              style: TextStyle(
                                color: Colors.white,
                                fontSize: 14,
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                          ),
                          const SizedBox(height: 6),
                          const Icon(Icons.arrow_downward_rounded,
                              color: kAccent, size: 28),
                        ],
                      ),
                    ),
                  );
                },
              ),
            ),
            // Dismiss hint
            Positioned(
              bottom: size.height * 0.42,
              left: 0,
              right: 0,
              child: Center(
                child: Text(
                  AppStrings.tutorialDismiss,
                  style: TextStyle(
                    color: Colors.white.withAlpha(140),
                    fontSize: 12,
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
