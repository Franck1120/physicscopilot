// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

import 'dart:async';
import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../main.dart' show kAccent, kBgCard, kBgCardBorder, kTextMuted;
import '../providers/session_provider.dart';
import '../utils/strings.dart';
import 'feedback_bar.dart';
import 'multi_step_view.dart';
import 'streaming_text.dart';

/// Bottom panel that displays AI responses and the text-input row.
///
/// Watches [sessionProvider] to switch between idle, processing, streaming,
/// error, and typewriter-response views. When [isOffline] is true and
/// [cachedResponse] is available, shows the last known answer instead.
class GuidancePanel extends ConsumerWidget {
  const GuidancePanel({
    super.key,
    required this.textController,
    required this.onSendText,
    this.cachedResponse,
    this.isOffline = false,
  });
  final TextEditingController textController;
  final VoidCallback onSendText;

  /// Last known AI response, shown as an offline fallback when [isOffline].
  final String? cachedResponse;
  final bool isOffline;

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
          Expanded(child: _ResponseArea(
            session: session,
            cachedResponse: cachedResponse,
            isOffline: isOffline,
          )),
          _TextInputRow(controller: textController, onSend: onSendText),
        ],
      ),
    );
  }
}

// ── Response area with animations ───────────────────────────────────────────

class _ResponseArea extends StatelessWidget {
  const _ResponseArea({
    required this.session,
    this.cachedResponse,
    this.isOffline = false,
  });
  final SessionState session;
  final String? cachedResponse;
  final bool isOffline;

  @override
  Widget build(BuildContext context) {
    Widget child;

    if (session.isProcessing) {
      child = const ThinkingIndicator();
    } else if (session.isStreaming && session.streamingText != null) {
      // True streaming: show accumulated chunks with a blinking cursor.
      child = StreamingText(
        key: const ValueKey('streaming'),
        text: session.streamingText!,
      );
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
    } else if (isOffline && cachedResponse != null) {
      // Offline fallback: show last cached response when disconnected.
      child = Column(
        key: const ValueKey('offline'),
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            width: double.infinity,
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
            color: Colors.orangeAccent.withAlpha(25),
            child: const Row(
              children: [
                Icon(Icons.cloud_off_outlined,
                    color: Colors.orangeAccent, size: 14),
                SizedBox(width: 8),
                Text(
                  'Modalità offline — ultima risposta disponibile',
                  style: TextStyle(
                    color: Colors.orangeAccent,
                    fontSize: 11,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ],
            ),
          ),
          Expanded(
            child: SingleChildScrollView(
              padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
              child: Text(
                cachedResponse!,
                style: const TextStyle(
                    color: Colors.white70, fontSize: 14, height: 1.5),
              ),
            ),
          ),
        ],
      );
    } else if (session.responseText != null) {
      final steps = parseSteps(session.responseText!);
      if (steps != null) {
        // Multi-step response: render interactive step cards.
        child = MultiStepView(
          key: ValueKey('steps_${session.responseText.hashCode}'),
          steps: steps,
        );
      } else {
        // Plain text: animate char-by-char with typewriter effect.
        child = _TypewriterResponse(
          key: ValueKey(session.responseText),
          text: session.responseText!,
        );
      }
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

// ── Thinking indicator — three pulsing dots ──────────────────────────────────

/// Animated three-dot indicator shown while the AI is processing a frame.
class ThinkingIndicator extends StatefulWidget {
  const ThinkingIndicator({super.key});
  @override
  State<ThinkingIndicator> createState() => _ThinkingIndicatorState();
}

class _ThinkingIndicatorState extends State<ThinkingIndicator>
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
        // Copy + feedback row — visible once typing is complete.
        AnimatedOpacity(
          opacity: done ? 1.0 : 0.0,
          duration: const Duration(milliseconds: 300),
          child: Padding(
            padding: const EdgeInsets.only(left: 8, right: 8, bottom: 4),
            child: Row(
              children: [
                // Feedback (thumbs up/down)
                FeedbackBar(responseText: widget.text),
                const Spacer(),
                // Copy button
                Semantics(
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
              ],
            ),
          ),
        ),
      ],
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
