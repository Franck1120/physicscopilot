import 'dart:async';
import 'dart:typed_data';

import 'package:camera/camera.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/camera_provider.dart';
import '../providers/session_provider.dart';
import '../providers/websocket_provider.dart';
import '../services/websocket_service.dart';

/// Full-screen camera view with AI overlay, mic button and connection indicator.
class CameraScreen extends ConsumerStatefulWidget {
  const CameraScreen({super.key});

  @override
  ConsumerState<CameraScreen> createState() => _CameraScreenState();
}

class _CameraScreenState extends ConsumerState<CameraScreen> {
  StreamSubscription<Uint8List>? _frameSubscription;
  StreamSubscription<Map<String, dynamic>>? _messageSubscription;
  bool _isMicActive = false;

  @override
  void dispose() {
    _frameSubscription?.cancel();
    _messageSubscription?.cancel();
    super.dispose();
  }

  // ── Bridge camera frames → WebSocket ──────────────────────────────────────

  void _startForwarding() {
    _frameSubscription?.cancel();
    final cameraService = ref.read(cameraServiceProvider);
    final wsService = ref.read(webSocketServiceProvider);
    _frameSubscription = cameraService.frames.listen(wsService.sendFrame);
  }

  void _startListeningMessages() {
    _messageSubscription?.cancel();
    final wsService = ref.read(webSocketServiceProvider);
    _messageSubscription = wsService.messages.listen((json) {
      if (json['type'] == 'response') {
        ref.read(sessionProvider.notifier).updateFromResponse(json);
      }
    });
  }

  void _toggleMic() {
    setState(() => _isMicActive = !_isMicActive);
    // TODO(franck): wire up speech_to_text when voice task is implemented.
  }

  // ── Build ─────────────────────────────────────────────────────────────────

  @override
  Widget build(BuildContext context) {
    final cameraInit = ref.watch(cameraInitProvider);
    final connectionStatus = ref.watch(connectionStatusProvider);
    final session = ref.watch(sessionProvider);

    // Start bridging once the camera has finished initialising.
    ref.listen<AsyncValue<void>>(cameraInitProvider, (_, next) {
      next.whenData((_) {
        _startForwarding();
        _startListeningMessages();
      });
    });

    return Scaffold(
      backgroundColor: Colors.black,
      body: cameraInit.when(
        loading: () => const Center(
          child: CircularProgressIndicator(color: Colors.white),
        ),
        error: (e, _) => Center(
          child: Text(
            'Errore camera: $e',
            style: const TextStyle(color: Colors.white),
          ),
        ),
        data: (_) => _CameraView(
          controller: ref.read(cameraServiceProvider).controller!,
          connectionStatus: connectionStatus,
          session: session,
          isMicActive: _isMicActive,
          onMicTap: _toggleMic,
        ),
      ),
    );
  }
}

// ── Private sub-widgets ─────────────────────────────────────────────────────

class _CameraView extends StatelessWidget {
  final CameraController controller;
  final AsyncValue<ConnectionStatus> connectionStatus;
  final SessionState session;
  final bool isMicActive;
  final VoidCallback onMicTap;

  const _CameraView({
    required this.controller,
    required this.connectionStatus,
    required this.session,
    required this.isMicActive,
    required this.onMicTap,
  });

  @override
  Widget build(BuildContext context) {
    final padding = MediaQuery.of(context).padding;

    return Stack(
      fit: StackFit.expand,
      children: [
        // ── Camera preview (fills screen) ───────────────────────────────
        CameraPreview(controller),

        // ── AI guidance overlay (bottom, above mic button) ──────────────
        Positioned(
          left: 16,
          right: 16,
          bottom: padding.bottom + 140,
          child: _AIOverlay(session: session),
        ),

        // ── Connection status indicator (top-right) ─────────────────────
        Positioned(
          top: padding.top + 16,
          right: 16,
          child: _ConnectionIndicator(status: connectionStatus),
        ),

        // ── Mic button (bottom-centre) ──────────────────────────────────
        Positioned(
          bottom: padding.bottom + 40,
          left: 0,
          right: 0,
          child: Center(
            child: _MicButton(isActive: isMicActive, onTap: onMicTap),
          ),
        ),
      ],
    );
  }
}

// ── AI overlay ──────────────────────────────────────────────────────────────

class _AIOverlay extends StatelessWidget {
  final SessionState session;

  const _AIOverlay({required this.session});

  @override
  Widget build(BuildContext context) {
    final text = session.responseText;
    final processing = session.isProcessing;

    if (text == null && !processing) return const SizedBox.shrink();

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        color: const Color(0x99000000), // black ~60 % opacity
        borderRadius: BorderRadius.circular(12),
      ),
      child: processing
          ? const Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                SizedBox(
                  width: 16,
                  height: 16,
                  child: CircularProgressIndicator(
                    strokeWidth: 2,
                    color: Colors.white,
                  ),
                ),
                SizedBox(width: 8),
                Text(
                  'Elaboro...',
                  style: TextStyle(color: Colors.white, fontSize: 14),
                ),
              ],
            )
          : Text(
              text!,
              style: const TextStyle(color: Colors.white, fontSize: 15),
            ),
    );
  }
}

// ── Connection indicator ─────────────────────────────────────────────────────

class _ConnectionIndicator extends StatelessWidget {
  final AsyncValue<ConnectionStatus> status;

  const _ConnectionIndicator({required this.status});

  @override
  Widget build(BuildContext context) {
    final (color, label) = status.when(
      data: (s) => switch (s) {
        ConnectionStatus.connected => (Colors.greenAccent, 'Online'),
        ConnectionStatus.connecting =>
          (Colors.orangeAccent, 'Connessione...'),
        ConnectionStatus.disconnected => (Colors.redAccent, 'Offline'),
      },
      loading: () => (Colors.orangeAccent, 'Connessione...'),
      error: (_, __) => (Colors.redAccent, 'Errore'),
    );

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: Colors.black54,
        borderRadius: BorderRadius.circular(20),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            width: 8,
            height: 8,
            decoration: BoxDecoration(color: color, shape: BoxShape.circle),
          ),
          const SizedBox(width: 6),
          Text(
            label,
            style: const TextStyle(color: Colors.white, fontSize: 12),
          ),
        ],
      ),
    );
  }
}

// ── Mic button ───────────────────────────────────────────────────────────────

class _MicButton extends StatelessWidget {
  final bool isActive;
  final VoidCallback onTap;

  const _MicButton({required this.isActive, required this.onTap});

  @override
  Widget build(BuildContext context) {
    final bgColor = isActive ? Colors.redAccent : Colors.white;
    final iconColor = isActive ? Colors.white : Colors.black87;
    final glowColor =
        isActive ? Colors.red : Colors.white;

    return GestureDetector(
      onTap: onTap,
      child: Container(
        width: 72,
        height: 72,
        decoration: BoxDecoration(
          shape: BoxShape.circle,
          color: bgColor,
          boxShadow: [
            BoxShadow(
              color: glowColor.withAlpha(100),
              blurRadius: 14,
              spreadRadius: 2,
            ),
          ],
        ),
        child: Icon(
          isActive ? Icons.mic : Icons.mic_none,
          size: 32,
          color: iconColor,
        ),
      ),
    );
  }
}
