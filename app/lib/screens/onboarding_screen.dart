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
          child: Column(
            children: [
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
          Icon(icon, size: 80, color: Colors.white),
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
          const SizedBox(height: 16),
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
          margin: const EdgeInsets.symmetric(horizontal: 4),
          width: isActive ? 24 : 8,
          height: 8,
          decoration: BoxDecoration(
            color: isActive ? Colors.white : const Color(0x66FFFFFF),
            borderRadius: BorderRadius.circular(4),
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
    return SizedBox(
      width: double.infinity,
      height: 52,
      child: ElevatedButton(
        onPressed: onPressed,
        style: ElevatedButton.styleFrom(
          backgroundColor: Colors.white,
          foregroundColor: const Color(0xFF1B4F72),
          disabledBackgroundColor: const Color(0x66FFFFFF),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(26),
          ),
          elevation: 0,
          textStyle: const TextStyle(
            fontSize: 16,
            fontWeight: FontWeight.w600,
            letterSpacing: 0.5,
          ),
        ),
        child: Text(label),
      ),
    );
  }
}
