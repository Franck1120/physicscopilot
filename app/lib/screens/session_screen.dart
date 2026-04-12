import 'dart:async';
import 'dart:ui' as ui;

import 'package:flutter/foundation.dart' show kDebugMode;
import 'package:flutter/scheduler.dart';

import 'package:camera/camera.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:flutter/rendering.dart' show RenderRepaintBoundary;
import 'package:share_plus/share_plus.dart';

import '../main.dart' show kAccent, kBgPrimary, kBgCard, kBgCardBorder, kTextMuted;
import '../services/camera_service.dart' show FrameQuality;
import '../models/session_record.dart';
import '../providers/camera_provider.dart';
import '../providers/equipment_provider.dart';
import '../providers/prefs_provider.dart';
import '../providers/session_history_provider.dart';
import '../providers/session_provider.dart';
import '../providers/settings_provider.dart';
import '../providers/voice_provider.dart';
import '../providers/websocket_provider.dart';
import '../services/api_service.dart';
import '../services/websocket_service.dart';
import '../services/notification_service.dart';
import '../utils/strings.dart';
import '../widgets/connection_banner.dart';
import '../widgets/guidance_panel.dart';
import '../widgets/safe_screen.dart';
import '../widgets/tutorial_overlay.dart';

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
  Uint8List? _lastFrame;

  static const _kCachedResponseKey = 'offline_last_ai_response';

  final DateTime _sessionStart = DateTime.now();
  Duration _elapsed = Duration.zero;
  Timer? _ticker;
  String? _firstUserMessage;
  bool _showTutorial = false;
  String? _lastVoiceText; // for play/pause replay
  String? _cachedResponse; // last AI response from previous session (offline fallback)

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
      final prefs = ref.read(sharedPrefsProvider);
      final shown = prefs.getBool(_kTutorialKey) ?? false;
      if (!shown) setState(() => _showTutorial = true);
      // Load last cached AI response for offline fallback.
      final cached = prefs.getString(_kCachedResponseKey);
      if (cached != null && cached.isNotEmpty) {
        setState(() => _cachedResponse = cached);
      }
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
    if (type == 'chunk') {
      // Streaming: accumulate partial text without restarting typewriter.
      final chunk = msg['text'] as String?;
      if (chunk != null && chunk.isNotEmpty) {
        ref.read(sessionProvider.notifier).appendChunk(chunk);
      }
    } else if (type == 'response') {
      ref.read(sessionProvider.notifier).updateFromResponse(msg);
      // Cache response for offline mode.
      final responseText = msg['text'] as String?;
      if (responseText != null && responseText.isNotEmpty) {
        setState(() => _cachedResponse = responseText);
        ref.read(sharedPrefsProvider).setString(_kCachedResponseKey, responseText);
      }
      // Auto-read voice_text if voice guidance is enabled.
      final voiceText = msg['voice_text'] as String?;
      if (voiceText != null && voiceText.isNotEmpty) {
        _lastVoiceText = voiceText;
        final voiceEnabled = ref.read(settingsProvider).voiceEnabled;
        if (voiceEnabled) {
          ref.read(voiceProvider.notifier).speak(voiceText);
        }
      }
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
      if (frame != null) {
        setState(() => _lastFrame = frame);
        wsService.sendFrame(frame);
      }
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
    try {
      return _buildContent(context);
    } catch (e) {
      return screenError(e, context);
    }
  }

  Widget _buildContent(BuildContext context) {
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
          // Voice play/pause toggle — only shown when voice guidance is enabled.
          Consumer(
            builder: (context, ref, _) {
              final voiceEnabled = ref.watch(settingsProvider).voiceEnabled;
              if (!voiceEnabled) return const SizedBox.shrink();
              final voiceState = ref.watch(voiceProvider);
              final isSpeaking = voiceState.isSpeaking;
              return Semantics(
                label: isSpeaking ? 'Ferma voce AI' : 'Riproduci voce AI',
                button: true,
                child: IconButton(
                  icon: Icon(
                    isSpeaking ? Icons.pause_circle_outline : Icons.volume_up_outlined,
                    color: isSpeaking ? kAccent : Colors.white54,
                    size: 20,
                  ),
                  tooltip: isSpeaking ? 'Ferma voce' : 'Riproduci istruzione',
                  onPressed: () {
                    if (isSpeaking) {
                      ref.read(voiceProvider.notifier).stopSpeaking();
                    } else if (_lastVoiceText != null) {
                      ref.read(voiceProvider.notifier).speak(_lastVoiceText!);
                    }
                  },
                ),
              );
            },
          ),
          if (_lastFrame != null)
            IconButton(
              icon: const Icon(Icons.draw_outlined,
                  color: Colors.white54, size: 20),
              tooltip: 'Annota immagine',
              onPressed: () => showDialog<void>(
                context: context,
                builder: (_) =>
                    _ImageAnnotationDialog(frame: _lastFrame!),
              ),
            ),
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
                    ? ConnectionBanner(status: s)
                    : const SizedBox.shrink(),
                loading: () => const ConnectionBanner(
                    status: ConnectionStatus.connecting),
                error: (_, __) => const ConnectionBanner(
                    status: ConnectionStatus.disconnected),
              ),
              Expanded(
                flex: 6,
                child: _CameraSection(onCapture: _captureAndSend),
              ),
              Expanded(
                flex: 4,
                child: GuidancePanel(
                  textController: _textController,
                  onSendText: _sendText,
                  cachedResponse: _cachedResponse,
                  isOffline: wsStatus.value != ConnectionStatus.connected,
                ),
              ),
            ],
          ),
          if (_showTutorial)
            TutorialOverlay(onDismiss: _dismissTutorial),
        ],
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

// ── Tutorial overlay ──────────────────────────────────────────────────────────

// ── Image annotation dialog ───────────────────────────────────────────────────

class _ImageAnnotationDialog extends StatefulWidget {
  const _ImageAnnotationDialog({required this.frame});
  final Uint8List frame;

  @override
  State<_ImageAnnotationDialog> createState() => _ImageAnnotationDialogState();
}

class _ImageAnnotationDialogState extends State<_ImageAnnotationDialog> {
  final _repaintKey = GlobalKey();
  final List<Offset> _pins = [];
  bool _sharing = false;

  Future<void> _share() async {
    if (_sharing) return;
    setState(() => _sharing = true);
    try {
      final boundary = _repaintKey.currentContext!.findRenderObject()
          as RenderRepaintBoundary;
      final image = await boundary.toImage(pixelRatio: 3.0);
      final bytes =
          await image.toByteData(format: ui.ImageByteFormat.png);
      if (bytes == null) return;
      final pngBytes = bytes.buffer.asUint8List();
      await Share.shareXFiles(
        [XFile.fromData(pngBytes,
            name: 'annotazione.png', mimeType: 'image/png')],
        subject: 'Annotazione PhysicsCopilot',
      );
    } finally {
      if (mounted) setState(() => _sharing = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Dialog(
      backgroundColor: kBgCard,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(16),
        side: const BorderSide(color: kBgCardBorder),
      ),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Header
            Row(
              children: [
                const Icon(Icons.draw_outlined, color: kAccent, size: 18),
                const SizedBox(width: 8),
                const Text(
                  'Annota immagine',
                  style: TextStyle(
                    color: Colors.white,
                    fontWeight: FontWeight.w600,
                    fontSize: 15,
                  ),
                ),
                const Spacer(),
                IconButton(
                  icon: const Icon(Icons.close,
                      color: Colors.white54, size: 18),
                  onPressed: () => Navigator.of(context).pop(),
                  padding: EdgeInsets.zero,
                  constraints: const BoxConstraints(),
                ),
              ],
            ),
            const SizedBox(height: 12),
            // Annotatable image area
            ClipRRect(
              borderRadius: BorderRadius.circular(10),
              child: GestureDetector(
                onTapUp: (d) =>
                    setState(() => _pins.add(d.localPosition)),
                child: RepaintBoundary(
                  key: _repaintKey,
                  child: Stack(
                    children: [
                      AspectRatio(
                        aspectRatio: 4 / 3,
                        child: Image.memory(
                          widget.frame,
                          fit: BoxFit.cover,
                        ),
                      ),
                      Positioned.fill(
                        child: CustomPaint(
                          painter: _PinPainter(pins: _pins),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
            const SizedBox(height: 8),
            const Text(
              'Tocca sull\'immagine per aggiungere un pin',
              style: TextStyle(color: kTextMuted, fontSize: 11),
            ),
            const SizedBox(height: 12),
            // Actions
            Row(
              children: [
                if (_pins.isNotEmpty) ...[
                  OutlinedButton.icon(
                    onPressed: () => setState(() => _pins.clear()),
                    icon: const Icon(Icons.delete_outline, size: 16),
                    label: const Text('Cancella'),
                    style: OutlinedButton.styleFrom(
                      foregroundColor: Colors.redAccent,
                      side: const BorderSide(color: Colors.redAccent),
                      padding: const EdgeInsets.symmetric(
                          horizontal: 12, vertical: 8),
                    ),
                  ),
                  const SizedBox(width: 8),
                ],
                Expanded(
                  child: ElevatedButton.icon(
                    onPressed: _sharing ? null : _share,
                    icon: _sharing
                        ? const SizedBox(
                            width: 14,
                            height: 14,
                            child: CircularProgressIndicator(
                                strokeWidth: 2, color: Colors.white),
                          )
                        : const Icon(Icons.ios_share_outlined, size: 16),
                    label:
                        Text(_sharing ? 'Preparazione…' : 'Condividi'),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: kAccent,
                      foregroundColor: Colors.white,
                      disabledBackgroundColor: kAccent.withAlpha(60),
                      padding: const EdgeInsets.symmetric(vertical: 10),
                    ),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _PinPainter extends CustomPainter {
  const _PinPainter({required this.pins});
  final List<Offset> pins;

  @override
  void paint(Canvas canvas, Size size) {
    for (var i = 0; i < pins.length; i++) {
      final p = pins[i];
      // Drop shadow
      canvas.drawCircle(
        p + const Offset(1, 2),
        14,
        Paint()..color = Colors.black.withAlpha(80),
      );
      // Filled circle
      canvas.drawCircle(
        p,
        13,
        Paint()..color = kAccent,
      );
      // Border
      canvas.drawCircle(
        p,
        13,
        Paint()
          ..color = Colors.white
          ..style = PaintingStyle.stroke
          ..strokeWidth = 2,
      );
      // Number
      final tp = TextPainter(
        text: TextSpan(
          text: '${i + 1}',
          style: const TextStyle(
            color: Colors.white,
            fontSize: 12,
            fontWeight: FontWeight.bold,
          ),
        ),
        textDirection: TextDirection.ltr,
      )..layout();
      tp.paint(
        canvas,
        p - Offset(tp.width / 2, tp.height / 2),
      );
    }
  }

  @override
  bool shouldRepaint(_PinPainter old) => old.pins != pins;
}

