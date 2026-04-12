// Copyright (c) 2026 Franck1120. All rights reserved.
// Use of this source code is governed by a MIT license that can be
// found in the LICENSE file.

import 'package:flutter/material.dart';

import '../main.dart' show kAccent;
import '../utils/strings.dart';

/// Shown once on the first use of the session screen.
/// Tapping anywhere dismisses it and marks it as seen in SharedPreferences.
class TutorialOverlay extends StatefulWidget {
  const TutorialOverlay({super.key, required this.onDismiss});

  final VoidCallback onDismiss;

  @override
  State<TutorialOverlay> createState() => _TutorialOverlayState();
}

class _TutorialOverlayState extends State<TutorialOverlay>
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
