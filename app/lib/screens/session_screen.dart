import 'dart:async';
import 'dart:typed_data';

import 'package:camera/camera.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../main.dart' show kAccent, kBgPrimary, kBgCard, kBgCardBorder, kTextMuted;
import '../providers/camera_provider.dart';
import '../providers/session_provider.dart';
import '../providers/websocket_provider.dart';

/// Active repair session screen.
///
/// Layout:
///   - Top 60 %: live camera preview with manual capture button
///   - Bottom 40 %: AI guidance panel (response text + text input)
///
/// Frames are forwarded automatically; the capture button forces an
/// immediate analysis of the current frame.
class SessionScreen extends ConsumerStatefulWidget {
  const SessionScreen({super.key});

  @override
  ConsumerState<SessionScreen> createState() => _SessionScreenState();
}

class _SessionScreenState extends ConsumerState<SessionScreen> {
  StreamSubscription<Uint8List>? _frameSubscription;
  StreamSubscription<Map<String, dynamic>>? _messageSubscription;
  final TextEditingController _textController = TextEditingController();

  @override
  void initState() {
    super.initState();
    _startListening();
  }

  @override
  void dispose() {
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

  Future<void> _captureAndSend() async {
    HapticFeedback.mediumImpact();
    final cameraService = ref.read(cameraServiceProvider);
    final wsService = ref.read(webSocketServiceProvider);
    ref.read(sessionProvider.notifier).setProcessing();
    try {
      final frame = await cameraService.captureFrame();
      if (frame != null) wsService.sendFrame(frame);
    } catch (_) {
      ref.read(sessionProvider.notifier).setError('Impossibile acquisire il frame');
    }
  }

  void _sendText() {
    final text = _textController.text.trim();
    if (text.isEmpty) return;
    final wsService = ref.read(webSocketServiceProvider);
    ref.read(sessionProvider.notifier).setProcessing();
    wsService.sendText(text);
    _textController.clear();
    FocusScope.of(context).unfocus();
  }

  @override
  Widget build(BuildContext context) {
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
          onPressed: () => Navigator.of(context).pop(),
        ),
      ),
      body: SafeArea(
        top: false, // AppBar already handles top
        child: Column(
          children: [
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
  Widget build(BuildContext context) {
    return Container(
      color: const Color(0xFF0D0D0D),
      child: const Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            CircularProgressIndicator(color: kAccent),
            SizedBox(height: 16),
            Text(
              'Inizializzazione camera…',
              style: TextStyle(color: kTextMuted, fontSize: 13),
            ),
          ],
        ),
      ),
    );
  }
}

class _CameraError extends StatelessWidget {
  const _CameraError();

  @override
  Widget build(BuildContext context) {
    return Container(
      color: const Color(0xFF0D0D0D),
      child: const Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.camera_alt_outlined, color: kTextMuted, size: 48),
            SizedBox(height: 12),
            Text(
              'Camera non disponibile',
              style: TextStyle(color: kTextMuted, fontSize: 13),
            ),
          ],
        ),
      ),
    );
  }
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

class _ResponseArea extends StatelessWidget {
  const _ResponseArea({required this.session});

  final SessionState session;

  @override
  Widget build(BuildContext context) {
    if (session.isProcessing) {
      return const Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            CircularProgressIndicator(color: kAccent, strokeWidth: 2),
            SizedBox(height: 10),
            Text(
              'Analisi in corso…',
              style: TextStyle(color: kTextMuted, fontSize: 13),
            ),
          ],
        ),
      );
    }

    if (session.errorText != null) {
      return Padding(
        padding: const EdgeInsets.all(16),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Icon(Icons.warning_amber_rounded,
                color: Colors.orangeAccent, size: 18),
            const SizedBox(width: 8),
            Expanded(
              child: Text(
                session.errorText!,
                style: const TextStyle(
                    color: Colors.orangeAccent, fontSize: 13),
              ),
            ),
          ],
        ),
      );
    }

    if (session.responseText != null) {
      return SingleChildScrollView(
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
                  color: Colors.white,
                  fontSize: 14,
                  height: 1.5,
                ),
              ),
            ),
          ],
        ),
      );
    }

    return const Center(
      child: Text(
        'Punta la camera sulla stampante\nper avviare l\'analisi AI.',
        textAlign: TextAlign.center,
        style: TextStyle(color: kTextMuted, fontSize: 13, height: 1.5),
      ),
    );
  }
}

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
                hintStyle:
                    const TextStyle(color: kTextMuted, fontSize: 14),
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
                contentPadding: const EdgeInsets.symmetric(
                    horizontal: 16, vertical: 10),
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
