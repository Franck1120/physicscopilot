import 'dart:async';
import 'dart:typed_data';

import 'package:camera/camera.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../main.dart' show kAccent, kAccentDark;

import '../providers/camera_provider.dart';
import '../providers/overlay_provider.dart';
import '../providers/session_provider.dart';
import '../providers/step_provider.dart';
import '../providers/voice_provider.dart';
import '../providers/websocket_provider.dart';
import '../services/websocket_service.dart';
import '../widgets/ar_overlay.dart';
import '../widgets/step_progress.dart';

/// Full-screen camera view with AR overlay, voice I/O, step progress, and chat.
class CameraScreen extends ConsumerStatefulWidget {
  const CameraScreen({super.key});

  @override
  ConsumerState<CameraScreen> createState() => _CameraScreenState();
}

class _CameraScreenState extends ConsumerState<CameraScreen>
    with SingleTickerProviderStateMixin {
  StreamSubscription<Uint8List>? _frameSubscription;
  StreamSubscription<Map<String, dynamic>>? _messageSubscription;
  StreamSubscription<String>? _sttSubscription;

  late final AnimationController _chatController;
  late final Animation<Offset> _chatSlide;
  bool _isChatOpen = false;

  final TextEditingController _textInput = TextEditingController();

  @override
  void initState() {
    super.initState();
    _chatController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 280),
    );
    _chatSlide = Tween<Offset>(
      begin: const Offset(1.0, 0.0),
      end: Offset.zero,
    ).animate(CurvedAnimation(
      parent: _chatController,
      curve: Curves.easeInOut,
    ));
  }

  @override
  void dispose() {
    _frameSubscription?.cancel();
    _messageSubscription?.cancel();
    _sttSubscription?.cancel();
    _chatController.dispose();
    _textInput.dispose();
    super.dispose();
  }

  // ── Bridge camera frames → WebSocket ──────────────────────────────────────

  void _startForwarding() {
    _frameSubscription?.cancel();
    final cameraService = ref.read(cameraServiceProvider);
    final wsService = ref.read(webSocketServiceProvider);
    _frameSubscription = cameraService.frames.listen(wsService.sendFrame);
  }

  // ── Listen to server messages ─────────────────────────────────────────────

  void _startListeningMessages() {
    _messageSubscription?.cancel();
    final wsService = ref.read(webSocketServiceProvider);
    _messageSubscription = wsService.messages.listen((json) {
      if (json['type'] == 'response') {
        ref.read(sessionProvider.notifier).updateFromResponse(json);
        ref.read(stepProvider.notifier).updateFromResponse(json);
        // Speak the AI text response aloud.
        final text = json['text'] as String?;
        if (text != null && text.isNotEmpty) {
          ref.read(voiceProvider.notifier).speak(text);
        }
      }
    });
  }

  // ── Forward STT results → WebSocket ──────────────────────────────────────

  void _startSttForwarding() {
    _sttSubscription?.cancel();
    final voiceService = ref.read(voiceServiceProvider);
    _sttSubscription = voiceService.recognizedText.listen((text) {
      ref.read(voiceProvider.notifier).onRecognized(text);
      ref.read(webSocketServiceProvider).sendText(text);
      ref.read(sessionProvider.notifier).setProcessing();
    });
  }

  // ── UI actions ────────────────────────────────────────────────────────────

  Future<void> _toggleMic() async {
    HapticFeedback.mediumImpact();
    await ref.read(voiceProvider.notifier).toggleListening();
  }

  void _toggleChat() {
    setState(() => _isChatOpen = !_isChatOpen);
    if (_isChatOpen) {
      _chatController.forward();
    } else {
      _chatController.reverse();
    }
  }

  void _sendText() {
    final text = _textInput.text.trim();
    if (text.isEmpty) return;
    _textInput.clear();
    ref.read(webSocketServiceProvider).sendText(text);
    ref.read(sessionProvider.notifier).setProcessing();
  }

  // ── Build ─────────────────────────────────────────────────────────────────

  @override
  Widget build(BuildContext context) {
    final cameraInit = ref.watch(cameraInitProvider);
    final connectionStatus = ref.watch(connectionStatusProvider);
    final session = ref.watch(sessionProvider);
    final voiceState = ref.watch(voiceProvider);
    final overlayData = ref.watch(overlayDataProvider);
    final stepState = ref.watch(stepProvider);

    // Start bridging once camera is ready.
    ref.listen<AsyncValue<void>>(cameraInitProvider, (_, next) {
      next.whenData((_) {
        _startForwarding();
        _startListeningMessages();
        _startSttForwarding();
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
            'Camera error: $e',
            style: const TextStyle(color: Colors.white),
          ),
        ),
        data: (_) => _CameraBody(
          controller: ref.read(cameraServiceProvider).controller!,
          connectionStatus: connectionStatus,
          session: session,
          voiceState: voiceState,
          overlayData: overlayData,
          stepState: stepState,
          isChatOpen: _isChatOpen,
          chatSlide: _chatSlide,
          textInput: _textInput,
          onMicTap: _toggleMic,
          onChatTap: _toggleChat,
          onSendText: _sendText,
        ),
      ),
    );
  }
}

// ── Body ─────────────────────────────────────────────────────────────────────

class _CameraBody extends StatelessWidget {
  final CameraController controller;
  final AsyncValue<ConnectionStatus> connectionStatus;
  final SessionState session;
  final VoiceState voiceState;
  final OverlayData? overlayData;
  final ProcedureState stepState;
  final bool isChatOpen;
  final Animation<Offset> chatSlide;
  final TextEditingController textInput;
  final VoidCallback onMicTap;
  final VoidCallback onChatTap;
  final VoidCallback onSendText;

  const _CameraBody({
    required this.controller,
    required this.connectionStatus,
    required this.session,
    required this.voiceState,
    required this.overlayData,
    required this.stepState,
    required this.isChatOpen,
    required this.chatSlide,
    required this.textInput,
    required this.onMicTap,
    required this.onChatTap,
    required this.onSendText,
  });

  @override
  Widget build(BuildContext context) {
    final padding = MediaQuery.of(context).padding;
    final screenSize = MediaQuery.of(context).size;

    return Stack(
      fit: StackFit.expand,
      children: [
        // ── Camera preview ───────────────────────────────────────────────
        CameraPreview(controller),

        // ── AR overlay (fills entire feed) ───────────────────────────────
        ArOverlay(data: overlayData),

        // ── Step progress (bottom) ───────────────────────────────────────
        if (stepState.steps.isNotEmpty)
          Positioned(
            left: 0,
            right: isChatOpen ? screenSize.width * 0.45 : 0,
            bottom: 0,
            child: StepProgress(
              steps: stepState.steps,
              currentStep: stepState.currentIndex,
            ),
          ),

        // ── AI processing indicator (bottom-centre, above step bar) ──────
        if (session.isProcessing)
          Positioned(
            bottom: padding.bottom +
                (stepState.steps.isNotEmpty ? 120 : 40),
            left: 0,
            right: 0,
            child: const Center(child: _AnalyzingIndicator()),
          ),

        // ── Connection status (top-right) ────────────────────────────────
        Positioned(
          top: padding.top + 16,
          right: isChatOpen ? screenSize.width * 0.45 + 8 : 16,
          child: _ConnectionIndicator(status: connectionStatus),
        ),

        // ── Chat toggle button (top-left) ────────────────────────────────
        Positioned(
          top: padding.top + 16,
          left: 16,
          child: _ChatToggleButton(isOpen: isChatOpen, onTap: onChatTap),
        ),

        // ── Mic button (bottom-centre, above step progress) ──────────────
        Positioned(
          bottom: padding.bottom +
              (stepState.steps.isNotEmpty ? 110 : 40),
          left: isChatOpen ? 0 : null,
          right: 0,
          child: Center(
            child: _MicButton(
              isActive: voiceState.isListening,
              isSpeaking: voiceState.isSpeaking,
              onTap: onMicTap,
            ),
          ),
        ),

        // ── Slide-in chat panel (right side) ─────────────────────────────
        SlideTransition(
          position: chatSlide,
          child: Align(
            alignment: Alignment.centerRight,
            child: SizedBox(
              width: screenSize.width * 0.45,
              child: _ChatPanel(
                session: session,
                textInput: textInput,
                onSend: onSendText,
              ),
            ),
          ),
        ),
      ],
    );
  }
}

// ── Analyzing indicator ───────────────────────────────────────────────────────

class _AnalyzingIndicator extends StatelessWidget {
  const _AnalyzingIndicator();

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
      decoration: BoxDecoration(
        color: const Color(0x99000000),
        borderRadius: BorderRadius.circular(20),
      ),
      child: const Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          SizedBox(
            width: 14,
            height: 14,
            child: CircularProgressIndicator(
              strokeWidth: 2,
              color: Colors.white70,
            ),
          ),
          SizedBox(width: 8),
          Text(
            'AI sta analizzando...',
            style: TextStyle(color: Colors.white70, fontSize: 13),
          ),
        ],
      ),
    );
  }
}

// ── Chat panel ────────────────────────────────────────────────────────────────

class _ChatPanel extends StatelessWidget {
  final SessionState session;
  final TextEditingController textInput;
  final VoidCallback onSend;

  const _ChatPanel({
    required this.session,
    required this.textInput,
    required this.onSend,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      color: const Color(0xEE111111),
      child: Column(
        children: [
          // AI response display
          Expanded(
            child: Padding(
              padding: const EdgeInsets.all(12),
              child: session.responseText != null
                  ? SingleChildScrollView(
                      child: Text(
                        session.responseText!,
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 14,
                          height: 1.5,
                        ),
                      ),
                    )
                  : const Center(
                      child: Text(
                        'Nessuna risposta ancora.',
                        style: TextStyle(color: Colors.white38, fontSize: 13),
                      ),
                    ),
            ),
          ),
          // Text input
          Container(
            padding: const EdgeInsets.fromLTRB(8, 8, 8, 16),
            decoration: const BoxDecoration(
              border: Border(top: BorderSide(color: Colors.white12)),
            ),
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: textInput,
                    style: const TextStyle(color: Colors.white, fontSize: 14),
                    decoration: InputDecoration(
                      hintText: 'Scrivi una domanda…',
                      hintStyle: const TextStyle(color: Colors.white38),
                      filled: true,
                      fillColor: Colors.white10,
                      contentPadding: const EdgeInsets.symmetric(
                        horizontal: 12,
                        vertical: 10,
                      ),
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(20),
                        borderSide: BorderSide.none,
                      ),
                    ),
                    onSubmitted: (_) => onSend(),
                  ),
                ),
                const SizedBox(width: 6),
                GestureDetector(
                  onTap: onSend,
                  child: Container(
                    width: 38,
                    height: 38,
                    decoration: BoxDecoration(
                      color: kAccent,
                      shape: BoxShape.circle,
                    ),
                    child: const Icon(
                      Icons.send,
                      color: Colors.white,
                      size: 18,
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

// ── Connection indicator ──────────────────────────────────────────────────────

class _ConnectionIndicator extends StatelessWidget {
  final AsyncValue<ConnectionStatus> status;

  const _ConnectionIndicator({required this.status});

  @override
  Widget build(BuildContext context) {
    final (color, label) = status.when(
      data: (s) => switch (s) {
        ConnectionStatus.connected => (Colors.greenAccent, 'Online'),
        ConnectionStatus.connecting => (Colors.orangeAccent, 'Connessione...'),
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

// ── Chat toggle button ────────────────────────────────────────────────────────

class _ChatToggleButton extends StatelessWidget {
  final bool isOpen;
  final VoidCallback onTap;

  const _ChatToggleButton({required this.isOpen, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        width: 40,
        height: 40,
        decoration: BoxDecoration(
          color: isOpen ? kAccent : Colors.black54,
          shape: BoxShape.circle,
        ),
        child: Icon(
          isOpen ? Icons.close : Icons.chat_bubble_outline,
          color: Colors.white,
          size: 20,
        ),
      ),
    );
  }
}

// ── Mic button ────────────────────────────────────────────────────────────────

class _MicButton extends StatelessWidget {
  final bool isActive;
  final bool isSpeaking;
  final VoidCallback onTap;

  const _MicButton({
    required this.isActive,
    required this.isSpeaking,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final Color bgColor;
    final Color iconColor;
    final Color glowColor;

    if (isSpeaking) {
      bgColor = kAccent;
      iconColor = Colors.white;
      glowColor = kAccent;
    } else if (isActive) {
      bgColor = Colors.redAccent;
      iconColor = Colors.white;
      glowColor = Colors.red;
    } else {
      bgColor = Colors.white;
      iconColor = Colors.black87;
      glowColor = Colors.white;
    }

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
          isSpeaking
              ? Icons.volume_up
              : isActive
                  ? Icons.mic
                  : Icons.mic_none,
          size: 32,
          color: iconColor,
        ),
      ),
    );
  }
}
