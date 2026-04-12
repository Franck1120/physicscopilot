// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

import 'package:flutter/material.dart';
import '../main.dart' show kAccent;

/// Semi-transparent overlay that displays the current AI repair instruction
/// over the camera preview. Animates in when [text] changes.
class GuidanceOverlay extends StatefulWidget {
  const GuidanceOverlay({
    super.key,
    required this.text,
    this.isProcessing = false,
  });

  /// Current AI instruction text. Null hides the overlay.
  final String? text;

  /// While true, shows a pulsing "analysing" shimmer instead of text.
  final bool isProcessing;

  @override
  State<GuidanceOverlay> createState() => _GuidanceOverlayState();
}

class _GuidanceOverlayState extends State<GuidanceOverlay>
    with SingleTickerProviderStateMixin {
  late final AnimationController _ctrl;
  late final Animation<double> _fade;
  late final Animation<Offset> _slide;

  String? _displayedText;

  @override
  void initState() {
    super.initState();
    _ctrl = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 350),
    );
    _fade = CurvedAnimation(parent: _ctrl, curve: Curves.easeOut);
    _slide = Tween<Offset>(
      begin: const Offset(0, 0.3),
      end: Offset.zero,
    ).animate(CurvedAnimation(parent: _ctrl, curve: Curves.easeOutCubic));

    if (widget.text != null || widget.isProcessing) {
      _displayedText = widget.text;
      _ctrl.forward();
    }
  }

  @override
  void didUpdateWidget(GuidanceOverlay old) {
    super.didUpdateWidget(old);
    final hasContent = widget.text != null || widget.isProcessing;
    final hadContent = old.text != null || old.isProcessing;

    if (hasContent && !hadContent) {
      _displayedText = widget.text;
      _ctrl.forward();
    } else if (!hasContent && hadContent) {
      _ctrl.reverse();
    } else if (widget.text != old.text && widget.text != null) {
      // New text: quick cross-fade
      _ctrl.reverse().then((_) {
        if (mounted) {
          setState(() => _displayedText = widget.text);
          _ctrl.forward();
        }
      });
    }
  }

  @override
  void dispose() {
    _ctrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return FadeTransition(
      opacity: _fade,
      child: SlideTransition(
        position: _slide,
        child: Container(
          margin: const EdgeInsets.fromLTRB(16, 0, 16, 16),
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: Colors.black.withAlpha(178), // 70% opacity
            borderRadius: BorderRadius.circular(16),
            border: Border.all(
              color: kAccent.withAlpha(60),
              width: 1,
            ),
            boxShadow: [
              BoxShadow(
                color: Colors.black.withAlpha(100),
                blurRadius: 20,
                offset: const Offset(0, 4),
              ),
            ],
          ),
          child: widget.isProcessing
              ? _AnalysingRow()
              : Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // Accent icon
                    Container(
                      margin: const EdgeInsets.only(top: 2, right: 12),
                      width: 28,
                      height: 28,
                      decoration: BoxDecoration(
                        shape: BoxShape.circle,
                        color: kAccent.withAlpha(30),
                      ),
                      child: const Icon(
                        Icons.auto_fix_high_rounded,
                        size: 16,
                        color: kAccent,
                      ),
                    ),
                    Expanded(
                      child: Text(
                        _displayedText ?? '',
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 14,
                          height: 1.5,
                        ),
                      ),
                    ),
                  ],
                ),
        ),
      ),
    );
  }
}

/// Pulsing row shown while the AI is processing a frame.
class _AnalysingRow extends StatefulWidget {
  @override
  State<_AnalysingRow> createState() => _AnalysingRowState();
}

class _AnalysingRowState extends State<_AnalysingRow>
    with SingleTickerProviderStateMixin {
  late final AnimationController _pulse;
  late final Animation<double> _opacity;

  @override
  void initState() {
    super.initState();
    _pulse = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 900),
    )..repeat(reverse: true);
    _opacity = Tween<double>(begin: 0.4, end: 1.0).animate(_pulse);
  }

  @override
  void dispose() {
    _pulse.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return FadeTransition(
      opacity: _opacity,
      child: const Row(
        children: [
          SizedBox(
            width: 14,
            height: 14,
            child: CircularProgressIndicator(
              strokeWidth: 2,
              color: kAccent,
            ),
          ),
          SizedBox(width: 12),
          Text(
            'AI sta analizzando...',
            style: TextStyle(
              color: Colors.white70,
              fontSize: 14,
            ),
          ),
        ],
      ),
    );
  }
}
