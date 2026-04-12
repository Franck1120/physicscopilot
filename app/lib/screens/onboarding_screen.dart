// Copyright (c) 2026 Franck1120. All rights reserved.
// Use of this source code is governed by a MIT license that can be
// found in the LICENSE file.

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:permission_handler/permission_handler.dart';

import '../main.dart' show onboardingCompletedProvider;
import '../providers/prefs_provider.dart' show sharedPrefsProvider;

/// Onboarding flow shown once on first launch.
/// Calls [onComplete] after saving the completion flag so the router
/// can navigate to the next screen.
class OnboardingScreen extends ConsumerStatefulWidget {
  final VoidCallback onComplete;

  const OnboardingScreen({super.key, required this.onComplete});

  @override
  ConsumerState<OnboardingScreen> createState() => _OnboardingScreenState();
}

class _OnboardingScreenState extends ConsumerState<OnboardingScreen> {
  static const String _prefKey = 'onboarding_completed';
  static const int _pageCount = 3;

  final PageController _pageController = PageController();
  int _currentPage = 0;
  bool _isCompleting = false;

  @override
  void dispose() {
    _pageController.dispose();
    super.dispose();
  }

  void _goToNextPage() {
    _pageController.nextPage(
      duration: const Duration(milliseconds: 350),
      curve: Curves.easeInOut,
    );
  }

  Future<void> _handleComplete() async {
    if (_isCompleting) return;
    setState(() => _isCompleting = true);

    await _requestPermissions();
    await _saveOnboardingCompleted();

    widget.onComplete();
  }

  Future<void> _requestPermissions() async {
    final statuses = await [
      Permission.camera,
      Permission.microphone,
    ].request();

    final cameraDenied = statuses[Permission.camera]?.isDenied ?? true;
    final micDenied = statuses[Permission.microphone]?.isDenied ?? true;

    if ((cameraDenied || micDenied) && mounted) {
      final denied = <String>[
        if (cameraDenied) 'camera',
        if (micDenied) 'microfono',
      ];
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            'Permessi ${denied.join(" e ")} negati. '
            'Alcune funzioni potrebbero non funzionare.',
          ),
          duration: const Duration(seconds: 4),
          behavior: SnackBarBehavior.floating,
        ),
      );
    }
  }

  /// Persists the completion flag to SharedPreferences and immediately
  /// updates the Riverpod [StateProvider] so the GoRouter redirect guard
  /// sees the new value without waiting for a provider recomputation.
  Future<void> _saveOnboardingCompleted() async {
    final prefs = ref.read(sharedPrefsProvider);
    await prefs.setBool(_prefKey, true);
    ref.read(onboardingCompletedProvider.notifier).state = true;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: DecoratedBox(
        decoration: const BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topCenter,
            end: Alignment.bottomCenter,
            colors: [Color(0xFF1B4F72), Color(0xFF000000)],
          ),
        ),
        child: SafeArea(
          child: Stack(
            children: [
              Column(
                children: [
                  // Reserve space at the top so PageView doesn't overlap
                  // the skip button area.
                  const SizedBox(height: 48),
                  Expanded(
                    child: PageView(
                      controller: _pageController,
                      onPageChanged: (index) =>
                          setState(() => _currentPage = index),
                      children: const [
                        _OnboardingPage(
                          icon: Icons.camera_alt,
                          title: 'Punta la camera',
                          subtitle: 'PhysicsCopilot vede cosa stai facendo',
                        ),
                        _OnboardingPage(
                          icon: Icons.volume_up,
                          title: 'Segui la voce',
                          subtitle: 'Ti guida passo-passo a voce',
                        ),
                        _OnboardingPage(
                          icon: Icons.check_circle_outline,
                          title: 'Risolvi il problema',
                          subtitle: 'Diagnostica e fix in tempo reale',
                        ),
                      ],
                    ),
                  ),
                  _PageIndicator(
                    pageCount: _pageCount,
                    currentPage: _currentPage,
                  ),
                  const SizedBox(height: 32),
                  Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 40),
                    child: _currentPage < _pageCount - 1
                        ? _PrimaryButton(
                            label: 'Avanti',
                            onPressed: _goToNextPage,
                          )
                        : _PrimaryButton(
                            label: 'Inizia',
                            onPressed: _isCompleting ? null : _handleComplete,
                          ),
                  ),
                  const SizedBox(height: 48),
                ],
              ),

              // Skip button — only visible on the first two slides.
              if (_currentPage < 2)
                Positioned(
                  top: 8,
                  right: 16,
                  child: GestureDetector(
                    onTap: _handleComplete,
                    child: const Padding(
                      padding: EdgeInsets.all(8),
                      child: Text(
                        'Salta',
                        style: TextStyle(
                          color: Colors.white54,
                          fontSize: 14,
                        ),
                      ),
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

// ---------------------------------------------------------------------------
// Private widgets
// ---------------------------------------------------------------------------

class _OnboardingPage extends StatelessWidget {
  final IconData icon;
  final String title;
  final String subtitle;

  const _OnboardingPage({
    required this.icon,
    required this.title,
    required this.subtitle,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 40),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          // Icon in a decorated circle for visual depth.
          Container(
            width: 120,
            height: 120,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              color: const Color(0xFF1B4F72).withAlpha(120),
              border: Border.all(
                color: Colors.white.withAlpha(60),
                width: 1.5,
              ),
            ),
            child: Icon(icon, size: 56, color: Colors.white),
          ),
          const SizedBox(height: 40),
          Text(
            title,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 28,
              fontWeight: FontWeight.bold,
              letterSpacing: 0.5,
            ),
            textAlign: TextAlign.center,
          ),

          // Decorative separator between title and subtitle.
          Container(
            width: 40,
            height: 2,
            margin: const EdgeInsets.symmetric(vertical: 16),
            decoration: BoxDecoration(
              color: Colors.white.withAlpha(80),
              borderRadius: BorderRadius.circular(1),
            ),
          ),

          Text(
            subtitle,
            style: const TextStyle(
              color: Color(0xFFADD8E6),
              fontSize: 16,
              height: 1.5,
            ),
            textAlign: TextAlign.center,
          ),
        ],
      ),
    );
  }
}

class _PageIndicator extends StatelessWidget {
  final int pageCount;
  final int currentPage;

  const _PageIndicator({
    required this.pageCount,
    required this.currentPage,
  });

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.center,
      children: List.generate(pageCount, (index) {
        final isActive = index == currentPage;
        return AnimatedContainer(
          duration: const Duration(milliseconds: 250),
          curve: Curves.easeInOut,
          margin: const EdgeInsets.symmetric(horizontal: 5),
          width: isActive ? 20 : 8,
          height: 6,
          decoration: BoxDecoration(
            color: isActive ? Colors.white : const Color(0x66FFFFFF),
            borderRadius: BorderRadius.circular(3),
          ),
        );
      }),
    );
  }
}

class _PrimaryButton extends StatelessWidget {
  final String label;
  final VoidCallback? onPressed;

  const _PrimaryButton({
    required this.label,
    required this.onPressed,
  });

  @override
  Widget build(BuildContext context) {
    // Gradient CTA — wraps an InkWell in a decorated container so the
    // green gradient is visible while still providing ink ripple feedback.
    return SizedBox(
      width: double.infinity,
      height: 52,
      child: DecoratedBox(
        decoration: BoxDecoration(
          gradient: onPressed != null
              ? const LinearGradient(
                  colors: [Color(0xFF10B981), Color(0xFF059669)],
                )
              : const LinearGradient(
                  colors: [Color(0x66FFFFFF), Color(0x66FFFFFF)],
                ),
          borderRadius: BorderRadius.circular(26),
        ),
        child: Material(
          color: Colors.transparent,
          borderRadius: BorderRadius.circular(26),
          child: InkWell(
            onTap: onPressed,
            borderRadius: BorderRadius.circular(26),
            child: Center(
              child: Text(
                label,
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 16,
                  fontWeight: FontWeight.w600,
                  letterSpacing: 0.5,
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
